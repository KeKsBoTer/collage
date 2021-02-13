package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
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

	// image Copy instruction information
	type imgCopy struct {
		image  image.Image
		bounds image.Rectangle
	}

	var wg sync.WaitGroup
	var jobs chan imgCopy = make(chan imgCopy)

	// define worker function that writes images to target image
	worker := func(jobs <-chan imgCopy) {
		defer wg.Done()
		for c := range jobs {
			size := c.bounds.Size()
			smallImage := resize.Resize(uint(size.X), uint(size.Y), c.image, resize.Lanczos3)
			draw.Draw(targetImage, c.bounds, smallImage, image.Pt(0, 0), draw.Src)
		}
	}

	for w := 1; w <= runtime.NumCPU(); w++ {
		wg.Add(1)
		go worker(jobs)
	}

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
		jobs <- imgCopy{
			image:  img,
			bounds: image.Rect(x, y, x+targetWidth, y+targetHeight),
		}
		x += int(actualWidth)
		if x == targetImage.Rect.Size().X {
			y += int(targetHeight)
			x = 0
		}
	}

	close(jobs)
	wg.Wait()

	outputFile, err := os.Create(outputFilePath)
	defer outputFile.Close()

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
