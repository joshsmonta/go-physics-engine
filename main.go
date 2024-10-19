package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Particle struct {
	current  rl.Vector3
	previous rl.Vector3
	accel    rl.Vector3
	radius   float64
}

func UpdatePositions(verlet_objects []Particle, delta_time float32) {
	for i := range verlet_objects {
		p := &verlet_objects[i]
		displacement := rl.Vector3Subtract(p.current, p.previous)
		p.previous = p.current
		p.current = rl.Vector3Add(rl.Vector3Add(p.current, displacement),
			rl.Vector3Scale(p.accel, delta_time*delta_time))
		p.accel = rl.Vector3{
			X: 0.0,
			Y: 0.0,
			Z: 0.0,
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
			frictionForce := rl.Vector3Scale(n, float32(-100))
			p.accel = rl.Vector3Add(p.accel, frictionForce)
		}
	}
}

func HandleCollision(a Particle, b Particle) {
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

func ApplyConstraints(radius float64, verlet_objects []Particle) {
	// Floor constraint
	// for i := range verlet_objects {
	// 	p := &verlet_objects[i]
	// 	if p.current.Y <= -2 {
	// 		displacement := p.current.Y - p.previous.Y
	// 		p.current.Y = -2
	// 		p.previous.Y = p.current.Y + displacement
	// 	}
	// }
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

func main() {
	// rand.NewSource(time.Now().UnixNano()) // Seed the random generator
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

	particles := make([]Particle, 20)
	sphereRadius := 15.0
	for i := range particles {
		random_position := RandomPointInSphere(float32(sphereRadius))
		particles[i] = Particle{
			current:  random_position,
			previous: random_position,
			accel: rl.Vector3{
				X: 0.0,
				Y: 0.0,
				Z: 0.0,
			},
			radius: 0.3,
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

		rl.DrawSphereWires(rl.Vector3{
			X: 0.0,
			Y: 0.0,
			Z: 0.0,
		}, float32(sphereRadius), 16, 26, rl.DarkGray)

		for _, p := range particles {
			rl.DrawSphere(p.current, float32(p.radius), rl.Beige)
		}
		sub_dt := float32(time_step) / float32(sub_steps)

		ApplyForces(particles, gravity)
		UpdatePositions(particles, sub_dt)
		HandleCollision()
		ApplyConstraints(sphereRadius, particles)

		rl.EndMode3D()
		rl.EndDrawing()
	}
}
