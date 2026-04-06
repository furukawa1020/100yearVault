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

	// 電脳の深緑 (NeuralMode)
	NeuralVault   *vault.Vault
	NeuralText    string
	NeuralSurface widget.Clickable
	
	// アニメーション用計測値
	Rotation     float32
	PulsePhase   float32
}

func (s *AppState) RotateNeural() {
	if len(s.Vaults) == 0 {
		s.NeuralVault = nil
		s.NeuralText = "NO_DATA_IN_CURRENT_VOID"
		return
	}
	
	idx := rand.Intn(len(s.Vaults))
	v := s.Vaults[idx]
	s.NeuralVault = v
	
	if v.State == vault.StateOpened {
		s.NeuralText = "CORE_FRAGMENT: [ " + v.Title + " ]"
	} else if time.Now().After(v.UnlockAt) {
		s.NeuralText = "STATUS: READY_TO_LOAD\nTARGET: " + v.Title + "\n[ PULSE_TO_INITIALIZE ]"
	} else {
		s.NeuralText = "STATUS: ENCRYPTED_STASIS\nTARGET: " + v.Title + "\nDECRYPTION_IMPOSSIBLE"
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
// 電脳の深淵: Neural Void
// ───────────────────────────────────────────────
func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	fillBackground(gtx, ColorBackground)

	return material.Clickable(gtx, &s.NeuralSurface, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			// 1. 幾何学的な背景：回転するマハラノビス楕円体（簡易ワイヤーフレーム）
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
				
				// 3つの異なる軸で回転する楕円を描画
				for i := 0; i < 3; i++ {
					angle := s.Rotation + float32(i)*math.Pi/3
					size := float32(300 + i*50)
					
					// 楕円のパス
					var p clip.Path
					p.Begin(gtx.Ops)
					
					r1 := size * float32(math.Abs(math.Cos(float64(angle))))
					r2 := size * float32(math.Abs(math.Sin(float64(angle*0.5))))
					
					p.MoveTo(f32.Pt(center.X+r1, center.Y))
					for a := 0.0; a < 2*math.Pi; a += 0.1 {
						x := center.X + r1*float32(math.Cos(a))
						y := center.Y + r2*float32(math.Sin(a))
						p.LineTo(f32.Pt(x, y))
					}
					p.Close()
					
					paint.FillShape(gtx.Ops, ColorPrimaryDim, clip.Stroke{
						Path:  p.End(),
						Width: 2,
					}.Op())
				}
				
				// 走査線（スキャンライン）
				scanY := int(float32(gtx.Constraints.Max.Y) * (float32(math.Sin(float64(s.Rotation*2))) + 1) / 2)
				rect := image.Rect(0, scanY, gtx.Constraints.Max.X, scanY+2)
				paint.FillShape(gtx.Ops, ColorPrimary, clip.Rect(rect).Op())
				
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}),
			
			// 2. 中央のシステムメッセージ
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(80)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					txt := s.NeuralText
					lbl := material.H4(s.Theme, txt)
					lbl.Color = ColorPrimary
					lbl.TextSize = unit.Sp(48)
					lbl.Alignment = text.Middle
					// フォントを少し「機械的」に（ monospace っぽく見せる工夫）
					return lbl.Layout(gtx)
				})
			}),

			// 3. システム・インジェクション
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(60)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(s.Theme, &s.NewVaultBtn, "[ INJECT_NEW_FRAGMENT ]")
						btn.Background = color.NRGBA{A: 0}
						btn.Color = ColorPrimary
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
