package colorthief

import (
	"github.com/wattb/imt"
)

const sigbits = 5
const rshift = 8 - sigbits
const max_iteration = 1000
const fract_by_populations = 0.75

func getColorIndex(r, g, b int) int {
	return (r << 2 * sigbits) + (g << sigbits) + b
}

// getHistogram outputs a map with the number of pixels in each quantized region of colour space.
func getHistogram(pixels []imt.Color) map[int]int {
	histo := make(map[int]int)
	for _, p := range pixels {
		r, g, b, _ := p.RGBA()
		rval := int(r >> rshift)
		gval := int(g >> rshift)
		bval := int(b >> rshift)

		v := getColorIndex(rval, gval, bval)
		if _, ok := histo[v]; ok {
			histo[v] += 1
		} else {
			histo[v] = 0
		}
	}

	return histo
}

func min(nums ...int) int {
	min := 0
	for _, num := range nums {
		if num <= min {
			min = num
		}
	}
	return min
}

func max(nums ...int) int {
	max := 0
	for _, num := range nums {
		if num >= max {
			max = num
		}
	}
	return max
}

type vbox struct {
	r1    int
	r2    int
	g1    int
	g2    int
	b1    int
	b2    int
	histo map[int]int
}

func (v *vbox) volume() int {
	sub_r := v.r2 - v.r1 + 1
	sub_g := v.g2 - v.g1 + 1
	sub_b := v.b2 - v.b1 + 1

	return sub_r * sub_g * sub_b
}

func (v *vbox) avg() (r_avg, g_avg, b_avg int) {
	var ntot, r_sum, g_sum, b_sum float32
	mult := float32(1 << (8 - sigbits))

	for i := v.r1; i < v.r2+1; i++ {
		for j := v.g1; j < v.g2+1; j++ {
			for k := v.b1; k < v.b2+1; k++ {

				index := getColorIndex(i, j, k)
				hval := float32(v.histo[index])

				ntot += hval

				r_sum += hval * (float32(i) + 0.5) * mult
				g_sum += hval * (float32(j) + 0.5) * mult
				b_sum += hval * (float32(k) + 0.5) * mult

			}
		}
	}

	if ntot > 0 {
		r_avg = int(r_sum / ntot)
		g_avg = int(g_sum / ntot)
		b_avg = int(b_sum / ntot)
	} else {
		r_avg = int(int(mult) * (v.r1 + v.r2 + 1) / 2)
		g_avg = int(int(mult) * (v.g1 + v.g2 + 1) / 2)
		b_avg = int(int(mult) * (v.b1 + v.b2 + 1) / 2)
	}
	return r_avg, g_avg, b_avg
}

func (v *vbox) contains(pixel imt.Color) bool {
	r, g, b, _ := pixel.RGBAint()

	return r >= v.r1 && r <= v.r2 && g >= v.g1 && r <= v.g2 && b >= v.b1 && b <= v.b2
}

func (v *vbox) count() int {
	npix := 0

	for i := v.r1; i < v.r2+1; i++ {
		for j := v.g1; j < v.g2+1; j++ {
			for k := v.b1; k < v.b2+1; k++ {

				index := getColorIndex(i, j, k)
				npix += v.histo[index]
			}
		}
	}

	return npix
}

func vboxFromPixels(pixels []imt.Color, histo map[int]int) vbox {
	rmin, gmin, bmin := 1000000, 1000000, 1000000
	rmax, gmax, bmax := 0, 0, 0

	for _, p := range pixels {
		r, g, b, _ := p.RGBA()

		rval := int(r >> rshift)
		gval := int(g >> rshift)
		bval := int(b >> rshift)

		rmin = min(rval, rmin)
		rmax = max(rval, rmax)

		gmin = min(gval, gmin)
		gmax = max(gval, gmax)

		bmin = min(bval, bmin)
		bmax = max(bval, bmax)

	}

	return vbox{
		r1:    rmin,
		r2:    rmax,
		g1:    gmin,
		g2:    gmax,
		b1:    bmin,
		b2:    bmax,
		histo: histo,
	}
}

func medianCutApply(v vbox) (v1, v2 vbox) {
	if v.count() == 0 {
		return v1, v2
	}

	rw := v.r2 - v.r1 + 1
	gw := v.g2 - v.g1 + 1
	bw := v.b2 - v.b1 + 1
	maxw := max(rw, gw, bw)
	if v.count() == 1 {
		return v, v2
	}

	var cutColor string
	var total, sum int
	partialsum := make(map[int]int)
	lookaheadsum := make(map[int]int)

	if maxw == rw {
		cutColor = "r"

		for i := v.r1; i < v.r2+1; i++ {
			sum = 0
			for j := v.g1; j < v.g2+1; j++ {
				for k := v.b1; k < v.b2+1; k++ {
					index := getColorIndex(i, j, k)
					sum += v.histo[index]
				}
			}
			total += sum
			partialsum[i] = total
		}

	} else if maxw == gw {
		cutColor = "g"

		for i := v.g1; i < v.g2+1; i++ {
			sum = 0
			for j := v.r1; j < v.r2+1; j++ {
				for k := v.b1; k < v.b2+1; k++ {
					index := getColorIndex(j, i, k)
					sum += v.histo[index]
				}
			}
			total += sum
			partialsum[i] = total
		}

	} else {
		cutColor = "b"

		for i := v.b1; i < v.b2+1; i++ {
			sum = 0
			for j := v.r1; j < v.r2+1; j++ {
				for k := v.g1; k < v.g2+1; k++ {
					index := getColorIndex(j, k, i)
					sum += v.histo[index]
				}
			}
			total += sum
			partialsum[i] = total
		}

	}

	for i, d := range partialsum {
		lookaheadsum[i] = total - d
	}

	var dim1, dim2 int
	switch cutColor {
	case "r":
		dim1, dim2 = v.r1, v.r2
	case "g":
		dim1, dim2 = v.g1, v.g2
	case "b":
		dim1, dim2 = v.b1, v.b2
	default:
	}
	for i := dim1; i < dim2+1; i++ {
		if partialsum[i] > (total / 2) {
			vbox1 := v
			vbox2 := v
			left := i - dim1
			right := dim2 - i
			if left <= right {
				d2 := min(dim2-1, i+(right/2))
			} else {
				d2 := max(dim1, i-1-(left/2))
			}
			// Avoid 0-count boxes

		}
	}
	return v1, v2
}
