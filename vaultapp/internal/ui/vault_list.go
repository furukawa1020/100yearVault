package ui

import (
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
	Mass, Drag          float32 // Advanced properties
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
	FaceHistory   [][]f32.Point
	PulseStrength float32
	FaceScale     float32
	GazeActive    bool
	NeuralSurface widget.Clickable
	FrameCount    int
	FaceMu        sync.Mutex

	// Advanced Physics tracking
	InputHistory [128]f32.Point
	HistPtr      int

	// Screen Navigation
	CurrentScreen Screen
	Compose       ComposeState
	Ritual        RitualState
}

const (
	TotalParticles = 16384 // Ultra High Density
)

func (s *AppState) initNeuralSpace() {
	if s.InitOnce { return }
	s.Particles = make([]Particle, TotalParticles)
	for i := range s.Particles {
		p := &s.Particles[i]
		a1, a2 := rand.Float64()*2*math.Pi, rand.Float64()*math.Pi
		dX, dY, dZ := 100+rand.Float64()*650, 100+rand.Float64()*550, 200+rand.Float64()*1000
		p.BaseX = float32(math.Sin(a2)*math.Cos(a1) * dX)
		p.BaseY = float32(math.Sin(a2)*math.Sin(a1) * dY)
		p.BaseZ = float32(math.Cos(a2) * dZ)
		p.Mass = 0.5 + rand.Float32()*2.0
		p.Drag = 0.92 + rand.Float32()*0.06
		p.Color = ColorDataFragments[rand.Intn(len(ColorDataFragments))]
	}
	s.InitOnce = true
}

func (s *AppState) RotateNeural() {
	s.initNeuralSpace()
	num := len(s.Memories)
	if num == 0 { return }
	for i := range s.Particles {
		p, m := &s.Particles[i], s.Memories[i%num]
		hash := float32(0)
		for _, c := range m.ID { hash += float32(c) }
		a, d := float64(hash*0.1), 200.0+math.Mod(float64(hash), 400.0)
		spr := 60.0 + math.Mod(float64(hash), 100.0)
		p.BaseX = float32(math.Cos(a)*d) + float32((rand.Float64()-0.5)*spr)
		p.BaseY = float32(math.Sin(a)*d) + float32((rand.Float64()-0.5)*spr)
		p.BaseZ = float32(math.Sin(a*0.5)*150.0) + float32((rand.Float64()-0.5)*spr)
		if m.Aura == vault.StateRadiant { p.Color = color.NRGBA{255, 255, 200, 255} }
	}
}

