package ui

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"gioui.org/f32"
	"gioui.org/io/pointer"
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
	BaseX, BaseY, BaseZ float32 // クラスタ中心座標 (非破壊的)
	X, Y, Z             float32 // 投影用 temporary
	VX, VY, VZ          float32 // 速度
	Color               color.NRGBA
}

type AppState struct {
	Theme *material.Theme

	// Neural Mirror Core
	Memories     []*vault.MemoryFragment
	NeuralMemory *vault.MemoryFragment // 同調中の記憶
	SelectBtns   []widget.Clickable
	NewVaultBtn  widget.Clickable

	// Cosmic Galaxy (4096 記憶の瞬き)
	Particles []Particle
	InitOnce  bool
	Rotation  float32
	
	// Interaction
	MousePos      f32.Point
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
		// 初期配置：宇宙の塵
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
	if numMemories == 0 {
		return 
	}

	for i := range s.Particles {
		p := &s.Particles[i]
		m := s.Memories[i%numMemories]
		
		// 記憶ごとの固有ハッシュによる座標決定
		hash := float32(0)
		for _, c := range m.ID { hash += float32(c) }
		
		angle := float64(hash * 0.1)
		dist  := 200.0 + math.Mod(float64(hash), 300.0)
		
		targetX := float32(math.Cos(angle) * dist)
		targetY := float32(math.Sin(angle) * dist)
		targetZ := float32(math.Sin(angle*0.5) * 100.0)
		
		// 記憶の「星団」を形成
		spread := 40.0 + math.Mod(float64(hash), 60.0)
		p.BaseX = targetX + float32((rand.Float64()-0.5)*spread)
		p.BaseY = targetY + float32((rand.Float64()-0.5)*spread)
		p.BaseZ = targetZ + float32((rand.Float64()-0.5)*spread)
		
		// 記憶の状態（Aura）を色に反映
		if m.Aura == vault.StateRadiant {
			p.Color = color.NRGBA{R: 255, G: 255, B: 200, A: 255} 
		}
	}
}

func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	s.FrameCount++

	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		// Base Layer: Interactive Void (Manual Control)
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			// 【零の鏡】全画面をクリップ領域として定義
			rect := image.Rectangle{Max: gtx.Constraints.Max}
			defer clip.Rect(rect).Push(gtx.Ops).Pop()
			
			// ポインタ・フィルタを直接登録
			pointer.Filter{
				Target: &s.NeuralSurface,
				Kinds:  pointer.Move | pointer.Drag | pointer.Press,
			}.Add(gtx.Ops)
			
			paint.FillShape(gtx.Ops, ColorBackground, clip.Rect(rect).Op())
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		// Middle Layer: 3D Memory Galaxy
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
			focalLength := float32(1000) 

			s.NeuralMemory = nil 
			
			for i := range s.Particles {
				p := &s.Particles[i]
				
				// 3D 回転演算
				rot := float64(s.Rotation)
				sinR, cosR := float32(math.Sin(rot)), float32(math.Cos(rot))
				tx := p.BaseX*cosR - p.BaseZ*sinR
				tz := p.BaseX*sinR + p.BaseZ*cosR
				ty := p.BaseY
				
				// 投影
				scale := focalLength / (focalLength + tz)
				sx, sy := center.X+tx*scale, center.Y+ty*scale
				if tz < -focalLength+50 { continue }

				// 同調判定 (Gazing)
				dx, dy := sx-s.MousePos.X, sy-s.MousePos.Y
				distSq := dx*dx + dy*dy
				pSize := 1.5 * scale
				pColor := p.Color
				
				if distSq < 1600 { 
					pSize *= 3.5
					pColor.A = 255
					if len(s.Memories) > 0 {
						s.NeuralMemory = s.Memories[i%len(s.Memories)]
					}
				} else {
					pColor.A = uint8(180 * scale)
				}

				rect := image.Rect(int(sx), int(sy), int(sx+float32(pSize)), int(sy+float32(pSize)))
				paint.FillShape(gtx.Ops, pColor, clip.Rect(rect).Op())
			}
			return layout.Dimensions{Size: gtx.Constraints.Max}
		}),

		// Upper Layer: HUD (Eternal Presence)
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(40)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return drawRawLabel(gtx, s.Theme, "NEURAL_MIRROR_CONNECTION_ESTABLISHED", 24, ColorPrimary)
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
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								txt := "SCANNING_THE_MEMORY_GALAXY..."
								if s.NeuralMemory != nil {
									txt = "RESONANCE_DETECTED: " + s.NeuralMemory.Title
								}
								return drawRawLabel(gtx, s.Theme, txt, 32, ColorPrimary)
							}),
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
