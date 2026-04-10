package main

import (
	"fmt"
	"time"
	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
)

func main() {
	fmt.Println("WAKING UP MIRROR SENSORS...")
	devices := mediadevices.EnumerateDevices()
	for i, d := range devices {
		fmt.Printf("[%d] Device: %s (Kind: %s) [ID: %s]\n", i, d.Label, d.Kind, d.DeviceID)
		
		if d.Kind == mediadevices.VideoInput {
			fmt.Printf("   -> Attempting to open %s...\n", d.Label)
			stream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
				Video: func(c *mediadevices.MediaTrackConstraints) {
					c.DeviceID = mediadevices.StringConstraint(d.DeviceID)
				},
			})
			if err != nil {
				fmt.Printf("   !! FAILED: %v\n", err)
			} else {
				fmt.Printf("   ++ SUCCESS: Opened %s\n", d.Label)
				for _, t := range stream.GetTracks() {
					t.Close()
				}
			}
		}
	}
	fmt.Println("DIAGNOSTIC COMPLETE.")
}
