package ui

import (
	"image/color"
	"os"
	"gioui.org/font/opentype"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	// 「電脳領域の原点 (Neural Origin v9.0) - ZERO FINAL」パレット
	ColorBackground  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}    // 絶対漆黒
	ColorSurface     = color.NRGBA{R: 0, G: 0, B: 0, A: 255}    // 深淵
	ColorSurfaceHigh = color.NRGBA{R: 0, G: 30, B: 60, A: 255}   // 境界面
	ColorPrimary     = color.NRGBA{R: 0, G: 255, B: 255, A: 255} // ネオンブルー
	ColorSecondary   = color.NRGBA{R: 255, G: 0, B: 255, A: 255} // マゼンタ
	ColorTertiary    = color.NRGBA{R: 0, G: 255, B: 150, A: 255} // グリーン
	ColorQuaternary  = color.NRGBA{R: 255, G: 200, B: 0, A: 255} // ゴールド
	ColorPrimaryDim  = color.NRGBA{R: 0, G: 120, B: 150, A: 255} // 暗い信号
	
	ColorDataFragments = []color.NRGBA{
		{R: 0, G: 255, B: 255, A: 255}, // Blue
		{R: 255, G: 0, B: 255, A: 255}, // Magenta
		{R: 0, G: 255, B: 150, A: 255}, // Green
		{R: 255, G: 200, B: 0, A: 255}, // Gold
		{R: 255, G: 255, B: 255, A: 255}, // White
	}

	ColorText    = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Opaque White
	ColorTextDim = color.NRGBA{R: 0, G: 180, B: 200, A: 255}    // 背景データ
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
