package verlet

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GRAVITY          = -15.0
	THREAD_COUNT     = 8
	VERLET_RADIUS    = 0.5
	CONTAINER_RADIUS = 29.0 // Half of container side length
)

type Particle struct {
	current  rl.Vector3
	previous rl.Vector3
	accel    rl.Vector3
	radius   float32
	color    color.RGBA
}

// parallelFor distributes [0, n) across NumCPU goroutines.
// For small n (<4× cores) it runs single-threaded to avoid overhead.
func parallelFor(n int, fn func(start, end int)) {
	numWorkers := runtime.NumCPU()
	if n < numWorkers*4 {
		fn(0, n)
		return
	}
	var wg sync.WaitGroup
	chunkSize := (n + numWorkers - 1) / numWorkers
	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		if start >= n {
			break
		}
		end := start + chunkSize
		if end > n {
			end = n
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			fn(s, e)
		}(start, end)
	}
	wg.Wait()
}

// CellKey is a 3D integer coordinate used as the hash map key.
// Using a struct as a map key is valid in Go — it's compared by value.
type CellKey struct {
	x, y, z int
}

// SpatialHash only allocates storage for cells that contain particles.
// With 500 particles this means ~100-200 occupied cells vs the old
// 58^3 = 195,112 cell array — a ~1000x reduction in Clear/iteration cost.
type SpatialHash struct {
	cells    map[CellKey][]*Particle
	cellSize float32
	bucketCap int
}

func UpdatePositions(verlet_objects []Particle, delta_time float32) {
	parallelFor(len(verlet_objects), func(start, end int) {
		for i := start; i < end; i++ {
			p := &verlet_objects[i]
			displacement := rl.Vector3Subtract(p.current, p.previous)
			p.previous = p.current
			p.current = rl.Vector3Add(rl.Vector3Add(p.current, displacement), rl.Vector3Scale(p.accel, delta_time*delta_time))
			p.accel = rl.Vector3{
				X: 0.0,
				Y: 0.0,
				Z: 0.0,
			}
		}
	})
}

func Magnet(particles []Particle, activeM bool) {
	if !activeM {
		return
	}
	center := rl.Vector3{X: 0.0, Y: 20.0, Z: 0.0}
	strength := float32(5000.0)
	parallelFor(len(particles), func(start, end int) {
		for i := start; i < end; i++ {
			p := &particles[i]
			direction := rl.Vector3Subtract(center, p.current)
			distance := math.Sqrt(float64(rl.Vector3DotProduct(direction, direction)))
			if distance > 0 {
				n := rl.Vector3Scale(direction, float32(1.0/distance))
				p.accel = rl.Vector3Add(p.accel, rl.Vector3Scale(n, strength))
			}
		}
	})
}

func ApplyForces(verlet_objects []Particle, gravity rl.Vector3) {
	parallelFor(len(verlet_objects), func(start, end int) {
		for i := start; i < end; i++ {
			p := &verlet_objects[i]
			p.accel = rl.Vector3Add(p.accel, gravity)
			displacement := rl.Vector3Subtract(p.current, p.previous)
			distance := math.Sqrt(float64(rl.Vector3DotProduct(displacement, displacement)))
			if distance > 0 {
				// Normalize the displacement vector
				n := rl.Vector3Scale(displacement, float32(1.0/distance))
				// Apply friction: accel += n * friction
				frictionForce := rl.Vector3Scale(n, float32(-200))
				p.accel = rl.Vector3Add(p.accel, frictionForce)
			}
		}
	})
}

// NewSpatialHash creates a hash with the given cell size.
// cellSize should be >= 2 * particle_radius so each particle
// occupies at most one cell and neighbors cover the collision range.
func NewSpatialHash(cellSize float32, bucketCap int) *SpatialHash {
	return &SpatialHash{
		cells:     make(map[CellKey][]*Particle, 2048),
		cellSize:  cellSize,
		bucketCap: bucketCap,
	}
}

// Clear resets the map. Only touches previously-occupied cells.
// Cost: O(occupied cells) instead of O(58^3).
func (sh *SpatialHash) Clear() {
	for k := range sh.cells {
		delete(sh.cells, k)
	}
}

