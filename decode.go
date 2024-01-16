package blurhash

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/FlamingTree/go-blurhash/base83"
)

// Components returns the X and Y components of a blurhash.
// func Components(hash string) (x, y int, err error) {
// 	sizeFlag, err := base83.Decode(string(hash[0]))
// 	if err != nil {
// 		return 0, 0, err
// 	}

// 	x = (sizeFlag % 9) + 1
// 	y = (sizeFlag / 9) + 1

// 	expectedLength := 4 + 2*x*y
// 	actualLength := len(hash)
// 	if expectedLength != actualLength {
// 		return 0, 0, lengthError(expectedLength, actualLength)
// 	}

//		return x, y, nil
//	}
func Components(hash string) (x, y int, err error) {
	x, err = base83.Decode(string(hash[0]))
	if err != nil {
		return 0, 0, err
	}

	y, err = base83.Decode(string(hash[1]))
	if err != nil {
		return 0, 0, err
	}

	return x, y, nil
}

// Decode returns an NRGBA image of the given hash with the given size.
func Decode(hash string, width, height int, punch int) (image.Image, error) {
	newImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	if err := DecodeDraw(newImg, hash, float64(punch)); err != nil {
		return nil, err
	}
	return newImg, nil
}

type drawImageNRGBA interface {
	SetNRGBA(x, y int, c color.NRGBA)
}

type drawImageRGBA interface {
	SetRGBA(x, y int, c color.RGBA)
}

// DecodeDraw decodes the given hash into the given image.
func DecodeDraw(dst draw.Image, hash string, punch float64) error {
	numX, numY, err := Components(hash)
	if err != nil {
		return err
	}

	// quantisedMaximumValue, err := base83.Decode(string(hash[1]))
	quantisedMaximumValue, err := base83.Decode(string(hash[2]))
	if err != nil {
		return err
	}
	maximumValue := float64(quantisedMaximumValue+1) / 166
	fmt.Printf("quantisedMaximumValue: %d, maximumValue: %f\n", quantisedMaximumValue, maximumValue)

	// for each component
	colors := make([][3]float64, numX*numY)
	for i := range colors {
		if i == 0 {
			// val, err := base83.Decode(hash[2:6])
			val, err := base83.Decode(hash[3:7])
			if err != nil {
				return err
			}
			colors[i] = decodeDC(val)
			fmt.Println(fmt.Sprintf("DC: %s, %f, %f, %f", hash[3:7], colors[i][0], colors[i][1], colors[i][2]))
		} else {
			// val, err := base83.Decode(hash[4+i*2 : 6+i*2])
			val, err := base83.Decode(hash[4+i*3 : 7+i*3])
			if err != nil {
				return err
			}
			colors[i] = decodeAC(float64(val), maximumValue*punch)
		}
	}

	for y := 0; y < numY; y++ {
		for x := 0; x < numX; x++ {
			if y == 0 && x == 0 {
				continue
			}
			dc := colors[x+y*numX]
			fmt.Printf("AC[%d][%d]: %f, %f, %f\n", y, x, dc[0], dc[1], dc[2])
		}
	}

	bounds := dst.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b float64

			for j := 0; j < numY; j++ {
				for i := 0; i < numX; i++ {
					basis := math.Cos(math.Pi*float64(x)*float64(i)/float64(width)) * math.Cos(math.Pi*float64(y)*float64(j)/float64(height))
					compColor := colors[i+j*numX]
					r += compColor[0] * basis
					g += compColor[1] * basis
					b += compColor[2] * basis
				}
			}

			sR := uint8(linearTosRGB(r))
			sG := uint8(linearTosRGB(g))
			sB := uint8(linearTosRGB(b))
			sA := uint8(255)

			// interface smuggle
			switch d := dst.(type) {
			case drawImageNRGBA:
				d.SetNRGBA(x, y, color.NRGBA{sR, sG, sB, sA})
			case drawImageRGBA:
				d.SetRGBA(x, y, color.RGBA{sR, sG, sB, sA})
			default:
				d.Set(x, y, color.NRGBA{sR, sG, sB, sA})
			}
		}
	}

	return nil
}

func decodeDC(val int) (c [3]float64) {
	c[0] = sRGBToLinear(val >> 16)
	c[1] = sRGBToLinear(val >> 8 & 255)
	c[2] = sRGBToLinear(val & 255)
	return c
}

//	func decodeAC(val, maximumValue float64) (c [3]float64) {
//		quantR := math.Floor(val / (19 * 19))
//		quantG := math.Mod(math.Floor(val/19), 19)
//		quantB := math.Mod(val, 19)
//		c[0] = signPow((quantR-9)/9, 2.0) * maximumValue
//		c[1] = signPow((quantG-9)/9, 2.0) * maximumValue
//		c[2] = signPow((quantB-9)/9, 2.0) * maximumValue
//		return c
//	}
func decodeAC(val, maximumValue float64) (c [3]float64) {
	quantR := math.Floor(val / (83 * 83))
	quantG := math.Mod(math.Floor(val/83), 83)
	quantB := math.Mod(val, 83)
	c[0] = signPow((quantR-41)/41, 2.0) * maximumValue
	c[1] = signPow((quantG-41)/41, 2.0) * maximumValue
	c[2] = signPow((quantB-41)/41, 2.0) * maximumValue
	return c
}
