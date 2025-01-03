package ndcd

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/effect"
	"github.com/nfnt/resize"
	"github.com/samber/lo"
)

const maxHeight = 256

type NdcdOption struct {
	ImageHeight int
	BlurSize    float64
	BlurType    string
	Contrast    float64
	Ganmma      float64
	Sharpen     bool
}

func New(r io.Reader, optFunc ...func(opt *NdcdOption)) (*Ndcd, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		// return nil, fmt.Errorf("failed to decode image: %w", err)
		return nil, fmt.Errorf("failed to decode image: %w, format: %s", err, format)
	}

	var baseImage image.Image
	if img.Bounds().Max.Y-img.Bounds().Min.Y > 1080 {
		baseImage = resize.Resize(0, 1080, img, resize.NearestNeighbor)
	} else {
		baseImage = img
	}

	imageHeight := baseImage.Bounds().Max.Y - baseImage.Bounds().Min.Y
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

	if defaultOpt.Contrast != 0 {
		baseImage = adjust.Contrast(baseImage, defaultOpt.Contrast)
	}

	if defaultOpt.BlurSize != 0 {
		switch defaultOpt.BlurType {
		case "box":
			baseImage = blur.Box(baseImage, defaultOpt.BlurSize)
		default:
			baseImage = blur.Gaussian(baseImage, defaultOpt.BlurSize)
		}
	}

	if defaultOpt.Ganmma != 0 {
		baseImage = adjust.Gamma(baseImage, defaultOpt.Ganmma)
	}

	if defaultOpt.Sharpen {
		baseImage = effect.Sharpen(baseImage)
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

type addedColor struct {
	r uint64
	g uint64
	b uint64
	a uint64
}

func addColor(c color.Color, acc *addedColor) *addedColor {
	if acc == nil {
		r, g, b, a := c.RGBA()
		return &addedColor{r: uint64(r), g: uint64(g), b: uint64(b), a: uint64(a)}
	}

	r, g, b, a := c.RGBA()
	return &addedColor{
		r: acc.r + uint64(r),
		g: acc.g + uint64(g),
		b: acc.b + uint64(b),
		a: acc.a + uint64(a),
	}
}

func colorAvg(colors []color.Color) color.Color {
	added := lo.Reduce(colors, func(agg *addedColor, item color.Color, _ int) *addedColor {
		return addColor(item, agg)
	}, nil)

	return color.RGBA64{
		uint16(added.r / uint64(len(colors))),
		uint16(added.g / uint64(len(colors))),
		uint16(added.b / uint64(len(colors))),
		uint16(added.a / uint64(len(colors))),
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

	minX := orgBounds.Min.X
	minY := orgBounds.Min.Y
	maxX := orgBounds.Max.X
	maxY := orgBounds.Max.Y

	newHeight := (maxY - minY) / impl.pixSize
	newWidth := (maxX - minX) / impl.pixSize

	return image.Rect(minX, minY, minX+newWidth, minY+newHeight)
}
