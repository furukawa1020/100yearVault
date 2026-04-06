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
	// 「電脳の深淵 (Neural Void)」パレット - ハード・サイバー
	ColorBackground  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}     // 漆黒 (Pitch Black)
	ColorSurface     = color.NRGBA{R: 0, G: 20, B: 20, A: 255}    // 深淵の底
	ColorSurfaceHigh = color.NRGBA{R: 0, G: 40, B: 40, A: 255}    // データの境界
	ColorPrimary     = color.NRGBA{R: 0, G: 255, B: 255, A: 255}  // ネオンシアン (Cyber Cyan)
	ColorPrimaryDim  = color.NRGBA{R: 0, G: 150, B: 150, A: 255}  // 減衰した信号
	ColorText        = color.NRGBA{R: 220, G: 255, B: 255, A: 255} // 高輝度データ (High Contrast)
	ColorTextDim     = color.NRGBA{R: 100, G: 200, B: 200, A: 255} // 背景データ
	ColorLocked      = color.NRGBA{R: 0, G: 0, B: 0, A: 255}      // NULL
	ColorDanger      = color.NRGBA{R: 255, G: 0, B: 0, A: 255}     // ERROR
)

func NewVaultTheme(fontPath string) *material.Theme {
	th := material.NewTheme()

	// デフォルトのフォント設定
	th.TextSize = unit.Sp(16)

	data, err := os.ReadFile(fontPath)
	if err != nil {
		log.Printf("フォント読み込み失敗: %v, システムデフォルトで続行", err)
	} else {
		face, err := opentype.Parse(data)
		if err != nil {
			log.Printf("フォント解析失敗: %v", err)
		} else {
			fonts := []font.FontFace{
				{Font: font.Font{Typeface: "IIS-Legacy"}, Face: face},
				{Font: font.Font{Typeface: "IIS-Legacy", Weight: font.Bold}, Face: face},
			}
			th.Shaper = text.NewShaper(text.WithCollection(fonts))
			th.Face = "IIS-Legacy"
		}
	}

	// 2126年標準: 高コントラスト・モノリス配色
	th.Palette.Bg = ColorBackground
	th.Palette.Fg = ColorText
	th.Palette.ContrastBg = ColorPrimary
	th.Palette.ContrastFg = ColorBackground

	return th
}
