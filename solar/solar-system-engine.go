package solar

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Planet struct {
	current  rl.Vector3
	previous rl.Vector3
	accel    rl.Vector3
	radius   float64
	// mass     float64
	color color.RGBA
}

func UpdatePositions(verlet_objects []Planet, delta_time float32) {
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

func Run() {
	rl.InitWindow(1280, 800, "Gravitational Pull Simulation")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	planets := make([]Planet, 1)
	sphere_radius := 15.0

	const G float64 = 6.67430e-11 // Gravitational constant
	time_step := 0.0015
	sub_steps := 1

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

		for _, p := range planets {
			rl.DrawSphere(p.current, float32(p.radius), p.color)
		}
		rl.DrawSphereWires(rl.Vector3{
			X: 0.0,
			Y: 0.0,
			Z: 0.0,
		}, float32(sphere_radius), 16, 16, rl.DarkGray)

		sub_dt := float32(time_step) / float32(sub_steps)
		UpdatePositions(planets, sub_dt)
		// ApplyGravity(planets)
		// ApplyGravity2(planets, G)
		// ApplyConstraints(sphere_radius, planets)
		rl.EndMode3D()
		rl.EndDrawing()
	}
}
