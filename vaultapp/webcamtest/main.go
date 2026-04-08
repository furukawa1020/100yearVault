package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"

	pigo "github.com/esimov/pigo/core"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"

	// Note: You may need a specific driver for Windows.
	// This is the generic camera driver.
	_ "github.com/pion/mediadevices/pkg/driver/camera"
)

func main() {
	fmt.Println("EYE_OF_THE_COSMOS: RESEARCH_PROTOTYPE_v0.1")
	fmt.Println("Initializing Neural Vision Engine (pigo)...")

	// 1. Load Pigo Cascade
	cascadeFile, err := ioutil.ReadFile("facefinder.bin")
	if err != nil {
		log.Fatalf("Error reading cascade file: %v", err)
	}

	p := pigo.NewPigo()
	classifier, err := p.Unpack(cascadeFile)
	if err != nil {
		log.Fatalf("Error unpacking cascade: %v", err)
	}

	fmt.Println("Neural Vision Engine: READY.")
	fmt.Println("Attempting to connect to Mirror Surface (Webcam)...")

	// 2. Setup Webcam
	stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
			c.FrameFormat = prop.FrameFormat(frame.FormatI420) // Common internal format
		},
	})
	if err != nil {
		log.Fatalf("Mirror Connection Failed: %v\n(Check if webcam is connected and not used by another app)", err)
	}

	videoTrack := stream.GetVideoTracks()[0]
	// Type assertion to access NewReader on VideoTrack
	vTrack, ok := videoTrack.(*mediadevices.VideoTrack)
	if !ok {
		log.Fatalf("Track is not a VideoTrack")
	}
	defer vTrack.Close()

	fmt.Println("Mirror Connection: ESTABLISHED.")
	fmt.Println("Starting Real-time Gaze Extraction (Press Ctrl+C to stop)...")

	// 3. Start Reader
	reader := vTrack.NewReader(false)
	defer reader.Close()
	
	// Pigo params
	pigoParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
	}

	for {
		f, release, err := reader.Read()
		if err != nil {
			fmt.Printf("Visual Data Interrupted: %v\n", err)
			break
		}

		// Convert frame to image.Image
		// pion/mediadevices usually provides image.YCbCr or image.RGBA
		img := f

		// Prepare image for pigo (grayscale)
		var pixels []uint8
		var rows, cols int

		// mediadevices frame usually implements image.Image
		// For I420, we can extract the Y plane (grayscale)
		if ycc, ok := img.(*image.YCbCr); ok {
			pixels = ycc.Y
			rows = ycc.Rect.Dy()
			cols = ycc.Rect.Dx()
		} else {
			// Fallback: convert to grayscale
			pixels = pigo.RgbToGrayscale(img)
			rows = img.Bounds().Max.Y
			cols = img.Bounds().Max.X
		}

		// Run Face Detection
		pigoParams.ImageParams = pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		}
		
		results := classifier.RunCascade(pigoParams, 0.0)
		results = classifier.ClusterDetections(results, 0.2)

		if len(results) > 0 {
			// Pick the most prominent face
			face := results[0]
			if face.Q > 5.0 { // Quality threshold
				// Face Center
				fx := float32(face.Col)
				fy := float32(face.Row)
				
				// Normalize Coordinates (0.0 to 1.0)
				nx := fx / 640.0
				ny := fy / 480.0

				fmt.Printf("\rNEURAL_RESONANCE_DETECTED: [X: %.2f, Y: %.2f] Intensity: %.1f", nx, ny, face.Q)
			}
		} else {
			fmt.Print("\rNEURAL_RESONANCE_LOST: Searching for user...            ")
		}

		release()
	}
}
