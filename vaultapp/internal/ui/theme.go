package ui

import (
	"image/color"
	"os"
	"gioui.org/font/opentype"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	ColorBackground  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}      // Absolute Void
	ColorSurface     = color.NRGBA{R: 5, G: 5, B: 10, A: 255}     // Monolith Surface
	ColorSurfaceHigh = color.NRGBA{R: 20, G: 20, B: 30, A: 255}   // High Resolution
	ColorPrimary     = color.NRGBA{R: 0, G: 255, B: 255, A: 255} // Neo Cyan
	ColorSecondary   = color.NRGBA{R: 255, G: 0, B: 255, A: 255} // Vivid Magenta
	ColorTertiary    = color.NRGBA{R: 0, G: 255, B: 150, A: 255} // Electric Green
	ColorQuaternary  = color.NRGBA{R: 255, G: 200, B: 0, A: 255} // Solar Gold
	ColorPrimaryDim  = color.NRGBA{R: 0, G: 120, B: 150, A: 255} 
	
	ColorDataFragments = []color.NRGBA{
		{R: 0, G: 255, B: 255, A: 255},   // Cyan
		{R: 255, G: 0, B: 255, A: 255},   // Magenta
		{R: 0, G: 255, B: 150, A: 255},   // Green
		{R: 255, G: 200, B: 0, A: 255},   // Gold
		{R: 255, G: 255, B: 255, A: 255}, // White
	}

	ColorText    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	ColorTextDim = color.NRGBA{R: 0, G: 180, B: 200, A: 255}
	ColorLocked  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	ColorDanger  = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
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
