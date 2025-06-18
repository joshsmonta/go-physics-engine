package ballbox

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GRAVITY          = 8000 // Earth gravity in m/s²
	BOUNCE_DAMPENING = 0.95 // Energy retained after bounce
	AIR_RESISTANCE   = 0.99 // Air friction coefficient
	GROUND_FRICTION  = 0.85 // Ground friction coefficient
)

type Particle struct {
	position rl.Vector2
	velocity rl.Vector2
	radius   float32
	mass     float32
}

func (p *Particle) Accelerate(x float32, y float32, delta float32) {
	// Update position using velocity
	p.position.X += (p.velocity.X + x) * delta
	p.position.Y += (p.velocity.Y + y) * delta
}

// func (p *Particle) UpdatePos(delta float32) {
// 	p.position.X = p.velocity.X * delta
// 	p.position.Y = p.velocity.Y * delta
// }

func (p *Particle) Update(delta float32, width, height float32) {
	// Apply gravity (converted to pixel/s² scale)
	// p.velocity.Y += GRAVITY * delta

	// Apply air resistance
	// p.velocity.X *= AIR_RESISTANCE
	// p.velocity.Y *= AIR_RESISTANCE

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

// Check collision between two particles
func (p *Particle) CollidesWith(other *Particle) bool {
	dx := p.position.X - other.position.X
	dy := p.position.Y - other.position.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	return distance < p.radius+other.radius
}

// Resolve collision between two particles
func resolveCollision(p1, p2 *Particle) {
	// Calculate collision normal (direction from p1 to p2)
	dx := p2.position.X - p1.position.X
	dy := p2.position.Y - p1.position.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	// Avoid division by zero
	if distance == 0 {
		return
	}

	// Normalize the direction vector
	nx := dx / distance
	ny := dy / distance

	// Calculate relative velocity
	rvx := p2.velocity.X - p1.velocity.X
	rvy := p2.velocity.Y - p1.velocity.Y

	// Calculate velocity along the normal
	velocityAlongNormal := rvx*nx + rvy*ny

	// Don't resolve if particles are moving apart
	if velocityAlongNormal > 0 {
		return
	}

	// Calculate impulse scalar
	impulse := -(1.0 + BOUNCE_DAMPENING) * velocityAlongNormal
	impulse /= 1/p1.mass + 1/p2.mass

	// Apply impulse
	impulseX := impulse * nx
	impulseY := impulse * ny

	p1.velocity.X -= impulseX / p1.mass
	p1.velocity.Y -= impulseY / p1.mass
	p2.velocity.X += impulseX / p2.mass
	p2.velocity.Y += impulseY / p2.mass

	// Position correction to prevent sticking
	overlap := (p1.radius + p2.radius) - distance
	if overlap > 0 {
		correctionX := nx * overlap * 0.5
		correctionY := ny * overlap * 0.5
		p1.position.X -= correctionX
		p1.position.Y -= correctionY
		p2.position.X += correctionX
		p2.position.Y += correctionY
	}
}

type Box struct {
	width  float32
	height float32
}

func Run() {
	screenWidth := float32(1280)
	screenHeight := float32(800)

	rl.InitWindow(1280, 800, "2D Physics with Mass")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	balls := make([]Particle, 2)
	// Create balls with different masses
	// balls[0] = Particle{
	// 	position: rl.Vector2{X: 200.0, Y: 100.0}, // Start near top
	// 	velocity: rl.Vector2{X: float32(100 + 50), Y: 0.0},
	// 	radius:   15.0,
	// 	mass:     float32((float64(7.35) * math.Pow(10, 22))),
	// }
	// balls[1] = Particle{
	// 	position: rl.Vector2{X: 800.0, Y: 100.0}, // Start near top
	// 	velocity: rl.Vector2{X: -float32(100 + 50), Y: 0.0},
	// 	radius:   15.0,
	// 	mass:     float32((float64(7.35) * math.Pow(10, 22))),
	// }
	balls[0] = Particle{
		position: rl.Vector2{X: 400, Y: 400},
		velocity: rl.Vector2{X: 10, Y: 40},
		radius:   30,
		mass:     1000, // Large mass
	}
	balls[1] = Particle{
		position: rl.Vector2{X: 800, Y: 400},
		velocity: rl.Vector2{X: -10, Y: 50}, // Initial velocity
		radius:   20,
		mass:     100, // Smaller mass
	}

	for !rl.WindowShouldClose() {
		delta := rl.GetFrameTime() // Get delta time
		for i := range balls {
			for j := range balls {
				if i == j {
					continue
				}
				dx := balls[j].position.X - balls[i].position.X
				dy := balls[j].position.Y - balls[i].position.Y
				distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
				// Skip if balls are touching
				if distance < balls[i].radius+balls[j].radius {
					continue
				}
				direction := rl.Vector2{X: dx / distance, Y: dy / distance}
				distance *= 1000
				gForce := (GRAVITY * balls[j].mass * balls[i].mass) / (float32(distance) * float32(distance))
				acc1 := gForce / balls[i].mass
				acc := rl.Vector2{X: acc1 * direction.X, Y: acc1 * direction.Y}
				balls[i].Accelerate(acc.X, acc.Y, delta)
				balls[j].Accelerate(acc.X, acc.Y, delta)
			}
			balls[i].Update(delta, screenWidth, screenHeight)
			// balls[i].Accelerate(0.0, -9.81, delta)
		}
		for i := range balls {
			for j := i + 1; j < len(balls); j++ {
				if balls[i].CollidesWith(&balls[j]) {
					resolveCollision(&balls[i], &balls[j])
				}

			}
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
