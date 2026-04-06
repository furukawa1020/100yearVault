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
	// 2126年標準: 生命の琥珀規格 (LAP) 用パレット
	ColorBackground  = color.NRGBA{R: 12, G: 8, B: 0, A: 255}    // 深い夜 (Midnight Amber)
	ColorSurface     = color.NRGBA{R: 28, G: 20, B: 5, A: 255}   // 琥珀の核 (Amber Core)
	ColorSurfaceHigh = color.NRGBA{R: 45, G: 32, B: 8, A: 255}   // 樹脂 (Resin)
	ColorPrimary     = color.NRGBA{R: 255, G: 191, B: 0, A: 255} // 琥珀光 (Amber Glow)
	ColorPrimaryDim  = color.NRGBA{R: 180, G: 135, B: 0, A: 255} // 燻る金 (Smoldering Gold)
	ColorText        = color.NRGBA{R: 255, G: 245, B: 230, A: 255}// 生命の白 (Vital White)
	ColorTextDim     = color.NRGBA{R: 160, G: 140, B: 110, A: 255}// 記憶の灰 (Memory Ash)
	ColorLocked      = color.NRGBA{R: 50, G: 35, B: 10, A: 255}   // 封印された時間
	ColorDanger      = color.NRGBA{R: 220, G: 60, B: 20, A: 255}   // 生命の鼓動
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
