package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	if len(vaults) == 0 {
		// Mock data for first run
		v1 := &vault.Vault{
			ID: "v1", Title: "Memory of Summer 2024", State: vault.StateSealed,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UnlockAt:  time.Now().Add(10 * time.Second), // For demo, make it quick
		}
		store.SaveVault(v1)
		vaults, _ = store.ListVaults()
	}

	// UI State
	fontPath := filepath.Join(".", "assets", "fonts", "font.ttf")
	th := ui.NewVaultTheme(fontPath)
	state := &ui.AppState{
		Theme:      th,
		Vaults:     vaults,
		SelectBtns: make([]widget.Clickable, len(vaults)),
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
			state.CurrentScreen = ui.ScreenVaultList
			w.Invalidate()
		}
		if state.Compose.SealBtn.Clicked(gtx) {
			// Ritual is real! Let's seal it.
			title := state.Compose.Title.Text()
			body := state.Compose.Body.Text()
			pass := state.Compose.Passphrase.Text()
			if title != "" && pass != "" {
				vid := fmt.Sprintf("v%d", time.Now().Unix())
				cipherPath := filepath.Join("vaults", vid+".age")
				
				// Encrypt
				ciphertext, err := crypto.Encrypt([]byte(body), pass)
				if err == nil {
					os.WriteFile(cipherPath, ciphertext, 0600)
					
					v := &vault.Vault{
						ID:                vid,
						Title:             title,
						State:             vault.StateSealed,
						CreatedAt:         time.Now(),
						UnlockAt:          time.Now().Add(10 * time.Second), // Demo 10s
						CipherPath:        cipherPath,
						RequirePassphrase: true,
						PassphraseHash:    pass, // For MVP, check plain; should be hash in real
					}
					store.SaveVault(v)
					// Refresh
					state.Vaults, _ = store.ListVaults()
					state.SelectBtns = make([]widget.Clickable, len(state.Vaults))
					state.CurrentScreen = ui.ScreenVaultList
					w.Invalidate()
				}
			}
		}
	}

	// Ritual Screen Logic
	if state.CurrentScreen == ui.ScreenRitual {
		if state.Ritual.CancelBtn.Clicked(gtx) {
			state.Ritual.IsProcessing = false
			state.CurrentScreen = ui.ScreenVaultList
			w.Invalidate()
		}
		if state.Ritual.UnlockBtn.Clicked(gtx) && !state.Ritual.IsProcessing {
			state.Ritual.IsProcessing = true
			state.Ritual.ProcessingSince = time.Now()
			w.Invalidate()
		}

		if state.Ritual.IsProcessing {
			if time.Since(state.Ritual.ProcessingSince) > 2*time.Second {
				v := state.Ritual.ActiveVault
				pass := state.Ritual.Password.Text()
				if time.Now().After(v.UnlockAt) && pass == v.PassphraseHash {
					// Decrypt to verify
					cipher, err := os.ReadFile(v.CipherPath)
					if err == nil {
						decrypted, err := crypto.Decrypt(cipher, pass)
						if err == nil {
							// SUCCESS
							v.State = vault.StateOpened
							v.OpenedAt = time.Now()
							v.PreviewHint = string(decrypted)
							store.SaveVault(v)
							
							state.Vaults, _ = store.ListVaults()
							state.SelectBtns = make([]widget.Clickable, len(state.Vaults))
							state.Ritual.IsProcessing = false
							state.Ritual.Password.SetText("") // Clear
							state.CurrentScreen = ui.ScreenVaultList
							w.Invalidate()
						}
					}
				} else {
					// Failure or just condition not met
					state.Ritual.IsProcessing = false
					w.Invalidate()
				}
			} else {
				// Keep invalidating to simulate animation frame
				w.Invalidate()
			}
		}
	}
}
