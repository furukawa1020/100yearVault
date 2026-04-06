package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"vaultapp/internal/vault"
)

type Screen int

const (
	ScreenVaultList Screen = iota
	ScreenCompose
	ScreenRitual
)

type Particle struct {
	BaseX, BaseY, BaseZ float32 // 初期配置 (非破壊)
	X, Y, Z             float32 // 計算後の投影用
	VX, VY, VZ          float32 // 反応的なオフセット
	Life                float32
	Color               color.NRGBA // 極彩色データ
}

type AppState struct {
	Theme         *material.Theme
	Vaults        []*vault.Vault
	CurrentScreen Screen

	// Sub-states
	Compose ComposeState
	Ritual  RitualState

	// Widgets
	NewVaultBtn widget.Clickable
	VaultList   layout.List
	SelectBtns  []widget.Clickable

	// 電脳領域 (NeuralSingularity v4.0)
	NeuralVault   *vault.Vault
	NeuralText    string
	NeuralSurface widget.Clickable
	
	// アニメーション・物理演算用計測値
	Rotation     float32
	PulsePhase   float32
	Particles    []Particle
	InitOnce     bool
	FrameCount   int
	MousePos     f32.Point
}

func (s *AppState) initNeuralSpace() {
	if s.InitOnce {
		return
	}
	s.Particles = make([]Particle, 4096) // 4096 粒子 (極限密度)
	for i := range s.Particles {
		// 3D マハラノビス空間：歪んだ楕円体状に分布
		angle1 := rand.Float64() * 2 * math.Pi
		angle2 := rand.Float64() * math.Pi
		// 軸ごとに異なる分散（共分散行列のイメージ）
		distX := 150 + rand.Float64()*450
		distY := 100 + rand.Float64()*300
		distZ := 200 + rand.Float64()*600
		
		s.Particles[i] = Particle{
			BaseX: float32(math.Sin(angle2)*math.Cos(angle1) * distX),
			BaseY: float32(math.Sin(angle2)*math.Sin(angle1) * distY),
			BaseZ: float32(math.Cos(angle2) * distZ),
			VX:    float32(rand.NormFloat64() * 0.1),
			VY:    float32(rand.NormFloat64() * 0.1),
			VZ:    float32(rand.NormFloat64() * 0.1),
			Life:  rand.Float32(),
			Color: ColorDataFragments[rand.Intn(len(ColorDataFragments))],
		}
	}
	s.InitOnce = true
}

func (s *AppState) RotateNeural() {
	if len(s.Vaults) == 0 {
		s.NeuralVault = nil
		s.NeuralText = "CRITICAL: NO_DATA_STREAM"
		return
	}
	
	idx := rand.Intn(len(s.Vaults))
	v := s.Vaults[idx]
	s.NeuralVault = v
	
	if v.State == vault.StateOpened {
		s.NeuralText = "FRAGMENT_LOADED: [" + v.Title + "]"
	} else if time.Now().After(v.UnlockAt) {
		s.NeuralText = "SIGNAL_DETECTED: [" + v.Title + "]\nREADY_TO_INITIALIZE_LINK"
	} else {
		s.NeuralText = "STASIS_LOCKED: [" + v.Title + "]\nTIMELOCK_ACTIVE"
	}
}

// ───────────────────────────────────────────────
// 背景を塗る共通ヘルパー
// ───────────────────────────────────────────────
func fillBackground(gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(gtx.Ops, c, clip.Rect(dr).Op())
}

