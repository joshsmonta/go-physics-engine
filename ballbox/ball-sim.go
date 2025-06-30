package ballbox

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	GRAVITY          = 1000.0 // Scaled gravitational constant
	BOUNCE_DAMPENING = 0.95   // Energy retained after bounce
	AIR_RESISTANCE   = 0.999  // Minimal air resistance
	GROUND_FRICTION  = 0.85   // Ground friction coefficient
)

type Particle struct {
	position rl.Vector2
	velocity rl.Vector2
	radius   float32
	mass     float32
	color    rl.Color
}

// Update handles wall collisions
func (p *Particle) Update(delta float32, width, height float32) {
	// Apply minimal air resistance
	// p.velocity.X *= AIR_RESISTANCE
	// p.velocity.Y *= AIR_RESISTANCE

	// Handle collisions with screen edges
	if p.position.Y > height-p.radius {
		p.position.Y = height - p.radius
		p.velocity.Y *= -BOUNCE_DAMPENING
	}
	if p.position.Y < p.radius {
		p.position.Y = p.radius
		p.velocity.Y *= -BOUNCE_DAMPENING
	}
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
	dx := p2.position.X - p1.position.X
	dy := p2.position.Y - p1.position.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if distance == 0 {
		return
	}

	nx := dx / distance
	ny := dy / distance

	rvx := p2.velocity.X - p1.velocity.X
	rvy := p2.velocity.Y - p1.velocity.Y

	velocityAlongNormal := rvx*nx + rvy*ny

	if velocityAlongNormal > 0 {
		return
	}

	impulse := -(1.0 + BOUNCE_DAMPENING) * velocityAlongNormal
	impulse /= 1/p1.mass + 1/p2.mass

	impulseX := impulse * nx
	impulseY := impulse * ny

	p1.velocity.X -= impulseX / p1.mass
	p1.velocity.Y -= impulseY / p1.mass
	p2.velocity.X += impulseX / p2.mass
	p2.velocity.Y += impulseY / p2.mass

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

func calculateOrbitalVelocity(mass float32) float32 {
	initialDistance := float32(300.0)
	orbitalVelocity := float32(math.Sqrt(float64(GRAVITY * mass / (initialDistance))))
	return orbitalVelocity
}

func Run() {
	screenWidth := float32(1280)
	screenHeight := float32(800)

	rl.InitWindow(1280, 800, "Orbiting Moons")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	initialDistance := float32(300.0)

	// Masses scaled down by 10^20 for simulation stability
	// sunMass := float32(1.989e30 / 1e20)   // Mass of the Sun: ~1.989 × 10³⁰ kg
	// earthMass := float32(5.972e24 / 1e20) // Mass of the Earth: ~5.972 × 10²⁴ kg
	moonMass := float32(7.35e22 / 1e20)

	balls := make([]Particle, 2)
	balls[0] = Particle{
		position: rl.Vector2{X: screenWidth/2 - initialDistance/2, Y: screenHeight / 2},
		velocity: rl.Vector2{X: 0, Y: -calculateOrbitalVelocity(moonMass)}, // Initial orbital velocity
		radius:   30,
		mass:     moonMass,
		color:    rl.Yellow,
	}
	balls[1] = Particle{
		position: rl.Vector2{X: screenWidth/2 + initialDistance/2, Y: screenHeight / 2},
		velocity: rl.Vector2{X: 0, Y: calculateOrbitalVelocity(moonMass)}, // Opposite direction
		radius:   30,
		mass:     moonMass,
		color:    rl.SkyBlue,
	}

	for !rl.WindowShouldClose() {
		subDelta := rl.GetFrameTime()
		// Apply gravitational forces
		for i := range balls {
			for j := range balls {
				if i == j {
					continue
				}
				dx := balls[j].position.X - balls[i].position.X
				dy := balls[j].position.Y - balls[i].position.Y
				distanceSq := dx*dx + dy*dy
				distance := float32(math.Sqrt(float64(distanceSq)))

				if distance == 0 {
					continue
				}

				// Gravitational acceleration (a = F/m = G*M/distance^2)
				acceleration := GRAVITY * balls[j].mass / distanceSq
				directionX := dx / distance
				directionY := dy / distance

				// Update velocity with acceleration
				balls[i].velocity.X += acceleration * directionX * subDelta
				balls[i].velocity.Y += acceleration * directionY * subDelta
			}
		}

		// Update positions
		for i := range balls {
			balls[i].position.X += balls[i].velocity.X * subDelta
			balls[i].position.Y += balls[i].velocity.Y * subDelta
			balls[i].Update(subDelta, screenWidth, screenHeight)
		}

		// Check collisions
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
		for _, ball := range balls {
			rl.DrawCircleV(ball.position, ball.radius, ball.color)
		}
		rl.EndDrawing()
	}
}
