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
	// 2126年標準: 百年の灯火 (Zero-UI) 用パレット
	ColorBackground  = color.NRGBA{R: 0, G: 4, B: 17, A: 255}    // 深い夜空 (Deep Midnight)
	ColorSurface     = color.NRGBA{R: 5, G: 10, B: 30, A: 255}    // 闇の帳
	ColorSurfaceHigh = color.NRGBA{R: 20, G: 30, B: 60, A: 255}   // 遠い銀河
	ColorPrimary     = color.NRGBA{R: 255, G: 215, B: 0, A: 255}  // 太陽の金 (Sunlight Amber)
	ColorPrimaryDim  = color.NRGBA{R: 180, G: 150, B: 0, A: 255}  // 燻る残り火
	ColorText        = color.NRGBA{R: 255, G: 250, B: 240, A: 255} // 純白の光 (High Contrast)
	ColorTextDim     = color.NRGBA{R: 200, G: 200, B: 180, A: 255} // 温かな記憶
	ColorLocked      = color.NRGBA{R: 10, G: 10, B: 20, A: 255}   // 静寂
	ColorDanger      = color.NRGBA{R: 255, G: 50, B: 50, A: 255}   // 鼓動 (Pulse)
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
