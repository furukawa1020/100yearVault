package ui

import (
	"image"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"vaultapp/internal/vault"
)

type RitualState struct {
	ActiveVault     *vault.Vault
	Password        widget.Editor
	UnlockBtn       widget.Clickable
	CancelBtn       widget.Clickable
	IsProcessing    bool
	ProcessingSince time.Time
	RevealedText    string
	IsRevealed      bool
	ErrorMessage    string
	AddLayerBtn     widget.Clickable
}

func (s *AppState) LayoutRitual(gtx layout.Context, r *RitualState) layout.Dimensions {
	fillBackground(gtx, ColorBackground)

	if r.ActiveVault == nil {
		return layout.Dimensions{}
	}

	if r.IsRevealed {
		return s.layoutRevealed(gtx, r)
	}
	if r.IsProcessing {
		return s.layoutProcessing(gtx, r)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// 巨大な戻るエリア
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(32)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(s.Theme, &r.CancelBtn, "← やめる (CANCEL)")
				btn.Background = ColorSurfaceHigh
				btn.Color = ColorTextDim
				btn.TextSize = unit.Sp(32)
				btn.Inset = layout.Inset{Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(40), Right: unit.Dp(40)}
				return btn.Layout(gtx)
			})
		}),

		// メイン：記憶との対面
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H4(s.Theme, r.ActiveVault.Title)
							lbl.Color = ColorPrimary
							lbl.TextSize = unit.Sp(64)
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(80)}.Layout),
						
						// パスフレーズ入力（巨大化）
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.labeledField(gtx, "合言葉を入力してください (KEY)", func(gtx layout.Context) layout.Dimensions {
								ed := material.Editor(s.Theme, &r.Password, "...")
								ed.TextSize = unit.Sp(64)
								ed.Color = ColorText
								return ed.Layout(gtx)
							})
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(80)}.Layout),

						// 巨大な解凍ボタン
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(s.Theme, &r.UnlockBtn, "記憶を呼び覚ます (OPEN)")
							btn.Background = ColorPrimary
							btn.Color = ColorBackground
							btn.TextSize = unit.Sp(56)
							btn.Inset = layout.Inset{Top: unit.Dp(50), Bottom: unit.Dp(50)}
							dim := btn.Layout(gtx)
							
							if r.ErrorMessage != "" {
								layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									lbl := material.H6(s.Theme, r.ErrorMessage)
									lbl.Color = ColorDanger
									return layout.Inset{Top: unit.Dp(60)}.Layout(gtx, lbl.Layout)
								})
							}
							return dim
						}),
					)
				})
			})
		}),
	)
}

// 処理中（自動的に進行）
func (s *AppState) layoutProcessing(gtx layout.Context, r *RitualState) layout.Dimensions {
	elapsed := time.Since(r.ProcessingSince).Seconds()
	progress := math.Min(elapsed/2.0, 1.0)

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H2(s.Theme, "想い出を繋いでいます...")
				lbl.Color = ColorPrimary
				lbl.TextSize = unit.Sp(60)
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(100)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				width := gtx.Dp(600)
				height := gtx.Dp(20)
				barRect := image.Rectangle{Max: image.Pt(width, height)}
				paint.FillShape(gtx.Ops, ColorSurfaceHigh, clip.Rect(barRect).Op())
				
				progressWidth := int(float64(width) * progress)
				progressRect := image.Rectangle{Max: image.Pt(progressWidth, height)}
				paint.FillShape(gtx.Ops, ColorPrimary, clip.Rect(progressRect).Op())
				
				return layout.Dimensions{Size: image.Pt(width, height)}
			}),
		)
	})
}

// 開封後：巨大表示
func (s *AppState) layoutRevealed(gtx layout.Context, r *RitualState) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H4(s.Theme, r.ActiveVault.Title)
					lbl.Color = ColorPrimaryDim
					lbl.TextSize = unit.Sp(48)
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(60)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H2(s.Theme, r.RevealedText)
					lbl.Color = ColorText
					lbl.TextSize = unit.Sp(72)
					lbl.Alignment = text.Middle
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(100)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(s.Theme, &r.CancelBtn, "もどる (BACK)")
					btn.Background = ColorSurfaceHigh
					btn.Color = ColorPrimary
					btn.TextSize = unit.Sp(40)
					btn.Inset = layout.Inset{Top: unit.Dp(40), Bottom: unit.Dp(40), Left: unit.Dp(60), Right: unit.Dp(60)}
					return btn.Layout(gtx)
				}),
			)
		})
	})
}
