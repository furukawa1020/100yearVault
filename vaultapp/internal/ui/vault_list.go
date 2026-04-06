package ui

import (
	"image"
	"image/color"
	"math/rand"
	"strconv"
	"time"

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

	// 木漏れ日の記憶 (EchoMode)
	EchoVault   *vault.Vault
	EchoText    string
	EchoSurface widget.Clickable
}

func (s *AppState) RotateEcho() {
	if len(s.Vaults) == 0 {
		s.EchoVault = nil
		s.EchoText = "静かな森の中です。\n今はただ、ここにいてください。"
		return
	}
	
	idx := rand.Intn(len(s.Vaults))
	v := s.Vaults[idx]
	s.EchoVault = v
	
	if v.State == vault.StateOpened {
		s.EchoText = "かつてのあなたの想い：\n\n「" + v.Title + "」"
	} else if time.Now().After(v.UnlockAt) {
		s.EchoText = v.Title + "\n\n(そっと触れて、記憶を呼び覚ます)"
	} else {
		// 未来への約束
		s.EchoText = v.Title + "\n\n(遠い未来への、大切な約束です)"
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
// 回想の風景: 木漏れ日の記憶
// ───────────────────────────────────────────────
func (s *AppState) LayoutEcho(gtx layout.Context) layout.Dimensions {
	fillBackground(gtx, ColorBackground)

	return material.Clickable(gtx, &s.EchoSurface, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			// 背景の演出：光の粒子（オーブ）を模した大きな円
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				// ここに将来的にパーティクルアニメーションを追加可能
				return layout.Dimensions{}
			}),
			
			// 中央のメッセージ（木漏れ日のように浮かび上がる）
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(80)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					txt := s.EchoText
					if txt == "" {
						txt = "穏やかな時間が流れています。"
					}
					lbl := material.H4(s.Theme, txt)
					lbl.Color = ColorPrimary
					// 開封可能な場合はより明るく
					if s.EchoVault != nil && s.EchoVault.State == vault.StateSealed && time.Now().After(s.EchoVault.UnlockAt) {
						lbl.Color = ColorText
					}
					lbl.TextSize = unit.Sp(64)
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				})
			}),

			// 下部の「想いを残す」誘導（ボタン感を出さず、静かな導線に）
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(60)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(s.Theme, &s.NewVaultBtn, "想いをそっと残す")
						btn.Background = color.NRGBA{A: 0} // 透明
						btn.Color = ColorPrimaryDim
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
