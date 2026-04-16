package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"math"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/f32"

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
	
	// Initialize Face History for Kinetic Echoes
	state.FaceHistory = make([][]f32.Point, 8)
	for i := range state.FaceHistory {
		state.FaceHistory[i] = make([]f32.Point, 4)
	}

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
			// Kinetic resonance: speed up rotation when user is close
			rotSpeed := float32(0.005)
			if state.GazeActive {
				rotSpeed = 0.005 + (state.FaceScale-1.0)*0.01
				if rotSpeed < 0.002 { rotSpeed = 0.002 }
			}
			state.Rotation += rotSpeed

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
	// Singularity interaction (Gaze/Mouse based Zero-UI)
	if state.CurrentScreen == ui.ScreenVaultList && state.IsSingularityFocused {
		// No button needed, just check for Press event in the pointer loop below
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
					v := f32.Pt(xev.Position.X-state.MousePos.X, xev.Position.Y-state.MousePos.Y)
					state.MouseVelocity = f32.Pt(state.MouseVelocity.X*0.7+v.X*0.3, state.MouseVelocity.Y*0.7+v.Y*0.3)
					state.MousePos = xev.Position
				}
				
				if xev.Kind == pointer.Press {
					state.IsGrabbing = true
					state.GrabForce = 1.0
					if state.CurrentScreen == ui.ScreenVaultList {
						if state.IsSingularityFocused {
							state.CurrentScreen = ui.ScreenCompose
							w.Invalidate()
						} else if state.NeuralMemory != nil {
							m := state.NeuralMemory
							// 記憶への同調
							state.Ritual.ActiveMemory = m
							state.CurrentScreen = ui.ScreenRitual
							w.Invalidate()
						}
					}
				}
				
				if xev.Kind == pointer.Release {
					state.IsGrabbing = false
					state.GrabForce = 0
				}
			}
		}
	}
	
	// Map interaction center for physics
	if state.IsGrabbing {
		state.GrabPos = state.MousePos
		state.GrabVelocity = state.MouseVelocity
	} else if state.GazeActive && state.FaceScale > 1.2 {
		// Leaning in triggers camera-based grabbing
		state.IsGrabbing = true
		state.GrabPos = state.GazePos
		state.GrabVelocity = state.GazeVelocity
		state.GrabForce = (state.FaceScale - 1.2) * 2.0
		if state.GrabForce > 1.5 { state.GrabForce = 1.5 }
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

func startWebcamGazeTracking(state *ui.AppState) {
	fmt.Println("Initializing Neural Vision Engine...")

	cascadeFile, err := os.ReadFile(filepath.Join(".", "assets", "models", "facefinder.bin"))
	if err != nil {
		fmt.Printf("GazeTracking Disabled (Cascade Not Found): %v\n", err)
		return
	}

	p := pigo.NewPigo()
	classifier, err := p.Unpack(cascadeFile)
	if err != nil {
		fmt.Printf("GazeTracking Disabled (Cascade Corrupt): %v\n", err)
		return
	}

	// Load Pupil Localization Cascade
	puplocFile, err := os.ReadFile(filepath.Join(".", "assets", "models", "puploc.bin"))
	if err != nil {
		fmt.Printf("FaceMask Refining Disabled (Puploc Not Found): %v\n", err)
	}
	var puplocCascade *pigo.PuplocCascade
	if len(puplocFile) > 0 {
		puplocCascade = pigo.NewPuplocCascade()
		puplocCascade, err = puplocCascade.UnpackCascade(puplocFile)
		if err != nil {
			fmt.Printf("Puploc Unpack Failed: %v\n", err)
			puplocCascade = nil
		}
	}

	fmt.Println("Attempting to connect to Mirror Surface (Webcam)...")

	var stream mediadevices.MediaStream
	var captureErr error

	// Multi-Device Scanning Logic
	devices := mediadevices.EnumerateDevices()
	found := false
	for _, d := range devices {
		if d.Kind != mediadevices.VideoInput { continue }
		fmt.Printf("Testing Device: %s...\n", d.Label)
		stream, captureErr = mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(c *mediadevices.MediaTrackConstraints) {
				// No strict constraints for max compatibility
			}, 
		})
		if captureErr == nil {
			fmt.Printf("Mirror Connection: ESTABLISHED on %s\n", d.Label)
			found = true
			break
		}
		fmt.Printf("Device %s Busy/Failed: %v\n", d.Label, captureErr)
	}

	if !found {
		fmt.Println("No physical Mirror Surface found. WAKING UP GHOST AVATAR (Demo Mode)...")
		runAvatarSimulator(state)
		return
	}

	videoTracks := stream.GetVideoTracks()
	if len(videoTracks) == 0 {
		state.GazeActive = false
		return
	}
	videoTrack := videoTracks[0]
	vTrack, ok := videoTrack.(*mediadevices.VideoTrack)
	if !ok {
		state.GazeActive = false
		return
	}
	defer func() {
		vTrack.Close()
		fmt.Println("Mirror Surface Released.")
	}()

	fmt.Println("Mirror Connection: ESTABLISHED.")
	reader := vTrack.NewReader(false)

	pigoParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
	}

	// Motion Analysis State
	var prevPixels []uint8
	gridRows, gridCols := 32, 40 // Higher resolution
	
	for {
		f, release, err := reader.Read()
		if err != nil {
			state.GazeActive = false
			break
		}

		pixels := pigo.RgbToGrayscale(f)
		rows := f.Bounds().Max.Y
		cols := f.Bounds().Max.X

		// --- NEURAL MOTION ANALYSIS (Tactile Grid with Face Masking) ---
		if prevPixels != nil && len(prevPixels) == len(pixels) {
			state.MotionMu.Lock()
			state.GridActive = true
			
			cellH := rows / gridRows
			cellW := cols / gridCols

			// Pre-calculate face mask bounds in grid coords
			var fR0, fR1, fC0, fC1 int
			if faceSize.X > 0 {
				fR0 = int((faceCenter.Y - faceSize.Y/2.0) / float32(cellH))
				fR1 = int((faceCenter.Y + faceSize.Y/2.0) / float32(cellH))
				fC0 = int((faceCenter.X - faceSize.X/2.0) / float32(cellW))
				fC1 = int((faceCenter.X + faceSize.X/2.0) / float32(cellW))
			}
			
			for r := 0; r < gridRows; r++ {
				for c := 0; c < gridCols; c++ {
					// Apply Face Mask: Skip if cell is inside face region
					if faceSize.X > 0 && r >= fR0 && r <= fR1 && c >= fC0 && c <= fC1 {
						mirroredC := gridCols - 1 - c
						state.MotionGrid[r][mirroredC] = 0
						continue
					}

					diffSum := uint32(0)
					count := uint32(0)
					
					// Sub-sampling for performance
					for y := r * cellH; y < (r+1)*cellH; y += 4 {
						for x := c * cellW; x < (c+1)*cellW; x += 4 {
							idx := y*cols + x
							d := int(pixels[idx]) - int(prevPixels[idx])
							if d < 0 { d = -d }
							if d > 25 { // Sensitivity threshold
								diffSum += uint32(d)
							}
							count++
						}
					}
					
					intensity := float32(diffSum) / float32(count*50.0) 
					if intensity > 1.0 { intensity = 1.0 }
					
					// Mirror horizontally for natural feel
					mirroredC := gridCols - 1 - c
					
					// Temporal smoothing (Persistence)
					oldVal := state.MotionGrid[r][mirroredC]
					if intensity > oldVal {
						state.MotionGrid[r][mirroredC] = intensity
						state.MotionVelocity[r][mirroredC] = f32.Pt(state.GazeVelocity.X*0.2, state.GazeVelocity.Y*0.2)
					} else {
						state.MotionGrid[r][mirroredC] = oldVal * 0.85 
						state.MotionVelocity[r][mirroredC].X *= 0.8
						state.MotionVelocity[r][mirroredC].Y *= 0.8
					}
				}
			}
			state.MotionMu.Unlock()
		}
		
		if len(prevPixels) != len(pixels) {
			prevPixels = make([]uint8, len(pixels))
		}
		copy(prevPixels, pixels)

		pigoParams.ImageParams = pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		}

		results := classifier.RunCascade(pigoParams, 0.0)
		results = classifier.ClusterDetections(results, 0.2)

		var faceCenter f32.Point
		var faceSize f32.Point

		if len(results) > 0 {
			face := results[0]
			if face.Q > 5.0 {
				faceCenter = f32.Pt(float32(face.Col), float32(face.Row))
				faceSize = f32.Pt(float32(face.Scale)*1.2, float32(face.Scale)*1.5) // Mask slightly larger than face
				
				// Base Face Center (Normalized)
				nx := faceCenter.X / float32(cols)
				ny := faceCenter.Y / float32(rows)
				
				targetX := nx * 1000.0 
				targetY := ny * 800.0 
				
				vx := targetX - state.GazePos.X
				vy := targetY - state.GazePos.Y
				
				state.FaceMu.Lock()
				// Shift history (Circular buffer style)
				for i := len(state.FaceHistory) - 1; i > 0; i-- {
					copy(state.FaceHistory[i], state.FaceHistory[i-1])
				}
				copy(state.FaceHistory[0], state.FacePoints)

				state.GazeVelocity.X = state.GazeVelocity.X*0.6 + vx*0.4
				state.GazeVelocity.Y = state.GazeVelocity.Y*0.6 + vy*0.4
				
				// Pulse Detection: Rapid head movement
				speed := float32(math.Sqrt(float64(vx*vx + vy*vy)))
				if speed > 20 { 
					state.PulseStrength = state.PulseStrength*0.8 + (speed/50.0)*0.2
					if state.PulseStrength > 1.5 { state.PulseStrength = 1.5 }
				} else {
					state.PulseStrength *= 0.95
				}

				state.GazePos.X += vx * 0.35 // Snappier tracking
				state.GazePos.Y += vy * 0.35
				
				// --- Z-Axis (Depth) Sensing ---
				rawScale := float32(face.Scale)
				targetFaceScale := rawScale / 250.0 
				if targetFaceScale < 0.5 { targetFaceScale = 0.5 }
				if targetFaceScale > 2.5 { targetFaceScale = 2.5 }
				state.FaceScale = state.FaceScale*0.9 + targetFaceScale*0.1

				state.GazeActive = true

				if len(state.FacePoints) < 4 {
					state.FacePoints = make([]f32.Point, 4)
				}
				state.FaceMu.Unlock()

				fScale := float32(face.Scale)
				
				state.FaceMu.Lock()
				// 1 & 2: Eyes 
				if puplocCascade != nil {
					puplocBase := pigo.Puploc{
						Row:   face.Row,
						Col:   face.Col,
						Scale: float32(face.Scale),
					}
					lp := puplocCascade.RunDetector(puplocBase, pigoParams.ImageParams, 0.0, false)
					if lp != nil {
						state.FacePoints[0] = f32.Pt(float32(lp.Col)/float32(cols)*1000.0, float32(lp.Row)/float32(rows)*800.0)
					} else {
						state.FacePoints[0] = f32.Pt(targetX-fScale*0.22, targetY-fScale*0.15)
					}
					rp := puplocCascade.RunDetector(puplocBase, pigoParams.ImageParams, 0.0, true)
					if rp != nil {
						state.FacePoints[1] = f32.Pt(float32(rp.Col)/float32(cols)*1000.0, float32(rp.Row)/float32(rows)*800.0)
					} else {
						state.FacePoints[1] = f32.Pt(targetX+fScale*0.22, targetY-fScale*0.15)
					}
				} else {
					state.FacePoints[0] = f32.Pt(targetX-fScale*0.22, targetY-fScale*0.15)
					state.FacePoints[1] = f32.Pt(targetX+fScale*0.22, targetY-fScale*0.15)
				}

				// 3: Nose (Center of face usually)
				state.FacePoints[2] = f32.Pt(targetX, targetY+fScale*0.05)

				// 4: Mouth (Lower part)
				state.FacePoints[3] = f32.Pt(targetX, targetY+fScale*0.3)
				state.FaceMu.Unlock()

			} else {
				state.GazeActive = false
			}
		} else {
			state.GazeActive = false
		}
		release()
		
		// Prevent CPU hogging
		time.Sleep(30 * time.Millisecond)
	}
}

