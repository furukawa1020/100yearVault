package ui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"vaultapp/internal/vault"
)

type AppState struct {
	Theme  *material.Theme
	Vaults []*vault.Vault
	
	// Widgets
	NewVaultBtn widget.Clickable
	VaultList   layout.List
}

func (s *AppState) LayoutList(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(20), Bottom: unit.Dp(20), Left: unit.Dp(20), Right: unit.Dp(20)}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Spacing: layout.SpaceBetween, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							h2 := material.H4(s.Theme, "Hundred-Year Vault")
							h2.Color = ColorPrimary
							return h2.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							btn := material.Button(s.Theme, &s.NewVaultBtn, "Seal New Memory")
							return btn.Layout(gtx)
						}),
					)
				},
			)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(s.Vaults) == 0 {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(s.Theme, "The vault is empty. No memories are sealed yet.")
					lbl.Color = ColorLocked
					return lbl.Layout(gtx)
				})
			}
			return s.VaultList.Layout(gtx, len(s.Vaults), func(gtx layout.Context, i int) layout.Dimensions {
				return s.layoutVaultItem(gtx, s.Vaults[i])
			})
		}),
	)
}

func (s *AppState) layoutVaultItem(gtx layout.Context, v *vault.Vault) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(10), Left: unit.Dp(20), Right: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Card Background
		dr := image.Rectangle{Max: gtx.Constraints.Max}
		paint.FillShape(gtx.Ops, ColorSurface, clip.UniformRRect(dr, 8).Op(gtx.Ops))
		
		return layout.UniformInset(unit.Dp(15)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							title := material.H6(s.Theme, v.Title)
							if v.State == vault.StateSealed {
								title.Color = ColorLocked
							}
							return title.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							state := material.Caption(s.Theme, string(v.State))
							state.Color = ColorPrimary
							return state.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					var info string
					if v.State == vault.StateSealed {
						remaining := time.Until(v.UnlockAt)
						if remaining > 0 {
							info = fmt.Sprintf("Locked. Unlockable in %v", remaining.Truncate(time.Second))
						} else {
							info = "Locked. Conditions met. Ready to Ritual Open."
						}
					} else {
						info = fmt.Sprintf("Opened on %v", v.OpenedAt.Format("2006/01/02 15:04"))
					}
					lbl := material.Caption(s.Theme, info)
					lbl.Color = color.NRGBA{R: 150, G: 150, B: 160, A: 255}
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}
