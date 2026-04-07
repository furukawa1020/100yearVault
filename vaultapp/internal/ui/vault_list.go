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
	Mass                float32 
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
	
	// Interaction & 3D Analytics
	MousePos      f32.Point
	PrevMousePos  f32.Point
	MouseVelocity f32.Point
	MouseDepth    float32 // Virtual Z-depth driven by speed
	
	// Online 3x3 Covariance Tracker
	MuX, MuY, MuZ                       float32
	Cxx, Cyy, Czz, Cxy, Cxz, Cyz        float32

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
		p.Mass = float32(1.0 + rand.Float64()*4.0) // Mass range 1.0 to 5.0
		p.Color = ColorDataFragments[rand.Intn(len(ColorDataFragments))]
	}
	s.InitOnce = true

	// Initial pseudo-identity for Covariance to avoid zero determinant mapping
	if s.Cxx == 0 {
		s.Cxx, s.Cyy, s.Czz = 100.0, 100.0, 100.0 
	}
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
			
			// --- 3D Online Covariance & Analytics Tracker ---
			vX := s.MousePos.X - s.PrevMousePos.X
			vY := s.MousePos.Y - s.PrevMousePos.Y
			s.MouseVelocity.X = s.MouseVelocity.X*0.7 + vX*0.3
			s.MouseVelocity.Y = s.MouseVelocity.Y*0.7 + vY*0.3
			s.PrevMousePos = s.MousePos

			velSq := s.MouseVelocity.X*s.MouseVelocity.X + s.MouseVelocity.Y*s.MouseVelocity.Y
			speed := float32(math.Sqrt(float64(velSq)))
			
			// Virtual Z-depth (diving into galaxy) mapped to cursor speed
			targetDepth := -speed * 4.0
			s.MouseDepth = s.MouseDepth*0.8 + targetDepth*0.2
			vWZ := s.MouseDepth - targetDepth 

			cursorX := s.MousePos.X - center.X
			cursorY := s.MousePos.Y - center.Y
			cursorZ := s.MouseDepth 

			// Online Covariance Learning (learning rate alpha = 0.05)
			alpha := float32(0.05)
			s.MuX = s.MuX*(1-alpha) + cursorX*alpha
			s.MuY = s.MuY*(1-alpha) + cursorY*alpha
			s.MuZ = s.MuZ*(1-alpha) + cursorZ*alpha

			dX := cursorX - s.MuX
			dY := cursorY - s.MuY
			dZ := cursorZ - s.MuZ

			s.Cxx = s.Cxx*(1-alpha) + (dX*dX)*alpha
			s.Cyy = s.Cyy*(1-alpha) + (dY*dY)*alpha
			s.Czz = s.Czz*(1-alpha) + (dZ*dZ)*alpha
			s.Cxy = s.Cxy*(1-alpha) + (dX*dY)*alpha
			s.Cxz = s.Cxz*(1-alpha) + (dX*dZ)*alpha
			s.Cyz = s.Cyz*(1-alpha) + (dY*dZ)*alpha

			// 3x3 Determinant and Matrix Inverse
			reg := float32(50.0) // Small Tikhonov regularization
			cxx := s.Cxx + reg + speed*speed*8.0 // Field naturally stretches rapidly along high variance directions
			cyy := s.Cyy + reg + speed*speed*8.0
			czz := s.Czz + reg + speed*speed*4.0

			det := cxx*(cyy*czz - s.Cyz*s.Cyz) - s.Cxy*(s.Cxy*czz - s.Cyz*s.Cxz) + s.Cxz*(s.Cxy*s.Cyz - cyy*s.Cxz)
			var Ixx, Iyy, Izz, Ixy, Ixz, Iyz float32
			if det > 0.0001 {
				invDet := 1.0 / det
				Ixx = (cyy*czz - s.Cyz*s.Cyz) * invDet
				Iyy = (cxx*czz - s.Cxz*s.Cxz) * invDet
				Izz = (cxx*cyy - s.Cxy*s.Cxy) * invDet
				Ixy = (s.Cxz*s.Cyz - s.Cxy*czz) * invDet
				Ixz = (s.Cxy*s.Cyz - cyy*s.Cxz) * invDet
				Iyz = (s.Cxy*s.Cxz - cxx*s.Cyz) * invDet
			} else {
				Ixx, Iyy, Izz = 1.0/reg, 1.0/reg, 1.0/reg // Fallback to spherical
			}

			closestDistSq := float32(math.MaxFloat32)

			for i := range s.Particles {
				p := &s.Particles[i]
				rot := float64(s.Rotation)
				sinR, cosR := float32(math.Sin(rot)), float32(math.Cos(rot))
				
				// 1. True 3D Base Orbit Coordinates
				tx := p.BaseX*cosR - p.BaseZ*sinR
				ty := p.BaseY
				tz := p.BaseX*sinR + p.BaseZ*cosR
				
				// 2. 3D Spring Physics (Restoring force is mass-dependent)
				p.VX *= 0.88 // 3D Friction
				p.VY *= 0.88 
				p.VZ *= 0.88
				
				springK := 0.08 / p.Mass
				p.X += (0 - p.X) * springK
				p.Y += (0 - p.Y) * springK
				p.Z += (0 - p.Z) * springK

				// 3. Fluid Electromagnetic Mechanics in true Mahalanobis Space
				worldPX := tx + p.X
				worldPY := ty + p.Y
				worldPZ := tz + p.Z
				
				vecX := worldPX - cursorX
				vecY := worldPY - cursorY
				vecZ := worldPZ - cursorZ

				// Quadratic form for Mahalanobis field check
				mahaD2 := vecX*(Ixx*vecX + Ixy*vecY + Ixz*vecZ) + 
						  vecY*(Ixy*vecX + Iyy*vecY + Iyz*vecZ) + 
						  vecZ*(Ixz*vecX + Iyz*vecY + Izz*vecZ)

				if speed > 0.5 && mahaD2 < 1.0 {
					// Impact strength governed by matrix proximity and mass inertia
					intensity := float32(math.Pow(float64(1.0 - mahaD2), 0.5)) * speed * (1.0 / p.Mass)
					
					// Outward Hydrodynamic Repulsion 
					p.VX += vecX * intensity * 0.15
					p.VY += vecY * intensity * 0.15
					p.VZ += vecZ * intensity * 0.15
					
					// Vector Curl (Biot-Savart Lorentz Force proxy: v x d)
					// Generates topological vortex twisting the space behind pointer
					curlX := s.MouseVelocity.Y*vecZ - vWZ*vecY
					curlY := vWZ*vecX - s.MouseVelocity.X*vecZ
					curlZ := s.MouseVelocity.X*vecY - s.MouseVelocity.Y*vecX
					
					p.VX += curlX * intensity * 0.05
					p.VY += curlY * intensity * 0.05
					p.VZ += curlZ * intensity * 0.05
				}

				// Soft gravity simulation (Hover interaction in 2D Screen projection logic)
				screenDX := (worldPX+center.X) - s.MousePos.X
				screenDY := (worldPY+center.Y) - s.MousePos.Y
				screenDistSq := screenDX*screenDX + screenDY*screenDY

				if screenDistSq < closestDistSq {
					closestDistSq = screenDistSq
				}

				if speed < 0.5 && screenDistSq < 4000 && len(s.Memories) > 0 {
					p.VX -= vecX * 0.003 / p.Mass
					p.VY -= vecY * 0.003 / p.Mass
					p.VZ -= vecZ * 0.003 / p.Mass
				}

				// Accumulate integrations
				p.X += p.VX
				p.Y += p.VY
				p.Z += p.VZ

				// 4. Perspective Camera Projection 
				finalZ := tz + p.Z
				scale := focalLength / (focalLength + finalZ)
				if finalZ < -focalLength+50 { continue }
				
				sx := center.X + (tx+p.X)*scale
				sy := center.Y + (ty+p.Y)*scale

				pSize := 1.5 * scale * float32(math.Max(1.0, float64(p.Mass)*0.4))
				pColor := p.Color
				
				if screenDistSq < 2500 { 
					pSize *= 2.5
					pColor.A = 255
				} else {
					pColor.A = uint8(180 * scale)
				}

				// Assign Memory to locked target
				if screenDistSq < 2500 && len(s.Memories) > 0 && screenDistSq == closestDistSq {
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
