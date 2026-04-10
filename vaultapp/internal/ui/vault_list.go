package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"vaultapp/internal/vault"
)

type Screen int

const (
	ScreenVaultList Screen = iota
	ScreenCompose
	ScreenRitual
)

type Particle struct {
	BaseX, BaseY, BaseZ float32 
	X, Y, Z             float32 
	VX, VY, VZ          float32 
	Color               color.NRGBA
}

type AppState struct {
	Theme *material.Theme

	// Neural Mirror Core
	Memories     []*vault.MemoryFragment
	NeuralMemory *vault.MemoryFragment 
	SelectBtns   []widget.Clickable
	NewVaultBtn  widget.Clickable

	// Cosmic Galaxy
	Particles []Particle
	InitOnce  bool
	Rotation  float32
	
	// Interaction
	MousePos      f32.Point
	PrevMousePos  f32.Point
	MouseVelocity f32.Point
	GazePos       f32.Point
	GazeVelocity  f32.Point
	FacePoints    []f32.Point
	FaceHistory   [][]f32.Point // Added for trails [History][LandmarkID]
	PulseStrength float32       // Added for resonant pulsing
	FaceScale     float32
	GazeActive    bool
	NeuralSurface widget.Clickable
	FrameCount    int
	FaceMu        sync.Mutex // Added for thread-safe landmark updates

	// Screen Navigation
	CurrentScreen Screen
	Compose       ComposeState
	Ritual        RitualState
}

func (s *AppState) initNeuralSpace() {
	if s.InitOnce {
		return
	}
	s.Particles = make([]Particle, 4096)
	for i := range s.Particles {
		p := &s.Particles[i]
		angle1 := rand.Float64() * 2 * math.Pi
		angle2 := rand.Float64() * math.Pi
		distX := 150 + rand.Float64()*450
		distY := 100 + rand.Float64()*300
		distZ := 200 + rand.Float64()*600
		
		p.BaseX = float32(math.Sin(angle2)*math.Cos(angle1) * distX)
		p.BaseY = float32(math.Sin(angle2)*math.Sin(angle1) * distY)
		p.BaseZ = float32(math.Cos(angle2) * distZ)
		p.Color = ColorDataFragments[rand.Intn(len(ColorDataFragments))]
	}
	s.InitOnce = true
}

func (s *AppState) RotateNeural() {
	s.initNeuralSpace()
	
	numMemories := len(s.Memories)
	if numMemories == 0 { return }

	for i := range s.Particles {
		p := &s.Particles[i]
		m := s.Memories[i%numMemories]
		
		hash := float32(0)
		for _, c := range m.ID { hash += float32(c) }
		
		angle := float64(hash * 0.1)
		dist  := 200.0 + math.Mod(float64(hash), 300.0)
		
		targetX := float32(math.Cos(angle) * dist)
		targetY := float32(math.Sin(angle) * dist)
		targetZ := float32(math.Sin(angle*0.5) * 100.0)
		
		spread := 40.0 + math.Mod(float64(hash), 60.0)
		p.BaseX = targetX + float32((rand.Float64()-0.5)*spread)
		p.BaseY = targetY + float32((rand.Float64()-0.5)*spread)
		p.BaseZ = targetZ + float32((rand.Float64()-0.5)*spread)
		
		if m.Aura == vault.StateRadiant {
			p.Color = color.NRGBA{R: 255, G: 255, B: 200, A: 255} 
		}
	}
}

