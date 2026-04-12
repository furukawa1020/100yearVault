package ui

import (
	"image/color"
	"math"
	"math/rand"
	"sync"
	"fmt"

	"gioui.org/f32"
	"gioui.org/layout"
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
	Mass, Drag          float32
	ColorIdx            int 
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

	// 5D Statistical Manifold [x, y, z, vx, vy]
	History5D [128][5]float32
	HistPtr   int
	InvS      [3][3]float32 
	EigenV    [3]float32 

	// Screen Navigation
	CurrentScreen Screen
	Compose       ComposeState
	Ritual        RitualState
}

const (
	TotalParticles = 6144 
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
		p.Mass = 0.5 + rand.Float32()*0.6
		p.Drag = 0.94 + rand.Float32()*0.02
		p.ColorIdx = rand.Intn(len(ColorDataFragments))
	}
	s.InitOnce = true
	fmt.Println("GALAXY HEARTBEAT: HYPER-FLUID POINT-BETA")
}

func (s *AppState) RotateNeural() {
	s.initNeuralSpace()
}

func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	s.FrameCount++

	// --- 1. PARALLEL PHYSICS (Stable Core) ---
	tPos := s.MousePos
	tZ := 600.0 * (1.0 - s.FaceScale)
	s.History5D[s.HistPtr] = [5]float32{tPos.X, tPos.Y, tZ, 0, 0}
	s.HistPtr = (s.HistPtr + 1) % 128

	var m3 [3]float32; for _, v := range s.History5D { m3[0],m3[1],m3[2] = m3[0]+v[0],m3[1]+v[1],m3[2]+v[2] }
	m3[0], m3[1], m3[2] = m3[0]/128, m3[1]/128, m3[2]/128
	var c3 [3][3]float32; for _, v := range s.History5D {
		d0, d1, d2 := v[0]-m3[0], v[1]-m3[1], v[2]-m3[2]
		c3[0][0]+=d0*d0; c3[1][1]+=d1*d1; c3[2][2]+=d2*d2; c3[0][1]+=d0*d1; c3[0][2]+=d0*d2; c3[1][2]+=d1*d2
	}
	for i:=0;i<3;i++ { for j:=0; j<3; j++ { c3[i][j] /= 128 } }
	c3[0][0]+=500; c3[1][1]+=500; c3[2][2]+=500; c3[1][0], c3[2][0], c3[2][1] = c3[0][1], c3[0][2], c3[1][2]

	det := c3[0][0]*(c3[1][1]*c3[2][2]-c3[1][2]*c3[2][1]) - c3[0][1]*(c3[1][0]*c3[2][2]-c3[1][2]*c3[2][0]) + c3[0][2]*(c3[1][0]*c3[2][1]-c3[1][1]*c3[2][0])
	if det < 0.1 { det = 0.1 }
	s.InvS[0][0], s.InvS[1][1], s.InvS[2][2] = (c3[1][1]*c3[2][2]-c3[1][2]*c3[2][1])/det, (c3[0][0]*c3[2][2]-c3[0][2]*c3[2][0])/det, (c3[0][0]*c3[1][1]-c3[0][1]*c3[1][0])/det
	s.InvS[0][1] = (c3[0][2]*c3[2][1]-c3[0][1]*c3[2][2])/det

	vE := [3]float32{1, 1, 1}
	for k:=0; k<3; k++ {
		v2 := [3]float32{c3[0][0]*vE[0]+c3[0][1]*vE[1]+c3[0][2]*vE[2], c3[1][0]*vE[0]+c3[1][1]*vE[1]+c3[1][2]*vE[2], c3[2][0]*vE[0]+c3[2][1]*vE[1]+c3[2][2]*vE[2]}
		mag := float32(math.Sqrt(float64(v2[0]*v2[0] + v2[1]*v2[1] + v2[2]*v2[2])))
		if mag > 0 { vE[0],vE[1],vE[2] = v2[0]/mag, v2[1]/mag, v2[2]/mag }
	}
	s.EigenV = vE

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, ColorBackground, clip.Rect{Max: gtx.Constraints.Max}.Op())
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
			focalLength := float32(1000)
			cosR, sinR := float32(math.Cos(float64(s.Rotation))), float32(math.Sin(float64(s.Rotation)))
			s.Rotation += 0.0025

			s.FaceMu.Lock(); fP, fS := s.FacePoints, s.FaceScale; s.FaceMu.Unlock()
			bRad := float32(85.0) * fS * (1.0 + s.PulseStrength*0.3)
			angle := float32(math.Atan2(float64(s.EigenV[1]), float64(s.EigenV[0])))
			csA, snA := float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))

			var wg sync.WaitGroup
			numG := 4; batchSize := TotalParticles / numG
			type screenPt struct { pos f32.Point; scale float32; force float32; mDist float32; colorIdx int }
			pts := make([]screenPt, TotalParticles)

			for g := 0; g < numG; g++ {
				wg.Add(1)
				go func(start, end int) {
					defer wg.Done()
					for i := start; i < end; i++ {
						p := &s.Particles[i]
						tx, ty, tz := p.BaseX*cosR-p.BaseZ*sinR, p.BaseY, p.BaseX*sinR+p.BaseZ*cosR
						scale := focalLength / (focalLength + tz)
						if scale > 1.3 { scale = 1.3 }; if scale < 0.2 { scale = 0.2 }
						bSx, bSy := center.X+tx*scale, center.Y+ty*scale
						dx, dy, dz := tPos.X-(bSx+p.X), tPos.Y-(bSy+p.Y), tZ-tz
						mDist := dx*dx*s.InvS[0][0] + dy*dy*s.InvS[1][1] + dz*dz*s.InvS[2][2] + 2*dx*dy*s.InvS[0][1]

						if mDist < 18.0 {
							f := (18.0 - mDist) * 1.55 / p.Mass 
							p.VX -= (dx / (float32(math.Sqrt(float64(dx*dx+dy*dy+1)))+0.1)) * f
							p.VY -= (dy / (float32(math.Sqrt(float64(dx*dx+dy*dy+1)))+0.1)) * f
						}
						p.VX, p.VY = p.VX*p.Drag + (0-p.X)*0.07, p.VY*p.Drag + (0-p.Y)*0.07
						p.X, p.Y = p.X+p.VX, p.Y+p.VY
						dScl := (1.0 - tz/1500.0); if dScl < 0.5 { dScl = 0.5 }
						
						resF := float32(0)
						if s.GazeActive {
							for _, fp := range fP {
								if fp.X == 0 { continue }
								fdx, fdy := (bSx+p.X)-fp.X, (bSy+p.Y)-fp.Y
								fD2 := fdx*fdx + fdy*fdy
								if fD2 < bRad*bRad {
									fdst := float32(math.Sqrt(float64(fD2)))
									lF := (1.0 - fdst/bRad) * 6.0 / p.Mass
									p.VX += (fdx / (fdst+0.1)) * lF; p.VY += (fdy / (fdst+0.1)) * lF
									if lF > resF { resF = lF }
								}
							}
						}
						pts[i] = screenPt{pos: f32.Pt(bSx+p.X, bSy+p.Y), scale: scale * dScl, force: resF, mDist: mDist, colorIdx: p.ColorIdx}
					}
				}(g*batchSize, (g+1)*batchSize)
			}
			wg.Wait()

			// --- 2. SECURE SEQUENTIAL BATCHING (Immortality Guard) ---
			for cIdx := 0; cIdx <= len(ColorDataFragments); cIdx++ {
				var pth clip.Path
				isBegun := false
				glowMode := cIdx == len(ColorDataFragments)
				
				for _, pt := range pts {
					isGlow := pt.mDist < 5.0 || pt.force > 0.4
					if (glowMode && isGlow) || (!glowMode && !isGlow && pt.colorIdx == cIdx) {
						if !isBegun {
							pth.Begin(gtx.Ops)
							isBegun = true
						}
						
						sz := 1.8 * pt.scale 
						if sz > 4.5 { sz = 4.5 }
						if glowMode { sz *= 1.8 }
						
						hSz := sz * 1.5
						sx, sy := pt.pos.X, pt.pos.Y
						pth.MoveTo(f32.Pt(sx + hSz*csA, sy + hSz*snA))
						pth.LineTo(f32.Pt(sx - sz*snA, sy + sz*csA))
						pth.LineTo(f32.Pt(sx - hSz*csA, sy - hSz*snA))
						pth.LineTo(f32.Pt(sx + sz*snA, sy - sz*csA))
						pth.Close()
					}
				}
				
				if isBegun {
					col := color.NRGBA{255, 255, 255, 255}
					if !glowMode {
						col = ColorDataFragments[cIdx]
						col.A = 160
					}
					// CRITICAL: End() is now guaranteed to be called only if Begin() was.
					paint.FillShape(gtx.Ops, col, clip.Outline{Path: pth.End()}.Op())
				}
			}

			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
	)
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
