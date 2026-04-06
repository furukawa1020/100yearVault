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

	// STP v2126 用拡張修飾
	AddLayerMode bool
	TargetVault  *vault.Vault
}

func (s *AppState) LayoutCompose(gtx layout.Context, c *ComposeState) layout.Dimensions {
	fillBackground(gtx, ColorBackground)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// ヘッダー（戻るボタンを巨大化）
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(32)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(s.Theme, &c.BackBtn, "← やめる (CANCEL)")
				btn.Background = ColorSurfaceHigh
				btn.Color = ColorTextDim
				btn.TextSize = unit.Sp(32)
				btn.Inset = layout.Inset{Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(40), Right: unit.Dp(40)}
				return btn.Layout(gtx)
			})
		}),
		
		// 入力エリア（スクロール可能な巨大フィールド）
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(40)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// タイトル（宛先）
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.labeledField(gtx, "どなたへの想いですか？ (TO)", func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(s.Theme, &c.Title, "未来の自分へ...")
							ed.TextSize = unit.Sp(48)
							ed.Color = ColorPrimary
							return ed.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(60)}.Layout),
					
					// 本文
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.labeledField(gtx, "今、伝えたいこと (MESSAGE) ※端末の音声入力も使えます", func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(s.Theme, &c.Body, "ここに指を置いて、話すようにお書きください...")
							ed.TextSize = unit.Sp(56)
							ed.Color = ColorText
							return ed.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(80)}.Layout),

					// 封印ボタン（画面最下部に巨大配置）
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := "想いを灯す (ETCH)"
						if c.AddLayerMode {
							label = "想いを重ねる (SYNC)"
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
		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
		layout.Rigid(field),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(4))
			paint.FillShape(gtx.Ops, ColorPrimaryDim, clip.Rect(image.Rectangle{Max: size}).Op())
			return layout.Dimensions{Size: size}
		}),
	)
}
