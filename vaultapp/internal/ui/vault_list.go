package ui

import (
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
	X, Y float32
	VX, VY float32
	Life float32
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

	// 電脳領域 (NeuralDeep-Dive v2.0)
	NeuralVault   *vault.Vault
	NeuralText    string
	NeuralSurface widget.Clickable
	
	// アニメーション用計測値
	Rotation     float32
	PulsePhase   float32
	Particles    []Particle
	InitOnce     bool
	FrameCount   int
}

func (s *AppState) initNeuralSpace() {
	if s.InitOnce {
		return
	}
	s.Particles = make([]Particle, 128)
	for i := range s.Particles {
		// マハラノビス的な楕円軌道を持つ粒子群
		angle := rand.Float64() * 2 * math.Pi
		dist := 50 + rand.Float64() * 300
		s.Particles[i] = Particle{
			X: float32(math.Cos(angle) * dist),
			Y: float32(math.Sin(angle) * dist),
			VX: float32(math.Cos(angle+math.Pi/2)) * (0.5 + rand.Float32()*1.5),
			VY: float32(math.Sin(angle+math.Pi/2)) * (0.5 + rand.Float32()*1.5),
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
func fillBackground(gtx l// ───────────────────────────────────────────────
// 電脳領域の深淵: Neural Deep-Dive v2.0 (Extreme)
// ───────────────────────────────────────────────
func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	s.FrameCount++
	
	// 漆黒の強制
	paint.FillShape(gtx.Ops, ColorBackground, clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y)).Op())

	return material.Clickable(gtx, &s.NeuralSurface, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			// 1. 幾何学演算：マハラノビス点群 & 回転ヘキサゴン
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
				
				// A. 粒子演算 (Mahalanobis Cloud)
				for i := range s.Particles {
					p := &s.Particles[i]
					p.X += p.VX
					p.Y += p.VY
					// 中心へ引き寄せる力とノイズ
					p.VX -= p.X * 0.0002
					p.VY -= p.Y * 0.0002
					
					pos := f32.Pt(center.X+p.X, center.Y+p.Y)
					// グリッチ的な明滅
					alpha := uint8(100 + rand.Intn(155))
					c := ColorPrimary
					c.A = alpha
					
					rect := image.Rect(int(pos.X), int(pos.Y), int(pos.X)+2, int(pos.Y)+2)
					paint.FillShape(gtx.Ops, c, clip.Rect(rect).Op())
				}

				// B. 三重回転ヘキサゴン
				for i := 1; i <= 3; i++ {
					rot := s.Rotation * float32(i) * 0.4
					if i == 2 { rot = -rot } // 逆回転
					
					size := float32(100 * i)
					var p clip.Path
					p.Begin(gtx.Ops)
					
					for a := 0.0; a < 2*math.Pi; a += math.Pi/3 {
						curA := a + float64(rot)
						x := center.X + size*float32(math.Cos(curA))
						y := center.Y + size*float32(math.Sin(curA))
						if a == 0 { p.MoveTo(f32.Pt(x, y)) } else { p.LineTo(f32.Pt(x, y)) }
					}
					p.Close()
					
					paint.FillShape(gtx.Ops, ColorPrimaryDim, clip.Stroke{
						Path:  p.End(),
						Width: 2,
					}.Op())
				}

				// C. システム・走査線 (High Pulse)
				scanY := int(float32(gtx.Constraints.Max.Y) * (float32(math.Sin(float64(s.Rotation*5))) + 1) / 2)
				paint.FillShape(gtx.Ops, ColorSecondary, clip.Rect(image.Rect(0, scanY, gtx.Constraints.Max.X, scanY+2)).Op())

				return layout.Dimensions{Size: gtx.Constraints.Max}
			}),
			
			// 2. システム・マトリクス (座標データ)
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							msg := fmt.Sprintf("SYSTEM_TIME: 2126.04.06.%04X | NODE: %X", s.FrameCount, rand.Uint32())
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
							// 等幅フォントへの強制
							lbl.Font.Variant = "monospace"
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
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(s.Theme, &s.NewVaultBtn, "[ INJECT_NEURAL_STREAMS ]")
								btn.Background = color.NRGBA{A: 0}
								btn.Color = ColorSecondary
								btn.TextSize = unit.Sp(24)
								return btn.Layout(gtx)
							}),
						)
					})
				})
			}),
		)
	})
}pacer{Height: unit.Dp(20)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							txt := s.NeuralText
							lbl := material.H4(s.Theme, txt)
							lbl.Color = ColorText
							lbl.TextSize = unit.Sp(48)
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
					)
				})
			}),

			// 3. コマンド・オーバーレイ
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(40)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(s.Theme, &s.NewVaultBtn, "[ INITIATE_UPLINK ]")
						btn.Background = color.NRGBA{A: 0}
						btn.Color = ColorSecondary
						btn.TextSize = unit.Sp(32)
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
