package ballbox

import (
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GRAVITY          = 90.81 // Earth gravity in m/s²
	BOUNCE_DAMPENING = 0.7   // Energy retained after bounce
	AIR_RESISTANCE   = 0.99  // Air friction coefficient
	GROUND_FRICTION  = 0.85  // Ground friction coefficient
)

type Particle struct {
	position   rl.Vector2
	velocity   rl.Vector2
	accel      rl.Vector2
	radius     float32
	mass       float32
	color      rl.Color
	elasticity float32 // Bounciness factor (0-1)
}

type Box struct {
	width  float32
	height float32
}

func NewParticle(pos rl.Vector2, radius, mass float32, color rl.Color) *Particle {
	return &Particle{
		position:   pos,
		velocity:   rl.Vector2Zero(),
		accel:      rl.Vector2{X: 0, Y: GRAVITY},
		radius:     radius,
		mass:       mass,
		color:      color,
		elasticity: 0.8 + rand.Float32()*0.2, // Random bounciness between 0.8-1.0
	}
}

func UpdatePhysics(p *Particle, delta float32) {
	// Apply forces (gravity scaled by mass)
	force := rl.Vector2Scale(p.accel, p.mass)

	// Calculate acceleration (F = ma => a = F/m)
	acceleration := rl.Vector2Scale(force, 1/p.mass)

	// Update velocity with acceleration
	p.velocity.X += acceleration.X * delta
	p.velocity.Y += acceleration.Y * delta

	// Apply air resistance
	p.velocity = rl.Vector2Scale(p.velocity, AIR_RESISTANCE)

	// Update position (Verlet integration would be better for more accuracy)
	p.position.X += p.velocity.X * delta
	p.position.Y += p.velocity.Y * delta
}

func HandleCollision(p *Particle, box *Box) {
	// Bottom collision (floor)
	if p.position.Y+p.radius > box.height {
		p.position.Y = box.height - p.radius
		p.velocity.Y = -p.velocity.Y * p.elasticity
		p.velocity.X *= GROUND_FRICTION // More friction when on ground
	}

	// Top collision (ceiling)
	if p.position.Y-p.radius < 0 {
		p.position.Y = p.radius
		p.velocity.Y = -p.velocity.Y * p.elasticity
	}

	// Right collision
	if p.position.X+p.radius > box.width {
		p.position.X = box.width - p.radius
		p.velocity.X = -p.velocity.X * p.elasticity
	}

	// Left collision
	if p.position.X-p.radius < 0 {
		p.position.X = p.radius
		p.velocity.X = -p.velocity.X * p.elasticity
	}
}

func Run() {
	screenWidth := int32(1280)
	screenHeight := int32(800)

	rl.InitWindow(screenWidth, screenHeight, "2D Physics with Mass")
	defer rl.CloseWindow()
	rl.SetTargetFPS(120)

	// Create balls with different masses
	balls := []*Particle{
		NewParticle(
			rl.Vector2{X: float32(screenWidth) / 4, Y: float32(screenHeight) / 4},
			25.0, 500.0, rl.Red), // Heavy ball (mass=5)
		NewParticle(
			rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 4},
			15.0, 100.0, rl.Blue), // Medium ball (mass=1)
		NewParticle(
			rl.Vector2{X: float32(screenWidth) * 3 / 4, Y: float32(screenHeight) / 4},
			10.0, 50.0, rl.Green), // Light ball (mass=0.3)
	}

	// Give initial velocities
	balls[0].velocity.X = 100.0
	balls[1].velocity.X = 100.0
	balls[2].velocity.X = 100.0

	box := Box{
		width:  float32(screenWidth),
		height: float32(screenHeight),
	}

	for !rl.WindowShouldClose() {
		delta := rl.GetFrameTime() // Get actual frame delta time

		// Handle input
		if rl.IsKeyPressed(rl.KeyR) {
			// Reset balls
			balls[0].position = rl.Vector2{X: float32(screenWidth) / 4, Y: float32(screenHeight) / 4}
			balls[1].position = rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 4}
			balls[2].position = rl.Vector2{X: float32(screenWidth) * 3 / 4, Y: float32(screenHeight) / 4}

			balls[0].velocity = rl.Vector2{X: 100, Y: 0}
			balls[1].velocity = rl.Vector2{X: 100, Y: 0}
			balls[2].velocity = rl.Vector2{X: 100, Y: 0}
		}

		// Update physics for all balls
		for _, ball := range balls {
			UpdatePhysics(ball, delta)
			HandleCollision(ball, &box)
		}

		// Draw
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Draw all balls
		for _, ball := range balls {
			rl.DrawCircleV(ball.position, ball.radius, ball.color)

			// Draw velocity vector
			rl.DrawLineV(
				ball.position,
				rl.Vector2Add(ball.position, rl.Vector2Scale(ball.velocity, 0.1)),
				rl.Fade(rl.White, 0.5),
			)
		}

		// Draw stats
		rl.DrawText("Physics with Mass Demonstration", 10, 10, 20, rl.White)
		rl.DrawText("Red: Heavy (5kg) | Blue: Medium (1kg) | Green: Light (0.3kg)", 10, 30, 20, rl.White)
		rl.DrawText("Press R to reset balls", 10, 50, 20, rl.White)

		rl.EndDrawing()
	}
}
