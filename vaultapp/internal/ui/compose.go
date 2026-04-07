package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"vaultapp/internal/vault"
)

type ComposeState struct {
	Title      widget.Editor
	Body       widget.Editor
	UnlockDays widget.Editor
	Passphrase widget.Editor

	SealBtn widget.Clickable
	BackBtn widget.Clickable
	ErrorMessage string

	// Neural Mirror Uplink 拡張
	AddLayerMode bool
	TargetMemory *vault.MemoryFragment
}

func (s *AppState) LayoutCompose(gtx layout.Context, c *ComposeState) layout.Dimensions {
	// 【零の鏡】各スクリーンレベルでも物理的漆黒クリアを再実行
	paint.FillShape(gtx.Ops, ColorBackground, clip.Rect{Max: gtx.Constraints.Max}.Op())

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// ヘッダー
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(32)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(s.Theme, &c.BackBtn, "← CLOSE_LETTER_AND_BACK")
				btn.Background = ColorSurfaceHigh
				btn.Color = ColorPrimary
				btn.TextSize = unit.Sp(32)
				btn.Inset = layout.Inset{Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(40), Right: unit.Dp(40)}
				return btn.Layout(gtx)
			})
		}),
		
		// 入力エリア
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(40)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// 宛先
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.labeledField(gtx, "DEAR_MY_FUTURE_SELF", func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(s.Theme, &c.Title, "未来の私へ...")
							ed.TextSize = unit.Sp(48)
							ed.Color = ColorPrimary
							return ed.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
					
					// 本文
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.labeledField(gtx, "WORDS_TO_BE_STARS", func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(s.Theme, &c.Body, "今の想いを綴ってください（星となって瞬き続けます）...")
							ed.TextSize = unit.Sp(42) // 読みやすさ重視
							ed.Color = ColorText
							return ed.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),

					// 合言葉
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.labeledField(gtx, "RESONANCE_KEY_OPTIONAL", func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(s.Theme, &c.Passphrase, "合言葉...")
							ed.TextSize = unit.Sp(32)
							ed.Color = ColorPrimaryDim
							return ed.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(60)}.Layout),

					// 放流ボタン
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := "RELEASE_MEMORY_TO_COSMOS"
						if c.AddLayerMode {
							label = "SYNCHRONIZE_NEW_REFLECTION"
						}
						btn := material.Button(s.Theme, &c.SealBtn, label)
						btn.Background = ColorPrimary
						btn.Color = ColorBackground
						btn.TextSize = unit.Sp(48)
						btn.Inset = layout.Inset{Top: unit.Dp(50), Bottom: unit.Dp(50)}
						dim := btn.Layout(gtx)
						
						if c.ErrorMessage != "" {
							layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.H6(s.Theme, c.ErrorMessage)
								lbl.Color = ColorDanger
								return layout.Inset{Top: unit.Dp(60)}.Layout(gtx, lbl.Layout)
							})
						}
						return dim
					}),
				)
			})
		}),
	)
}

// labeledField は 100 歳用：巨大ラベル＋下線
func (s *AppState) labeledField(gtx layout.Context, label string, field func(layout.Context) layout.Dimensions) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(s.Theme, label)
			lbl.Color = ColorPrimaryDim
			lbl.TextSize = unit.Sp(32)
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(field),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(2)) // 繊細な境界線
			paint.FillShape(gtx.Ops, ColorPrimaryDim, clip.Rect(image.Rectangle{Max: size}).Op())
			return layout.Dimensions{Size: size}
		}),
	)
}
