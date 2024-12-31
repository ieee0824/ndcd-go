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
)

type NdcdOption struct {
	ImageHeight int
}

func New(r io.Reader, optFunc ...func(opt *NdcdOption)) (*Ndcd, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	resizedImage := resize.Resize(0, 256, img, resize.Lanczos3)

	imageHeight := resizedImage.Bounds().Max.Y - resizedImage.Bounds().Min.Y
	defaultOpt := &NdcdOption{
		ImageHeight: 64,
	}

	defaultPixSize := imageHeight / 64
	for _, f := range optFunc {
		f(defaultOpt)
		if defaultOpt.ImageHeight > 64 {
			return nil, fmt.Errorf("image height must be less than or equal to 64")
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
		NumColors: 128,
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

func colorAvg(c1, c2 color.Color) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return color.RGBA64{
		uint16((r1 + r2) / 2),
		uint16((g1 + g2) / 2),
		uint16((b1 + b2) / 2),
		uint16((a1 + a2) / 2),
	}
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

	if dist1 > dist2 {
		// c1, c2の平均色を返す
		return colorAvg(c1, c2)
	}

	// c3, c4の平均色を返す
	return colorAvg(c3, c4)
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
