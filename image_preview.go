package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	ansisixel "github.com/charmbracelet/x/ansi/sixel"
	"github.com/nfnt/resize"
)

type terminalImageProtocol string

const (
	terminalImageAuto  terminalImageProtocol = "auto"
	terminalImageKitty terminalImageProtocol = "kitty"
	terminalImageSixel terminalImageProtocol = "sixel"
	terminalImageNone  terminalImageProtocol = "none"
)

type terminalImageCacheKey struct {
	Path                string
	X, Y, Columns, Rows int
	Protocol            terminalImageProtocol
}

var terminalImageSequenceCache sync.Map

func parseTerminalImageProtocol(value string) terminalImageProtocol {
	switch terminalImageProtocol(strings.ToLower(strings.TrimSpace(value))) {
	case terminalImageKitty:
		return terminalImageKitty
	case terminalImageSixel:
		return terminalImageSixel
	case terminalImageNone:
		return terminalImageNone
	default:
		return terminalImageAuto
	}
}

func selectTerminalImageProtocol(configured, override, insideEmacs, termProgram, term string) terminalImageProtocol {
	if selected := parseTerminalImageProtocol(override); selected != terminalImageAuto {
		return selected
	}
	if selected := parseTerminalImageProtocol(configured); selected != terminalImageAuto {
		return selected
	}
	insideEmacs = strings.ToLower(insideEmacs)
	termProgram = strings.ToLower(termProgram)
	term = strings.ToLower(term)
	if strings.Contains(insideEmacs, ",eat") || strings.HasPrefix(term, "eat-") || strings.Contains(term, "sixel") {
		return terminalImageSixel
	}
	// iTerm 3.6+ and Kitty both understand the Kitty graphics protocol. Keep
	// Kitty as the compatibility default for terminals that do not identify
	// themselves, matching the behavior before protocol auto-detection.
	if strings.Contains(termProgram, "iterm") || strings.Contains(term, "kitty") {
		return terminalImageKitty
	}
	return terminalImageKitty
}

func resolveTerminalImageProtocol() terminalImageProtocol {
	configured := "auto"
	if cfg := LoadConfig(); cfg != nil && cfg.TerminalImageProtocol != nil {
		configured = *cfg.TerminalImageProtocol
	}
	return selectTerminalImageProtocol(
		configured,
		os.Getenv("TEAMS_TUI_GO_IMAGE_PROTOCOL"),
		os.Getenv("INSIDE_EMACS"),
		os.Getenv("TERM_PROGRAM"),
		os.Getenv("TERM"),
	)
}

// isImageAttachment checks if the attachment is an image based on ContentType or file extension.
func isImageAttachment(att MessageAttachment) bool {
	if att.ContentType != nil && strings.HasPrefix(strings.ToLower(*att.ContentType), "image/") {
		return true
	}
	if att.Name != nil {
		ext := strings.ToLower(filepath.Ext(*att.Name))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
			return true
		}
	}
	if att.ContentURL != nil {
		ext := strings.ToLower(filepath.Ext(*att.ContentURL))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
			return true
		}
	}
	return false
}

// getAttachmentCachePath returns the local cached path for an attachment.
func getAttachmentCachePath(att MessageAttachment) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	previewsDir := filepath.Join(cacheDir, "previews")
	if err := os.MkdirAll(previewsDir, 0o700); err != nil {
		return "", err
	}

	var urlStr string
	if att.ContentURL != nil {
		urlStr = *att.ContentURL
	} else if att.Content != nil {
		urlStr = *att.Content
	} else {
		urlStr = att.ID
	}

	hash := sha256.Sum256([]byte(urlStr))
	hashStr := hex.EncodeToString(hash[:])

	ext := ".png"
	if att.Name != nil {
		if e := filepath.Ext(*att.Name); e != "" {
			ext = e
		}
	}

	return filepath.Join(previewsDir, hashStr+ext), nil
}

// MsgPreviewDownloaded is sent when a background preview image download completes.
type MsgPreviewDownloaded struct {
	DestPath string
	Err      error
}

// downloadPreviewCmd downloads a file attachment to cache silently.
func downloadPreviewCmd(clientID, fileURL, destPath string) tea.Cmd {
	return func() tea.Msg {
		token, err := GetValidTokenSilent(clientID)
		if err != nil {
			return MsgPreviewDownloaded{Err: err}
		}
		err = DownloadFile(token, fileURL, destPath)
		return MsgPreviewDownloaded{DestPath: destPath, Err: err}
	}
}