// AddObject hashes the particle's position into a cell key.
func (sh *SpatialHash) AddObject(p *Particle) {
	ix := int(math.Floor(float64(p.current.X / sh.cellSize)))
	iy := int(math.Floor(float64(p.current.Y / sh.cellSize)))
	iz := int(math.Floor(float64(p.current.Z / sh.cellSize)))
	key := CellKey{ix, iy, iz}

	cell := sh.cells[key]
	if len(cell) >= sh.bucketCap {
		return
	}
	sh.cells[key] = append(cell, p)
}

// forwardNeighbors contains the 13 neighbor offsets that are
// lexicographically greater than (0,0,0). By only checking these
// (plus the self-cell), every pair of adjacent cells is visited
// exactly once — no duplicate checks, no unsafe.Pointer comparison.
//
// dx=1: 9 neighbors (all dy,dz combos)
// dx=0, dy=1: 3 neighbors (all dz combos)
// dx=0, dy=0, dz=1: 1 neighbor
var forwardNeighbors = [13][3]int{
	{1, -1, -1}, {1, -1, 0}, {1, -1, 1},
	{1, 0, -1}, {1, 0, 0}, {1, 0, 1},
	{1, 1, -1}, {1, 1, 0}, {1, 1, 1},
	{0, 1, -1}, {0, 1, 0}, {0, 1, 1},
	{0, 0, 1},
}

// ApplyCollisions iterates only occupied cells and checks the 13
// forward neighbors. Single-threaded — eliminates the data race from
// the old parallel grid approach (two threads could mutate the same
// particle at strip boundaries). For 500 particles this is faster
// than goroutine creation + sync overhead.
func (sh *SpatialHash) ApplyCollisions() {
	for key, cell := range sh.cells {
		if len(cell) == 0 {
			continue
		}

		// Intra-cell collisions
		for a := 0; a < len(cell); a++ {
			for b := a + 1; b < len(cell); b++ {
				handleCollision(cell[a], cell[b])
			}
		}

		// Inter-cell collisions with forward neighbors only
		for _, offset := range forwardNeighbors {
			neighborKey := CellKey{
				x: key.x + offset[0],
				y: key.y + offset[1],
				z: key.z + offset[2],
			}
			neighborCell, exists := sh.cells[neighborKey]
			if !exists || len(neighborCell) == 0 {
				continue
			}
			for a := range cell {
				for b := range neighborCell {
					handleCollision(cell[a], neighborCell[b])
				}
			}
		}
	}
}

func handleCollision(a *Particle, b *Particle) {
	axis := rl.Vector3Subtract(a.current, b.current)
	distSq := float64(rl.Vector3DotProduct(axis, axis))
	minDist := float64(a.radius + b.radius)
	// Early-out using squared distance — avoids sqrt for non-colliding pairs
	if distSq == 0 || distSq >= minDist*minDist {
		return
	}
	distance := math.Sqrt(distSq)
	n := rl.Vector3Scale(axis, float32(1.0/distance))
	delta := minDist - distance
	a.current = rl.Vector3Add(a.current, rl.Vector3Scale(n, float32(0.5*delta)))
	b.current = rl.Vector3Subtract(b.current, rl.Vector3Scale(n, float32(0.5*delta)))
}

func BruteForceCollision(verlet_objects []Particle) {
	for i := range verlet_objects {
		for j := i + 1; j < len(verlet_objects); j++ {
			a := &verlet_objects[i]
			b := &verlet_objects[j]
			if a != b {
				handleCollision(a, b)
			}
		}
	}
}

func ApplyConstraints(radius float64, verlet_objects []Particle) {
	containerPos := rl.Vector3{X: 0.0, Y: 0.0, Z: 0.0}
	parallelFor(len(verlet_objects), func(start, end int) {
		for i := start; i < end; i++ {
			p := &verlet_objects[i]
			displacement := rl.Vector3Subtract(p.current, containerPos)
			distance := math.Sqrt(float64(rl.Vector3DotProduct(displacement, displacement)))
			if distance > radius-float64(p.radius) {
				n := rl.Vector3Scale(displacement, float32(1.0/distance))
				p.current = rl.Vector3Add(containerPos, rl.Vector3Scale(n, float32(float32(radius)-p.radius)))
			}
		}
	})
}

