package ui

import (
	"image/color"
	"log"
	"os"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"
	"gioui.org/widget/material"
)

var (
	// 深海・金庫を連想する配色
	ColorBackground  = color.NRGBA{R: 8, G: 8, B: 12, A: 255}   // 漆黒
	ColorSurface     = color.NRGBA{R: 16, G: 15, B: 22, A: 255}  // 暗礁
	ColorSurfaceHigh = color.NRGBA{R: 28, G: 26, B: 38, A: 255}  // 少し浮き上がり
	ColorPrimary     = color.NRGBA{R: 196, G: 168, B: 96, A: 255} // 古びた金
	ColorPrimaryDim  = color.NRGBA{R: 120, G: 100, B: 50, A: 255} // 錆びた金
	ColorText        = color.NRGBA{R: 200, G: 198, B: 210, A: 255} // 月光
	ColorTextDim     = color.NRGBA{R: 100, G: 98, B: 115, A: 255}  // 霧
	ColorLocked      = color.NRGBA{R: 55, G: 52, B: 72, A: 255}    // 施錠の青
	ColorDanger      = color.NRGBA{R: 160, G: 60, B: 60, A: 255}   // 危険
	ColorUnlockable  = color.NRGBA{R: 80, G: 160, B: 120, A: 255}  // 解錠可能・緑青
)

func NewVaultTheme(fontPath string) *material.Theme {
	th := material.NewTheme()

	data, err := os.ReadFile(fontPath)
	if err != nil {
		log.Printf("フォント読み込み失敗: %v, システムデフォルトで続行", err)
	} else {
		face, err := opentype.Parse(data)
		if err != nil {
			log.Printf("フォント解析失敗: %v", err)
		} else {
			fonts := []font.FontFace{
				{Font: font.Font{Typeface: "Mincho"}, Face: face},
				// Regular / Bold / Italic variants for same face
				{Font: font.Font{Typeface: "Mincho", Weight: font.Bold}, Face: face},
			}
			th.Shaper = text.NewShaper(text.WithCollection(fonts))
			th.Face = "Mincho"
		}
	}

	// カラーパレット適用
	th.Palette.Bg = ColorBackground
	th.Palette.Fg = ColorText
	th.Palette.ContrastBg = ColorPrimary
	th.Palette.ContrastFg = ColorBackground

	return th
}
