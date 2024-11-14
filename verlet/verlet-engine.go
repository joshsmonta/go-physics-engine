package verlet

import (
	"image/color"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Particle struct {
	current  rl.Vector3
	previous rl.Vector3
	accel    rl.Vector3
	radius   float64
	color    color.RGBA
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
	strength := float32(2000.0)
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

func HandleCollision(a *Particle, b *Particle) {
	axis := rl.Vector3Subtract(a.current, b.current)
	distance := math.Sqrt(float64(rl.Vector3DotProduct(axis, axis)))
	if distance == 0 {
		// if distance is zero, do nothing to avoid divistion by zero
		return
	}
	if distance < (a.radius + b.radius) {
		n := rl.Vector3Scale(axis, float32(1.0/distance))
		delta := (a.radius + b.radius) - distance // calucalte penetration depth
		a.current = rl.Vector3Add(a.current, rl.Vector3Scale(n, float32(0.5*delta)))
		b.current = rl.Vector3Subtract(b.current, rl.Vector3Scale(n, float32(0.5*delta)))
	}
}

func BruteForceCollision(verlet_objects []Particle) {
	for i := range verlet_objects {
		for j := range verlet_objects {
			a := &verlet_objects[i]
			b := &verlet_objects[j]
			HandleCollision(a, b)
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
		if distance > radius-p.radius {
			n := rl.Vector3Scale(displacement, float32(1.0/distance))
			p.current = rl.Vector3Add(container_pos, rl.Vector3Scale(n, float32(radius-p.radius)))
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
		Y: -1000.0,
		Z: 0.0,
	}
	time_step := 0.0015
	sub_steps := 1
	rl.InitWindow(1280, 800, "3D Particle Simulation Inside a Sphere")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	magnetActive := false
	particles := make([]Particle, 500)
	sphereRadius := 15.0
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
		BruteForceCollision(particles)
		ApplyConstraints(sphereRadius, particles)

		rl.EndMode3D()
		// rl.DrawText("SubSteps: "+str, 10, 35, 20, rl.RayWhite)
		if magnetActive {
			rl.DrawText("Magnet: ON", 10, 10, 20, rl.Green)
		} else {
			rl.DrawText("Magnet: OFF (Press SPACE to toggle)", 10, 10, 20, rl.Red)
		}
		rl.EndDrawing()
	}
}
