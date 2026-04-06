package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"

	"vaultapp/internal/crypto"
	"vaultapp/internal/db"
	"vaultapp/internal/ui"
	"vaultapp/internal/vault"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Hundred-Year Vault"))
		w.Option(app.Size(unit.Dp(1000), unit.Dp(800)))
		
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	// Initialize Store
	dbPath := filepath.Join(".", "vault.db")
	store, err := db.NewStore(dbPath)
	if err != nil {
		return err
	}
	defer store.Close()

	// Initial Load
	vaults, _ := store.ListVaults()

	// UI State
	fontPath := filepath.Join(".", "assets", "fonts", "font.ttf")
	th := ui.NewVaultTheme(fontPath)
	state := &ui.AppState{
		Theme:      th,
		Vaults:     vaults,
		SelectBtns: make([]widget.Clickable, len(vaults)),
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

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Logic Handling
			updateLogic(gtx, state, store, w)

			// Main Layout (Neural Interface)
			switch state.CurrentScreen {
			case ui.ScreenVaultList:
				state.LayoutNeural(gtx)
			case ui.ScreenCompose:
				state.LayoutCompose(gtx, &state.Compose)
			case ui.ScreenRitual:
				state.LayoutRitual(gtx, &state.Ritual)
			}

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

	// Neural Void Logic (Deep-Dive v2.0)
	if state.CurrentScreen == ui.ScreenVaultList {
		if state.NeuralSurface.Clicked(gtx) {
			if state.NeuralVault != nil {
				v := state.NeuralVault
				// 刻が満ちているか、既に開封済みの場合は「システム・アクセス」へ
				if v.State == vault.StateOpened || time.Now().After(v.UnlockAt) {
					state.Ritual.ActiveVault = v
					state.CurrentScreen = ui.ScreenRitual
					w.Invalidate()
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
			daysInput := state.Compose.UnlockDays.Text()

			if title == "" && !state.Compose.AddLayerMode {
				state.Compose.ErrorMessage = "タイトルを入力してください。"
				w.Invalidate()
			} else if pass == "" {
				state.Compose.ErrorMessage = "封印のための合言葉（パスフレーズ）が必要です。"
				w.Invalidate()
			} else if len(pass) < 4 {
				state.Compose.ErrorMessage = "合言葉は少なくとも4文字以上必要です。"
				w.Invalidate()
			} else {
				vid := ""
				if state.Compose.AddLayerMode && state.Compose.TargetVault != nil {
					vid = state.Compose.TargetVault.ID
				} else {
					vid = fmt.Sprintf("v%d", time.Now().Unix())
				}

				layerID := fmt.Sprintf("l%d", time.Now().UnixNano())
				cipherPath := filepath.Join("vaults", layerID+".age")
				
				// Ensure vaults directory exists
				os.MkdirAll("vaults", 0700)

				// Encrypt
				ciphertext, err := crypto.Encrypt([]byte(body), pass)
				if err != nil {
					state.Compose.ErrorMessage = "封印に失敗しました: " + err.Error()
					w.Invalidate()
				} else {
					err = os.WriteFile(cipherPath, ciphertext, 0600)
					if err != nil {
						state.Compose.ErrorMessage = "ファイルの保存に失敗しました。"
						w.Invalidate()
					} else {
						if state.Compose.AddLayerMode && state.Compose.TargetVault != nil {
							l := &vault.Layer{
								ID:         layerID,
								ParentID:   vid,
								CipherPath: cipherPath,
								CreatedAt:  time.Now(),
							}
							if err := store.SaveLayer(l); err != nil {
								log.Printf("STP LAYER SAVE ERROR: %v", err)
								state.Compose.ErrorMessage = "地層の保存に失敗しました。"
								w.Invalidate()
								return
							}
						} else {
							// 新規残響の作成 (残響の刻印)
							days, _ := strconv.ParseFloat(daysInput, 64)
							if days <= 0 {
								days = 36500 
							}

							// 2126年 EEP: 思考の漂流 (意志の残響)
							maxSeconds := days * 24 * 60 * 60
							randomSeconds := float64(time.Now().UnixNano()%int64(maxSeconds))
							
							unlockAt := time.Now().Add(time.Duration(randomSeconds * float64(time.Second)))

							v := &vault.Vault{
								ID:                vid,
								Title:             title,
								State:             vault.StateSealed,
								CreatedAt:         time.Now(),
								UnlockAt:          unlockAt,
								CipherPath:        cipherPath,
								RequirePassphrase: true,
								PreviewHint:       body, 
							}
							if err := store.SaveVault(v); err != nil {
								log.Printf("QSP VAULT SAVE ERROR: %v", err)
								state.Compose.ErrorMessage = "時空への放流に失敗しました。"
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
						state.Compose.TargetVault = nil
						
						// Refresh List
						var err error
						state.Vaults, err = store.ListVaults()
						if err != nil {
							log.Printf("LIST REFRESH ERROR: %v", err)
						}
						state.SelectBtns = make([]widget.Clickable, len(state.Vaults))
						state.CurrentScreen = ui.ScreenVaultList
						w.Invalidate()
					}
				}
			}
		}
	}

	// Ritual Screen Logic
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
			state.Compose.TargetVault = state.Ritual.ActiveVault
			state.Compose.Title.SetText("RE: " + state.Ritual.ActiveVault.Title)
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
			if time.Since(state.Ritual.ProcessingSince) > 2*time.Second {
				v := state.Ritual.ActiveVault
				pass := state.Ritual.Password.Text()
				
				if time.Now().Before(v.UnlockAt) {
					state.Ritual.ErrorMessage = "まだ刻（とき）が満ちていません。"
					state.Ritual.IsProcessing = false
					w.Invalidate()
				} else {
					// 実際に復号を試みることでパスフレーズの正当性を確認する
					cipherData, err := os.ReadFile(v.CipherPath)
					if err != nil {
						state.Ritual.ErrorMessage = "封印ファイルの読み込みに失敗しました。"
						state.Ritual.IsProcessing = false
						w.Invalidate()
					} else {
						decrypted, err := crypto.Decrypt(cipherData, pass)
						if err != nil {
							// 復号失敗 = パスフレーズが違う
							state.Ritual.ErrorMessage = "合言葉が違います。記憶はまだ閉ざされています。"
							state.Ritual.IsProcessing = false
							w.Invalidate()
						} else {
							// 復号成功
							v.State = vault.StateOpened
							v.OpenedAt = time.Now()
							v.PreviewHint = "" 
							store.SaveVault(v)
							
							state.Ritual.IsProcessing = false
							state.Ritual.IsRevealed = true
							
							// 全レイヤーの復号
							layers, _ := store.ListLayers(v.ID)
							fullText := "--- 2026 ORIGINAL ---\n" + string(decrypted)
							for i, l := range layers {
								lData, _ := os.ReadFile(l.CipherPath)
								// 同じパスフレーズを期待 (TODO: レイヤーごとのパス追跡)
								lDec, err := crypto.Decrypt(lData, pass)
								if err == nil {
									fullText += fmt.Sprintf("\n\n--- LAYER %d (%s) ---\n%s", i+1, l.CreatedAt.Format("2006/01/02"), string(lDec))
								}
							}
							state.Ritual.RevealedText = fullText
							state.Ritual.Password.SetText("") 
							
							// 2126 RESONANCE: Update state and rotate
							state.RotateNeural()

							// リストの同期
							state.Vaults, _ = store.ListVaults()
							state.SelectBtns = make([]widget.Clickable, len(state.Vaults))
							w.Invalidate()
						}
					}
				}
			} else {
				// アニメーション継続
				w.Invalidate()
			}
		}
	}
}
