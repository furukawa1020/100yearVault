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
	// 「電脳領域の特異点 (Neural Singularity v6.0)」パレット
	ColorBackground  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}     // 真・漆黒 (Absolute Void)
	ColorSurface     = color.NRGBA{R: 0, G: 0, B: 0, A: 255}     // 深淵
	ColorSurfaceHigh = color.NRGBA{R: 0, G: 40, B: 50, A: 255}    // 境界面
	ColorPrimary     = color.NRGBA{R: 0, G: 255, B: 255, A: 255}  // ネオンシアン (Neural Blue)
	ColorSecondary   = color.NRGBA{R: 0, G: 180, B: 255, A: 255}  // パルスブルー
	ColorPrimaryDim  = color.NRGBA{R: 0, G: 100, B: 120, A: 255}  // 暗い信号
	ColorText        = color.NRGBA{R: 220, G: 255, B: 255, A: 255} // 高輝度データ
	ColorTextDim     = color.NRGBA{R: 0, G: 150, B: 180, A: 255}  // 背景データ
	ColorLocked      = color.NRGBA{R: 0, G: 0, B: 0, A: 255}      // NULL_STATE
	ColorDanger      = color.NRGBA{R: 255, G: 0, B: 0, A: 255}     // SYSTEM_ERROR
)

func NewVaultTheme(fontPath string) *material.Theme {
	// 2126年標準: 高コントラスト・モノリス
	th := material.NewTheme()
	th.Palette.Bg = ColorBackground
	th.Palette.Fg = ColorText
	th.Palette.ContrastBg = ColorPrimary
	th.Palette.ContrastFg = ColorBackground
	th.TextSize = unit.Sp(16)

	// フォント設定の強制（サイバーパンク/等幅）
	data, err := os.ReadFile(fontPath)
	if err != nil {
		log.Printf("フォント読み込み失敗: システム等幅フォントを明示的に強制")
		// システムの monospace / Consolas を優先するコレクション
		fonts := []font.FontFace{
			{Font: font.Font{Typeface: "Consolas"}},
			{Font: font.Font{Typeface: "monospace"}},
		}
		th.Shaper = text.NewShaper(text.WithCollection(fonts))
		th.Face = "Consolas"
	} else {
		face, err := opentype.Parse(data)
		if err != nil {
			log.Printf("フォント解析失敗: %v", err)
			th.Face = "monospace"
		} else {
			fonts := []font.FontFace{
				{Font: font.Font{Typeface: "Neural-Logic"}, Face: face},
				{Font: font.Font{Typeface: "Neural-Logic", Weight: font.Bold}, Face: face},
			}
			th.Shaper = text.NewShaper(text.WithCollection(fonts))
			th.Face = "Neural-Logic"
		}
	}

	return th
}
