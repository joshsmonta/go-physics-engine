package threebodyp

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GRAVITY        = 1000.0 // Scaled gravitational constant
	SUB_STEPS      = 10     // Physics sub-steps for accuracy
	TRAIL_LEN      = 500    // Length of orbit trails
	SIM_SPEED_MIN  = 0.1    // Minimum simulation speed
	SIM_SPEED_MAX  = 5.0    // Maximum simulation speed
	SIM_SPEED_INIT = 1.0    // Initial simulation speed
	ZOOM_MIN       = 0.1    // Minimum zoom level
	ZOOM_MAX       = 3.0    // Maximum zoom level
)

type Body struct {
	position   rl.Vector2
	velocity   rl.Vector2
	radius     float32
	mass       float32
	color      rl.Color
	trail      [TRAIL_LEN]rl.Vector2
	trailIdx   int
	name       string
	initialPos rl.Vector2
	initialVel rl.Vector2
}

func Run() {
	screenWidth := int32(1280)
	screenHeight := int32(800)

	rl.InitWindow(screenWidth, screenHeight, "Three-Body Problem: Equal Masses")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// Create three bodies with equal mass
	bodies := make([]*Body, 3)
	centerX := float32(screenWidth) / 2
	centerY := float32(screenHeight) / 2
	mass := float32(1000) // Equal mass for all bodies

	// Body 1 (Blue)
	bodies[0] = &Body{
		position:   rl.Vector2{X: centerX - 200, Y: centerY - 100},
		velocity:   rl.Vector2{X: 0, Y: -50},
		radius:     15,
		mass:       mass,
		color:      rl.Blue,
		name:       "Alpha",
		initialPos: rl.Vector2{X: centerX - 200, Y: centerY - 100},
		initialVel: rl.Vector2{X: 0, Y: -50},
	}

	// Body 2 (Green)
	bodies[1] = &Body{
		position:   rl.Vector2{X: centerX + 200, Y: centerY - 100},
		velocity:   rl.Vector2{X: 0, Y: 60},
		radius:     15,
		mass:       mass,
		color:      rl.Green,
		name:       "Beta",
		initialPos: rl.Vector2{X: centerX + 200, Y: centerY - 100},
		initialVel: rl.Vector2{X: 0, Y: 60},
	}

	// Body 3 (Red)
	bodies[2] = &Body{
		position:   rl.Vector2{X: centerX, Y: centerY + 200},
		velocity:   rl.Vector2{X: -40, Y: -40},
		radius:     15,
		mass:       mass,
		color:      rl.Red,
		name:       "Gamma",
		initialPos: rl.Vector2{X: centerX, Y: centerY + 200},
		initialVel: rl.Vector2{X: -40, Y: -40},
	}

	// Initialize trails
	for _, body := range bodies {
		for i := range body.trail {
			body.trail[i] = body.position
		}
	}

	font := rl.GetFontDefault()
	camera := rl.Camera2D{}
	camera.Target = rl.Vector2{X: centerX, Y: centerY}
	camera.Offset = rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2}
	camera.Zoom = 1.0

	simSpeed := SIM_SPEED_INIT
	paused := false
	showInfo := true
	showVectors := true

	for !rl.WindowShouldClose() {
		// Toggle pause
		if rl.IsKeyPressed(rl.KeySpace) {
			paused = !paused
		}

		// Toggle info display
		if rl.IsKeyPressed(rl.KeyI) {
			showInfo = !showInfo
		}

		// Toggle velocity vectors
		if rl.IsKeyPressed(rl.KeyV) {
			showVectors = !showVectors
		}

		// Reset simulation
		if rl.IsKeyPressed(rl.KeyR) {
			for _, body := range bodies {
				body.position = body.initialPos
				body.velocity = body.initialVel
				body.trailIdx = 0
				for j := range body.trail {
					body.trail[j] = body.initialPos
				}
			}
			camera.Target = rl.Vector2{X: centerX, Y: centerY}
			camera.Zoom = 1.0
			simSpeed = SIM_SPEED_INIT
			paused = false
		}

		// Adjust simulation speed
		if rl.IsKeyDown(rl.KeyEqual) || rl.IsKeyDown(rl.KeyKpAdd) {
			simSpeed += 0.05
			if simSpeed > SIM_SPEED_MAX {
				simSpeed = SIM_SPEED_MAX
			}
		}
		if rl.IsKeyDown(rl.KeyMinus) || rl.IsKeyDown(rl.KeyKpSubtract) {
			simSpeed -= 0.05
			if simSpeed < SIM_SPEED_MIN {
				simSpeed = SIM_SPEED_MIN
			}
		}

		// Handle camera controls
		if rl.IsKeyDown(rl.KeyRight) {
			camera.Target.X += 10 / camera.Zoom
		}
		if rl.IsKeyDown(rl.KeyLeft) {
			camera.Target.X -= 10 / camera.Zoom
		}
		if rl.IsKeyDown(rl.KeyDown) {
			camera.Target.Y += 10 / camera.Zoom
		}
		if rl.IsKeyDown(rl.KeyUp) {
			camera.Target.Y -= 10 / camera.Zoom
		}

		// Handle zoom
		camera.Zoom += rl.GetMouseWheelMove() * 0.05
		if camera.Zoom < ZOOM_MIN {
			camera.Zoom = ZOOM_MIN
		}
		if camera.Zoom > ZOOM_MAX {
			camera.Zoom = ZOOM_MAX
		}

		delta := rl.GetFrameTime() * float32(simSpeed)

		if !paused {
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

						// Prevent division by zero and extreme forces
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
		if showVectors {
			for _, body := range bodies {
				end := rl.Vector2Add(body.position, rl.Vector2Scale(rl.Vector2Normalize(body.velocity), 30))
				rl.DrawLineEx(body.position, end, 2, rl.White)
				rl.DrawCircleV(end, 3, rl.White)
			}
		}

		rl.EndMode2D()

		// Draw UI
		if showInfo {
			// Header
			rl.DrawRectangle(0, 0, screenWidth, 30, rl.Fade(rl.DarkGray, 0.7))
			rl.DrawTextEx(font, "Three-Body Problem: Equal Masses", rl.Vector2{X: 10, Y: 5}, 20, 1, rl.White)

			// Info panel
			panelHeight := float32(120)
			rl.DrawRectangle(0, int32(float32(screenHeight)-panelHeight), int32(float32(screenWidth)), int32(panelHeight), rl.Fade(rl.DarkGray, 0.7))

			// Controls
			controlsText := "Controls: SPACE=Pause/Resume  R=Reset  I=Toggle Info  V=Toggle Vectors  +/-=Speed"
			rl.DrawTextEx(font, controlsText, rl.Vector2{X: 10, Y: float32(screenHeight) - 25}, 20, 1, rl.LightGray)

			// Body info
			yPos := float32(screenHeight) - panelHeight + 10
			for _, body := range bodies {
				speed := math.Sqrt(float64(body.velocity.X*body.velocity.X + body.velocity.Y*body.velocity.Y))
				infoText := fmt.Sprintf("%s: Mass=%.0f | Speed=%.1f | Position=(%.0f, %.0f)",
					body.name, body.mass, speed, body.position.X, body.position.Y)
				rl.DrawTextEx(font, infoText, rl.Vector2{X: 10, Y: yPos}, 20, 1, body.color)
				yPos += 25
			}
			// Simulation info
			simInfo := fmt.Sprintf("Simulation Speed: %.1fx | Zoom: %.1fx | %s",
				simSpeed, camera.Zoom, map[bool]string{true: "PAUSED", false: "RUNNING"}[paused])
			rl.DrawTextEx(font, simInfo, rl.Vector2{X: 10, Y: yPos}, 20, 1, rl.Gold)
		} else {
			// Minimal info when full info is hidden
			rl.DrawTextEx(font, "I: Show Info", rl.Vector2{X: 10, Y: 10}, 20, 1, rl.LightGray)
		}

		rl.EndDrawing()
	}
}
