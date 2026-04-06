package ui

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"vaultapp/internal/vault"
)

type RitualState struct {
	ActiveVault     *vault.Vault
	Password        widget.Editor
	UnlockBtn       widget.Clickable
	CancelBtn       widget.Clickable
	IsProcessing    bool
	ProcessingSince time.Time
	// 開封後のメッセージ表示
	RevealedText string
	IsRevealed   bool
	ErrorMessage string
}

func (s *AppState) LayoutRitual(gtx layout.Context, r *RitualState) layout.Dimensions {
	fillBackground(gtx, ColorBackground)

	if r.ActiveVault == nil {
		return layout.Dimensions{}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// ヘッダー
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(20), Bottom: unit.Dp(16),
				Left: unit.Dp(32), Right: unit.Dp(32),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Spacing: layout.SpaceStart, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if !r.IsProcessing && !r.IsRevealed {
							btn := material.Button(s.Theme, &r.CancelBtn, "← 戻る")
							btn.Background = ColorSurfaceHigh
							btn.Color = ColorTextDim
							return btn.Layout(gtx)
						}
						return layout.Dimensions{}
					}),
				)
			})
		}),
		// 区切り線
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
			paint.FillShape(gtx.Ops, ColorSurfaceHigh, clip.Rect(image.Rectangle{Max: size}).Op())
			return layout.Dimensions{Size: size}
		}),
		// メイン
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if r.IsRevealed {
				return s.layoutRevealed(gtx, r)
			}
			if r.IsProcessing {
				return s.layoutProcessing(gtx, r)
			}
			return s.layoutRitualInput(gtx, r)
		}),
	)
}

// 開封条件入力画面
func (s *AppState) layoutRitualInput(gtx layout.Context, r *RitualState) layout.Dimensions {
	v := r.ActiveVault
	remaining := time.Until(v.UnlockAt)
	isLocked := remaining > 0

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(480)
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			// アイコン的な区切り
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H2(s.Theme, "⊘")
				if isLocked {
					lbl.Color = ColorLocked
				} else {
					lbl.Color = ColorPrimary
				}
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			// タイトル
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H5(s.Theme, v.Title)
				lbl.Color = ColorText
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			// 作成日
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(s.Theme, fmt.Sprintf("封印日 %s", v.CreatedAt.Format("2006年01月02日")))
				lbl.Color = ColorTextDim
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
			// 状態表示
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if isLocked {
					return s.layoutLockedInfo(gtx, remaining)
				}
				return s.layoutUnlockableInfo(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
			// パスフレーズ入力（常に表示）
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.labeledField(gtx, "封印パスフレーズ", func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(s.Theme, &r.Password, "合言葉を入力してください")
					ed.Color = ColorPrimary
					return ed.Layout(gtx)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
			// ボタン
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if isLocked {
					lbl := material.Caption(s.Theme, "開封条件が未達成です。時が満ちるまで待ちなさい。")
					lbl.Color = ColorTextDim
					return lbl.Layout(gtx)
				}
				btn := material.Button(s.Theme, &r.UnlockBtn, "封印を解く")
				btn.Background = ColorPrimary
				btn.Color = ColorBackground
				btn.TextSize = unit.Sp(16)
				dim := btn.Layout(gtx)
				
				if r.ErrorMessage != "" {
					layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Caption(s.Theme, r.ErrorMessage)
						lbl.Color = ColorDanger
						return layout.Inset{Top: unit.Dp(40)}.Layout(gtx, lbl.Layout)
					})
				}
				return dim
			}),
		)
	})
}

func (s *AppState) layoutLockedInfo(gtx layout.Context, remaining time.Duration) layout.Dimensions {
	days := int(remaining.Hours() / 24)
	hours := int(remaining.Hours()) % 24
	mins := int(remaining.Minutes()) % 60

	var text string
	if days > 365 {
		years := days / 365
		text = fmt.Sprintf("あと約 %d 年", years)
	} else if days > 0 {
		text = fmt.Sprintf("あと %d 日 %d 時間", days, hours)
	} else {
		text = fmt.Sprintf("あと %d 時間 %d 分", hours, mins)
	}

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H4(s.Theme, text)
			lbl.Color = ColorLocked
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Caption(s.Theme, "この封印はまだ開けられません")
			lbl.Color = ColorTextDim
			return lbl.Layout(gtx)
		}),
	)
}

