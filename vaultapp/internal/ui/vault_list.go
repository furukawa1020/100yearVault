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
	InvS      [3][3]float32 // Inverse Spatial Cov (3x3)
	InvV      [2][2]float32 // Inverse Velocity Cov (2x2)
	EigenAxis [3]f32.Point  // Principal Spatial Components

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
		p.Mass = 0.4 + rand.Float32()*1.0
		p.Drag = 0.93 + rand.Float32()*0.03
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
	s.Rotation += 0.002

	// --- 1. 5D STATISTICAL CORE ---
	tPos := s.MousePos
	tVel := s.MouseVelocity
	if s.GazeActive && tVel.X*tVel.X+tVel.Y*tVel.Y < 1.5 {
		tPos, tVel = s.GazePos, s.GazeVelocity
	}
	tZ := 500.0 * (1.0 - s.FaceScale) // Synthetic Z from face scale
	s.History5D[s.HistPtr] = [5]float32{tPos.X, tPos.Y, tZ, tVel.X, tVel.Y}
	s.HistPtr = (s.HistPtr + 1) % 128

	// Mean & Covariance 3x3 (Spatial)
	var m3 [3]float32; for _, v := range s.History5D { m3[0],m3[1],m3[2] = m3[0]+v[0],m3[1]+v[1],m3[2]+v[2] }
	m3[0], m3[1], m3[2] = m3[0]/128, m3[1]/128, m3[2]/128
	var c3 [3][3]float32
	for _, v := range s.History5D {
		d0, d1, d2 := v[0]-m3[0], v[1]-m3[1], v[2]-m3[2]
		c3[0][0]+=d0*d0; c3[1][1]+=d1*d1; c3[2][2]+=d2*d2
		c3[0][1]+=d0*d1; c3[0][2]+=d0*d2; c3[1][2]+=d1*d2
	}
	for i := 0; i < 3; i++ { for j := 0; j < 3; j++ { c3[i][j] /= 128.0 } }
	c3[0][0]+=400; c3[1][1]+=400; c3[2][2]+=400; // Stabilizers
	c3[1][0], c3[2][0], c3[2][1] = c3[0][1], c3[0][2], c3[1][2]

	// 3x3 Matrix Inversion (Analytic)
	det := c3[0][0]*(c3[1][1]*c3[2][2]-c3[1][2]*c3[2][1]) - c3[0][1]*(c3[1][0]*c3[2][2]-c3[1][2]*c3[2][0]) + c3[0][2]*(c3[1][0]*c3[2][1]-c3[1][1]*c3[2][0])
	if det < 0.1 { det = 0.1 }
	s.InvS[0][0] = (c3[1][1]*c3[2][2]-c3[1][2]*c3[2][1])/det
	s.InvS[0][1] = (c3[0][2]*c3[2][1]-c3[0][1]*c3[2][2])/det
	s.InvS[0][2] = (c3[0][1]*c3[1][2]-c3[0][2]*c3[1][1])/det
	// (Symmetric result assumed for other indices for physics)
	s.InvS[1][1] = (c3[0][0]*c3[2][2]-c3[0][2]*c3[2][0])/det
	s.InvS[2][2] = (c3[0][0]*c3[1][1]-c3[0][1]*c3[1][0])/det

	// Mean & Covariance 2x2 (Velocity)
	var m2 [2]float32; for _, v := range s.History5D { m2[0],m2[1] = m2[0]+v[3],m2[1]+v[4] }
	m2[0], m2[1] = m2[0]/128, m2[1]/128
	var c2 [2][2]float32; for _, v := range s.History5D {
		d3, d4 := v[3]-m2[0], v[4]-m2[1]
		c2[0][0]+=d3*d3; c2[1][1]+=d4*d4; c2[0][1]+=d3*d4
	}
	for i:=0;i<2;i++ { for j:=0;j<2;j++ { c2[i][j]/=128 } }
	c2[0][0]+=10; c2[1][1]+=10; c2[1][0]=c2[0][1]
	det2 := c2[0][0]*c2[1][1]-c2[0][1]*c2[1][0]; if det2 < 0.1 { det2 = 0.1 }
	s.InvV[0][0], s.InvV[1][1], s.InvV[0][1] = c2[1][1]/det2, c2[0][0]/det2, -c2[0][1]/det2

	// Eigen-Axis Extraction (Simplified Power Iteration for 3D guide)
	ax := f32.Pt(1, 0) // Approximation focus
	s.EigenAxis[0] = f32.Pt(c3[0][0]/400, c3[0][1]/400)

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
				pos f32.Point; color color.NRGBA; scale float32; force float32; mDist float32; vDist float32
			}
			pts := make([]screenPt, TotalParticles)

			// --- 5D PARALLEL MANIFOLD ENGINE ---
			var wg sync.WaitGroup
			numG := 4; batchSize := TotalParticles / numG
			s.FaceMu.Lock(); fP, fH, fS := s.FacePoints, s.FaceHistory, s.FaceScale; s.FaceMu.Unlock()
			bRad := float32(80.0) * fS * (1.0 + s.PulseStrength*0.3)

			for g := 0; g < numG; g++ {
				wg.Add(1)
				go func(start, end int) {
					defer wg.Done()
					for i := start; i < end; i++ {
						p := &s.Particles[i]
						tx, ty, tz := p.BaseX*cosR-p.BaseZ*sinR, p.BaseY, p.BaseX*sinR+p.BaseZ*cosR
						scale := focalLength / (focalLength + tz)
						bSx, bSy := center.X+tx*scale, center.Y+ty*scale
						
						// --- Stereoscopic Mahalanobis Manifold ---
						dx, dy, dz := tPos.X-(bSx+p.X), tPos.Y-(bSy+p.Y), tZ-tz
						mDist := dx*dx*s.InvS[0][0] + dy*dy*s.InvS[1][1] + dz*dz*s.InvS[2][2] + 2*dx*dy*s.InvS[0][1]
						
						// --- Velocity Mahalanobis Manifold ---
						vDist := p.VX*p.VX*s.InvV[0][0] + p.VY*p.VY*s.InvV[1][1] + 2*p.VX*p.VY*s.InvV[0][1]

						if mDist < 16.0 {
							f := (16.0 - mDist) * 0.45 / p.Mass
							p.VX -= (dx / (float32(math.Sqrt(float64(dx*dx+dy*dy)))+0.1)) * f
							p.VY -= (dy / (float32(math.Sqrt(float64(dx*dx+dy*dy)))+0.1)) * f
							p.Spin += f * 0.05
						}
						
						// Fluid Restorative Force
						p.VX, p.VY = p.VX + (0 - p.X)*0.045, p.VY + (0 - p.Y)*0.045
						
						resF := float32(0)
						if s.GazeActive {
							for _, fp := range fP {
								if fp.X == 0 { continue }
								fdx, fdy := (bSx+p.X)-fp.X, (bSy+p.Y)-fp.Y
								fD2 := fdx*fdx + fdy*fdy; rad := bRad
								if fD2 < rad*rad {
									fdst := float32(math.Sqrt(float64(fD2)))
									lF := (1.0 - fdst/rad) * 4.8 / p.Mass
									p.VX += (fdx / (fdst+0.1)) * lF; p.VY += (fdy / (fdst+0.1)) * lF
									if lF > resF { resF = lF }
								}
							}
						}

						p.VX, p.VY = p.VX*p.Drag, p.VY*p.Drag
						p.X, p.Y = p.X+p.VX, p.Y+p.VY
						dScl := (1.0 - tz/1500.0); if dScl < 0.3 { dScl = 0.3 }
						
						pts[i] = screenPt{pos: f32.Pt(bSx+p.X, bSy+p.Y), color: p.Color, scale: scale * dScl, force: resF, mDist: mDist, vDist: vDist}
					}
				}(g*batchSize, (g+1)*batchSize)
			}
			wg.Wait()

			// 2. Eigen-Axis / Tensor Guide Layer
			if s.FrameCount % 2 == 0 {
				var pth clip.Path; pth.Begin(gtx.Ops)
				pth.MoveTo(tPos); pth.LineTo(f32.Pt(tPos.X+s.EigenAxis[0].X*100, tPos.Y+s.EigenAxis[0].Y*100))
				paint.FillShape(gtx.Ops, color.NRGBA{255, 255, 255, 30}, clip.Stroke{Path: pth.End(), Width: 0.5}.Op())
			}

			// 3. Chromatic Manifold Fragments
			for i, pt := range pts {
				sz := (2.0 + pt.vDist*2.0) * pt.scale * (1.0 + s.PulseStrength*0.2)
				pCl := pt.color
				
				// Doppler Shift Chromatic Resonance
				if pt.vDist > 1.0 { pCl = lerpColor(pCl, ColorSecondary, 0.4) }

				if pt.mDist < 4.0 || pt.force > 0.5 {
					sz *= (3.5 + pt.force*3.0); pCl.A = 255
					if pt.force > 0.5 { pCl = lerpColor(pCl, ColorPrimary, 0.6) }
				} else {
					pCl.A = uint8(160 * pt.scale)
				}
				
				if pt.mDist < 4.0 {
					var pth clip.Path; pth.Begin(gtx.Ops); sx, sy := pt.pos.X, pt.pos.Y
					// Spin-aware rendering
					sp := s.Particles[i].Spin * float32(s.FrameCount)
					cs, sn := float32(math.Cos(float64(sp))), float32(math.Sin(float64(sp)))
					pth.MoveTo(f32.Pt(sx+sz*cs, sy+sz*sn))
					pth.LineTo(f32.Pt(sx-sz*cs, sy-sz*sn))
					paint.FillShape(gtx.Ops, pCl, clip.Stroke{Path: pth.End(), Width: 1.5}.Op())
				} else {
					paint.FillShape(gtx.Ops, pCl, clip.Rect{Min: image.Pt(int(pt.pos.X), int(pt.pos.Y)), Max: image.Pt(int(pt.pos.X)+int(sz+1), int(pt.pos.Y)+int(sz+1))}.Op())
				}
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
