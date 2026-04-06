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

	// 電脳領域 (NeuralMode v2.0)
	NeuralVault   *vault.Vault
	NeuralText    string
	NeuralSurface widget.Clickable
	
	// アニメーション用計測値
	Rotation     float32
	PulsePhase   float32
	Particles    []Particle
	InitOnce     bool
}

func (s *AppState) initNeuralSpace() {
	if s.InitOnce {
		return
	}
	s.Particles = make([]Particle, 64)
	for i := range s.Particles {
		// マハラノビス空間をイメージした楕円状の初速と配置
		angle := rand.Float64() * 2 * math.Pi
		dist := rand.Float64() * 200
		s.Particles[i] = Particle{
			X: float32(math.Cos(angle) * dist),
			Y: float32(math.Sin(angle) * dist),
			VX: float32(math.Cos(angle+math.Pi/2)) * 0.5,
			VY: float32(math.Sin(angle+math.Pi/2)) * 0.5,
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
// 基盤レイアウト: 百年の灯火
// ───────────────────────────────────────────────
// ───────────────────────────────────────────────
// 電脳領域の深淵: Neural Deep-Dive v2.0
// ───────────────────────────────────────────────
func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	fillBackground(gtx, ColorBackground)

	return material.Clickable(gtx, &s.NeuralSurface, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			// 1. 幾何学・粒子演算レイヤー
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
				
				// A. パーティクル (Mahalanobis Point Cloud)
				for i := range s.Particles {
					p := &s.Particles[i]
					p.X += p.VX
					p.Y += p.VY
					// 中心に引き寄せる力（統計的収束）
					p.VX -= p.X * 0.0001
					p.VY -= p.Y * 0.0001
					
					pos := f32.Pt(center.X + p.X, center.Y + p.Y)
					rect := image.Rect(int(pos.X), int(pos.Y), int(pos.X)+2, int(pos.Y)+2)
					paint.FillShape(gtx.Ops, ColorPrimary, clip.Rect(rect).Op())
				}

				// B. 回転ヘキサゴン・ゲート
				for i := 1; i <= 3; i++ {
					angle := s.Rotation * float32(i) * 0.5
					size := float32(150 * i)
					
					var path clip.Path
					path.Begin(gtx.Ops)
					for a := 0.0; a < 2*math.Pi; a += math.Pi / 3 {
						curA := a + float64(angle)
						x := center.X + size*float32(math.Cos(curA))
						y := center.Y + size*float32(math.Sin(curA))
						if a == 0 {
							path.MoveTo(f32.Pt(x, y))
						} else {
							path.LineTo(f32.Pt(x, y))
						}
					}
					path.Close()
					
					paint.FillShape(gtx.Ops, ColorPrimaryDim, clip.Stroke{
						Path:  path.End(),
						Width: 1,
					}.Op())
				}
				
				// C. スキャンライン (高輝度)
				scanY := int(float32(gtx.Constraints.Max.Y) * (float32(math.Sin(float64(s.Rotation*3))) + 1) / 2)
				paint.FillShape(gtx.Ops, ColorPrimary, clip.Rect(image.Rect(0, scanY, gtx.Constraints.Max.X, scanY+1)).Op())
				
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}),
			
			// 2. システム・コア・テキスト
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(s.Theme, ">>> NEURAL_LINK_ESTABLISHED")
							lbl.Color = ColorPrimary
							lbl.TextSize = unit.Sp(24)
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
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