func runAvatarSimulator(state *ui.AppState) {
	state.FaceMu.Lock()
	state.GazeActive = true
	state.FaceScale = 1.0
	state.FaceMu.Unlock()

	ticker := time.NewTicker(16 * time.Millisecond)
	angle := 0.0
	for range ticker.C {
		angle += 0.05
		// Simulate a face floating in a slow infinity pattern
		centerX := float32(500 + math.Cos(angle*0.7)*150)
		centerY := float32(400 + math.Sin(angle*1.3)*100)
		
		fScale := float32(200 + math.Sin(angle*0.5)*50)

		state.FaceMu.Lock()
		state.FaceScale = fScale / 250.0
		state.GazePos = f32.Pt(centerX, centerY)
		
		if len(state.FacePoints) < 4 {
			state.FacePoints = make([]f32.Point, 4)
		}
		
		// Simulate Landmarks
		state.FacePoints[0] = f32.Pt(centerX-fScale*0.22, centerY-fScale*0.15) // Eye L
		state.FacePoints[1] = f32.Pt(centerX+fScale*0.22, centerY-fScale*0.15) // Eye R
		state.FacePoints[2] = f32.Pt(centerX, centerY+fScale*0.05)             // Nose
		state.FacePoints[3] = f32.Pt(centerX, centerY+fScale*0.3)              // Mouth
		
		// Pulse simulation
		state.PulseStrength = float32(math.Max(0, math.Sin(angle*2.0))) * 0.5
		
		state.GazeActive = true
		state.FaceMu.Unlock()
	}
}
