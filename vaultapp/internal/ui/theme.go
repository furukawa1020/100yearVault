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
	// 2126年標準: 百年の残響規格 (EEP) 用パレット
	ColorBackground  = color.NRGBA{R: 2, G: 5, B: 12, A: 255}    // 深宇宙の紺碧 (Midnight Echo)
	ColorSurface     = color.NRGBA{R: 10, G: 15, B: 30, A: 255}   // 思考の深淵 (Neural Void)
	ColorSurfaceHigh = color.NRGBA{R: 25, G: 35, B: 60, A: 255}   // 銀河の縁 (Galactic Rim)
	ColorPrimary     = color.NRGBA{R: 192, G: 200, B: 220, A: 255} // 流体シルバー (Liquid Silver)
	ColorPrimaryDim  = color.NRGBA{R: 100, G: 120, B: 160, A: 255} // 星屑の灰 (Stardust Ash)
	ColorText        = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 純粋な光 (Pure Echo)
	ColorTextDim     = color.NRGBA{R: 150, G: 170, B: 200, A: 255} // 遠い記憶の光 (Distant Star)
	ColorLocked      = color.NRGBA{R: 15, G: 25, B: 45, A: 255}   // 未同調の沈黙
	ColorDanger      = color.NRGBA{R: 255, G: 80, B: 100, A: 255}  // 生命の臨界 (Critical Pulse)
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
