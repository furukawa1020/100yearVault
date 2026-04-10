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
	Spin, Mass, Drag    float32
	Color               color.NRGBA
}

type AppState struct {
	Theme *material.Theme

	// Digital Mirror Core
	Memories     []*vault.MemoryFragment
	NeuralMemory *vault.MemoryFragment 
	SelectBtns   []widget.Clickable
	NewVaultBtn  widget.Clickable

	// Cosmic Manifold
	Particles []Particle
	InitOnce  bool
	Rotation  float32
	
	// Interaction
	MousePos, PrevMousePos f32.Point
	MouseVelocity          f32.Point
	GazePos, GazeVelocity  f32.Point
	FacePoints             []f32.Point
	FaceHistory            [][]f32.Point
	PulseStrength          float32
	FaceScale              float32
	GazeActive             bool
	NeuralSurface          widget.Clickable
	FrameCount             int
	FaceMu                 sync.Mutex

	// 5D Statistical Tracking [x, y, z, vx, vy]
	History5D [128][5]float32
	HistPtr   int
	InvS      [3][3]float32 
	InvV      [2][2]float32 

	// Screen Navigation
	CurrentScreen Screen
	Compose       ComposeState
	Ritual        RitualState
}

const (
	TotalParticles = 10240 
)

func (s *AppState) initNeuralSpace() {
	if s.InitOnce { return }
	s.Particles = make([]Particle, TotalParticles)
	for i := range s.Particles {
		p := &s.Particles[i]
		a1, a2 := rand.Float64()*2*math.Pi, rand.Float64()*math.Pi
		dX, dY, dZ := 100+rand.Float64()*650, 100+rand.Float64()*550, 250+rand.Float64()*900
		p.BaseX = float32(math.Sin(a2)*math.Cos(a1) * dX)
		p.BaseY = float32(math.Sin(a2)*math.Sin(a1) * dY)
		p.BaseZ = float32(math.Cos(a2) * dZ)
		p.Mass = 0.4 + rand.Float32()*0.6 // Lighter for better fluidity
		p.Drag = 0.94 + rand.Float32()*0.02
		p.Spin = rand.Float32() * 0.1
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
	s.Rotation += 0.0025

	// 1. Statistical Context
	tPos := s.MousePos
	tVel := s.MouseVelocity
	if s.GazeActive && tVel.X*tVel.X+tVel.Y*tVel.Y < 1.0 {
		tPos, tVel = s.GazePos, s.GazeVelocity
	}
	tZ := 500.0 * (1.0 - s.FaceScale)
	s.History5D[s.HistPtr] = [5]float32{tPos.X, tPos.Y, tZ, tVel.X, tVel.Y}
	s.HistPtr = (s.HistPtr + 1) % 128

	var m3 [3]float32; for _, v := range s.History5D { m3[0],m3[1],m3[2] = m3[0]+v[0],m3[1]+v[1],m3[2]+v[2] }
	m3[0], m3[1], m3[2] = m3[0]/128, m3[1]/128, m3[2]/128
	var c3 [3][3]float32
	for _, v := range s.History5D {
		d0, d1, d2 := v[0]-m3[0], v[1]-m3[1], v[2]-m3[2]
		c3[0][0]+=d0*d0; c3[1][1]+=d1*d1; c3[2][2]+=d2*d2; c3[0][1]+=d0*d1; c3[0][2]+=d0*d2; c3[1][2]+=d1*d2
	}
	for i := 0; i < 3; i++ { for j := 0; j < 3; j++ { c3[i][j] /= 128.0 } }
	c3[0][0]+=200; c3[1][1]+=200; c3[2][2]+=200; // Stabilizers lowered for snappier warping
	c3[1][0], c3[2][0], c3[2][1] = c3[0][1], c3[0][2], c3[1][2]

	det := c3[0][0]*(c3[1][1]*c3[2][2]-c3[1][2]*c3[2][1]) - c3[0][1]*(c3[1][0]*c3[2][2]-c3[1][2]*c3[2][0]) + c3[0][2]*(c3[1][0]*c3[2][1]-c3[1][1]*c3[2][0])
	if det < 0.1 { det = 0.1 }
	s.InvS[0][0], s.InvS[1][1], s.InvS[2][2] = (c3[1][1]*c3[2][2]-c3[1][2]*c3[2][1])/det, (c3[0][0]*c3[2][2]-c3[0][2]*c3[2][0])/det, (c3[0][0]*c3[1][1]-c3[0][1]*c3[1][0])/det
	s.InvS[0][1] = (c3[0][2]*c3[2][1]-c3[0][1]*c3[2][2])/det

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
				pos f32.Point; color color.NRGBA; scale float32; force float32; mDist float32
			}
			pts := make([]screenPt, TotalParticles)

			var wg sync.WaitGroup
			numG := 4
			batchSize := TotalParticles / numG
			
			s.FaceMu.Lock(); fP, fH, fS := s.FacePoints, s.FaceHistory, s.FaceScale; s.FaceMu.Unlock()
			bRad := float32(85.0) * fS * (1.0 + s.PulseStrength*0.3)

			for g := 0; g < numG; g++ {
				wg.Add(1)
				go func(start, end int) {
					defer wg.Done()
					for i := start; i < end; i++ {
						p := &s.Particles[i]
						tx, ty, tz := p.BaseX*cosR-p.BaseZ*sinR, p.BaseY, p.BaseX*sinR+p.BaseZ*cosR
						
						// Perspective Clamping
						scale := focalLength / (focalLength + tz)
						if scale > 1.8 { scale = 1.8 } 

						bSx, bSy := center.X+tx*scale, center.Y+ty*scale
						dx, dy, dz := tPos.X-(bSx+p.X), tPos.Y-(bSy+p.Y), tZ-tz
						mDist := dx*dx*s.InvS[0][0] + dy*dy*s.InvS[1][1] + dz*dz*s.InvS[2][2] + 2*dx*dy*s.InvS[0][1]

						// Drastic Interaction Sharpening
						if mDist < 16.0 {
							f := (16.0 - mDist) * 1.25 / p.Mass // Power increased
							p.VX -= (dx / (float32(math.Sqrt(float64(dx*dx+dy*dy)))+0.1)) * f
							p.VY -= (dy / (float32(math.Sqrt(float64(dx*dx+dy*dy)))+0.1)) * f
							p.Spin += f * 0.1
						}
						
						// Elastic Restoration
						p.VX, p.VY = p.VX + (0 - p.X)*0.07, p.VY + (0 - p.Y)*0.07
						
						resF := float32(0)
						if s.GazeActive {
							runA := func(target []f32.Point, w float32) {
								for _, fp := range target {
									if fp.X == 0 { continue }
									fdx, fdy := (bSx+p.X)-fp.X, (bSy+p.Y)-fp.Y
									fD2 := fdx*fdx + fdy*fdy; rad := bRad * w
									if fD2 < rad*rad {
										fdst := float32(math.Sqrt(float64(fD2)))
										lF := (1.0 - fdst/rad) * 5.5 * w / p.Mass
										p.VX += (fdx / (fdst+0.1)) * lF; p.VY += (fdy / (fdst+0.1)) * lF
										if lF > resF { resF = lF }
									}
								}
							}
							runA(fP, 1.0); for _, h := range fH { runA(h, 0.4) }
						}

						p.VX, p.VY = p.VX*p.Drag, p.VY*p.Drag
						p.X, p.Y = p.X+p.VX, p.Y+p.VY
						dScl := (1.0 - tz/1500.0); if dScl < 0.4 { dScl = 0.4 }
						
						pts[i] = screenPt{pos: f32.Pt(bSx+p.X, bSy+p.Y), color: p.Color, scale: scale * dScl, force: resF, mDist: mDist}
					}
				}(g*batchSize, (g+1)*batchSize)
			}
			wg.Wait()

			// 2. Fragment Constellations
			for i := 0; i < len(pts); i += 12 {
				for j := i + 1; j < i+12 && j < len(pts); j++ {
					dx, dy := pts[i].pos.X-pts[j].pos.X, pts[i].pos.Y-pts[j].pos.Y
					if dx*dx+dy*dy < 1200 {
						lineC := lerpColor(pts[i].color, pts[j].color, 0.5)
						lineC.A = uint8(35 * pts[i].scale)
						var pth clip.Path; pth.Begin(gtx.Ops); pth.MoveTo(pts[i].pos); pth.LineTo(pts[j].pos)
						paint.FillShape(gtx.Ops, lineC, clip.Stroke{Path: pth.End(), Width: 0.6}.Op())
					}
				}
			}

			// 3. Fluid Fragment Rendering (Size-Clamped & Rect-Free)
			for i, pt := range pts {
				// Final size clamping for foreground sanity
				sz := 2.5 * pt.scale * (1.0 + s.PulseStrength*0.2)
				if sz < 0.8 { sz = 0.8 }
				if sz > 12.0 { sz = 12.0 } // CLAMPED
				pCl := pt.color
				
				isNear := pt.mDist < 4.0 || pt.force > 0.5
				if isNear {
					sz *= (3.5 + pt.force*2.0); pCl.A = 255
					if pt.force > 0.5 { pCl = lerpColor(pCl, ColorPrimary, 0.6) }
					if pt.mDist < 4.0 { pCl = lerpColor(pCl, ColorSecondary, 0.6) }
				} else {
					pCl.A = uint8(180 * pt.scale)
				}
				
				// NO MORE RECTANGLES - ALL FRAGMENTS
				var pth clip.Path; pth.Begin(gtx.Ops); sx, sy := pt.pos.X, pt.pos.Y
				sp := s.Particles[i].Spin * float32(s.FrameCount)
				cs, sn := float32(math.Cos(float64(sp))), float32(math.Sin(float64(sp)))
				
				if i%3 == 0 { // Triangle Fragment
					pth.MoveTo(f32.Pt(sx+sz*cs, sy+sz*sn))
					pth.LineTo(f32.Pt(sx+sz*sn, sy-sz*cs))
					pth.LineTo(f32.Pt(sx-sz*cs, sy-sz*sn))
					pth.Close()
				} else { // Diamond Fragment
					pth.MoveTo(f32.Pt(sx+sz*cs, sy))
					pth.LineTo(f32.Pt(sx, sy+sz*sn))
					pth.LineTo(f32.Pt(sx-sz*cs, sy))
					pth.LineTo(f32.Pt(sx, sy-sz*sn))
					pth.Close()
				}
				paint.FillShape(gtx.Ops, pCl, clip.Outline{Path: pth.End()}.Op())
			}
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
	)
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
