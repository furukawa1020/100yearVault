package ui

import (
	"image/color"
	"log"
	"os"

	"gioui.org/font/opentype"
	"gioui.org/text"
	"gioui.org/widget/material"
)

var (
	ColorBackground = color.NRGBA{R: 10, G: 10, B: 15, A: 255} // Dark Vault Blue-Black
	ColorSurface    = color.NRGBA{R: 20, G: 20, B: 25, A: 255}
	ColorPrimary    = color.NRGBA{R: 180, G: 160, B: 100, A: 255} // Gold/Ancient
	ColorText       = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	ColorLocked     = color.NRGBA{R: 80, G: 80, B: 90, A: 255}
)

func NewVaultTheme(fontPath string) *material.Theme {
	data, err := os.ReadFile(fontPath)
	if err != nil {
		log.Printf("failed to load font: %v, using default", err)
		return material.NewTheme()
	}
	
	face, err := opentype.Parse(data)
	if err != nil {
		log.Printf("failed to parse font: %v, using default", err)
		return material.NewTheme()
	}
	
	fonts := []text.FontFace{
		{
			Font: text.Font{Typeface: "Mincho"},
			Face: face,
		},
	}
	
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(fonts))
	
	// Customize colors
	th.Palette.Bg = ColorBackground
	th.Palette.Fg = ColorText
	th.Palette.ContrastBg = ColorPrimary
	th.Palette.ContrastFg = ColorBackground
	
	return th
}
