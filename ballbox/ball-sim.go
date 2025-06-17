package ballbox

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GRAVITY          = 90.81 // Earth gravity in m/s²
	BOUNCE_DAMPENING = 0.95  // Energy retained after bounce
	AIR_RESISTANCE   = 0.99  // Air friction coefficient
	GROUND_FRICTION  = 0.85  // Ground friction coefficient
)

type Particle struct {
	position rl.Vector2
	velocity rl.Vector2
	radius   float32
	mass     float32
}

// func (p *Particle) Accelerate(x float32, y float32, delta float32) {
// 	p.velocity.X += x * delta
// 	p.velocity.Y += y * delta
// }

// func (p *Particle) UpdatePos(delta float32) {
// 	p.position.X = p.velocity.X * delta
// 	p.position.Y = p.velocity.Y * delta
// }

func (p *Particle) Update(delta float32, width, height float32) {
	// Apply gravity (converted to pixel/s² scale)
	p.velocity.Y += GRAVITY * delta

	// Apply air resistance
	// p.velocity.X *= AIR_RESISTANCE
	// p.velocity.Y *= AIR_RESISTANCE

	// Update position using velocity
	p.position.X += p.velocity.X * delta
	p.position.Y += p.velocity.Y * delta

	// Handle collisions with screen edges
	// Bottom collision
	if p.position.Y > height-p.radius {
		p.position.Y = height - p.radius
		p.velocity.Y *= -BOUNCE_DAMPENING
	}
	// Top collision
	if p.position.Y < p.radius {
		p.position.Y = p.radius
		p.velocity.Y *= -BOUNCE_DAMPENING
	}
	// Right/left collisions
	if p.position.X > width-p.radius {
		p.position.X = width - p.radius
		p.velocity.X *= -BOUNCE_DAMPENING
	} else if p.position.X < p.radius {
		p.position.X = p.radius
		p.velocity.X *= -BOUNCE_DAMPENING
	}
}

type Box struct {
	width  float32
	height float32
}

// func UpdatePhysics(p *Particle, delta float32) {
// 	// Apply forces (gravity scaled by mass)
// 	force := rl.Vector2Scale(p.accel, p.mass)

// 	// Calculate acceleration (F = ma => a = F/m)
// 	acceleration := rl.Vector2Scale(force, 1/p.mass)

// 	// Update velocity with acceleration
// 	p.velocity.X += acceleration.X * delta
// 	p.velocity.Y += acceleration.Y * delta

// 	// Apply air resistance
// 	p.velocity = rl.Vector2Scale(p.velocity, AIR_RESISTANCE)

// 	// Update position (Verlet integration would be better for more accuracy)
// 	p.position.X += p.velocity.X * delta
// 	p.position.Y += p.velocity.Y * delta
// }

func Run() {
	screenWidth := float32(1280)
	screenHeight := float32(800)

	rl.InitWindow(1280, 800, "2D Physics with Mass")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	balls := make([]Particle, 2)
	// Create balls with different masses
	for i := range balls {
		balls[i] = Particle{
			position: rl.Vector2{X: 200.0, Y: 100.0}, // Start near top
			velocity: rl.Vector2{X: float32(100 + i*50), Y: 0.0},
			radius:   15.0,
			mass:     15.0,
		}
	}

	for !rl.WindowShouldClose() {
		delta := rl.GetFrameTime() // Get delta time
		for i := range balls {
			balls[i].Update(delta, screenWidth, screenHeight)
			// balls[i].Accelerate(0, -9.81, delta)
			// balls[i].UpdatePos(delta)
			// if balls[i].position.Y < 0 || balls[i].position.Y > screenHeight {
			// 	balls[i].velocity.Y *= -0.95
			// }
			// if balls[i].position.X < 0 || balls[i].position.X > screenWidth {
			// 	balls[i].velocity.X *= -0.95
			// }
		}
		// Draw
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		// delta := rl.GetFrameTime() // Get actual frame delta time

		for _, ball := range balls {
			rl.DrawCircleV(ball.position, ball.radius, rl.White)
		}

		// Update physics for all balls
		rl.EndDrawing()
	}
}