// ───────────────────────────────────────────────
// 電脳領域の特異点: Neural Origin v8.0 (Colorful Final)
// ───────────────────────────────────────────────
func (s *AppState) LayoutNeural(gtx layout.Context) layout.Dimensions {
	s.initNeuralSpace()
	s.FrameCount++
	
	// 【二重防衛】レイアウト内部でも漆黒フラッシュを実行
	paint.ColorOp{Color: ColorBackground}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	return material.Clickable(gtx, &s.NeuralSurface, func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			// 1. 三次元演算レイヤー (Ultimate Color Engine)
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				center := f32.Pt(float32(gtx.Constraints.Max.X)/2, float32(gtx.Constraints.Max.Y)/2)
				focalLength := float32(900) // 焦点距離
				
				// A. 4096 多色粒子 3D シミュレーション
				for i := range s.Particles {
					p := &s.Particles[i]
					
					// 3D 回転 (AppState.Rotation に同期)
					rotX, rotY := float64(s.Rotation*0.3), float64(s.Rotation*0.7)
					sinX, cosX := float32(math.Sin(rotX)), float32(math.Cos(rotX))
					sinY, cosY := float32(math.Sin(rotY)), float32(math.Cos(rotY))
					
					tx := p.BaseX*cosY - p.BaseZ*sinY
					tz := p.BaseX*sinY + p.BaseZ*cosY
					ty := p.BaseY*cosX - tz*sinX
					tz = p.BaseY*sinX + tz*cosX

					p.VX *= 0.96
					p.VY *= 0.96
					p.X = tx + p.VX
					p.Y = ty + p.VY
					p.Z = tz

					// 高感度斥力
					dx := p.X - (s.MousePos.X - center.X)
					dy := p.Y - (s.MousePos.Y - center.Y)
					distSq := dx*dx + dy*dy
					if distSq < 25000 {
						force := (25000 - distSq) / 25000
						p.VX += dx * force * 0.4
						p.VY += dy * force * 0.4
					}
					
					// 3D -> 2D 投影
					zWorld := p.Z + 900
					if zWorld <= 50 { continue }
					
					sx := center.X + (p.X * focalLength) / zWorld
					sy := center.Y + (p.Y * focalLength) / zWorld
					
					// カラフルな光の粒
					size := 1.5 + (1000 / zWorld)
					if size > 10 { size = 10 }
					
					alphaVal := uint32(255 * (900 / zWorld))
					if alphaVal > 255 { alphaVal = 255 }
					
					c := p.Color // 個別色
					c.A = uint8(alphaVal)
					
					rect := image.Rect(int(sx), int(sy), int(sx+size), int(sy+size))
					paint.FillShape(gtx.Ops, c, clip.Rect(rect).Op())
				}

				// B. 3D 回転ヘキサゴン・ゲート (Neo Blue)
				for i := 1; i <= 3; i++ {
					rot := s.Rotation * float32(i) * 0.3
					size := float32(150 * i)
					
					var path clip.Path
					path.Begin(gtx.Ops)
					for a := 0.0; a < 2*math.Pi; a += math.Pi/3 {
						curA := a + float64(rot)
						vx := size * float32(math.Cos(curA))
						vy := size * float32(math.Sin(curA))
						vz := float32(math.Sin(float64(s.Rotation*3))) * 100
						
						tz := vz + 1000
						tsx := center.X + (vx * focalLength) / tz
						tsy := center.Y + (vy * focalLength) / tz
						
						if a == 0 { path.MoveTo(f32.Pt(tsx, tsy)) } else { path.LineTo(f32.Pt(tsx, tsy)) }
					}
					path.Close()
					paint.FillShape(gtx.Ops, ColorPrimary, clip.Stroke{Path: path.End(), Width: 1.5}.Op())
				}

				return layout.Dimensions{Size: gtx.Constraints.Max}
			}),
			
			// 2. システム情報オーバーレイ (Guaranteed Monospace)
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(60)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							msg := fmt.Sprintf("NEURAL_SYNC_V8: 0x%04X | NODE_HEALTH: 100%% | FRAGMENTS: 4096", s.FrameCount%0xFFFF)
							// material をバイパスし直接 Consolas で描画
							return drawRawLabel(gtx, s.Theme, msg, 12, ColorPrimaryDim)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							// 微妙な座標揺れ (Glitch)
							offset := image.Pt(int(rand.NormFloat64()*1.2), int(rand.NormFloat64()*1.2))
							stack := op.Offset(offset).Push(gtx.Ops)
							dims := drawRawLabel(gtx, s.Theme, s.NeuralText, 38, ColorText)
							stack.Pop()
							return dims
						}),
					)
				})
			}),

			// 3. インジェクション・ゲート
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return layout.S.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(30)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(s.Theme, &s.NewVaultBtn, "[ INITIATE_NEURAL_UPLINK ]")
						btn.Background = color.NRGBA{A: 0}
						btn.Color = ColorSecondary
						btn.TextSize = unit.Sp(24)
						return btn.Layout(gtx)
					})
				})
			}),
		)
	})
}

// ───────────────────────────────────────────────
// 【最終兵器】素材ウィジェットを頼らずに直接描画する (Serif 体追放)
// ───────────────────────────────────────────────
func drawRawLabel(gtx layout.Context, th *material.Theme, txt string, size int, clr color.NRGBA) layout.Dimensions {
	// Consolas を強制的に探す shaper
	label := material.Label(th, unit.Sp(float32(size)), txt)
	label.Color = clr
	label.Font.Typeface = "Consolas" // これが重要
	label.Alignment = text.Middle
	return label.Layout(gtx)
}

func drawBackground(ops *op.Ops, gtx layout.Context, c color.NRGBA) {
	dr := image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)}
	paint.FillShape(ops, c, clip.Rect(dr).Op())
}
