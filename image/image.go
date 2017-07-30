package image

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"

	_ "github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"github.com/zeropage/mukgoorm/setting"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	JPG_EXTEND  = "jpg"
	JPEG_EXTEND = "jpeg"
	PNG_EXTEND  = "png"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "../testdata/luxisr.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", 14, "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
)

func FileExtend(filename string) string {
	s := strings.Split(filename, ".")
	return s[len(s)-1]
}

func IsImage(filename string) bool {
	extend := FileExtend(filename)
	if extend == JPG_EXTEND || extend == PNG_EXTEND {
		return true
	}
	return false
}

func Resize(imagePath string, size uint) {
	t := signature(imagePath)
	if t != JPEG_EXTEND && t != PNG_EXTEND {
		return
	}

	d, err := ioutil.ReadFile(imagePath)
	if err != nil {
		panic(err)
	}
	Compress(size, imagePath, d)
}

func GenerateTextToImage(fileName string, filePath string) {
	// fontfile := "../testdata/luxisr.ttf"
	// dpi := 72
	// Read the font data.
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		panic(err)
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}

	// Draw the background and the guidelines.
	fg, bg := image.Black, image.White
	ruler := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	if *wonb {
		fg, bg = image.White, image.Black
		ruler = color.RGBA{0x22, 0x22, 0x22, 0xff}
	}
	const imgW, imgH = 300, 300
	rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	for i := 0; i < imgH-20; i++ {
		rgba.Set(10, 10+i, ruler)
		rgba.Set(imgW-10, 10+i, ruler)
	}
	for i := 0; i < imgW-20; i++ {
		rgba.Set(10+i, 10, ruler)
		rgba.Set(10+i, imgH-10, ruler)
	}

	// Draw the text.
	h := font.HintingNone
	// use font
	switch *hinting {
	case "full":
		h = font.HintingFull
	}
	d := &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    *size,
			DPI:     *dpi,
			Hinting: h,
		}),
	}
	y := 10 + int(math.Ceil(*size**dpi/72))
	dy := int(math.Ceil(*size * *spacing * *dpi / 72))
	d.Dot = fixed.Point26_6{
		X: (fixed.I(imgW) - d.MeasureString(fileName)) / 2,
		Y: fixed.I(y),
	}
	d.DrawString(fileName)
	y += dy

	file, _ := os.Open(filePath)
	scanner := bufio.NewScanner(file)
	// TODO count y sum
	for i := 0; i < 12; i++ {
		if !scanner.Scan() {
			break
		}
		fmt.Println(scanner.Text())
		d.Dot = fixed.P(20, y)
		d.DrawString(scanner.Text())
		y += dy
	}

	// Save that RGBA image to disk.
	outFile, err := os.Create(path.Join(ImagePath(), "out.jpeg"))
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	b := bufio.NewWriter(outFile)
	jpeg.Encode(b, rgba, nil)
	if err != nil {
		panic(err)
	}
	err = b.Flush()
	if err != nil {
		panic(err)
	}
}

func Compress(size uint, imagePath string, data []byte) {
	var img image.Image
	var err error
	if extend := FileExtend(imagePath); extend == PNG_EXTEND {
		img, err = png.Decode(bytes.NewReader(data))
	} else {
		img, _, err = image.Decode(bytes.NewReader(data))
	}
	if err != nil {
		panic(err)
	}
	newImg := resize.Resize(size, size, img, resize.Lanczos3)

	s := strings.Split(imagePath, "/")
	s = strings.Split(s[len(s)-1], ".")
	name := s[0] + "." + JPG_EXTEND
	dir := path.Join(ImagePath(), name)
	out, err := os.Create(dir)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	jpeg.Encode(out, newImg, nil)
}

func ImagePath() string {
	return path.Join(setting.GetDirectory().Path, ".images")
}

func MakeImageDir() {
	imageDir := ImagePath()
	if f, err := os.Stat(imageDir); f == nil {
		err = os.Mkdir(imageDir, 0770)
		if err != nil {
			panic(err)
		}
	}
}
