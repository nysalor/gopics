package handler

import (
	"os"
	"io"
	"fmt"
	"image/jpeg"
	"path/filepath"
	"crypto/md5"
	"github.com/nfnt/resize"
)

type Resizer struct {
	OrigPath  string
	Filename string
	OutDir   string
	Width    uint
	Height   uint
}

func (resizer *Resizer) OutPath(filename string) (path string) {
	path = filepath.Join(resizer.OutDir, filename)
	return
}

func baseName(path string) (baseName string) {
	baseName = filepath.Base(path[:len(path)-len(filepath.Ext(path))])
	return
}

func resizeImage(resizer Resizer) (outfile string) {
	file, err := os.Open(resizer.OrigPath)
	if err != nil {
		return
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		return
	}
	file.Close()

	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		return
	}
	csum := h.Sum(nil)
	baseName := baseName(resizer.Filename)
	outfile = fmt.Sprintf("%s-%x.jpg", baseName, csum)
	m := resize.Thumbnail(resizer.Width, resizer.Height, img, resize.Lanczos3)
	outpath := filepath.Join(resizer.OutPath(outfile), outfile)

	out, err := os.Create(outpath)
	if err != nil {
		return
	}
	defer out.Close()
	jpeg.Encode(out, m, nil)

	return
}
