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
		// ヘッダー
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(20), Bottom: unit.Dp(16),
				Left: unit.Dp(32), Right: unit.Dp(32),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(s.Theme, &c.BackBtn, "← 戻る")
						btn.Background = ColorSurfaceHigh
						btn.Color = ColorTextDim
						return btn.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := "時空への放流"
						if c.AddLayerMode && c.TargetVault != nil {
							title = "地層の重合: " + c.TargetVault.Title
						}
						h2 := material.H4(s.Theme, title)
						h2.Color = ColorPrimary
						return h2.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(80)}.Layout),
				)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(2))
			paint.FillShape(gtx.Ops, ColorPrimary, clip.Rect(image.Rectangle{Max: size}).Op())
			return layout.Dimensions{Size: size}
		}),
		// 本文エリア
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left: unit.Dp(80), Right: unit.Dp(80),
				Top: unit.Dp(32), Bottom: unit.Dp(32),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// タイトル
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.labeledField(gtx, "タイトル", func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(s.Theme, &c.Title, "この記憶のタイトルを入力…")
							ed.TextSize = unit.Sp(22)
							ed.Color = ColorText
							return ed.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
					// 本文
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return s.labeledField(gtx, "本文（封印後は読めません）", func(gtx layout.Context) layout.Dimensions {
							ed := material.Editor(s.Theme, &c.Body, "未来の自分へ伝えたいことを書いてください…")
							ed.Color = ColorText
							return ed.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
					// 開封条件行
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx,
							// 開封日数
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return s.labeledField(gtx, "開封可能まで（日数）", func(gtx layout.Context) layout.Dimensions {
									ed := material.Editor(s.Theme, &c.UnlockDays, "36500")
									ed.Color = ColorPrimary
									return ed.Layout(gtx)
								})
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(24)}.Layout),
							// パスフレーズ
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return s.labeledField(gtx, "封印パスフレーズ", func(gtx layout.Context) layout.Dimensions {
									ed := material.Editor(s.Theme, &c.Passphrase, "秘密の合言葉…")
									ed.Color = ColorPrimary
									return ed.Layout(gtx)
								})
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
					// 封印ボタン
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := "時空の海へ放流する (RELEASE)"
						if c.AddLayerMode {
							label = "この地層を同期させる (SYNC)"
						}
						btn := material.Button(s.Theme, &c.SealBtn, label)
						btn.TextSize = unit.Sp(18)
						btn.Background = ColorPrimary
						btn.Color = ColorBackground
						dim := btn.Layout(gtx)
						
						if c.ErrorMessage != "" {
							layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Caption(s.Theme, c.ErrorMessage)
								lbl.Color = ColorDanger
								return layout.Inset{Top: unit.Dp(40)}.Layout(gtx, lbl.Layout)
							})
						}
						return dim
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						hint := material.Caption(s.Theme, "封印後は、条件を満たすまで中身を読むことはできません。")
						hint.Color = ColorTextDim
						return hint.Layout(gtx)
					}),
				)
			})
		}),
	)
}

// labeledField はラベル＋下線付きフィールドを描画するヘルパー
func (s *AppState) labeledField(gtx layout.Context, label string, field func(layout.Context) layout.Dimensions) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Caption(s.Theme, label)
			lbl.Color = ColorPrimaryDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(field),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		// 2126年標準: 琥珀の境界線
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(2))
			paint.FillShape(gtx.Ops, ColorPrimaryDim, clip.Rect(image.Rectangle{Max: size}).Op())
			return layout.Dimensions{Size: size}
		}),
	)
}
