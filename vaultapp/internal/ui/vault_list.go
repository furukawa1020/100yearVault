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
						lbl := material.H3(s.Theme, "QUANTUM SYNCHRONICITY")
						lbl.Color = ColorPrimary
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						sub := material.H6(s.Theme, s.ConnectionStatus)
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
		// 琥珀の残り火（生命の輝き）表示
		
		// 背景ブロック
		dr := image.Rectangle{Max: gtx.Constraints.Max}
		paint.FillShape(gtx.Ops, ColorSurfaceHigh, clip.Rect(dr).Op())
		
		// 生命の鼓動
		accent := image.Rectangle{Max: image.Pt(gtx.Dp(8), gtx.Constraints.Max.Y)}
		paint.FillShape(gtx.Ops, ColorPrimary, clip.Rect(accent).Op())

		return layout.UniformInset(unit.Dp(32)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Caption(s.Theme, "琥珀の残り火 (LIVING EMBERS FROM THE PAST)")
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
	return layout.Inset{
		Bottom: unit.Dp(8), Left: unit.Dp(24), Right: unit.Dp(24),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 2126年標準: モノリス・ブロックデザイン
		thickness := gtx.Dp(unit.Dp(float32(v.LayerCount) * 2))
		if thickness > gtx.Dp(20) {
			thickness = gtx.Dp(20)
		}

		// 底面の影（厚み）
		shadowRect := image.Rectangle{
			Min: image.Pt(0, thickness),
			Max: gtx.Constraints.Max,
		}
		paint.FillShape(gtx.Ops, ColorSurfaceHigh, clip.Rect(shadowRect).Op())
		
		// メインの前面
		topRect := image.Rectangle{
			Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y-thickness),
		}
		paint.FillShape(gtx.Ops, ColorSurface, clip.Rect(topRect).Op())

		// 左端の「不変の境界」
		borderColor := ColorLocked
		if v.State == vault.StateUnlockable {
			borderColor = ColorPrimary
		} else if v.State == vault.StateOpened {
			borderColor = ColorPrimaryDim
		}
		
		borderWidth := gtx.Dp(4)
		borderRect := image.Rectangle{Max: image.Pt(borderWidth, gtx.Constraints.Max.Y-thickness)}
		paint.FillShape(gtx.Ops, borderColor, clip.Rect(borderRect).Op())

		// 世紀の刻印 (Century Pulse): 1年経過ごとに1本のノッチ
		yearsPassed := int(time.Since(v.CreatedAt).Hours() / (24 * 365))
		// デモ用に、1分=1年とみなしてノッチを表示する（実運用では1年に変更）
		// yearsPassed = int(time.Since(v.CreatedAt).Minutes()) 
		
		for n := 0; n < yearsPassed && n < 100; n++ {
			yOff := gtx.Dp(unit.Dp(float32(n)*3 + 2))
			if yOff > gtx.Constraints.Max.Y-2 {
				break
			}
			notchRect := image.Rectangle{
				Min: image.Pt(0, yOff),
				Max: image.Pt(borderWidth+gtx.Dp(4), yOff+gtx.Dp(1)),
			}
			paint.FillShape(gtx.Ops, ColorText, clip.Rect(notchRect).Op())
		}

		return material.Clickable(gtx, &s.SelectBtns[i], func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(16), Bottom: unit.Dp(16),
				Left: unit.Dp(24), Right: unit.Dp(18),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// タイトル行
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								title := material.H6(s.Theme, v.Title)
								if v.State == vault.StateSealed {
									title.Color = ColorTextDim
								} else if v.State == vault.StateOpened {
									title.Color = ColorPrimary
								}
								return title.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										if v.LayerCount > 0 {
											return s.layoutLayerBadge(gtx, v.LayerCount)
										}
										return layout.Dimensions{}
									}),
									layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return s.layoutStateBadge(gtx, v.State)
									}),
								)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					// サブ情報
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutVaultItemInfo(gtx, v)
					}),
				)
			})
		})
	})
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
