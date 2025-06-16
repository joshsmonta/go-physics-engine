package verlet

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sync"
	"unsafe"

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

type Grid struct {
	cells   [][][][]*Particle
	size    int
	cellNum int
}

func UpdatePositions(verlet_objects []Particle, delta_time float32) {
	for i := range verlet_objects {
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
}

func Magnet(particles []Particle, activeM bool) {
	if !activeM {
		return
	}
	center := rl.Vector3{X: 0.0, Y: 20.0, Z: 0.0}
	strength := float32(5000.0)
	for i := range particles {
		p := &particles[i]
		direction := rl.Vector3Subtract(center, p.current)
		distance := math.Sqrt(float64(rl.Vector3DotProduct(direction, direction)))
		if distance > 0 {
			n := rl.Vector3Scale(direction, float32(1.0/distance))
			p.accel = rl.Vector3Add(p.accel, rl.Vector3Scale(n, strength))
		}
	}
}

func ApplyForces(verlet_objects []Particle, gravity rl.Vector3) {
	for i := range verlet_objects {
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
}

func NewGrid(size, cellNum int) *Grid {
	grid := make([][][][]*Particle, cellNum)
	for i := range grid {
		grid[i] = make([][][]*Particle, cellNum)
		for j := range grid[i] {
			grid[i][j] = make([][]*Particle, cellNum)
			for k := range grid[i][j] {
				grid[i][j][k] = make([]*Particle, 0, size)
			}
		}
	}
	return &Grid{
		cells:   grid,
		size:    size,
		cellNum: cellNum,
	}
}

func (g *Grid) Clear() {
	for i := range g.cells {
		for j := range g.cells[i] {
			for k := range g.cells[i][j] {
				g.cells[i][j][k] = g.cells[i][j][k][:0]
			}
		}
	}
}

// Add object to grid based on position
func (g *Grid) AddObject(p *Particle) {
	cellSize := 2.0 * p.radius
	halfGrid := float32(g.cellNum) / 2.0

	ix := int(math.Floor(float64(p.current.X/cellSize + halfGrid)))
	iy := int(math.Floor(float64(p.current.Y/cellSize + halfGrid)))
	iz := int(math.Floor(float64(p.current.Z/cellSize + halfGrid)))

	if ix < 0 || ix >= g.cellNum ||
		iy < 0 || iy >= g.cellNum ||
		iz < 0 || iz >= g.cellNum {
		return // Particle outside grid boundaries
	}

	cell := &g.cells[ix][iy][iz]
	if len(*cell) >= g.size {
		return // Cell at capacity
	}

	*cell = append(*cell, p)
}

// Process collisions using the grid
func (g *Grid) ApplyCollisions() {
	var wg sync.WaitGroup
	wg.Add(THREAD_COUNT)

	// Process grid in parallel strips along X-axis
	for t := 0; t < THREAD_COUNT; t++ {
		go func(threadID int) {
			defer wg.Done()
			start := threadID * (g.cellNum / THREAD_COUNT)
			end := (threadID + 1) * (g.cellNum / THREAD_COUNT)
			if threadID == THREAD_COUNT-1 {
				end = g.cellNum
			}

			for x := start; x < end; x++ {
				for y := 0; y < g.cellNum; y++ {
					for z := 0; z < g.cellNum; z++ {
						currentCell := g.cells[x][y][z]
						if len(currentCell) == 0 {
							continue
						}

						// Check collisions with objects in current cell
						for a := 0; a < len(currentCell); a++ {
							for b := a + 1; b < len(currentCell); b++ {
								handleCollision(currentCell[a], currentCell[b])
							}
						}

						// Check collisions with neighboring cells
						for dx := -1; dx <= 1; dx++ {
							for dy := -1; dy <= 1; dy++ {
								for dz := -1; dz <= 1; dz++ {
									nx, ny, nz := x+dx, y+dy, z+dz
									if nx < 0 || nx >= g.cellNum || ny < 0 || ny >= g.cellNum || nz < 0 || nz >= g.cellNum || (dx == 0 && dy == 0 && dz == 0) {
										continue
									}

									neighborCell := g.cells[nx][ny][nz]
									if len(neighborCell) == 0 {
										continue
									}

									for a := range currentCell {
										for b := range neighborCell {
											// Ensure each pair is processed only once
											if uintptr(unsafe.Pointer(currentCell[a])) < uintptr(unsafe.Pointer(neighborCell[b])) {
												handleCollision(currentCell[a], neighborCell[b])
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}(t)
	}
	wg.Wait()
}

func handleCollision(a *Particle, b *Particle) {
	axis := rl.Vector3Subtract(a.current, b.current)
	distance := math.Sqrt(float64(rl.Vector3DotProduct(axis, axis)))
	if distance == 0 {
		// if distance is zero, do nothing to avoid divistion by zero
		return
	}
	if distance < (float64(a.radius + b.radius)) {
		n := rl.Vector3Scale(axis, float32(1.0/distance))
		delta := float64(a.radius+b.radius) - distance // calucalte penetration depth
		a.current = rl.Vector3Add(a.current, rl.Vector3Scale(n, float32(0.5*delta)))
		b.current = rl.Vector3Subtract(b.current, rl.Vector3Scale(n, float32(0.5*delta)))
	}
}

func BruteForceCollision(verlet_objects []Particle) {
	for i := 0; i < len(verlet_objects); i++ {
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
	for i := range verlet_objects {
		p := &verlet_objects[i]
		container_pos := rl.Vector3{
			X: 0.0,
			Y: 0.0,
			Z: 0.0,
		}
		displacement := rl.Vector3Subtract(p.current, container_pos)
		distance := math.Sqrt(float64(rl.Vector3DotProduct(displacement, displacement)))
		if distance > radius-float64(p.radius) {
			n := rl.Vector3Scale(displacement, float32(1.0/distance))
			p.current = rl.Vector3Add(container_pos, rl.Vector3Scale(n, float32(float32(radius)-p.radius)))
		}
	}
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
	sub_steps := 1
	rl.InitWindow(1280, 800, "3D Particle Simulation Inside a Sphere")
	defer rl.CloseWindow()
	rl.SetTargetFPS(120)
	magnetActive := false
	particles := make([]Particle, 1000)
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
	grid := NewGrid(4, 58)
	// containerPos := [3]float32{0, 0, 0}
	camera := rl.Camera3D{
		Position:   rl.Vector3{X: 25.0, Y: 25.0, Z: 25.0},
		Target:     rl.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
		Up:         rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0},
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.BeginMode3D(camera)
		if rl.IsKeyPressed(rl.KeySpace) {
			magnetActive = !magnetActive
		}
		for _, p := range particles {
			rl.DrawSphere(p.current, float32(p.radius), p.color)
		}
		rl.DrawSphereWires(rl.Vector3{
			X: 0.0,
			Y: 0.0,
			Z: 0.0,
		}, float32(sphereRadius), 16, 16, rl.DarkGray)

		sub_dt := float32(time_step) / float32(sub_steps)
		Magnet(particles, magnetActive)
		ApplyForces(particles, gravity)
		UpdatePositions(particles, sub_dt)
		// BruteForceCollision(particles)
		grid.Clear()
		for i := range particles {
			grid.AddObject(&particles[i])
		}
		grid.ApplyCollisions()

		ApplyConstraints(sphereRadius, particles)

		rl.EndMode3D()
		// rl.DrawText("SubSteps: "+str, 10, 35, 20, rl.RayWhite)
		rl.DrawText(fmt.Sprintf("FPS: %d", rl.GetFPS()), 10, 30, 20, rl.RayWhite)
		if magnetActive {
			rl.DrawText("Magnet: ON", 10, 10, 20, rl.Green)
		} else {
			rl.DrawText("Magnet: OFF (Press SPACE to toggle)", 10, 10, 20, rl.Red)
		}
		rl.EndDrawing()
	}
}
