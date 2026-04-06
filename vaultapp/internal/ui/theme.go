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
	// 2126年標準: 不変性整合性規格 (IIS) 用パレット
	ColorBackground  = color.NRGBA{R: 0, G: 0, B: 0, A: 255}     // 真実の黒 (Monolith)
	ColorSurface     = color.NRGBA{R: 26, G: 26, B: 26, A: 255}  // 鋳鉄 (Cast Iron)
	ColorSurfaceHigh = color.NRGBA{R: 45, G: 45, B: 45, A: 255}  // 鋼 (Steel)
	ColorPrimary     = color.NRGBA{R: 197, G: 160, B: 89, A: 255} // 鍛造金 (Forged Gold)
	ColorPrimaryDim  = color.NRGBA{R: 115, G: 90, B: 50, A: 255}  // 鈍色金 (Dull Gold)
	ColorText        = color.NRGBA{R: 242, G: 242, B: 242, A: 255} // 純白 (Invariant White)
	ColorTextDim     = color.NRGBA{R: 160, G: 160, B: 160, A: 255} // 刻印の灰 (Etched Grey)
	ColorLocked      = color.NRGBA{R: 40, G: 40, B: 40, A: 255}    // 閉ざされた鉄
	ColorDanger      = color.NRGBA{R: 180, G: 40, B: 40, A: 255}   // 警告の赤
	ColorUnlockable  = color.NRGBA{R: 197, G: 160, B: 89, A: 255} // 解錠可能時も金を使用

	// 2126年 QSP用: 熵の焔 (Entropy Glow)
	ColorGlow0 = color.NRGBA{R: 0, G: 255, B: 242, A: 255}   // 虚空の青
	ColorGlow1 = color.NRGBA{R: 255, G: 0, B: 200, A: 255}   // 記憶の紫
	ColorGlow2 = color.NRGBA{R: 121, G: 255, B: 0, A: 255}   // 意識の緑
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
