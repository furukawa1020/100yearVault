package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"

	pigo "github.com/esimov/pigo/core"
	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/camera"

	"vaultapp/internal/crypto"
	"vaultapp/internal/db"
	"vaultapp/internal/ui"
	"vaultapp/internal/vault"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("Hundred-Year Vault"),
			app.Size(unit.Dp(1000), unit.Dp(800)),
		)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	// Initialize Store inside loop to catch errors properly
	dbPath := filepath.Join(".", "vault.db")
	store, err := db.NewStore(dbPath)
	if err != nil {
		return err
	}
	defer store.Close()

	// Initial Load
	memories, _ := store.ListMemories()

	// UI State
	// UI State (Absolute Pathing)
	fontPath, _ := filepath.Abs(filepath.Join(".", "assets", "fonts", "font.ttf"))
	th := ui.NewVaultTheme(fontPath)
	state := &ui.AppState{
		Theme:      th,
		Memories:   memories,
		SelectBtns: make([]widget.Clickable, len(memories)),
	}
	state.Compose.UnlockDays.SetText("36500")

	state.RotateNeural()

	// 高速アニメーション・クロック (60FPS 極限稼働)
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond)
		for range ticker.C {
			state.Rotation += 0.03 // 回転速度を微増
			w.Invalidate()
		}
	}()
	
	// データ巡回クロック
	go func() {
		for {
			time.Sleep(12 * time.Second) // 12秒ごとにデータ断片を再構成
			if state.CurrentScreen == ui.ScreenVaultList {
				state.RotateNeural()
			}
		}
	}()

	// EYE-OF-THE-COSMOS: Webcam Integration
	go startWebcamGazeTracking(state)

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// 【絶対的回帰 v9.0】Neural Zero Layering
			// すべてのウィジェットの背後に漆黒の岩盤を物理的に固定。
			// これにより、透過やバッファ汚染による「灰色」を根絶する。
			layout.Stack{Alignment: layout.Center}.Layout(gtx,
				// Layer 0: Absolute Void (岩盤)
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					// 【二重防衛】各スクリーンレベルでも物理的漆黒クリアを再実行
					paint.FillShape(gtx.Ops, ui.ColorBackground, clip.Rect{Max: gtx.Constraints.Max}.Op())
					return layout.Dimensions{Size: gtx.Constraints.Max}
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					// 1. Absolute Manual Event Tracker (No Clickable overhead)
					// Covers entire screen and intercepts everything.
					defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
					event.Op(gtx.Ops, state)

					// ロジック実行（座標の読み出しもここで行う）
					updateLogic(gtx, state, store, w)

					// 描画実行
					switch state.CurrentScreen {
					case ui.ScreenVaultList:
						state.LayoutNeural(gtx)
					case ui.ScreenCompose:
						state.LayoutCompose(gtx, &state.Compose)
					case ui.ScreenRitual:
						state.LayoutRitual(gtx, &state.Ritual)
					}
					return layout.Dimensions{Size: gtx.Constraints.Max}
				}),
			)

			e.Frame(gtx.Ops)
		}
	}
}

