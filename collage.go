package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	glob "github.com/bmatcuk/doublestar/v2"
	"github.com/nfnt/resize"
)

func main() {

	width := flag.Int("width", 1920, "width of created image")
	height := flag.Uint("height", 1080, "width of created image")
	rows := flag.Uint("rows", 5, "number of image rows")

	flag.Usage = func() {
		fmt.Println("A tools for creating photo collages")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  collage <arguments> [source directory] [output file]")
		fmt.Println()
		fmt.Println("Arguments:")
		flag.PrintDefaults()
	}
	flag.Parse()

	dir := flag.Arg(0)
	if dir == "" {
		fmt.Println("missing source directory")
		os.Exit(1)
		return
	}
	outputFilePath := flag.Arg(1)
	if outputFilePath == "" {
		fmt.Println("missing output file")
		os.Exit(1)
		return
	}

	targetSize := image.Point{
		X: int(*width),
		Y: int(*height),
	}

	targetHeight := targetSize.Y / int(*rows)

	files, err := glob.Glob(filepath.Join(dir, "**/*.{jpeg,jpg,gif,png}"))
	if err != nil {
		log.Fatalf("error finding files: %v\n", err)
		os.Exit(1)
		return
	}

	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	targetImage := image.NewRGBA(image.Rectangle{Max: targetSize})

	var c chan int = make(chan int)
	images := 0

	x := 0
	y := 0
	for _, f := range files {
		imgFile, err := os.OpenFile(f, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Println("ERROR ", err)
			imgFile.Close()
			continue
		}

		img, _, err := image.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			log.Println("ERROR ", err)
			continue
		}
		size := img.Bounds().Size()
		targetWidth := int(float32(size.X) / float32(size.Y) * float32(targetHeight))

		actualWidth := targetWidth
		if x+int(targetWidth) > targetImage.Rect.Size().X {
			actualWidth = targetImage.Rect.Size().X - x
		}

		if y+int(targetHeight) > targetSize.Y {
			break
		}

		// copy async as it is only writing to a fixed place
		// (no goroutine will write to the same place in the image)
		go func(x int, y int) {
			//var copyImage func()
			switch input := img.(type) {
			case *image.RGBA:
				// copy directly for RGBA as this is faster
				sub := targetImage.SubImage(image.Rect(x, y, x+int(targetWidth), y+int(targetHeight))).(*image.RGBA)
				for r := 0; r < int(targetHeight); r++ {
					copy(
						sub.Pix[r*sub.Stride:r*sub.Stride+int(actualWidth)*4],
						input.Pix[r*input.Stride:r*input.Stride+int(actualWidth)*4],
					)
				}
			default:
				smallImage := resize.Resize(uint(targetWidth), uint(targetHeight), img, resize.Lanczos3)
				for i := 0; i < int(targetHeight); i++ {
					for j := 0; j < int(actualWidth); j++ {
						targetImage.Set(x+j, y+i, smallImage.At(j, i))
					}
				}
			}
			c <- 1
		}(x, y)
		images++
		x += int(actualWidth)
		if x == targetImage.Rect.Size().X {
			y += int(targetHeight)
			x = 0
		}
	}

	// wait for all images to copy
	for i := 0; i < images; i++ {
		<-c
	}

	outputFile, err := os.Create(outputFilePath)
	switch filepath.Ext(outputFilePath) {
	case ".png":
		err = png.Encode(outputFile, targetImage)
	case ".jpg":
		fallthrough
	case ".jpeg":
		err = jpeg.Encode(outputFile, targetImage, &jpeg.Options{Quality: 100})
	case ".gif":
		err = gif.Encode(outputFile, targetImage, &gif.Options{})
	}
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	fmt.Printf("saved image to %s\n", outputFilePath)
}
