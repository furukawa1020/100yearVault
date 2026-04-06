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

	// 2126年標準: 百年の灯火 (Zero-UI)
	LanternVault  *vault.Vault
	LanternText   string
	LanternBtn    widget.Clickable
}

func (s *AppState) RotateLantern() {
	if len(s.Vaults) == 0 {
		s.LanternVault = nil
		s.LanternText = "静かな朝です。\n今の想いを、未来に託しませんか。"
		return
	}
	
	idx := rand.Intn(len(s.Vaults))
	v := s.Vaults[idx]
	s.LanternVault = v
	
	if v.State == vault.StateOpened {
		s.LanternText = v.Title + "\n\n(開封された想い出)"
	} else if time.Now().After(v.UnlockAt) {
		s.LanternText = v.Title + "\n\n[タップして記憶を呼び覚ます]"
	} else {
		// まだ先の未来
		days := int(time.Until(v.UnlockAt).Hours() / 24)
		if days > 365 {
			years := days / 365
			s.LanternText = v.Title + "\n\n(約 " + strconv.Itoa(years) + " 年の熟成)"
		} else {
			s.LanternText = v.Title + "\n\n(約 " + strconv.Itoa(days) + " 日の熟成)"
		}
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
func (s *AppState) LayoutList(gtx layout.Context) layout.Dimensions {
	// 背景: 深い宇宙
	dr := image.Rectangle{Max: gtx.Constraints.Max}
	paint.FillShape(gtx.Ops, ColorBackground, clip.Rect(dr).Op())

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// 上部：過去の自分からの語りかけ（極大表示）
		layout.Flexed(0.7, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.Clickable(gtx, &s.LanternBtn, func(gtx layout.Context) layout.Dimensions {
						txt := s.LanternText
						if txt == "" {
							txt = "静かな朝です。\n今の想いを、未来に託しませんか。"
						}
						lbl := material.H4(s.Theme, txt)
						lbl.Color = ColorPrimary
						// 開封可能な場合は白く発光させる（コントラストを上げる）
						if s.LanternVault != nil && s.LanternVault.State == vault.StateSealed && time.Now().After(s.LanternVault.UnlockAt) {
							lbl.Color = ColorText
						}
						lbl.TextSize = unit.Sp(64)
						lbl.Alignment = text.Middle
						return lbl.Layout(gtx)
					})
				})
			})
		}),

		// 下部：巨大な「想いを残す」ボタン（どこでも触れる安心感）
		layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(40)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(s.Theme, &s.NewVaultBtn, "今の想いを灯す (RECORD)")
				btn.Background = ColorSurfaceHigh
				btn.Color = ColorPrimary
				btn.TextSize = unit.Sp(40)
				btn.Inset = layout.Inset{Top: unit.Dp(60), Bottom: unit.Dp(60)}
				return btn.Layout(gtx)
			})
		}),
	)
}

// ───────────────────────────────────────────────
// ヘルパー（互換性のために残すが、基本は未使用）
// ───────────────────────────────────────────────
func drawBackground(ops *op.Ops, gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(ops, c, clip.Rect(dr).Op())
}
