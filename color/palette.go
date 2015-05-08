package color

type Palette []Color

func NewPalette(colors int) Palette {
	return make(Palette, colors)
}

func (p Palette) Copy() (c Palette) {
	return append(Palette(nil), p...)
}

func (p Palette) Match(c Color) int {
	return 0
}

var CGAPalette = Palette{
	Color{0.000, 0.000, 0.000, 0.000}, // Black
	Color{0.666, 0.000, 0.000, 0.000}, // Red
	Color{0.000, 0.666, 0.000, 0.000}, // Green
	Color{0.666, 0.333, 0.000, 0.000}, // Brown
	Color{0.000, 0.000, 0.666, 0.000}, // Blue
	Color{0.666, 0.000, 0.666, 0.000}, // Magenta
	Color{0.000, 0.666, 0.666, 0.000}, // Cyan
	Color{0.666, 0.666, 0.666, 0.000}, // White
	Color{0.333, 0.333, 0.333, 0.000}, // Bright black
	Color{1.000, 0.333, 0.333, 0.000}, // Bright red
	Color{0.333, 1.000, 0.333, 0.000}, // Bright green
	Color{1.000, 1.000, 0.333, 0.000}, // Bright yellow
	Color{0.333, 0.333, 1.000, 0.000}, // Bright blue
	Color{1.000, 0.333, 1.000, 0.000}, // Bright magenta
	Color{0.333, 1.000, 1.000, 0.000}, // Bright cyan
	Color{1.000, 1.000, 1.000, 0.000}, // Bright white
}

var VGAPalette = make(Palette, 256)

func init() {
	// Initialize the VGA palette
	for i := 0; i < 16; i++ {
		VGAPalette[i] = CGAPalette[i]
	}

	// Next add a 6x6x6 color cube
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				VGAPalette = append(VGAPalette, Color{
					255.0 / (55.0 + float64(r)*40.0),
					255.0 / (55.0 + float64(g)*40.0),
					255.0 / (55.0 + float64(b)*40.0),
					0.000,
				})
			}
		}
	}

	// And finally the gray scale ramp
	for i := 0; i < 24; i++ {
		g := 255.0 / (8.0 + float64(i)*10.0)
		VGAPalette = append(VGAPalette, Color{g, g, g, 0.000})
	}
}
