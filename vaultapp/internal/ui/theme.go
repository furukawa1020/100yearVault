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
	// 「電脳領域の特異点 (Neural Singularity v4.0)」パレット
	ColorBackground  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}     // 真・漆黒 (Absolute Void)
	ColorSurface     = color.NRGBA{R: 0, G: 0, B: 0, A: 255}     // 深淵
	ColorSurfaceHigh = color.NRGBA{R: 0, G: 40, B: 40, A: 255}    // 境界面
	ColorPrimary     = color.NRGBA{R: 0, G: 255, B: 255, A: 255}  // ネオンシアン (Neural Cyan)
	ColorSecondary   = color.NRGBA{R: 255, G: 0, B: 255, A: 255}  // ネオンマゼンタ (Pulse Magenta)
	ColorPrimaryDim  = color.NRGBA{R: 0, G: 120, B: 120, A: 255}  // 暗い信号
	ColorText        = color.NRGBA{R: 200, G: 255, B: 255, A: 255} // 高輝度データ
	ColorTextDim     = color.NRGBA{R: 0, G: 150, B: 150, A: 255}  // 背景データ
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
		log.Printf("フォント読み込み失敗: システム等幅フォントを使用")
		// システムの monospace を優先するコレクションを作成
		th.Shaper = text.NewShaper(text.WithCollection([]font.FontFace{}))
		th.Face = "monospace"
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
