package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"vaultapp/internal/vault"
)

type RitualState struct {
	ActiveVault *vault.Vault
	Password    widget.Editor
	UnlockBtn   widget.Clickable
	CancelBtn   widget.Clickable
}

func (s *AppState) LayoutRitual(gtx layout.Context, r *RitualState) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: unit.Dp(100), Right: unit.Dp(100)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					h3 := material.H3(s.Theme, "UNLOCK RITUAL")
					h3.Color = ColorPrimary
					return h3.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(s.Theme, "Vault: "+r.ActiveVault.Title)
					return lbl.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(s.Theme, &r.Password, "Enter Secret Passphrase")
					ed.Color = ColorPrimary
					return ed.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(40)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(s.Theme, &r.CancelBtn, "ABANDON").Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(s.Theme, &r.UnlockBtn, "BREAK THE SEAL")
							btn.Background = ColorPrimary
							return btn.Layout(gtx)
						}),
					)
				}),
			)
		})
	})
}
