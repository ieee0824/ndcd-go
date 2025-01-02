package ndcd

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"

	"github.com/ericpauley/go-quantize/quantize"
	"github.com/nfnt/resize"
	"github.com/samber/lo"
)

const maxHeight = 256

type NdcdOption struct {
	ImageHeight int
}

func New(r io.Reader, optFunc ...func(opt *NdcdOption)) (*Ndcd, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	resizedImage := resize.Resize(0, 512, img, resize.NearestNeighbor)

	imageHeight := resizedImage.Bounds().Max.Y - resizedImage.Bounds().Min.Y
	defaultOpt := &NdcdOption{
		ImageHeight: maxHeight,
	}

	defaultPixSize := imageHeight / maxHeight
	for _, f := range optFunc {
		f(defaultOpt)
		if defaultOpt.ImageHeight > maxHeight {
			return nil, fmt.Errorf("image height must be less than or equal to %d", maxHeight)
		}
		defaultPixSize = imageHeight / defaultOpt.ImageHeight
	}

	tmpFile, err := os.CreateTemp("", "ndcd.gif")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := gif.Encode(tmpFile, resizedImage, &gif.Options{
		NumColors: 32,
		Quantizer: quantize.MedianCutQuantizer{},
	}); err != nil {
		return nil, fmt.Errorf("failed to encode gif: %w", err)
	}

	tmpFile.Seek(0, 0)

	baseImage, _, err := image.Decode(tmpFile)
	if err != nil {
		panic(err)
	}

	return &Ndcd{
		baseImg: baseImage,
		pixSize: defaultPixSize,
	}, nil
}

type Ndcd struct {
	baseImg image.Image
	pixSize int
}

func colorDist(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return math.Sqrt(float64((r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2)))
}

// gray scaleの濃い方を返す
func colorMax(c1, c2 color.Color) color.Color {
	g1 := color.Gray16Model.Convert(c1).(color.Gray16).Y
	g2 := color.Gray16Model.Convert(c2).(color.Gray16).Y

	if g1 > g2 {
		return c1
	}
	return c2
}

func colorAvgTwoClor(c1, c2 color.Color) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return color.RGBA64{
		uint16((r1 + r2) / 2),
		uint16((g1 + g2) / 2),
		uint16((b1 + b2) / 2),
		uint16((a1 + a2) / 2),
	}
}

func colorAvg(colors []color.Color) color.Color {
	return lo.Reduce(colors, func(agg, item color.Color, _ int) color.Color {
		if agg == nil {
			return item
		}

		return colorAvgTwoClor(agg, item)
	}, nil)
}

func (impl *Ndcd) At(x, y int) color.Color {
	// 4点の色を取得
	c1 := impl.baseImg.At(x*impl.pixSize, y*impl.pixSize)
	c2 := impl.baseImg.At(x*impl.pixSize+impl.pixSize-1, y*impl.pixSize)
	c3 := impl.baseImg.At(x*impl.pixSize, y*impl.pixSize+impl.pixSize-1)
	c4 := impl.baseImg.At(x*impl.pixSize+impl.pixSize-1, y*impl.pixSize+impl.pixSize-1)

	// 対角線の色差を計算
	dist1 := colorDist(c1, c4)
	dist2 := colorDist(c2, c3)

	var u, v color.Color
	if dist1 > dist2 {
		u = c1
		v = c4
	} else {
		u = c2
		v = c3
	}
	su := []color.Color{}
	sv := []color.Color{}

	for k := x * impl.pixSize; k < x*impl.pixSize+impl.pixSize; k++ {
		for l := y * impl.pixSize; l < y*impl.pixSize+impl.pixSize; l++ {
			if colorDist(impl.baseImg.At(k, l), u) < colorDist(impl.baseImg.At(k, l), v) {
				su = append(su, impl.baseImg.At(k, l))
			} else {
				sv = append(sv, impl.baseImg.At(k, l))
			}
		}
	}

	if len(su) > len(sv) {
		return colorAvg(su)
	} else {
		return colorAvg(sv)
	}
}

func (impl *Ndcd) ColorModel() color.Model {
	return impl.baseImg.ColorModel()
}

func (impl *Ndcd) Bounds() image.Rectangle {
	orgBounds := impl.baseImg.Bounds()

	_ = orgBounds

	minX := orgBounds.Min.X
	minY := orgBounds.Min.Y
	maxX := orgBounds.Max.X
	maxY := orgBounds.Max.Y

	newHeight := (maxY - minY) / impl.pixSize
	newWidth := (maxX - minX) / impl.pixSize

	return image.Rect(minX, minY, minX+newWidth, minY+newHeight)
}
