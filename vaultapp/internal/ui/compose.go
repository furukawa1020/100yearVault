package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type ComposeState struct {
	Title      widget.Editor
	Body       widget.Editor
	UnlockDays widget.Editor
	Passphrase widget.Editor
	
	SealBtn widget.Clickable
	BackBtn widget.Clickable
}

func (s *AppState) LayoutCompose(gtx layout.Context, c *ComposeState) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(20), Right: unit.Dp(20)}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(s.Theme, &c.BackBtn, "< Back").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							h2 := material.H4(s.Theme, "New Seal")
							h2.Color = ColorPrimary
							return h2.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(100)}.Layout),
					)
				},
			)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(100), Right: unit.Dp(100)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(s.Theme, &c.Title, "Memory Title")
						ed.TextSize = unit.Sp(24)
						return ed.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(s.Theme, &c.Body, "What memory do you want to seal?")
						return ed.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return material.Label(s.Theme, unit.Sp(12), "Unlock in (days)").Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return material.Editor(s.Theme, &c.UnlockDays, "36500").Layout(gtx)
									}),
								)
							}),
							layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return material.Label(s.Theme, unit.Sp(12), "Passphrase").Layout(gtx)
									}),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										ed := material.Editor(s.Theme, &c.Passphrase, "Secret Phrases")
										// Password filtering would be here
										return ed.Layout(gtx)
									}),
								)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(s.Theme, &c.SealBtn, "SEAL INTO THE VAULT")
						btn.TextSize = unit.Sp(20)
						btn.Background = ColorPrimary
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
				)
			})
		}),
	)
}