func (s *AppState) layoutUnlockableInfo(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H5(s.Theme, "開封条件が成立しています")
			lbl.Color = ColorUnlockable
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Caption(s.Theme, "パスフレーズを入力して封印を解いてください")
			lbl.Color = ColorTextDim
			return lbl.Layout(gtx)
		}),
	)
}

// 儀式演出中の画面 (2126年標準: 機械的解錠)
func (s *AppState) layoutProcessing(gtx layout.Context, r *RitualState) layout.Dimensions {
	elapsed := time.Since(r.ProcessingSince).Seconds()
	progress := math.Min(elapsed/2.0, 1.0)

	// 背景の機械的ギミック描画
	center := image.Pt(gtx.Constraints.Max.X/2, gtx.Constraints.Max.Y/2)
	
	// 2つの回転するリング（ラチェット風）
	for i := 0; i < 2; i++ {
		direction := 1.0
		if i%2 == 1 {
			direction = -1.0
		}
		angle := float64(elapsed) * 2.0 * direction
		radius := gtx.Dp(unit.Dp(100 + i*40))
		
		// 12個の「歯」を持つリング
		for n := 0; n < 12; n++ {
			toothAngle := angle + float64(n)*math.Pi/6
			tx := float64(center.X) + float64(radius)*math.Cos(toothAngle)
			ty := float64(center.Y) + float64(radius)*math.Sin(toothAngle)
			
			toothSize := gtx.Dp(unit.Dp(8))
			toothRect := image.Rectangle{
				Min: image.Pt(int(tx)-toothSize, int(ty)-toothSize),
				Max: image.Pt(int(tx)+toothSize, int(ty)+toothSize),
			}
			paint.FillShape(gtx.Ops, ColorSurfaceHigh, clip.Rect(toothRect).Op())
		}
	}

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H4(s.Theme, "UNSEALING")
				lbl.Color = ColorPrimary
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
			// プログレスバー
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				width := gtx.Dp(240)
				height := gtx.Dp(2)
				barRect := image.Rectangle{Max: image.Pt(width, height)}
				paint.FillShape(gtx.Ops, ColorSurfaceHigh, clip.Rect(barRect).Op())
				
				progressWidth := int(float64(width) * progress)
				progressRect := image.Rectangle{Max: image.Pt(progressWidth, height)}
				paint.FillShape(gtx.Ops, ColorPrimary, clip.Rect(progressRect).Op())
				
				return layout.Dimensions{Size: image.Pt(width, height)}
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
			// 整合性チェックログ
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				logs := []string{
					"BOOTING IIS-v2126...",
					"VERIFYING TEMPORAL INTEGRITY...",
					"ENTROPY CHECK: NOMINAL",
					"DECRYPTING CENTURY-BLOCK...",
					"IIS-STANDARD STATUS: OK",
				}
				idx := int(elapsed * 4)
				if idx >= len(logs) {
					idx = len(logs) - 1
				}
				lbl := material.Caption(s.Theme, logs[idx])
				lbl.Color = ColorTextDim
				return lbl.Layout(gtx)
			}),
		)
	})
}

// 開封後の内容表示画面
func (s *AppState) layoutRevealed(gtx layout.Context, r *RitualState) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = gtx.Dp(560)
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H2(s.Theme, "◈")
				lbl.Color = ColorPrimary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H5(s.Theme, "封印が解かれました")
				lbl.Color = ColorPrimary
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Caption(s.Theme, r.ActiveVault.Title)
				lbl.Color = ColorTextDim
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
			// 本文カード
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Dp(220))}
				paint.FillShape(gtx.Ops, ColorSurface, clip.UniformRRect(dr, 8).Op(gtx.Ops))
				return layout.UniformInset(unit.Dp(20)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(s.Theme, r.RevealedText)
					lbl.Color = ColorText
					return lbl.Layout(gtx)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),
			// 戻るボタン
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(s.Theme, &r.CancelBtn, "保管庫に戻る")
				btn.Background = ColorSurfaceHigh
				btn.Color = ColorText
				return btn.Layout(gtx)
			}),
		)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
