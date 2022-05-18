package common

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"strings"
)

type Resize struct {
	outputDir     string
	width, height uint
	quality       int
	isOverwrite   bool
	rotate        int
}

func New() *Resize {
	return &Resize{outputDir: ".", quality: 85}
}

// Resize 压缩图片，默认保存成jpg，压缩比例85%
// 如果长和宽都为0，则不改变长和宽
// 如果长和宽有一个不为0，则维持默认长宽比的情况下自动缩放
// 如果长和宽都不为0，则会固定长宽比，图片可能会变形
func (r *Resize) Resize(imgFile string) (string, error) {
	stat, err := os.Stat(imgFile)
	if err != nil {
		return "", err
	}
	imgName := stat.Name()
	if i := strings.LastIndex(imgName, "."); i != -1 {
		imgName = imgName[:strings.LastIndex(imgName, ".")]
	}
	//创建目录
	if _, err := os.Stat(r.outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(r.outputDir, 0755)
		if err != nil {
			return "", err
		}
	}
	outputFile := r.outputDir + fmt.Sprintf("/%s_resized.jpg", imgName)
	if r.isOverwrite {
		outputFile = r.outputDir + fmt.Sprintf("/%s.jpg", imgName)
	}
	file, err := os.Open(imgFile)
	if err != nil {
		return "", err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}
	//旋转图片
	if r.rotate != 0 {
		switch r.rotate {
		case 90:
			img = rotate90(img)
		case 180:
			img = rotate180(img)
		case 270:
			img = rotate270(img)
		}
	}
	imgResize := resize.Resize(r.width, r.height, img, resize.Lanczos3)
	output, err := os.Create(outputFile)
	if err != nil {
		return "", err
	}
	defer output.Close()
	return outputFile, jpeg.Encode(output, imgResize, &jpeg.Options{Quality: r.quality})
}

// SetOutputDir 设置结果输出目录
func (r *Resize) SetOutputDir(dir string) {
	r.outputDir = dir
}

// SetIsOverwrite 设置是否覆盖原文件
func (r *Resize) SetIsOverwrite(b bool) {
	r.isOverwrite = b
}

// SetQuality 设置压缩率
func (r *Resize) SetQuality(quality int) {
	r.quality = quality
}

// SetWidth 设置长
func (r *Resize) SetWidth(w uint) {
	r.width = w
}

// SetHeight 设置宽
func (r *Resize) SetHeight(h uint) {
	r.height = h
}

// SetRotate 设置旋转角度
func (r *Resize) SetRotate(rotate int) {
	r.rotate = rotate
}
