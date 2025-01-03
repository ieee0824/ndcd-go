// $ ndcd input.jpg output.png -pix-size 64
package main

import (
	"flag"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/ieee0824/ndcd-go"
	"github.com/nfnt/resize"
)

const maxHeight = 256

func main() {
	log.SetFlags(log.Lshortfile)
	imageHeight := flag.Int("oh", 64, "output image height")
	inputFileName := flag.String("i", "", "input file name")
	outputFileName := flag.String("o", "", "output file name")
	blurSize := flag.Float64("bs", 0.0, "blur size")
	blurType := flag.String("bt", "gaussian", "gaussian or box")
	contrast := flag.Float64("c", 0.0, "contrast")
	gamma := flag.Float64("g", 0.0, "gamma")
	sharpen := flag.Bool("s", false, "sharpen")
	colorQuantization := flag.Int("cq", 0, "color quantization")

	// 出力の拡大
	outputExpandSize := flag.Int("oe", 0, "output expand size")
	flag.Parse()

	if *inputFileName == "" || *outputFileName == "" {
		log.Fatal("input file name and output file name are required")
	}
	if *imageHeight > maxHeight {
		log.Fatalf("image height must be less than or equal to %d", maxHeight)
	}

	originalImage, err := os.Open(*inputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer originalImage.Close()

	converter, err := ndcd.New(originalImage, func(opt *ndcd.NdcdOption) {
		opt.ImageHeight = *imageHeight
		opt.BlurSize = *blurSize
		opt.BlurType = *blurType
		opt.Contrast = *contrast
		opt.Ganmma = *gamma
		opt.Sharpen = *sharpen
		opt.ColorQuantization = *colorQuantization
	})
	if err != nil {
		log.Fatal(err)
	}

	var outputImage image.Image
	if *outputExpandSize > 0 {
		outputImage = resize.Resize(uint(*outputExpandSize), 0, converter, resize.NearestNeighbor)
	} else {
		outputImage = converter
	}

	ex := filepath.Ext(*outputFileName)

	writeImage, err := os.Create(*outputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer writeImage.Close()

	switch ex {
	case ".png", ".PNG":
		if err := png.Encode(writeImage, outputImage); err != nil {
			log.Fatal(err)
		}
	case ".jpg", ".jpeg", ".JPG", ".JPEG":
		if err := jpeg.Encode(writeImage, outputImage, &jpeg.Options{
			Quality: 100,
		}); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unsupported file type: %s", ex)
	}
}
