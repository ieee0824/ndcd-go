// $ ndcd input.jpg output.png -pix-size 64
package main

import (
	"flag"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/ieee0824/ndcd-go"
)

const maxHeight = 256

func main() {
	log.SetFlags(log.Lshortfile)
	imageHeight := flag.Int("oh", 64, "output image height")
	inputFileName := flag.String("i", "", "input file name")
	outputFileName := flag.String("o", "", "output file name")
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
	})
	if err != nil {
		log.Fatal(err)
	}

	ex := filepath.Ext(*outputFileName)

	writeImage, err := os.Create(*outputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer writeImage.Close()

	switch ex {
	case ".png", ".PNG":
		if err := png.Encode(writeImage, converter); err != nil {
			log.Fatal(err)
		}
	case ".jpg", ".jpeg", ".JPG", ".JPEG":
		if err := jpeg.Encode(writeImage, converter, nil); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unsupported file type: %s", ex)
	}
}
