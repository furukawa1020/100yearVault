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
	// 「木漏れ日の記憶 (Komorebi Echoes)」パレット
	ColorBackground  = color.NRGBA{R: 5, G: 15, B: 5, A: 255}    // 深い森の静寂 (Deep Forest)
	ColorSurface     = color.NRGBA{R: 10, G: 30, B: 10, A: 255}   // 陽射しの隙間
	ColorSurfaceHigh = color.NRGBA{R: 30, G: 60, B: 30, A: 255}   // 葉の輝き
	ColorPrimary     = color.NRGBA{R: 255, G: 191, B: 0, A: 255}  // 琥珀色の陽光 (Amber Sunlight)
	ColorPrimaryDim  = color.NRGBA{R: 150, G: 110, B: 0, A: 255}  // 夕凪の光
	ColorText        = color.NRGBA{R: 250, G: 245, B: 230, A: 255} // 乳白色の記憶 (Antique Cream)
	ColorTextDim     = color.NRGBA{R: 200, G: 190, B: 160, A: 255} // セピアのささやき
	ColorLocked      = color.NRGBA{R: 5, G: 10, B: 5, A: 255}     // 凪
	ColorDanger      = color.NRGBA{R: 255, G: 80, B: 80, A: 255}   // 鼓動
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
