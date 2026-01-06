package icon

import (
	"fmt"
	"image"
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
	// Smaller size: 8 chars wide = 16 pixels, 6 chars tall = 24 pixels
	iconCharWidth  = 8
	iconCharHeight = 6
	iconWidth      = iconCharWidth * 2  // 16 pixels
	iconHeight     = iconCharHeight * 4 // 24 pixels
)

// FetchAndRender downloads a favicon and renders it as a colored bar
func FetchAndRender(url string) (string, error) {
	if url == "" {
		return renderPlaceholder(), nil
	}

	// Download image
	img, err := downloadImage(url)
	if err != nil {
		return renderPlaceholder(), nil
	}

	// Extract dominant color and render as a simple colored bar
	return renderColorBar(img), nil
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

// renderPlaceholder renders a default colored bar when no favicon is available
func renderPlaceholder() string {
	// Default color - nice green
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#86efac")).
		Bold(true)

	var result strings.Builder
	for i := 0; i < 6; i++ {
		result.WriteString(style.Render(" ███ "))
		result.WriteString("\n")
	}

	return result.String()
}

// renderColorBar extracts dominant color and renders a simple colored bar
func renderColorBar(img image.Image) string {
	// Get dominant color
	r, g, b := getDominantColor(img)

	// Create a 3-character wide colored bar, 6 lines tall
	colorHex := fmt.Sprintf("#%02x%02x%02x", r, g, b)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorHex)).
		Bold(true)

	var result strings.Builder
	for i := 0; i < 6; i++ {
		result.WriteString(style.Render(" ███ "))
		result.WriteString("\n")
	}

	return result.String()
}

// getDominantColor extracts the dominant color from an image
func getDominantColor(img image.Image) (uint8, uint8, uint8) {
	bounds := img.Bounds()

	// Sample multiple points and average
	var totalR, totalG, totalB uint64
	var count uint64

	// Sample in a grid pattern
	for y := bounds.Min.Y; y < bounds.Max.Y; y += bounds.Dy() / 10 {
		for x := bounds.Min.X; x < bounds.Max.X; x += bounds.Dx() / 10 {
			r, g, b, a := img.At(x, y).RGBA()

			// Skip transparent pixels
			if a < 32768 {
				continue
			}

			totalR += uint64(r >> 8)
			totalG += uint64(g >> 8)
			totalB += uint64(b >> 8)
			count++
		}
	}

	if count == 0 {
		// Fallback color
		return 134, 239, 172 // Nice green
	}

	return uint8(totalR / count), uint8(totalG / count), uint8(totalB / count)
}
