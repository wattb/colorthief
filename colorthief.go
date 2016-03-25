package colorthief

import (
	"fmt"
	"github.com/wattb/imt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

type HexColor string

type Params struct {
	Img     image.Image
	Quality int
	Count   int
}

// DominantColor gets the single most dominant color in the image.
// The Params.Count field is not used here.
func DominantColor(p Params) HexColor {
	return HexColor("FFFFFF")
}

// Palette gets Params.Count most frequent colors in the image.
func Palette(p Params) []HexColor {
	i := imt.UnpackImage(p.Img)
	pixel_count := i.Width * i.Height
	var validPixels []imt.Color

	for x := 0; x < pixel_count; x += p.Quality {
		r, g, b, a := i.Pixels[x].RGBA()
		// If pixel is mostly opaque and not white
		if a > 125 {
			if !(r > 250 && g > 250 && b > 250) {
				validPixels = append(validPixels, imt.Color{R: r, G: g, B: b, A: a})
			}
		}
	}

	return nil
}

func main() {
	path := os.Args[1]
	f, _ := os.Open(path)
	i, _, _ := image.Decode(f)
	fmt.Println(i.Bounds())
}