func updateLogic(gtx layout.Context, state *ui.AppState, store *db.Store, w *app.Window) {
	// Global: New Vault Button
	if state.NewVaultBtn.Clicked(gtx) {
		state.CurrentScreen = ui.ScreenCompose
		w.Invalidate()
	}

	// イベント・フィルタリング：マウス座標とクリックの完全手動捕捉
	{
		for {
			ev, ok := gtx.Event(
				pointer.Filter{
					Target: state,
					Kinds:  pointer.Enter | pointer.Leave | pointer.Move | pointer.Drag | pointer.Press | pointer.Release,
				},
			)
			if !ok {
				break
			}
			if xev, ok := ev.(pointer.Event); ok {
				if xev.Kind == pointer.Move || xev.Kind == pointer.Drag {
					state.MousePos = xev.Position
				}
				if xev.Kind == pointer.Press && state.CurrentScreen == ui.ScreenVaultList {
					if state.NeuralMemory != nil {
						m := state.NeuralMemory
						// 記憶への同調
						state.Ritual.ActiveMemory = m
						state.CurrentScreen = ui.ScreenRitual
						w.Invalidate()
					}
				}
			}
		}
	}

	// Compose Screen Logic
	if state.CurrentScreen == ui.ScreenCompose {
		if state.Compose.BackBtn.Clicked(gtx) {
			state.Compose.ErrorMessage = ""
			state.CurrentScreen = ui.ScreenVaultList
			w.Invalidate()
		}
		if state.Compose.SealBtn.Clicked(gtx) {
			// Validation
			title := state.Compose.Title.Text()
			body := state.Compose.Body.Text()
			pass := state.Compose.Passphrase.Text()
			
			if title == "" && !state.Compose.AddLayerMode {
				state.Compose.ErrorMessage = "記憶の題名が必要です。"
				w.Invalidate()
			} else {
				mid := ""
				if state.Compose.AddLayerMode && state.Compose.TargetMemory != nil {
					mid = state.Compose.TargetMemory.ID
				} else {
					mid = fmt.Sprintf("m%d", time.Now().Unix())
				}

				layerID := fmt.Sprintf("l%d", time.Now().UnixNano())
				cipherPath := filepath.Join("vaults", layerID+".age")
				
				// Ensure vaults directory exists
				os.MkdirAll("vaults", 0700)

				// Encrypt (深層としての保護)
				ciphertext, err := crypto.Encrypt([]byte(body), pass)
				if err != nil {
					state.Compose.ErrorMessage = "共鳴の失敗: " + err.Error()
					w.Invalidate()
				} else {
					err = os.WriteFile(cipherPath, ciphertext, 0600)
					if err != nil {
						state.Compose.ErrorMessage = "時層の安定化に失敗しました。"
						w.Invalidate()
					} else {
						if state.Compose.AddLayerMode && state.Compose.TargetMemory != nil {
							l := &vault.Layer{
								ID:         layerID,
								ParentID:   mid,
								CipherPath: cipherPath,
								CreatedAt:  time.Now(),
							}
							if err := store.SaveLayer(l); err != nil {
								state.Compose.ErrorMessage = "地層の保存に失敗しました。"
								w.Invalidate()
								return
							}
						} else {
							// 記憶の断片を宇宙に放流
							m := &vault.MemoryFragment{
								ID:                mid,
								Title:             title,
								Aura:              vault.StatePulse,
								CreatedAt:         time.Now(),
								Luminosity:        1.0,
								CipherPath:        cipherPath,
								RequirePassphrase: pass != "",
								PreviewHint:       body, 
							}
							if err := store.SaveMemory(m); err != nil {
								state.Compose.ErrorMessage = "記憶の放流に失敗しました。"
								w.Invalidate()
								return
							}
						}
						
						// Success Cleanup
						state.Compose.Title.SetText("")
						state.Compose.Body.SetText("")
						state.Compose.Passphrase.SetText("")
						state.Compose.ErrorMessage = ""
						state.Compose.AddLayerMode = false
						state.Compose.TargetMemory = nil
						
						// Refresh List
						state.Memories, _ = store.ListMemories()
						state.SelectBtns = make([]widget.Clickable, len(state.Memories))
						state.CurrentScreen = ui.ScreenVaultList
						w.Invalidate()
					}
				}
			}
		}
	}

	// Ritual Screen Logic (Memory Sync)
	if state.CurrentScreen == ui.ScreenRitual {
		if state.Ritual.CancelBtn.Clicked(gtx) {
			state.Ritual.IsProcessing = false
			state.Ritual.IsRevealed = false
			state.Ritual.ErrorMessage = ""
			state.CurrentScreen = ui.ScreenVaultList
			w.Invalidate()
		}
		if state.Ritual.AddLayerBtn.Clicked(gtx) {
			state.Compose.AddLayerMode = true
			state.Compose.TargetMemory = state.Ritual.ActiveMemory
			state.Compose.Title.SetText("Reflection: " + state.Ritual.ActiveMemory.Title)
			state.CurrentScreen = ui.ScreenCompose
			w.Invalidate()
		}
		if state.Ritual.UnlockBtn.Clicked(gtx) && !state.Ritual.IsProcessing && !state.Ritual.IsRevealed {
			state.Ritual.IsProcessing = true
			state.Ritual.ProcessingSince = time.Now()
			state.Ritual.ErrorMessage = ""
			w.Invalidate()
		}

		if state.Ritual.IsProcessing {
			if time.Since(state.Ritual.ProcessingSince) > (3*time.Second)/2 {
				m := state.Ritual.ActiveMemory
				pass := state.Ritual.Password.Text()
				
				// 記憶へのアクセス
				cipherData, err := os.ReadFile(m.CipherPath)
				if err != nil {
					state.Ritual.ErrorMessage = "記憶の断片を読み取れません。"
					state.Ritual.IsProcessing = false
					w.Invalidate()
				} else {
					decrypted, err := crypto.Decrypt(cipherData, pass)
					if err != nil {
						state.Ritual.ErrorMessage = "想いが一致しません（合言葉を確認してください）。"
						state.Ritual.IsProcessing = false
						w.Invalidate()
					} else {
						// 同調成功
						m.Aura = vault.StateRadiant
						m.Luminosity = 1.0
						store.SaveMemory(m)
						
						state.Ritual.IsProcessing = false
						state.Ritual.IsRevealed = true
						
						// 地層の読み込み
						layers, _ := store.ListLayers(m.ID)
						fullText := "--- [ORIGIN] ---\n" + string(decrypted)
						for i, l := range layers {
							lData, _ := os.ReadFile(l.CipherPath)
							lDec, err := crypto.Decrypt(lData, pass)
							if err == nil {
								fullText += fmt.Sprintf("\n\n--- [LAYER %d (%s)] ---\n%s", i+1, l.CreatedAt.Format("2006/01/02"), string(lDec))
							}
						}
						state.Ritual.RevealedText = fullText
						state.Ritual.Password.SetText("") 
						
						state.RotateNeural()
						state.Memories, _ = store.ListMemories()
						state.SelectBtns = make([]widget.Clickable, len(state.Memories))
						w.Invalidate()
					}
				}
			} else {
				w.Invalidate()
			}
		}
	}
}
