package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

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
	GazeActive    bool
	NeuralSurface widget.Clickable
	FrameCount    int

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
		// Background (No interactive handling here, main.go will do it)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, ColorBackground, clip.Rect{Max: gtx.Constraints.Max}.Op())
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		// 3D Galaxy Layer
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
			focalLength := float32(1000) 
			s.NeuralMemory = nil 
			
			// --- Frame Velocity Calculation ---
			// We calculate velocity here so it gracefully decays to 0 when mouse stops.
			vX := s.MousePos.X - s.PrevMousePos.X
			vY := s.MousePos.Y - s.PrevMousePos.Y
			// Exponential moving average for velocity to stabilize interactions
			s.MouseVelocity.X = s.MouseVelocity.X*0.7 + vX*0.3
			s.MouseVelocity.Y = s.MouseVelocity.Y*0.7 + vY*0.3
			s.PrevMousePos = s.MousePos

			velSq := s.MouseVelocity.X*s.MouseVelocity.X + s.MouseVelocity.Y*s.MouseVelocity.Y
			speed := float32(math.Sqrt(float64(velSq)))
			
			// Vector normalization for Mahalanobis field
			uX, uY := float32(0), float32(0)
			if speed > 0.1 {
				uX = s.MouseVelocity.X / speed
				uY = s.MouseVelocity.Y / speed
			}

			// Mahalanobis constants
			// n is perpendicular to u
			nX := -uY
			nY := uX

			// Minor/Major axis squares
			// Base radius is 80 (distSq 6400). Stretch along u axis based on speed.
			a2 := float32(6400.0) // Axis perpendicular to movement
			b2 := float32(6400.0) + speed*speed*50.0 // Axis along movement (gets stretched significantly)

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
				// d is vector from particle's projected position to mouse
				dx := baseSx + p.X - s.MousePos.X
				dy := baseSy + p.Y - s.MousePos.Y
				euclidDistSq := dx*dx + dy*dy

				if euclidDistSq < closestDistSq {
					closestDistSq = euclidDistSq
				}

				if speed > 0.1 {
					// Project dx, dy onto u and n axes
					du := dx*uX + dy*uY
					dn := dx*nX + dy*nY

					// Mahalanobis distance squared calculation
					mahaD2 := (du*du)/b2 + (dn*dn)/a2

					if mahaD2 < 1.0 { // Inside the hyper-ellipse
						// 速度向きに対するはんぱつ (Repulsion perpendicular to velocity + slight forward push)
						// 強さ (1.0 - mahaD2) に比例
						force := (1.0 - mahaD2) * speed * 2.0
						
						// 斥力の向き: 横への弾き飛ばし成分(nX, nY) をメインに、前方成分(uX, uY)も少し混ぜる
						pushDirX := nX
						pushDirY := nY
						// 左右どちらに避けるべきか判定 (dn の符号)
						if dn < 0 { 
							pushDirX = -nX
							pushDirY = -nY
						}
						
						p.VX += pushDirX * force * 0.5
						p.VY += pushDirY * force * 0.5
						p.VX += uX * force * 0.2 // Slight push forward
						p.VY += uY * force * 0.2
					}
				} else if euclidDistSq < 2500 && len(s.Memories) > 0 {
					// Static Hover effect
					p.VX += dx * 0.02
					p.VY += dy * 0.02
				}

				// Apply physics velocities
				p.X += p.VX
				p.Y += p.VY

				// Actual screen coordinates mapping including displacement
				sx := baseSx + p.X
				sy := baseSy + p.Y

				pSize := 1.5 * scale
				pColor := p.Color
				
				// Static Hover effect logic for highlighting and selection
				if euclidDistSq < 2500 { 
					pSize *= 3.5
					pColor.A = 255
				} else {
					pColor.A = uint8(180 * scale)
				}

				// Assign closest memory to lock-on logic
				if euclidDistSq < 2500 && len(s.Memories) > 0 && euclidDistSq == closestDistSq {
					s.NeuralMemory = s.Memories[i%len(s.Memories)]
				}

				rect := image.Rect(int(sx), int(sy), int(sx+float32(pSize)), int(sy+float32(pSize)))
				paint.FillShape(gtx.Ops, pColor, clip.Rect(rect).Op())
			}
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		// HUD Layer
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(40)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return drawRawLabel(gtx, s.Theme, "NEURAL_MIRROR_CONNECTION_ESTABLISHED", 24, ColorPrimary)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								txt := fmt.Sprintf("DEBUG_COORD: (%.0f, %.0f) | SCAN: %v", s.MousePos.X, s.MousePos.Y, s.NeuralMemory != nil)
								if s.NeuralMemory != nil {
									txt = fmt.Sprintf("RESONANCE_LOCKED: %s (SYNC_READY)", s.NeuralMemory.Title)
								}
								return drawRawLabel(gtx, s.Theme, txt, 12, ColorPrimaryDim)
							}),
						)
					})
				}),
				
				layout.Flexed(1, layout.Spacer{}.Layout),

				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(s.Theme, &s.NewVaultBtn, "RELEASE_NEW_MEMORY")
								btn.Background = ColorPrimary
								btn.Color = ColorBackground
								btn.TextSize = unit.Sp(32)
								btn.Inset = layout.Inset{Top: unit.Dp(30), Bottom: unit.Dp(30), Left: unit.Dp(60), Right: unit.Dp(60)}
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
