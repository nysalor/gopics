package model

import (
	"os"
	"fmt"
	"strings"
	"image/jpeg"
	"path/filepath"
	"github.com/nfnt/resize"
)

type Resizer struct {
	OrigPath  string
	OutDir   string
	Width    uint
	Height   uint
}

func (resizer *Resizer) OutPath(filename string) (path string) {
	path = filepath.Join(resizer.OutDir, filename)
	return
}

func baseName(path string) (baseName string) {
	strs := strings.Split(filepath.Base(path), ".")
	baseName = strings.Join(strs[:len(strs)-1], ".")
	return
}

func (resizer *Resizer) ResizeImage() (outfile string) {
	file, err := os.Open(resizer.OrigPath)
	img, err := jpeg.Decode(file)
	if err != nil {
		panic(err)
	}
	file.Close()

	csum := checkSum(resizer.OrigPath)
	baseName := baseName(resizer.OrigPath)
	outfile = fmt.Sprintf("%s-%s.jpg", baseName, csum)
	m := resize.Thumbnail(resizer.Width, resizer.Height, img, resize.Lanczos3)
	outpath := filepath.Join(conf.CacheDir, outfile)

	out, err := os.Create(outpath)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)

	return
}
