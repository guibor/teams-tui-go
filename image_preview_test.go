package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectTerminalImageProtocol(t *testing.T) {
	tests := []struct {
		name        string
		configured  string
		override    string
		insideEmacs string
		termProgram string
		term        string
		want        terminalImageProtocol
	}{
		{name: "EAT auto-detection", configured: "auto", insideEmacs: "30.2,eat", term: "eat-truecolor", want: terminalImageSixel},
		{name: "iTerm auto-detection", configured: "auto", termProgram: "iTerm.app", term: "xterm-256color", want: terminalImageKitty},
		{name: "explicit config", configured: "sixel", termProgram: "iTerm.app", want: terminalImageSixel},
		{name: "environment override", configured: "sixel", override: "kitty", insideEmacs: "30.2,eat", want: terminalImageKitty},
		{name: "disabled", configured: "none", want: terminalImageNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := selectTerminalImageProtocol(
				test.configured,
				test.override,
				test.insideEmacs,
				test.termProgram,
				test.term,
			)
			if got != test.want {
				t.Fatalf("protocol = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSixelImageSequence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preview.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(40 * x), G: uint8(40 * y), B: 180, A: 255})
		}
	}
	if err := png.Encode(file, img); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	sequence := sixelImageSequence(path, 2, 3, 10, 8)
	if !strings.Contains(sequence, "\x1bP0;1q") {
		t.Fatalf("sequence does not contain a Sixel DCS: %q", sequence)
	}
	if !strings.HasSuffix(sequence, "\x1b\\\x1b[u") {
		t.Fatalf("sequence does not terminate Sixel and restore the cursor")
	}
}
