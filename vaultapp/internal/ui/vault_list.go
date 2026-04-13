package ui

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"
	"fmt"

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
	MotionMu               sync.Mutex
	MotionGrid             [16][20]float32 // 16 rows, 20 columns
	MotionVelocity         [16][20]f32.Point
	GridActive             bool
	MemoryAnchors          []int        // Indices of particles that represent memories
	FocusIndex             int          // Current focused memory index (-1 if none)
	FocusStrength          float32      // How "awakened" the focus is (0.0 to 1.0)
	SingularityPos         f32.Point    // Position of the central singularity
	IsSingularityFocused   bool
	
	// Grabbing Mechanic
	IsGrabbing             bool
	GrabPos                f32.Point
	GrabVelocity           f32.Point
	GrabForce              float32

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
	
	// Assign Memory Anchors
	s.MemoryAnchors = nil
	if len(s.Memories) > 0 {
		for i := 0; i < len(s.Memories); i++ {
			// Spread them out by picking random indices, or specific patterns
			idx := (i * (TotalParticles / (len(s.Memories) + 1))) % TotalParticles
			s.MemoryAnchors = append(s.MemoryAnchors, idx)
		}
	}
	s.SingularityPos = f32.Pt(0, 0) // Center of galaxy local coords
	s.FocusIndex = -1
	
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

			s.FaceMu.Lock(); fP, fS := s.FacePoints, s.FaceScale; s.FaceMu.Unlock()
			bRad := float32(85.0) * fS * (1.0 + s.PulseStrength*0.3)
			angle := float32(math.Atan2(float64(s.EigenV[1]), float64(s.EigenV[0])))
			csA, snA := float32(math.Cos(float64(angle))), float32(math.Sin(float64(angle)))

			// --- Focus Logic (Gaze & Mouse) ---
			interactionCenters := []f32.Point{tPos}
			if s.GazeActive {
				interactionCenters = append(interactionCenters, s.GazePos)
			}

			newFocusIdx := -1
			minDist := float32(1000000.0)
			
			// Check Singularity (Center)
			sPos := center // Singularity is at the center in screen space for now
			for _, ic := range interactionCenters {
				dix, diy := ic.X-sPos.X, ic.Y-sPos.Y
				d2 := dix*dix + diy*diy
				if d2 < 40*40 { // 40px radius for singularity
					s.IsSingularityFocused = true
					break
				} else {
					s.IsSingularityFocused = false
				}
			}

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
						
						// --- Neural Hand Reflections (Grid Interaction) ---
						if s.GridActive {
							gridR := int(bSy / (float32(gtx.Constraints.Max.Y) / 16.0))
							gridC := int(bSx / (float32(gtx.Constraints.Max.X) / 20.0))
							
							if gridR >= 0 && gridR < 16 && gridC >= 0 && gridC < 20 {
								s.MotionMu.Lock()
								mIntensity := s.MotionGrid[gridR][gridC]
								s.MotionMu.Unlock()
								
								if mIntensity > 0.1 {
									// Tactile Adhesion: Particles "stick" to moving regions
									mVel := s.MotionVelocity[gridR][gridC]
									adhesion := mIntensity * 6.0 / p.Mass
									p.VX += mVel.X * adhesion
									p.VY += mVel.Y * adhesion
									
									// Jitter/Resonance for tactile feel
									p.VX += (rand.Float32() - 0.5) * 2.5 * mIntensity
									p.VY += (rand.Float32() - 0.5) * 2.5 * mIntensity
									
									if mIntensity > 0.5 {
										// Actual Grab feel: dampen velocity to "hold"
										p.VX *= 0.8
										p.VY *= 0.8
									}
									resF += mIntensity * 3.0
								}
							}
						}

						// --- Siphon Logic (Grabbing) ---
						if s.IsGrabbing {
							gdx, gdy := s.GrabPos.X-(bSx+p.X), s.GrabPos.Y-(bSy+p.Y)
							gDist2 := gdx*gdx + gdy*gdy
							if gDist2 < 300*300 {
								gDist := float32(math.Sqrt(float64(gDist2)))
								// Nonlinear attraction: peaks at radius 80, then plateaus
								attraction := (1.0 - gDist/300.0) * s.GrabForce / p.Mass
								if gDist > 0.1 {
									p.VX += (gdx / gDist) * attraction * 3.5
									p.VY += (gdy / gDist) * attraction * 3.5
								}
								// Capture zone: Inherit grab velocity
								if gDist < 60 {
									influence := (1.0 - gDist/60.0)
									p.VX = p.VX*(1-influence*0.1) + s.GrabVelocity.X*influence*0.2
									p.VY = p.VY*(1-influence*0.1) + s.GrabVelocity.Y*influence*0.2
									resF += influence * 2.0 // Visual glow
								}
							}
						}

						if s.GazeActive {
							for _, fp := range fP {
								if fp.X == 0 { continue }
								fdx, fdy := (bSx+p.X)-fp.X, (bSy+p.Y)-fp.Y
								fD2 := fdx*fdx + fdy*fdy
								if fD2 < bRad*bRad {
									fdst := float32(math.Sqrt(float64(fD2)))
									// Camera Resonance: Stronger attraction/repulsion based on face points
									lF := (1.0 - fdst/bRad) * 8.5 / p.Mass
									p.VX += (fdx / (fdst+0.1)) * lF; p.VY += (fdy / (fdst+0.1)) * lF
									if lF > resF { resF = lF }
								}
							}
						}
						
						pScl := float32(1.0)
						if resF > 0.5 { pScl = 1.0 + (resF-0.5)*0.5 }
						
						pts[i] = screenPt{pos: f32.Pt(bSx+p.X, bSy+p.Y), scale: scale * dScl * pScl, force: resF, mDist: mDist, colorIdx: p.ColorIdx}
					}
				}(g*batchSize, (g+1)*batchSize)
			}
			wg.Wait()
			
			// Post-Physics: Determine Focus
			for i, aIdx := range s.MemoryAnchors {
				apt := pts[aIdx]
				for _, ic := range interactionCenters {
					dx, dy := ic.X-apt.pos.X, ic.Y-apt.pos.Y
					d2 := dx*dx + dy*dy
					if d2 < 60*60 && d2 < minDist {
						minDist = d2
						newFocusIdx = i
					}
				}
			}
			if newFocusIdx != s.FocusIndex {
				s.FocusIndex = newFocusIdx
				if newFocusIdx != -1 {
					s.NeuralMemory = s.Memories[newFocusIdx]
				} else {
					s.NeuralMemory = nil
				}
			}
			if s.FocusIndex != -1 {
				s.FocusStrength = s.FocusStrength*0.8 + 0.2
			} else {
				s.FocusStrength *= 0.8
			}

			// --- 2. SECURE SEQUENTIAL BATCHING (Immortality Guard) ---
			for cIdx := 0; cIdx <= len(ColorDataFragments); cIdx++ {
				var pth clip.Path
				isBegun := false
				glowMode := cIdx == len(ColorDataFragments)
				
				for i, pt := range pts {
					// Memory Anchor Highlights
					isGlow := pt.mDist < 5.0 || pt.force > 0.4
					
					// Special glow for focused anchor
					isFocusedAnchor := false
					for _, aIdx := range s.MemoryAnchors {
						if aIdx == i && s.FocusIndex != -1 && s.MemoryAnchors[s.FocusIndex] == aIdx {
							isFocusedAnchor = true
							break
						}
					}
					if isFocusedAnchor { isGlow = true }

					if (glowMode && isGlow) || (!glowMode && !isGlow && pt.colorIdx == cIdx) {
						if !isBegun {
							pth.Begin(gtx.Ops)
							isBegun = true
						}
						
						sz := 1.8 * pt.scale 
						if isFocusedAnchor { sz *= 2.5 }
						if sz > 6.5 { sz = 6.5 }
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
					paint.FillShape(gtx.Ops, col, clip.Outline{Path: pth.End()}.Op())
				}
			}

			// --- 3. FLOATING TYPOGRAPHY (Zero-UI Awakening) ---
			if s.FocusIndex != -1 && s.FocusStrength > 0.1 {
				m := s.Memories[s.FocusIndex]
				apt := pts[s.MemoryAnchors[s.FocusIndex]]
				
				col := ColorPrimary
				col.A = uint8(255 * s.FocusStrength)
				
				// Title
				op.Offset(image.Pt(int(apt.pos.X+20), int(apt.pos.Y-10))).Add(gtx.Ops)
				titleLabel := material.H5(s.Theme, m.Title)
				titleLabel.Color = col
				titleLabel.Layout(gtx)
				
				// Hint/Aura
				op.Offset(image.Pt(0, 30)).Add(gtx.Ops)
				hintLabel := material.Body2(s.Theme, string(m.Aura) + " resonance detected...")
				hintLabel.Color = color.NRGBA{R: 200, G: 200, B: 200, A: col.A}
				hintLabel.Layout(gtx)
				
				op.Offset(image.Pt(0, -20)).Add(gtx.Ops) // Reset offset for next potential drawings
			}

			// --- 4. SINGULARITY (The Void of Creation) ---
			{
				sCol := ColorPrimary
				if s.IsSingularityFocused {
					sCol = color.NRGBA{255, 255, 255, 255}
					// Singularity Typography
					op.Offset(image.Pt(int(center.X - 60), int(center.Y + 60))).Add(gtx.Ops)
					composeLabel := material.Body1(s.Theme, "DESCEND INTO VOID")
					composeLabel.Color = color.NRGBA{255, 255, 255, 255}
					composeLabel.Layout(gtx)
				}
				
				// Draw Singularity (Rotating Cross)
				var sPth clip.Path
				sPth.Begin(gtx.Ops)
				rotS := s.Rotation * 2.0
				csS, snS := float32(math.Cos(float64(rotS))), float32(math.Sin(float64(rotS)))
				
				sz := float32(15.0)
				if s.IsSingularityFocused { sz = 25.0 }
				
				sPth.MoveTo(f32.Pt(center.X + sz*csS, center.Y + sz*snS))
				sPth.LineTo(f32.Pt(center.X - sz*snS, center.Y + sz*csS))
				sPth.LineTo(f32.Pt(center.X - sz*csS, center.Y - sz*snS))
				sPth.LineTo(f32.Pt(center.X + sz*snS, center.Y - sz*csS))
				sPth.Close()
				paint.FillShape(gtx.Ops, sCol, clip.Outline{Path: sPth.End()}.Op())
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
