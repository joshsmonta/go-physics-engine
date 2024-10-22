package ballbox

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Particle struct {
	current  rl.Vector2
	previous rl.Vector2
	velocity rl.Vector2
	accel    rl.Vector2
	radius   float32
}

// type Box struct {
// 	width  float32
// 	length float32
// }

func UpdateBallPosition(particle Particle, delta float32) {
	p := &particle
	p.velocity = rl.Vector2Add(p.velocity, rl.Vector2Scale(p.accel, delta))
	p.current = rl.Vector2Add(p.current, rl.Vector2Scale(p.velocity, delta))
}

// func HandleBoxCollision(particle Particle, box Box) {
// 	p := &particle
// 	box_pos := rl.Vector2{
// 		X: float32(box.width),
// 		Y: float32(box.length),
// 	}
// 	displacement := rl.Vector2Subtract(p.current, box_pos)
// 	distance := rl.Vector2Length(displacement)
// 	if distance < p.radius {
// 		n := rl.Vector2Scale(displacement, 1.0/distance)
// 		p.accel = rl.Vector2Add(p.accel, rl.Vector2Scale(n, p.radius-distance))
// 	}

// }

func Run() {
	width := 1280
	length := 800
	rl.InitWindow(1280, 800, "2D Ball Bounce")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	particle := Particle{
		current:  rl.Vector2{X: float32(width / 3), Y: float32(length / 3)},
		previous: rl.Vector2{X: float32(width / 3), Y: float32(length / 3)},
		velocity: rl.Vector2{X: 0.0, Y: 25.0},
		accel:    rl.Vector2{X: 0.0, Y: 20.0},
		radius:   15,
	}
	// box := Box{
	// 	width:  float32(width),
	// 	length: float32(length),
	// }

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.DrawCircle(int32(particle.current.X), int32(particle.current.Y), particle.radius, rl.Red)
		UpdateBallPosition(particle, 1.0/60.0)
		// HandleBoxCollision(particle, box)

		rl.EndDrawing()
	}
}