package ui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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

	// 2126年標準: CPP 用拡張画面
	DailyFragment    string
	ConnectionStatus string
}

// ───────────────────────────────────────────────
// 背景を塗る共通ヘルパー
// ───────────────────────────────────────────────
func fillBackground(gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(gtx.Ops, c, clip.Rect(dr).Op())
}

// ───────────────────────────────────────────────
// 保管庫一覧画面
// ───────────────────────────────────────────────
func (s *AppState) LayoutList(gtx layout.Context) layout.Dimensions {
	fillBackground(gtx, ColorBackground)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// 2126年標準: ダッシュボード・ヘッダー
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutDashboardHeader(gtx)
		}),
		// 時空の漂着物 (Driftwood from the Time-Sea)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if s.DailyFragment != "" {
				return s.layoutDriftwood(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
		// 保管庫本体（地層の重なり）
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(s.Vaults) == 0 {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(s.Theme, "記憶の種はまだ蒔かれていません。")
					lbl.Color = ColorTextDim
					return lbl.Layout(gtx)
				})
			}
			return s.VaultList.Layout(gtx, len(s.Vaults), func(gtx layout.Context, i int) layout.Dimensions {
				return s.layoutVaultItem(gtx, i, s.Vaults[i])
			})
		}),
	)
}

func (s *AppState) layoutDashboardHeader(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top: unit.Dp(32), Bottom: unit.Dp(24),
		Left: unit.Dp(40), Right: unit.Dp(40),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.H3(s.Theme, "ETERNAL ECHO")
						lbl.Color = ColorPrimary
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						sub := material.H6(s.Theme, "PERSONAL UNIVERSAL CONTINUUM")
						sub.Color = ColorTextDim
						return sub.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(s.Theme, &s.NewVaultBtn, "＋ 記憶を封じる")
				btn.Background = ColorPrimary
				btn.Color = ColorBackground
				btn.TextSize = unit.Sp(20)
				return btn.Layout(gtx)
			}),
		)
	})
}

func (s *AppState) layoutDriftwood(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Left: unit.Dp(40), Right: unit.Dp(40),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 百年の残響: 思考の星図背景
		dr := image.Rectangle{Max: gtx.Constraints.Max}
		paint.FillShape(gtx.Ops, ColorSurface, clip.UniformRRect(dr, 12).Op(gtx.Ops))
		
		return layout.UniformInset(unit.Dp(32)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Caption(s.Theme, "百年の残響 (THE ETERNAL ECHO)")
					lbl.Color = ColorPrimary
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H4(s.Theme, s.DailyFragment)
					lbl.Color = ColorText
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (s *AppState) layoutVaultItem(gtx layout.Context, i int, v *vault.Vault) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(32), Left: unit.Dp(40), Right: unit.Dp(40)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 思考の星図における一つの残響 (Echo as a Star)
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// 記憶の星（粒子）
				size := gtx.Dp(20)
				dr := image.Rectangle{Max: image.Pt(size, size)}
				
				color := ColorPrimaryDim
				if v.State == vault.StateOpened || v.State == vault.StateUnlockable {
					color = ColorPrimary
				}
				
				paint.FillShape(gtx.Ops, color, clip.Ellipse(dr).Op(gtx.Ops))
				return layout.Dimensions{Size: image.Pt(size, size)}
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(24)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.H5(s.Theme, v.Title)
						lbl.Color = ColorText
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutVaultItemInfo(gtx, v)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutInteractBtn(gtx, i, v)
			}),
		)
	})
}

func (s *AppState) layoutInteractBtn(gtx layout.Context, i int, v *vault.Vault) layout.Dimensions {
	label := "同調 (RESONATE)"
	if v.State == vault.StateOpened {
		label = "再臨 (REVISIT)"
	}
	btn := material.Button(s.Theme, &s.SelectBtns[i], label)
	btn.Background = ColorSurfaceHigh
	btn.Color = ColorPrimary
	btn.Inset = layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8), Left: unit.Dp(16), Right: unit.Dp(16)}
	btn.TextSize = unit.Sp(16)
	return btn.Layout(gtx)
}

func (s *AppState) layoutStateBadge(gtx layout.Context, state vault.State) layout.Dimensions {
	var label string
	var bg color.NRGBA
	switch state {
	case vault.StateSealed:
		label = "封印中"
		bg = ColorLocked
	case vault.StateUnlockable:
		label = "解凍可能"
		bg = ColorPrimary
	case vault.StateOpened:
		label = "開封済"
		bg = ColorPrimaryDim
	case vault.StateDestroyed:
		label = "破棄"
		bg = ColorDanger
	default:
		label = string(state)
		bg = ColorTextDim
	}

	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			dr := image.Rectangle{Max: gtx.Constraints.Min}
			paint.FillShape(gtx.Ops, bg, clip.UniformRRect(dr, 4).Op(gtx.Ops))
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(3), Bottom: unit.Dp(3),
				Left: unit.Dp(8), Right: unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(s.Theme, label)
				lbl.Color = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
				return lbl.Layout(gtx)
			})
		}),
	)
}

func (s *AppState) layoutVaultItemInfo(gtx layout.Context, v *vault.Vault) layout.Dimensions {
	var info string
	var infoColor color.NRGBA

	switch v.State {
	case vault.StateSealed:
		remaining := time.Until(v.UnlockAt)
		if remaining > 0 {
			days := int(remaining.Hours() / 24)
			hours := int(remaining.Hours()) % 24
			mins := int(remaining.Minutes()) % 60
			secs := int(remaining.Seconds()) % 60
			if days > 0 {
				info = fmt.Sprintf("封印中 ── あと %d日 %d時間で開封可能", days, hours)
			} else if hours > 0 {
				info = fmt.Sprintf("封印中 ── あと %d時間 %d分で開封可能", hours, mins)
			} else {
				info = fmt.Sprintf("封印中 ── あと %d分 %d秒で開封可能", mins, secs)
			}
			infoColor = ColorTextDim
		} else {
			info = "琥珀が十分な熱を持っています。解凍儀式を開始できます。"
			infoColor = ColorPrimary
		}
	case vault.StateOpened:
		info = fmt.Sprintf("開封日時: %s", v.OpenedAt.Format("2006年01月02日 15:04"))
		if v.PreviewHint != "" {
			preview := v.PreviewHint
			if len([]rune(preview)) > 60 {
				runes := []rune(preview)
				preview = string(runes[:60]) + "…"
			}
			info += "\n" + preview
		}
		infoColor = ColorTextDim
	default:
		info = fmt.Sprintf("作成: %s", v.CreatedAt.Format("2006年01月02日"))
		infoColor = ColorTextDim
	}

	lbl := material.Caption(s.Theme, info)
	lbl.Color = infoColor
	return lbl.Layout(gtx)
}

func (s *AppState) layoutLayerBadge(gtx layout.Context, count int) layout.Dimensions {
	label := fmt.Sprintf("%d LAYERS", count+1) // 最初の層 + 追記分
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			dr := image.Rectangle{Max: gtx.Constraints.Min}
			paint.FillShape(gtx.Ops, ColorSurfaceHigh, clip.Rect(dr).Op())
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(2), Bottom: unit.Dp(2),
				Left: unit.Dp(6), Right: unit.Dp(6),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(s.Theme, label)
				lbl.Color = ColorPrimary
				return lbl.Layout(gtx)
			})
		}),
	)
}

// ───────────────────────────────────────────────
// 描画ヘルパー: ops に保存して背景を全塗り
// ───────────────────────────────────────────────
func drawBackground(ops *op.Ops, gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(ops, c, clip.Rect(dr).Op())
}
