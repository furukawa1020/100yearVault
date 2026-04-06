package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

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
	X, Y, Z    float32
	VX, VY, VZ float32
	Life       float32
}

type AppState struct {
	Theme         *material.Theme
	Vaults        []*vault.Vault
	CurrentScreen Screen

	// Sub-states
	Compose ComposeState
	Ritual  RitualState

	// Widgets
	NewVaultBtn widget.Clickable
	VaultList   layout.List
	SelectBtns  []widget.Clickable

	// 電脳領域 (NeuralSingularity v4.0)
	NeuralVault   *vault.Vault
	NeuralText    string
	NeuralSurface widget.Clickable
	
	// アニメーション・物理演算用計測値
	Rotation     float32
	PulsePhase   float32
	Particles    []Particle
	InitOnce     bool
	FrameCount   int
	MousePos     f32.Point
}

func (s *AppState) initNeuralSpace() {
	if s.InitOnce {
		return
	}
	s.Particles = make([]Particle, 1024) // 1024 粒子
	for i := range s.Particles {
		// 3D マハラノビス空間：楕円体状に分布
		angle1 := rand.Float64() * 2 * math.Pi
		angle2 := rand.Float64() * math.Pi
		dist := 100 + rand.Float64()*400
		
		s.Particles[i] = Particle{
			X: float32(math.Sin(angle2)*math.Cos(angle1) * dist),
			Y: float32(math.Sin(angle2)*math.Sin(angle1) * dist),
			Z: float32(math.Cos(angle2) * dist),
			VX: float32(rand.NormFloat64() * 0.2),
			VY: float32(rand.NormFloat64() * 0.2),
			VZ: float32(rand.NormFloat64() * 0.2),
			Life: rand.Float32(),
		}
	}
	s.InitOnce = true
}

func (s *AppState) RotateNeural() {
	if len(s.Vaults) == 0 {
		s.NeuralVault = nil
		s.NeuralText = "CRITICAL: NO_DATA_STREAM"
		return
	}
	
	idx := rand.Intn(len(s.Vaults))
	v := s.Vaults[idx]
	s.NeuralVault = v
	
	if v.State == vault.StateOpened {
		s.NeuralText = "FRAGMENT_LOADED: [" + v.Title + "]"
	} else if time.Now().After(v.UnlockAt) {
		s.NeuralText = "SIGNAL_DETECTED: [" + v.Title + "]\nREADY_TO_INITIALIZE_LINK"
	} else {
		s.NeuralText = "STASIS_LOCKED: [" + v.Title + "]\nTIMELOCK_ACTIVE"
	}
}

// ───────────────────────────────────────────────
// 背景を塗る共通ヘルパー
// ───────────────────────────────────────────────
func fillBackground(gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(gtx.Ops, c, clip.Rect(dr).Op())
}

// ───────────────────────────────────────────────
// 電脳領域の特異点: Neural Singularity v4.0
// ───────────────────────────────────────────────
func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	s.FrameCount++
	
	// 漆黒の物理遮断
	paint.FillShape(gtx.Ops, ColorBackground, clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)).Op())

	return material.Clickable(gtx, &s.NeuralSurface, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			// 1. 三次元演算レイヤー
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
				focalLength := float32(800) // 焦点距離
				
				// A. 3D 粒子シミュレーション (Perspective Projection)
				for i := range s.Particles {
					p := &s.Particles[i]
					
					// 回転演算
					cosR, sinR := float32(math.Cos(0.01)), float32(math.Sin(0.01))
					nx := p.X*cosR - p.Z*sinR
					nz := p.X*sinR + p.Z*cosR
					p.X, p.Z = nx, nz

					// マウスインタラクション (斥力)
					dx := p.X - (s.MousePos.X - center.X)
					dy := p.Y - (s.MousePos.Y - center.Y)
					distSq := dx*dx + dy*dy
					if distSq < 10000 {
						force := (10000 - distSq) / 10000
						p.VX += dx * force * 0.1
						p.VY += dy * force * 0.1
					}

					p.X += p.VX
					p.Y += p.VY
					p.X *= 0.99 // 減衰
					p.Y *= 0.99
					
					// 3D -> 2D 投影
					zPos := p.Z + 1000 // 奥へオフセット
					if zPos <= 0 { continue }
					
					sx := center.X + (p.X * focalLength) / zPos
					sy := center.Y + (p.Y * focalLength) / zPos
					
					// 遠近による不透明度とサイズ
					pointSize := 1.0 + (500 / zPos)
					alphaVal := uint32(255 * (1000 / zPos))
					if alphaVal > 255 { alphaVal = 255 }
					alpha := uint8(alphaVal)
					
					c := ColorPrimary
					c.A = alpha
					
					rect := image.Rect(int(sx), int(sy), int(sx+pointSize), int(sy+pointSize))
					paint.FillShape(gtx.Ops, c, clip.Rect(rect).Op())
				}

				// B. 3D 回転ヘキサゴン・グリッド
				for i := 1; i <= 2; i++ {
					rot := s.Rotation * float32(i)
					size := float32(200 * i)
					zOffset := float32(500)
					
					var path clip.Path
					path.Begin(gtx.Ops)
					
					for a := 0.0; a < 2*math.Pi; a += math.Pi/3 {
						curA := a + float64(rot)
						// 3D空間の頂点
						vx := size * float32(math.Cos(curA))
						vy := size * float32(math.Sin(curA))
						vz := float32(math.Sin(float64(s.Rotation))) * 100
						
						// 投影
						tz := vz + zOffset + 500
						tsx := center.X + (vx * focalLength) / tz
						tsy := center.Y + (vy * focalLength) / tz
						
						if a == 0 { path.MoveTo(f32.Pt(tsx, tsy)) } else { path.LineTo(f32.Pt(tsx, tsy)) }
					}
					path.Close()
					paint.FillShape(gtx.Ops, ColorPrimaryDim, clip.Stroke{Path: path.End(), Width: 1}.Op())
				}

				return layout.Dimensions{Size: gtx.Constraints.Max}
			}),
			
			// 2. システム情報オーバーレイ
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							msg := fmt.Sprintf("SINGULARITY_SYNC: 0x%04X | LATENCY: 0.16ms | PARTICLES: 1024", s.FrameCount%0xFFFF)
							lbl := material.H6(s.Theme, msg)
							lbl.Color = ColorPrimaryDim
							lbl.TextSize = unit.Sp(14)
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H4(s.Theme, s.NeuralText)
							lbl.Color = ColorText
							lbl.TextSize = unit.Sp(42)
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
					)
				})
			}),

			// 3. インジェクション・ゲート
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(30)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(s.Theme, &s.NewVaultBtn, "[ INITIATE_NEURAL_UPLINK ]")
						btn.Background = color.NRGBA{A: 0}
						btn.Color = ColorSecondary
						btn.TextSize = unit.Sp(24)
						return btn.Layout(gtx)
					})
				})
			}),
		)
	})
}

// ───────────────────────────────────────────────
// ヘルパー（互換性のために残すが、基本は未使用）
// ───────────────────────────────────────────────
func drawBackground(ops *op.Ops, gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(ops, c, clip.Rect(dr).Op())
}
