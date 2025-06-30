package solar

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GRAVITY   = 1000.0 // Scaled gravitational constant
	SUB_STEPS = 10     // Physics sub-steps for accuracy
	TRAIL_LEN = 500    // Length of orbit trails
	SIM_SPEED = 1.0    // Simulation speed multiplier
)

type Body struct {
	position rl.Vector2
	velocity rl.Vector2
	radius   float32
	mass     float32
	color    rl.Color
	trail    [TRAIL_LEN]rl.Vector2
	trailIdx int
}

func Run() {
	screenWidth := int32(1280)
	screenHeight := int32(800)

	rl.InitWindow(screenWidth, screenHeight, "Three-Body Orbital Simulation")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// Create three bodies with different masses
	bodies := make([]*Body, 3)
	centerX := float32(screenWidth) / 2
	centerY := float32(screenHeight) / 2

	// Primary body (yellow star)
	bodies[0] = &Body{
		position: rl.Vector2{X: centerX, Y: centerY},
		velocity: rl.Vector2{X: 0, Y: 0},
		radius:   25,
		mass:     2000,
		color:    rl.Yellow,
	}

	// Secondary body (blue planet)
	bodies[1] = &Body{
		position: rl.Vector2{X: centerX - 300, Y: centerY},
		velocity: rl.Vector2{X: 0, Y: 80},
		radius:   15,
		mass:     100,
		color:    rl.Blue,
	}

	// Tertiary body (red planet)
	bodies[2] = &Body{
		position: rl.Vector2{X: centerX + 400, Y: centerY - 100},
		velocity: rl.Vector2{X: 0, Y: -70},
		radius:   12,
		mass:     50,
		color:    rl.Red,
	}

	// Initialize trails
	for _, body := range bodies {
		for i := range body.trail {
			body.trail[i] = body.position
		}
	}

	font := rl.GetFontDefault()
	camera := rl.Camera2D{}
	camera.Zoom = 1.0

	for !rl.WindowShouldClose() {
		// Handle camera controls
		if rl.IsKeyDown(rl.KeyRight) {
			camera.Target.X += 10
		}
		if rl.IsKeyDown(rl.KeyLeft) {
			camera.Target.X -= 10
		}
		if rl.IsKeyDown(rl.KeyDown) {
			camera.Target.Y += 10
		}
		if rl.IsKeyDown(rl.KeyUp) {
			camera.Target.Y -= 10
		}
		if rl.IsKeyPressed(rl.KeyR) {
			camera.Target = rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2}
			camera.Zoom = 1.0
		}

		// Handle zoom
		camera.Zoom += rl.GetMouseWheelMove() * 0.05
		if camera.Zoom < 0.1 {
			camera.Zoom = 0.1
		}
		if camera.Zoom > 3.0 {
			camera.Zoom = 3.0
		}

		delta := rl.GetFrameTime() * SIM_SPEED
		subDelta := delta / float32(SUB_STEPS)

		// Physics simulation with sub-stepping
		for step := 0; step < SUB_STEPS; step++ {
			// Calculate gravitational forces
			for i, body := range bodies {
				for j, other := range bodies {
					if i == j {
						continue
					}

					dx := other.position.X - body.position.X
					dy := other.position.Y - body.position.Y
					distanceSq := dx*dx + dy*dy
					distance := float32(math.Sqrt(float64(distanceSq)))

					if distance < 1e-5 {
						continue
					}

					// Gravitational acceleration (a = G * M / r^2)
					acceleration := GRAVITY * other.mass / distanceSq
					directionX := dx / distance
					directionY := dy / distance

					// Update velocity with acceleration
					body.velocity.X += acceleration * directionX * subDelta
					body.velocity.Y += acceleration * directionY * subDelta
				}
			}

			// Update positions
			for _, body := range bodies {
				body.position.X += body.velocity.X * subDelta
				body.position.Y += body.velocity.Y * subDelta
			}
		}

		// Update trails
		for _, body := range bodies {
			body.trail[body.trailIdx] = body.position
			body.trailIdx = (body.trailIdx + 1) % TRAIL_LEN
		}

		// Draw
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.BeginMode2D(camera)

		// Draw trails
		for _, body := range bodies {
			for i := 0; i < TRAIL_LEN-1; i++ {
				idx1 := (body.trailIdx + i) % TRAIL_LEN
				idx2 := (body.trailIdx + i + 1) % TRAIL_LEN

				// Skip if position hasn't been set yet
				if body.trail[idx1].X == 0 && body.trail[idx1].Y == 0 {
					continue
				}

				// Fade trail based on age
				alpha := uint8(255 * float32(TRAIL_LEN-i) / TRAIL_LEN)
				trailColor := rl.Color{
					R: body.color.R,
					G: body.color.G,
					B: body.color.B,
					A: alpha,
				}

				rl.DrawLineV(body.trail[idx1], body.trail[idx2], trailColor)
			}
		}

		// Draw bodies
		for _, body := range bodies {
			rl.DrawCircleV(body.position, body.radius, body.color)
			rl.DrawCircleLines(int32(body.position.X), int32(body.position.Y), body.radius, rl.Fade(rl.White, 0.5))
		}

		// Draw velocity vectors
		for _, body := range bodies {
			end := rl.Vector2Add(body.position, rl.Vector2Scale(body.velocity, 0.1))
			rl.DrawLineEx(body.position, end, 2, rl.Green)
			rl.DrawCircleV(end, 3, rl.Green)
		}

		rl.EndMode2D()

		// Draw UI
		rl.DrawRectangle(0, 0, screenWidth, 30, rl.Fade(rl.DarkGray, 0.7))
		rl.DrawTextEx(font, "Three-Body Orbital Simulation", rl.Vector2{X: 10, Y: 5}, 20, 1, rl.White)
		rl.DrawTextEx(
			font, "Controls: Arrow keys to pan, Mouse wheel to zoom, R to reset view",
			rl.Vector2{X: 10, Y: float32(screenHeight) - 25}, 20, 1, rl.LightGray)

		// Draw info
		// infoText := "Bodies:"
		// for i, body := range bodies {
		// 	speed := math.Sqrt(float64(body.velocity.X*body.velocity.X + body.velocity.Y*body.velocity.Y))
		// 	infoText += fmt.Printf(" Body %d: Mass=%.0f | Speed=%.0f", i+1, body.mass, speed)
		// }
		// rl.DrawTextEx(font, infoText, rl.Vector2{X: 10, Y: 35}, 20, 1, rl.LightGray)

		rl.EndDrawing()
	}
}