func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	s.FrameCount++
	s.Rotation += 0.0018

	// --- 1. Statistical Context Calculation (Real Mahalanobis) ---
	tPos := s.MousePos
	if s.GazeActive && s.MouseVelocity.X*s.MouseVelocity.X+s.MouseVelocity.Y*s.MouseVelocity.Y < 1.0 {
		tPos = s.GazePos
	}
	s.InputHistory[s.HistPtr] = tPos
	s.HistPtr = (s.HistPtr + 1) % len(s.InputHistory)

	var muX, muY float32
	for _, p := range s.InputHistory {
		muX += p.X
		muY += p.Y
	}
	muX /= float32(len(s.InputHistory))
	muY /= float32(len(s.InputHistory))

	var varX, varY, covXY float32
	for _, p := range s.InputHistory {
		dx, dy := p.X-muX, p.Y-muY
		varX += dx * dx
		varY += dy * dy
		covXY += dx * dy
	}
	N := float32(len(s.InputHistory))
	varX /= N
	varY /= N
	covXY /= N

	// Add noise to stability (Singularity Protection)
	varX += 100.0
	varY += 100.0

	// Matrix Inversion for 2x2 Covariance Sigma
	det := varX*varY - covXY*covXY
	if det < 0.1 { det = 0.1 }
	invVarX := varY / det
	invVarY := varX / det
	invCovXY := -covXY / det

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, ColorBackground, clip.Rect{Max: gtx.Constraints.Max}.Op())
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
			focalLength := float32(1000)
			cosR, sinR := float32(math.Cos(float64(s.Rotation))), float32(math.Sin(float64(s.Rotation)))

			type screenPt struct {
				pos f32.Point; color color.NRGBA; scale float32; force float32; distSq float32; mDist float32
			}
			pts := make([]screenPt, len(s.Particles))

			for i := range s.Particles {
				p := &s.Particles[i]
				tx := p.BaseX*cosR - p.BaseZ*sinR
				ty, tz := p.BaseY, p.BaseX*sinR + p.BaseZ*cosR
				
				scale := focalLength / (focalLength + tz)
				bSx, bSy := center.X + tx*scale, center.Y + ty*scale
				
				dx, dy := tPos.X-bSx, tPos.Y-bSy
				// --- Real Mahalanobis Quadratic Form ---
				d2 := dx*dx + dy*dy
				mDist := dx*dx*invVarX + dy*dy*invVarY + 2*dx*dy*invCovXY
				
				if mDist < 12.0 { // Repel based on statistical distance
					f := (12.0 - mDist) * 0.4 / p.Mass
					p.VX -= (dx / (float32(math.Sqrt(float64(d2)))+0.1)) * f
					p.VY -= (dy / (float32(math.Sqrt(float64(d2)))+0.1)) * f
				}
				
				p.VX, p.VY = p.VX*p.Drag, p.VY*p.Drag
				p.X, p.Y = p.X+p.VX, p.Y+p.VY
				
				dScl := (1.0 - tz/1500.0)
				if dScl < 0.3 { dScl = 0.3 } 
				pts[i] = screenPt{pos: f32.Pt(bSx+p.X, bSy+p.Y), color: p.Color, scale: scale * dScl, distSq: d2, mDist: mDist}
				
				// Avatar High-Energy Interaction
				resF := float32(0)
				if s.GazeActive {
					s.FaceMu.Lock()
					fP, fH, fS := s.FacePoints, s.FaceHistory, s.FaceScale
					s.FaceMu.Unlock()
					bRad := float32(75.0) * fS * (1.0 + s.PulseStrength*0.3)
					runA := func(target []f32.Point, w float32) {
						for _, fp := range target {
							if fp.X == 0 { continue }
							fdx, fdy := pts[i].pos.X-fp.X, pts[i].pos.Y-fp.Y
							fD2 := fdx*fdx + fdy*fdy
							rad := bRad * w
							if fD2 < rad*rad {
								fdst := float32(math.Sqrt(float64(fD2)))
								lF := (1.0 - fdst/rad) * 4.5 * w / p.Mass
								p.VX += (fdx / (fdst+0.1)) * lF
								p.VY += (fdy / (fdst+0.1)) * lF
								if lF > resF { resF = lF }
							}
						}
					}
					runA(fP, 1.0); for _, h := range fH { runA(h, 0.4) }
				}
				pts[i].force = resF
			}

			// 2. Chromatic Tensor Constellations
			for i := 0; i < len(pts); i += 6 {
				for j := i + 1; j < i+18 && j < len(pts); j++ {
					dx, dy := pts[i].pos.X-pts[j].pos.X, pts[i].pos.Y-pts[j].pos.Y
					if dx*dx+dy*dy < 1400 {
						lineC := lerpColor(pts[i].color, pts[j].color, 0.5)
						// Lines stretch based on mahalanobis proximity
						lineC.A = uint8(50 * pts[i].scale)
						if pts[i].mDist < 5.0 { lineC.A = 180 }
						var pth clip.Path; pth.Begin(gtx.Ops); pth.MoveTo(pts[i].pos); pth.LineTo(pts[j].pos)
						paint.FillShape(gtx.Ops, lineC, clip.Stroke{Path: pth.End(), Width: 0.9}.Op())
					}
				}
			}

			// 3. Chromatic Fragment Rendering
			for i, pt := range pts {
				sz := 2.0 * pt.scale * (1.0 + s.PulseStrength*0.2)
				if sz < 0.8 { sz = 0.8 }
				pCl := pt.color
				
				if pt.distSq < 3000 || pt.mDist < 4.0 || pt.force > 0.5 {
					sz *= (3.8 + pt.force*2.5)
					if pt.force > 0.5 || pt.mDist < 4.0 {
						ripC := ColorPrimary; if i%2 == 0 { ripC = ColorSecondary }
						pCl = lerpColor(pCl, ripC, 0.6)
					}
					pCl.A = 255
					if pt.distSq < 2500 { pCl = lerpColor(pCl, ColorQuaternary, 0.7) }
				} else {
					shim := uint8(math.Sin(float64(s.FrameCount)*0.1+float64(i)*0.01) * 40)
					alpha := float64(220*pt.scale) + float64(shim)
					pCl.A = uint8(math.Max(40, math.Min(255, alpha)))
				}
				var pth clip.Path; pth.Begin(gtx.Ops); sx, sy := pt.pos.X, pt.pos.Y
				if i%17 == 0 { // Glyph
					pth.MoveTo(f32.Pt(sx, sy)); pth.LineTo(f32.Pt(sx+sz, sy+sz/2))
					pth.MoveTo(f32.Pt(sx+sz/2, sy)); pth.LineTo(f32.Pt(sx+sz/2, sy+sz))
				} else if i%13 == 0 { // Diamond
					pth.MoveTo(f32.Pt(sx+sz/2, sy)); pth.LineTo(f32.Pt(sx+sz, sy+sz/2))
					pth.LineTo(f32.Pt(sx+sz/2, sy+sz)); pth.LineTo(f32.Pt(sx, sy+sz/2)); pth.Close()
				} else { // Fragment
					pth.MoveTo(f32.Pt(sx, sy+sz)); pth.LineTo(f32.Pt(sx+sz/2, sy)); pth.LineTo(f32.Pt(sx+sz, sy+sz)); pth.Close()
				}
				paint.FillShape(gtx.Ops, pCl, clip.Outline{Path: pth.End()}.Op())
			}
			return layout.Dimensions{Size: gtx.Constraints.Max}
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
	if t > 1 { t = 1 }; if t < 0 { t = 0 }
	return color.NRGBA{
		R: uint8(float32(c1.R)*(1-t) + float32(c2.R)*t),
		G: uint8(float32(c1.G)*(1-t) + float32(c2.G)*t),
		B: uint8(float32(c1.B)*(1-t) + float32(c2.B)*t),
		A: uint8(float32(c1.A)*(1-t) + float32(c2.A)*t),
	}
}
