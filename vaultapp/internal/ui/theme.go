package ui

import (
	"image/color"
	"log"
	"os"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	// 「電脳領域の特異点 (Neural Origin v8.0) - COLORFUL FINAL」パレット
	ColorBackground  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}    // 真・漆黒
	ColorSurface     = color.NRGBA{R: 0, G: 0, B: 0, A: 255}    // 深淵
	ColorSurfaceHigh = color.NRGBA{R: 0, G: 20, B: 40, A: 255}   // 境界面
	ColorPrimary     = color.NRGBA{R: 0, G: 255, B: 255, A: 255} // ネオンブルー
	ColorSecondary   = color.NRGBA{R: 255, G: 0, B: 255, A: 255} // マゼンタ
	ColorTertiary    = color.NRGBA{R: 0, G: 255, B: 120, A: 255} // エメラルド
	ColorQuaternary  = color.NRGBA{R: 255, G: 200, B: 0, A: 255} // ゴールド
	ColorPrimaryDim  = color.NRGBA{R: 0, G: 100, B: 120, A: 255} // 暗い信号
	
	ColorDataFragments = []color.NRGBA{
		{R: 0, G: 255, B: 255, A: 255}, // Blue
		{R: 255, G: 0, B: 255, A: 255}, // Magenta
		{R: 0, G: 255, B: 150, A: 255}, // Green
		{R: 255, G: 200, B: 0, A: 255}, // Gold
		{R: 255, G: 255, B: 255, A: 255}, // White
	}

	ColorText    = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 高輝度
	ColorTextDim = color.NRGBA{R: 0, G: 180, B: 200, A: 255}    // 背景データ
	ColorLocked  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	ColorDanger  = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
)

func NewVaultTheme(fontPath string) *material.Theme {
	// 2126年標準: 高コントラスト・モノリス
	th := material.NewTheme()
	th.Palette.Bg = ColorBackground
	th.Palette.Fg = ColorText
	th.Palette.ContrastBg = ColorPrimary
	th.Palette.ContrastFg = ColorBackground
	th.TextSize = unit.Sp(16)

	// フォント設定の強制（システム等幅 Consolas を優先）
	th.Face = "Consolas"
	
	data, err := os.ReadFile(fontPath)
	if err == nil {
		face, err := opentype.Parse(data)
		if err == nil {
			// カスタムフォントがあればセット（コレクションへの追加は行わない）
			th.Face = "Neural-Logic"
			_ = face // フォント解析成功
		}
	}

	return th
}
