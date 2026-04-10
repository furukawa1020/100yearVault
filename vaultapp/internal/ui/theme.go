package ui

import (
	"image/color"
	"os"
	"gioui.org/font/opentype"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	ColorBackground  = color.NRGBA{R: 2, G: 2, B: 4, A: 255}      // The Cold Void (Near Black)
	ColorSurface     = color.NRGBA{R: 0, G: 0, B: 0, A: 255}      // Absolute Stasis
	ColorSurfaceHigh = color.NRGBA{R: 15, G: 15, B: 20, A: 255}   // Monolith Edge
	ColorPrimary     = color.NRGBA{R: 212, G: 175, B: 55, A: 255} // Alchemy Gold
	ColorSecondary   = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Pure Logic (White)
	ColorTertiary    = color.NRGBA{R: 212, G: 175, B: 55, A: 200} // Dim Gold
	ColorQuaternary  = color.NRGBA{R: 255, G: 245, B: 200, A: 255} // Brilliant Aura (Sunlight)
	ColorPrimaryDim  = color.NRGBA{R: 80, G: 65, B: 20, A: 255}   // Oxidized Gold
	
	ColorDataFragments = []color.NRGBA{
		{R: 212, G: 175, B: 55, A: 255},  // Gold
		{R: 255, G: 255, B: 255, A: 255}, // White
		{R: 180, G: 150, B: 40, A: 255},  // Bronze
		{R: 255, G: 250, B: 240, A: 255}, // Ivory
	}

	ColorText    = color.NRGBA{R: 212, G: 175, B: 55, A: 255} // Golden Script
	ColorTextDim = color.NRGBA{R: 80, G: 65, B: 20, A: 255}    // Fading Inscription
	ColorLocked  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	ColorDanger  = color.NRGBA{R: 150, G: 0, B: 0, A: 255}    // Blood/Sacrifice Red
)

func NewVaultTheme(fontPath string) *material.Theme {
	// 2126年標準: 高コントラスト・モノリス
	th := material.NewTheme()
	
	// パレットの絶対適用
	th.Palette.Bg = ColorBackground
	th.Palette.Fg = ColorText
	th.Palette.ContrastBg = ColorPrimary
	th.Palette.ContrastFg = ColorBackground
	th.TextSize = unit.Sp(16)

	// 【物理的強制】等幅フォント Consolas を最優先に
	th.Face = "Consolas"
	
	// カスタムフォントがあれば試行するが、失敗しても Consolas を維持
	if fontPath != "" {
		data, err := os.ReadFile(fontPath)
		if err == nil {
			face, err := opentype.Parse(data)
			if err == nil {
				th.Face = "Neural-Logic"
				_ = face 
			}
		}
	}

	return th
}
