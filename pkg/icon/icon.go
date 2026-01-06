package icon

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nfnt/resize"
)

const (
	// With Braille: each character = 2x4 pixels
	// So 12 chars wide = 24 pixels, 8 chars tall = 32 pixels
	iconCharWidth  = 12
	iconCharHeight = 8
	iconWidth      = iconCharWidth * 2  // 24 pixels
	iconHeight     = iconCharHeight * 4 // 32 pixels
)

// FetchAndRender downloads a favicon and renders it as terminal art
func FetchAndRender(url string) (string, error) {
	if url == "" {
		return renderPlaceholder(), nil
	}

	// Download image
	img, err := downloadImage(url)
	if err != nil {
		return renderPlaceholder(), nil
	}

	// Render as terminal art
	return renderImage(img), nil
}

// downloadImage downloads an image from a URL
func downloadImage(url string) (image.Image, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		// Try reading body to see if it's too large
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if len(body) == 0 {
			return nil, fmt.Errorf("empty response")
		}
		return nil, err
	}

	return img, nil
}

// renderImage converts an image to Braille dot patterns for high-res rendering
// Each Braille character represents 2x4 pixels, giving 4x the resolution
func renderImage(img image.Image) string {
	// Resize to target dimensions
	resized := resize.Resize(iconWidth, iconHeight, img, resize.Lanczos3)

	var result strings.Builder

	// Braille Unicode offset
	const brailleOffset = 0x2800

	// Braille dot positions (in 2x4 grid):
	// 0 3
	// 1 4
	// 2 5
	// 6 7
	dotValues := []int{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}

	// Process 2x4 pixel blocks
	for charY := 0; charY < iconCharHeight; charY++ {
		for charX := 0; charX < iconCharWidth; charX++ {
			// Calculate Braille character for this 2x4 block
			brailleValue := 0
			var avgR, avgG, avgB uint32
			pixelCount := 0

			// Check each of the 8 dots in the Braille character
			for dotY := 0; dotY < 4; dotY++ {
				for dotX := 0; dotX < 2; dotX++ {
					pixelX := charX*2 + dotX
					pixelY := charY*4 + dotY

					if pixelX >= iconWidth || pixelY >= iconHeight {
						continue
					}

					r, g, b, a := resized.At(pixelX, pixelY).RGBA()

					// Accumulate colors for averaging
					avgR += r
					avgG += g
					avgB += b
					pixelCount++

					// Check if pixel is visible (not transparent)
					if a > 32768 {
						// Determine which dot this is (0-7)
						dotIndex := dotY*2 + dotX
						if dotIndex < 6 {
							brailleValue |= dotValues[dotIndex]
						} else {
							// Dots 6 and 7 are at the bottom
							brailleValue |= dotValues[6+dotX]
						}
					}
				}
			}

			// Calculate average color for this character
			if pixelCount > 0 {
				avgR /= uint32(pixelCount)
				avgG /= uint32(pixelCount)
				avgB /= uint32(pixelCount)

				r8 := uint8(avgR >> 8)
				g8 := uint8(avgG >> 8)
				b8 := uint8(avgB >> 8)

				// Create colored Braille character
				brailleChar := rune(brailleOffset + brailleValue)
				style := lipgloss.NewStyle().
					Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r8, g8, b8)))

				result.WriteString(style.Render(string(brailleChar)))
			} else {
				result.WriteString(" ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// renderPlaceholder renders a default icon when no favicon is available
func renderPlaceholder() string {
	// Create a simple radio icon
	lines := []string{
		"   ╭────────╮   ",
		"   │  ◉  ◉ │   ",
		"   │ ┌────┐│   ",
		"   │ │━━━━││   ",
		"   │ └────┘│   ",
		"   │▓▓▓▓▓▓▓│   ",
		"   ╰────────╯   ",
		"      📻        ",
	}

	var result strings.Builder
	for _, line := range lines {
		result.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Render(line))
		result.WriteString("\n")
	}

	return result.String()
}

// GetDominantColor returns the dominant color from an image
func GetDominantColor(img image.Image) color.Color {
	// Simple implementation: sample the center pixel
	bounds := img.Bounds()
	centerX := bounds.Min.X + (bounds.Max.X-bounds.Min.X)/2
	centerY := bounds.Min.Y + (bounds.Max.Y-bounds.Min.Y)/2
	return img.At(centerX, centerY)
}
