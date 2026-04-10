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
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, ColorBackground, clip.Rect{Max: gtx.Constraints.Max}.Op())
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
			focalLength := float32(1000)

			vX := s.MousePos.X - s.PrevMousePos.X
			vY := s.MousePos.Y - s.PrevMousePos.Y
			s.MouseVelocity.X = s.MouseVelocity.X*0.7 + vX*0.3
			s.MouseVelocity.Y = s.MouseVelocity.Y*0.7 + vY*0.3
			s.PrevMousePos = s.MousePos

			targetPos, targetVel := s.MousePos, s.MouseVelocity
			if s.GazeActive {
				mSpeedSq := s.MouseVelocity.X*s.MouseVelocity.X + s.MouseVelocity.Y*s.MouseVelocity.Y
				if mSpeedSq < 1.0 { targetPos, targetVel = s.GazePos, s.GazeVelocity }
			}
			
			velSq := targetVel.X*targetVel.X + targetVel.Y*targetVel.Y
			speed := float32(math.Sqrt(float64(velSq)))
			uX, uY := float32(0.0), float32(0.0)
			if speed > 0.1 { uX, uY = targetVel.X/speed, targetVel.Y/speed }
			nX, nY := -uY, uX
			a2, b2 := float32(6400.0), float32(6400.0)+speed*speed*50.0

			type screenPoint struct {
				pos f32.Point; color color.NRGBA; scale float32; force float32; distSq float32
			}
			points := make([]screenPoint, len(s.Particles))

			for i := range s.Particles {
				p := &s.Particles[i]
				scale := focalLength / (focalLength + p.Z)
				baseSx, baseSy := center.X + p.X*scale, center.Y + p.Y*scale
				dx, dy := targetPos.X - baseSx, targetPos.Y - baseSy
				euD2 := dx*dx + dy*dy
				
				if speed > 0.1 {
					du, dn := dx*uX+dy*uY, dx*nX+dy*nY
					mahaD2 := (du*du)/b2 + (dn*dn)/a2
					if mahaD2 < 1.0 {
						f := (1.0 - mahaD2) * speed * 2.0
						px, py := nX, nY
						if dn < 0 { px, py = -nX, -nY }
						p.VX, p.VY = p.VX + px*f*0.5, p.VY + py*f*0.5
					}
				}

				p.VX, p.VY = p.VX*0.94, p.VY*0.94
				p.X, p.Y = p.X+p.VX, p.Y+p.VY
				dScale := (1.0 - p.Z/1000.0)
				points[i] = screenPoint{pos: f32.Pt(baseSx+p.X, baseSy+p.Y), color: p.Color, scale: scale * dScale, distSq: euD2}
				
				avatarF := float32(0)
				if s.GazeActive {
					s.FaceMu.Lock()
					fPts, fHis, fScl := s.FacePoints, s.FaceHistory, s.FaceScale
					s.FaceMu.Unlock()
					bRad := float32(60.0) * fScl * (1.0 + s.PulseStrength*0.3)
					apply := func(pts []f32.Point, w float32) {
						for _, fp := range pts {
							if fp.X == 0 { continue }
							fdx, fdy := points[i].pos.X-fp.X, points[i].pos.Y-fp.Y
							fD2 := fdx*fdx + fdy*fdy
							rad := bRad * w
							if fD2 < rad*rad {
								fdist := float32(math.Sqrt(float64(fD2)))
								localF := (1.0 - fdist/rad) * 3.5 * w
								p.VX += (fdx / (fdist+0.1)) * localF
								p.VY += (fdy / (fdist+0.1)) * localF
								if localF > avatarF { avatarF = localF }
							}
						}
					}
					apply(fPts, 1.0)
					for _, h := range fHis { apply(h, 0.4) }
				}
				points[i].force = avatarF
			}

			for i := 0; i < len(points); i += 4 {
				for j := i + 1; j < i+15 && j < len(points); j++ {
					dx, dy := points[i].pos.X-points[j].pos.X, points[i].pos.Y-points[j].pos.Y
					if dx*dx+dy*dy < 1200 {
						lineC := ColorPrimary; lineC.A = uint8(30 * points[i].scale)
						var pth clip.Path; pth.Begin(gtx.Ops); pth.MoveTo(points[i].pos); pth.LineTo(points[j].pos)
						paint.FillShape(gtx.Ops, lineC, clip.Stroke{Path: pth.End(), Width: 0.5}.Op())
					}
				}
			}

			for i, pt := range points {
				sz := 1.5 * pt.scale * (1.0 + s.PulseStrength*0.2)
				pCl := pt.color
				if pt.distSq < 2500 || pt.force > 0.5 {
					sz *= (3.5 + pt.force*2.0); pCl.A = 255
					if pt.distSq < 2500 { pCl = lerpColor(pCl, ColorQuaternary, 0.6) }
				} else {
					sh := uint8(math.Sin(float64(s.FrameCount)*0.1+float64(i)*0.01) * 30)
					pCl.A = uint8(math.Max(0, math.Min(255, float64(180*pt.scale)+float64(sh))))
				}
				var pth clip.Path; pth.Begin(gtx.Ops); sx, sy := pt.pos.X, pt.pos.Y
				if i%17 == 0 {
					pth.MoveTo(f32.Pt(sx, sy)); pth.LineTo(f32.Pt(sx+sz, sy+sz/2))
					pth.MoveTo(f32.Pt(sx+sz/2, sy)); pth.LineTo(f32.Pt(sx+sz/2, sy+sz))
				} else if i%13 == 0 {
					pth.MoveTo(f32.Pt(sx+sz/2, sy)); pth.LineTo(f32.Pt(sx+sz, sy+sz/2))
					pth.LineTo(f32.Pt(sx+sz/2, sy+sz)); pth.LineTo(f32.Pt(sx, sy+sz/2)); pth.Close()
				} else {
					pth.MoveTo(f32.Pt(sx, sy+sz)); pth.LineTo(f32.Pt(sx+sz/2, sy)); pth.LineTo(f32.Pt(sx+sz, sy+sz)); pth.Close()
				}
				paint.FillShape(gtx.Ops, pCl, clip.Outline{Path: pth.End()}.Op())
			}
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),
	)
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