func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	s.FrameCount++

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		// Background
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, ColorBackground, clip.Rect{Max: gtx.Constraints.Max}.Op())
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		// 3D Galaxy Layer
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
			focalLength := float32(1000) 
			
			// Interaction Physics (Unchanged from existing version)
			vX := s.MousePos.X - s.PrevMousePos.X
			vY := s.MousePos.Y - s.PrevMousePos.Y
			s.MouseVelocity.X = s.MouseVelocity.X*0.7 + vX*0.3
			s.MouseVelocity.Y = s.MouseVelocity.Y*0.7 + vY*0.3
			s.PrevMousePos = s.MousePos

			targetPos := s.MousePos
			if s.GazeActive {
				mouseActiveSpeedSq := s.MouseVelocity.X*s.MouseVelocity.X + s.MouseVelocity.Y*s.MouseVelocity.Y
				uX = targetVel.X / speed
				uY = targetVel.Y / speed
			}

			// Mahalanobis constants
			// n is perpendicular to u
			nX := -uY
			nY := uX

			// Minor/Major axis squares
			// Base radius is 80 (distSq 6400). Stretch along u axis based on speed.
			a2 := float32(6400.0) 
			b2 := float32(6400.0) + speed*speed*50.0 
			
			// Depth-based modifications (Z-Axis Interaction)
			depthScale := s.FaceScale
			if depthScale < 0.1 { depthScale = 1.0 } // Default fallback
			a2 *= depthScale
			b2 *= depthScale

			closestDistSq := float32(math.MaxFloat32)

			for i := range s.Particles {
				p := &s.Particles[i]
				rot := float64(s.Rotation)
				sinR, cosR := float32(math.Sin(rot)), float32(math.Cos(rot))
				
				// 1. Base analytic rotation (Galaxy Orbit)
				tx := p.BaseX*cosR - p.BaseZ*sinR
				tz := p.BaseX*sinR + p.BaseZ*cosR
				ty := p.BaseY
				
				scale := focalLength / (focalLength + tz)
				if tz < -focalLength+50 { continue }

				baseSx := center.X + tx*scale
				baseSy := center.Y + ty*scale

				// 2. Spring Physics (Restoring force to base position)
				p.VX *= 0.85 // Friction
				p.VY *= 0.85 // Friction
				p.X += (0 - p.X) * 0.05 // Spring towards 0 local displacement
				p.Y += (0 - p.Y) * 0.05

				// 3. Fluid Repulsion (Mahalanobis Space)
				// Primary interaction point (Mouse or Gaze Center)
				dx := baseSx + p.X - targetPos.X
				dy := baseSy + p.Y - targetPos.Y
				euclidDistSq := dx*dx + dy*dy

				if euclidDistSq < closestDistSq {
					closestDistSq = euclidDistSq
				}

				if speed > 0.1 {
					du := dx*uX + dy*uY
					dn := dx*nX + dy*nY
					mahaD2 := (du*du)/b2 + (dn*dn)/a2

					if mahaD2 < 1.0 {
						force := (1.0 - mahaD2) * speed * 2.0
						pushDirX, pushDirY := nX, nY
						if dn < 0 { pushDirX, pushDirY = -nX, -nY }
						p.VX += pushDirX * force * 0.5
						p.VY += pushDirY * force * 0.5
						p.VX += uX * force * 0.2
						p.VY += uY * force * 0.2
					}
				} else if euclidDistSq < 2500 && len(s.Memories) > 0 {
					p.VX += dx * 0.02
					p.VY += dy * 0.02
				}

				pColor := p.Color
				
				// 4. Face Silhouette & Kinetic Echoes (Landmarks + Trails)
				avatarForce := float32(0)
				if s.GazeActive && speed < 1.0 { 
					s.FaceMu.Lock()
					fPoints := s.FacePoints
					fHistory := s.FaceHistory
					fScale := s.FaceScale
					fPulse := s.PulseStrength
					s.FaceMu.Unlock()

					// Main Landmarks + History Trails
					// We loop current points (weight 1.0) and history (decaying weight)
					baseRadius := float32(60.0) * fScale * (1.0 + fPulse*0.3)
					
					// Helper func for repulsion + coloring
					applyRepulsion := func(points []f32.Point, weight float32) {
						for i, fp := range points {
							if fp.X == 0 && fp.Y == 0 { continue }
							fdx := baseSx + p.X - fp.X
							fdy := baseSy + p.Y - fp.Y
							fDistSq := fdx*fdx + fdy*fdy
							
							radius := baseRadius * weight
							if fDistSq < radius*radius {
								fdist := float32(math.Sqrt(float64(fDistSq)))
								if fdist < 0.1 { fdist = 0.1 }
								
								// Force calculation with weight decay
								localForce := (1.0 - fdist/radius) * 3.5 * weight
								p.VX += (fdx / fdist) * localForce
								p.VY += (fdy / fdist) * localForce
								
								if localForce > avatarForce { avatarForce = localForce }
								
								// Chromatic Resonance + Spiritual Aura
								if i < 2 { // Eyes
									// If extremely close to center, turn into "Inner Light" (Aura)
									if localForce > 2.5 {
										pColor = lerpColor(pColor, color.NRGBA{255, 255, 240, 255}, (localForce-2.5)*2)
									} else {
										pColor = lerpColor(pColor, ColorPrimary, localForce*0.7)
									}
								} else if i == 3 { // Mouth
									pColor = lerpColor(pColor, ColorSecondary, localForce*0.7)
								}
							}
						}
					}

					// 1. Current Frame (Full weight)
					applyRepulsion(fPoints, 1.0)
					
					// 2. Kinetic Echoes (Historical frames with decay)
					for hIdx, hp := range fHistory {
						// hIdx 0 is newest, len-1 is oldest
						hWeight := float32(1.0 - float32(hIdx+1)/float32(len(fHistory)+1))
						applyRepulsion(hp, hWeight * 0.4) // Ghostly trails are 40% strength max
					}
				}

				// Apply physics velocities
				p.X += p.VX
				p.Y += p.VY

				// 5. Drawing (The Monolithic Materialization)
				// We transform the pixel into a material fragment.
				sx := baseSx + p.X
				sy := baseSy + p.Y
				
				pSize := 1.5 * scale * depthScale * (1.0 + s.PulseStrength*0.2)
				shimmer := float32(math.Sin(float64(s.FrameCount)*0.1 + float64(i)*0.01)) * 30
				
				// Static Hover / Interaction Highlighting
				if euclidDistSq < 2500 || avatarForce > 0.5 { 
					pSize *= (3.5 + avatarForce*2.0) // Aura particles grow even larger
					pColor.A = 255
					if euclidDistSq < 2500 {
						pColor = lerpColor(pColor, ColorQuaternary, 0.6) // Interaction Glow (Gold)
					}
				} else {
					pColor.A = uint8(math.Max(0, math.Min(255, float64(180*scale)+float64(shimmer))))
				}

				// Assign closest memory to lock-on logic
				if euclidDistSq < 2500 && len(s.Memories) > 0 && euclidDistSq == closestDistSq {
					s.NeuralMemory = s.Memories[i%len(s.Memories)]
				}

								return btn.Layout(gtx)
							}),
							layout.Flexed(1, layout.Spacer{}.Layout),
						)
					})
				}),
			)
		}),
	)
}

func drawRawLabel(gtx layout.Context, th *material.Theme, txt string, size int, clr color.NRGBA) layout.Dimensions {
	label := material.Label(th, unit.Sp(float32(size)), txt)
	label.Color = clr
	label.Font.Typeface = "Consolas"
	label.Alignment = text.Middle
	return label.Layout(gtx)
}

func drawBackground(ops *op.Ops, gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(ops, c, clip.Rect(dr).Op())
}

func lerpColor(c1, c2 color.NRGBA, t float32) color.NRGBA {
	if t > 1 { t = 1 }
	if t < 0 { t = 0 }
	return color.NRGBA{
		R: uint8(float32(c1.R)*(1-t) + float32(c2.R)*t),
		G: uint8(float32(c1.G)*(1-t) + float32(c2.G)*t),
		B: uint8(float32(c1.B)*(1-t) + float32(c2.B)*t),
		A: uint8(float32(c1.A)*(1-t) + float32(c2.A)*t),
	}
}