func prepareTerminalImage(filePath string, x, y, cols, rows int) (image.Image, int, int, int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	img, _, err := image.Decode(file)
	file.Close()
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}

	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// 1. Get exact cell size in pixels
	cellW, cellH := getCellSize()

	// 2. Compute maximum available pixel dimensions
	maxPixelW := cols * cellW
	maxPixelH := rows * cellH

	// 3. Scale original dimensions to fit within maximum pixel bounds
	scaleX := float64(maxPixelW) / float64(origW)
	scaleY := float64(maxPixelH) / float64(origH)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	newPixelW := int(float64(origW) * scale)
	newPixelH := int(float64(origH) * scale)
	if newPixelW < 1 {
		newPixelW = 1
	}
	if newPixelH < 1 {
		newPixelH = 1
	}

	// 4. Calculate cell columns and rows occupied by the new pixel dimensions
	c := int(float64(newPixelW)/float64(cellW) + 0.5)
	r := int(float64(newPixelH)/float64(cellH) + 0.5)
	if c < 1 {
		c = 1
	}
	if r < 1 {
		r = 1
	}
	if c > cols {
		c = cols
	}
	if r > rows {
		r = rows
	}

	// 5. Center the image inside the border box
	padX := (cols - c) / 2
	padY := (rows - r) / 2
	targetX := x + padX
	targetY := y + padY

	// 6. Resample the image client-side to the exact target pixels using high-quality Lanczos3
	resizedImg := resize.Resize(uint(newPixelW), uint(newPixelH), img, resize.Lanczos3)
	return resizedImg, targetX, targetY, c, r, nil
}

// kittyImageSequence generates the escape sequence to draw a centered image
// using the Kitty Graphics Protocol.
func kittyImageSequence(filePath string, x, y, cols, rows int) string {
	resizedImg, targetX, targetY, c, r, err := prepareTerminalImage(filePath, x, y, cols, rows)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, resizedImg); err != nil {
		return ""
	}
	pngBytes := buf.Bytes()
	encoded := base64.StdEncoding.EncodeToString(pngBytes)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\x1b[s\x1b[%d;%dH", targetY+1, targetX+1))

	const chunkSize = 4096
	totalLen := len(encoded)

	for i := 0; i < totalLen; i += chunkSize {
		end := i + chunkSize
		mVal := 1
		if end >= totalLen {
			end = totalLen
			mVal = 0
		}

		chunk := encoded[i:end]
		if i == 0 {
			sb.WriteString(fmt.Sprintf("\x1b_Ga=T,f=100,m=%d,c=%d,r=%d;%s\x1b\\", mVal, c, r, chunk))
		} else {
			sb.WriteString(fmt.Sprintf("\x1b_Gm=%d;%s\x1b\\", mVal, chunk))
		}
	}

	sb.WriteString("\x1b[u")
	return sb.String()
}

// sixelImageSequence generates a centered Sixel image for terminal emulators
// such as Emacs EAT that do not implement Kitty image placement.
func sixelImageSequence(filePath string, x, y, cols, rows int) string {
	resizedImg, targetX, targetY, _, _, err := prepareTerminalImage(filePath, x, y, cols, rows)
	if err != nil {
		return ""
	}
	var payload bytes.Buffer
	encoder := &ansisixel.Encoder{}
	if err := encoder.Encode(&payload, resizedImg); err != nil {
		return ""
	}
	return fmt.Sprintf(
		"\x1b[s\x1b[%d;%dH%s\x1b[u",
		targetY+1,
		targetX+1,
		ansi.SixelGraphics(0, 1, 0, payload.Bytes()),
	)
}

func terminalImageSequence(filePath string, x, y, cols, rows int) string {
	protocol := resolveTerminalImageProtocol()
	if protocol == terminalImageNone {
		return ""
	}
	key := terminalImageCacheKey{
		Path: filePath, X: x, Y: y, Columns: cols, Rows: rows, Protocol: protocol,
	}
	if cached, ok := terminalImageSequenceCache.Load(key); ok {
		return cached.(string)
	}
	var sequence string
	switch protocol {
	case terminalImageSixel:
		sequence = sixelImageSequence(filePath, x, y, cols, rows)
	default:
		sequence = kittyImageSequence(filePath, x, y, cols, rows)
	}
	if sequence != "" {
		terminalImageSequenceCache.Store(key, sequence)
	}
	return sequence
}

// clearTerminalImagesCmd clears Kitty placements. Sixel images are ordinary
// terminal cells and disappear when Bubble Tea redraws the underlying view.
func clearTerminalImagesCmd() tea.Cmd {
	return func() tea.Msg {
		if resolveTerminalImageProtocol() == terminalImageKitty {
			_, _ = os.Stdout.Write([]byte("\x1b_Ga=d,d=a\x1b\\"))
		}
		return nil
	}
}

// previewImage is used by the CLI subcommand "preview-image"
func previewImage(path string) {
	protocol := resolveTerminalImageProtocol()
	seq := terminalImageSequence(path, 0, 0, 80, 24)
	if seq != "" {
		if protocol == terminalImageKitty {
			fmt.Print("\x1b_Ga=d,d=a\x1b\\")
		}
		fmt.Printf("%s\n", seq)
	}
	fmt.Printf("Image preview loaded (%s). Press Enter to exit...\n", protocol)
	_, _ = os.Stdin.Read(make([]byte, 1))
}