func RandomPointInSphere(radius float32) rl.Vector3 {
	for {
		// Generate random point in a cube [-1, 1] on each axis
		x := rand.Float32()*2 - 1
		y := rand.Float32()*2 - 1
		z := rand.Float32()*2 - 1

		// Check if the point is within the sphere
		if x*x+y*y+z*z <= 1 {
			return rl.Vector3{
				X: float32(x * radius),
				Y: float32(y * radius),
				Z: float32(z * radius),
			}
		}
	}
}

func Run() {
	// var wg sync.WaitGroup
	gravity := rl.Vector3{
		X: 0.0,
		Y: -2000.0,
		Z: 0.0,
	}
	time_step := 0.0015
	sub_steps := 2
	rl.InitWindow(1280, 800, "3D Particle Simulation Inside a Sphere")
	defer rl.CloseWindow()
	rl.SetTargetFPS(120)
	magnetActive := false
	particles := make([]Particle, 10000)
	sphereRadius := 18.0
	colors := []color.RGBA{
		rl.NewColor(255, 182, 193, 255), // Light Pink
		rl.NewColor(173, 216, 230, 255), // Light Blue
		rl.NewColor(240, 230, 140, 255), // Light Khaki
		rl.NewColor(152, 251, 152, 255), // Pale Green
		rl.NewColor(221, 160, 221, 255), // Plum
	}
	for i := range particles {
		randColor := colors[rand.Intn(len(colors))]
		random_position := RandomPointInSphere(float32(sphereRadius))
		particles[i] = Particle{
			current:  random_position,
			previous: random_position,
			accel: rl.Vector3{
				X: 0.0,
				Y: 0.0,
				Z: 0.0,
			},
			radius: 0.5,
			color:  randColor,
		}
	}
	// cellSize = 2 * particleRadius = 1.0. This ensures each particle
	// fits in one cell and neighbor checks cover the collision radius.
	// bucketCap = 16: max particles per cell before overflow is ignored.
	hash := NewSpatialHash(1.0, 16)

	// Pre-generate low-poly sphere mesh and load as model.
	// 4 rings/slices = ~32 triangles per sphere vs DrawSphere's 256.
	// GenMeshSphere already uploads to GPU — do NOT call UploadMesh again.
	sphereMesh := rl.GenMeshSphere(0.5, 4, 4)
	sphereModel := rl.LoadModelFromMesh(sphereMesh)
	defer rl.UnloadModel(sphereModel)

	camera := rl.Camera3D{
		Position:   rl.Vector3{X: 25.0, Y: 25.0, Z: 25.0},
		Target:     rl.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
		Up:         rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0},
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}
	for !rl.WindowShouldClose() {
		// --- Input handling (before physics) ---
		if rl.IsKeyPressed(rl.KeySpace) {
			magnetActive = !magnetActive
		}

		// --- Physics (before rendering) ---
		// Sub-stepping: run physics 8 times per frame with a smaller dt.
		// Each sub-step resolves collisions incrementally. With 1 sub-step
		// (the old value), piled-up particles don't converge — overlaps
		// persist across frames. With 8, the iterative resolution has
		// 8 chances to push particles apart, producing stable stacking.
		sub_dt := float32(time_step) / float32(sub_steps)
		for s := 0; s < sub_steps; s++ {
			Magnet(particles, magnetActive)
			ApplyForces(particles, gravity)
			UpdatePositions(particles, sub_dt)

			hash.Clear()
			for i := range particles {
				hash.AddObject(&particles[i])
			}
			hash.ApplyCollisions()

			ApplyConstraints(sphereRadius, particles)
		}

		// --- Rendering (reads finalized positions) ---
		// Physics is done, so the GPU gets the draw calls immediately
		// without waiting for CPU physics computation.
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(camera)

		for i := range particles {
			p := &particles[i]
			rl.DrawModel(sphereModel, p.current, 1.0, p.color)
		}
		rl.DrawSphereWires(rl.Vector3{
			X: 0.0,
			Y: 0.0,
			Z: 0.0,
		}, float32(sphereRadius), 16, 16, rl.DarkGray)

		rl.EndMode3D()

		rl.DrawText(fmt.Sprintf("FPS: %d", rl.GetFPS()), 10, 30, 20, rl.RayWhite)
		if magnetActive {
			rl.DrawText("Magnet: ON", 10, 10, 20, rl.Green)
		} else {
			rl.DrawText("Magnet: OFF (Press SPACE to toggle)", 10, 10, 20, rl.Red)
		}
		rl.EndDrawing()
	}
}
