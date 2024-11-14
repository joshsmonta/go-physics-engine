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

type Box struct {
	width  float32
	length float32
}

func UpdateBallPosition(p *Particle, delta float32) {
	// p := &particle
	p.velocity = rl.Vector2Add(p.velocity, rl.Vector2Scale(p.accel, delta))
	p.current = rl.Vector2Add(p.current, rl.Vector2Scale(p.velocity, delta))
}

func HandleBoxCollision(p *Particle, box *Box) {
	ball_bottom := p.current.Y + p.radius
	ball_top := p.current.Y - p.radius
	ball_right := p.current.X + p.radius
	ball_left := p.current.X - p.radius
	if box.length <= ball_bottom || 0 >= ball_top {
		p.velocity.Y = -p.velocity.Y
		p.accel.Y = -p.accel.Y / 2
	}
	if box.width <= ball_right || 0 >= ball_left {
		p.velocity.X = -p.velocity.X
		p.accel.X = -p.accel.X / 2
	}
}

func Run() {
	width := 1280
	length := 800
	rl.InitWindow(1280, 800, "2D Ball Bounce")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	particle := Particle{
		current:  rl.Vector2{X: float32(width / 3), Y: float32(length / 3)},
		previous: rl.Vector2{X: float32(width / 3), Y: float32(length / 3)},
		velocity: rl.Vector2{X: 50.0, Y: 50.0},
		accel:    rl.Vector2{X: 25.0, Y: 50.0},
		radius:   15,
	}
	box := Box{
		width:  float32(width),
		length: float32(length),
	}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.DrawCircle(int32(particle.current.X), int32(particle.current.Y), particle.radius, rl.Red)
		UpdateBallPosition(&particle, 1.0/60.0)
		HandleBoxCollision(&particle, &box)

		rl.EndDrawing()
	}
}
