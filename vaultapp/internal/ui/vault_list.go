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

			// 1. Interaction State & Physics
			vX := s.MousePos.X - s.PrevMousePos.X
			vY := s.MousePos.Y - s.PrevMousePos.Y
			s.MouseVelocity.X = s.MouseVelocity.X*0.7 + vX*0.3
			s.MouseVelocity.Y = s.MouseVelocity.Y*0.7 + vY*0.3
			s.PrevMousePos = s.MousePos

			targetPos := s.MousePos
			targetVel := s.MouseVelocity
			if s.GazeActive {
				mouseActiveSpeedSq := s.MouseVelocity.X*s.MouseVelocity.X + s.MouseVelocity.Y*s.MouseVelocity.Y
				if mouseActiveSpeedSq < 1.0 { 
					targetPos = s.GazePos 
					targetVel = s.GazeVelocity
				}
			}
			
			velSq := targetVel.X*targetVel.X + targetVel.Y*targetVel.Y
			speed := float32(math.Sqrt(float64(velSq)))
			uX, uY := float32(0), float32(0)
			if speed > 0.1 {
				uX = targetVel.X / speed
				uY = targetVel.Y / speed
			}
			nX, nY := -uY, uX
			a2, b2 := float32(6400.0), float32(6400.0)+speed*speed*50.0

			type screenPoint struct {
				pos    f32.Point
				color  color.NRGBA
				scale  float32
				force  float32
				distSq float32
			}
			points := make([]screenPoint, len(s.Particles))

			for i := range s.Particles {
				p := &s.Particles[i]
				scale := focalLength / (focalLength + p.Z)
				baseSx := center.X + p.X*scale
				baseSy := center.Y + p.Y*scale
				
				dx := targetPos.X - baseSx
				dy := targetPos.Y - baseSy
				euclidDistSq := dx*dx + dy*dy
				
				if speed > 0.1 {
					du, dn := dx*uX+dy*uY, dx*nX+dy*nY
					mahaD2 := (du*du)/b2 + (dn*dn)/a2
					if mahaD2 < 1.0 {
						force := (1.0 - mahaD2) * speed * 2.0
					p.VX -= (dx / dist) * force
					p.VY -= (dy / dist) * force
				}

				p.VX *= 0.94
				p.VY *= 0.94
				p.X += p.VX
				p.Y += p.VY

				points[i].pos = f32.Pt(baseSx+p.X, baseSy+p.Y)
				points[i].color = p.Color
				points[i].distSq = euclidDistSq
				points[i].scale = scale * depthScale
				
				// Avatar Interaction
				avatarForce := float32(0)
				if s.GazeActive {
					s.FaceMu.Lock()
					fPoints := s.FacePoints
					fHistory := s.FaceHistory
					fScale := s.FaceScale
					s.FaceMu.Unlock()

					baseRadius := float32(60.0) * fScale * (1.0 + s.PulseStrength*0.3)
					checkAvatar := func(pts []f32.Point, weight float32) {
						for idx, fp := range pts {
							if fp.X == 0 && fp.Y == 0 { continue }
							fdx := points[i].pos.X - fp.X
							fdy := points[i].pos.Y - fp.Y
							fDistSq := fdx*fdx + fdy*fdy
							radius := baseRadius * weight
							if fDistSq < radius*radius {
								fdist := float32(math.Sqrt(float64(fDistSq)))
								if fdist < 0.1 { fdist = 0.1 }
								localForce := (1.0 - fdist/radius) * 3.5 * weight
								p.VX += (fdx / fdist) * localForce
								p.VY += (fdy / fdist) * localForce
								if localForce > avatarForce { avatarForce = localForce }
							}
						}
					}
					checkAvatar(fPoints, 1.0)
					for _, hp := range fHistory { checkAvatar(hp, 0.4) }
				}
				points[i].force = avatarForce
			}

			// --- DRAWING: Constellations (The "Meaning" Web) ---
			// We connect nearby particles with thin gold lines to increase information density.
			for i := 0; i < len(points); i += 4 { // Sample to keep 60FPS
				for j := i + 1; j < i+20 && j < len(points); j++ {
					dx := points[i].pos.X - points[j].pos.X
					dy := points[i].pos.Y - points[j].pos.Y
					distsq := dx*dx + dy*dy
					if distsq < 1600 { // Connection threshold
						opacity := uint8(40 * (1.0 - float32(math.Sqrt(float64(distsq)))/40.0) * points[i].scale)
						lineColor := ColorPrimary
						lineColor.A = opacity
						
						var path clip.Path
						path.Begin(gtx.Ops)
						path.MoveTo(points[i].pos)
						path.LineTo(points[j].pos)
						paint.FillShape(gtx.Ops, lineColor, clip.Stroke{
							Path:  path.End(),
							Width: 0.5,
						}.Op())
					}
				}
			}

			// --- DRAWING: Fragments (The Solid Matter) ---
			for i, pt := range points {
				sx, sy := pt.pos.X, pt.pos.Y
				pSize := 1.5 * pt.scale * (1.0 + s.PulseStrength*0.2)
				pColor := pt.color
				
				if pt.distSq < 2500 || pt.force > 0.5 { 
					pSize *= (3.5 + pt.force*2.0)
					pColor.A = 255
					if pt.distSq < 2500 { pColor = lerpColor(pColor, ColorQuaternary, 0.6) }
				} else {
					shimmer := uint8(math.Sin(float64(s.FrameCount)*0.1 + float64(i)*0.01) * 30)
					pColor.A = uint8(math.Max(0, math.Min(255, float64(180*pt.scale)+float64(shimmer))))
				}

				var path clip.Path
				path.Begin(gtx.Ops)
				if i%13 == 0 { // Crystalline Diamond
					path.MoveTo(f32.Pt(sx+pSize/2, sy))
					path.LineTo(f32.Pt(sx+pSize, sy+pSize/2))
					path.LineTo(f32.Pt(sx+pSize/2, sy+pSize))
					path.LineTo(f32.Pt(sx, sy+pSize/2))
					path.Close()
				} else { // Hard Edge Fragment
					path.MoveTo(f32.Pt(sx, sy+pSize))
					path.LineTo(f32.Pt(sx+pSize/2, sy))
					path.LineTo(f32.Pt(sx+pSize, sy+pSize))
					path.Close()
				}
				paint.FillShape(gtx.Ops, pColor, path.End().Op())
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
	if t > 1 { t = 1 }
	if t < 0 { t = 0 }
	return color.NRGBA{
		R: uint8(float32(c1.R)*(1-t) + float32(c2.R)*t),
		G: uint8(float32(c1.G)*(1-t) + float32(c2.G)*t),
		B: uint8(float32(c1.B)*(1-t) + float32(c2.B)*t),
		A: uint8(float32(c1.A)*(1-t) + float32(c2.A)*t),
	}
}
