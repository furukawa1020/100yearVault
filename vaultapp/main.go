package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/unit"

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

	// Seed dummy data for mockup if empty
	vaults, err := store.ListVaults()
	if err != nil {
		return err
	}
	if len(vaults) == 0 {
		mockVaults := []*vault.Vault{
			{
				ID: "v1", Title: "Memory of Summer 2024", State: vault.StateSealed,
				CreatedAt: time.Now().Add(-24 * time.Hour), 
				UnlockAt: time.Now().Add(100 * 365 * 24 * time.Hour), // 100 years
			},
			{
				ID: "v2", Title: "The Resolve of Start-Up", State: vault.StateOpened,
				CreatedAt: time.Now().Add(-48 * time.Hour), 
				OpenedAt: time.Now().Add(-1 * time.Hour),
			},
		}
		for _, v := range mockVaults {
			store.SaveVault(v)
		}
		vaults = mockVaults
	}

	// UI State
	fontPath := filepath.Join(".", "assets", "fonts", "font.ttf")
	th := ui.NewVaultTheme(fontPath)
	state := &ui.AppState{
		Theme: th,
		Vaults: vaults,
	}

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			
			// Main Layout
			state.LayoutList(gtx)
			
			e.Frame(gtx.Ops)
		}
	}
}
