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
		Theme:            th,
		Vaults:           vaults,
		SelectBtns:       make([]widget.Clickable, len(vaults)),
		ConnectionStatus: "CONNECTION TO 2026: STABLE [99.9%]",
	}
	state.Compose.UnlockDays.SetText("36500")

	// 2126 RESONANCE: Pick a random fragment from opened vaults
	for _, v := range vaults {
		if v.State == vault.StateOpened && v.PreviewHint != "" {
			state.DailyFragment = v.PreviewHint
			break // For now, just the first one found
		}
	}
	if state.DailyFragment == "" {
		state.DailyFragment = "接続待機中... 最初の記憶を封じなさい。"
	}

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Logic Handling
			updateLogic(gtx, state, store, w)

			// Main Layout Based on Screen
			switch state.CurrentScreen {
			case ui.ScreenVaultList:
				state.LayoutList(gtx)
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

	// List Screen Logic
	if state.CurrentScreen == ui.ScreenVaultList {
		for i := range state.SelectBtns {
			if state.SelectBtns[i].Clicked(gtx) {
				v := state.Vaults[i]
				state.Ritual.ActiveVault = v
				state.CurrentScreen = ui.ScreenRitual
				w.Invalidate()
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
							// Layerの追加
							l := &vault.Layer{
								ID:         layerID,
								ParentID:   vid,
								CipherPath: cipherPath,
								CreatedAt:  time.Now(),
							}
							store.SaveLayer(l)
						} else {
							// 新規Vaultの作成 (時空への放流)
							days, _ := strconv.ParseFloat(daysInput, 64)
							if days <= 0 {
								days = 36500 
							}

							// 2126年 QSP: 到着日の方流化 (ランダム性)
							// 1分後から指定日数の間でランダムに漂着
							maxSeconds := days * 24 * 60 * 60
							randomSeconds := float64(time.Now().UnixNano()%int64(maxSeconds))
							// デモ用に短縮（最大30分など）する場合:
							// randomSeconds = float64(time.Now().UnixNano() % (30 * 60))
							
							unlockAt := time.Now().Add(time.Duration(randomSeconds * float64(time.Second)))

							v := &vault.Vault{
								ID:                vid,
								Title:             title,
								State:             vault.StateSealed,
								CreatedAt:         time.Now(),
								UnlockAt:          unlockAt,
								CipherPath:        cipherPath,
								RequirePassphrase: true,
								PreviewHint:       body, // 後のキメラ合成用にヒントとして保持
							}
							store.SaveVault(v)
						}
						
						// Success Cleanup
						state.Compose.Title.SetText("")
						state.Compose.Body.SetText("")
						state.Compose.Passphrase.SetText("")
						state.Compose.ErrorMessage = ""
						state.Compose.AddLayerMode = false
						state.Compose.TargetVault = nil
						
						// Refresh List
						state.Vaults, _ = store.ListVaults()
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
							
							// 2126 RESONANCE: Update dashboard fragment
							state.DailyFragment = v.PreviewHint

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
