package common

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"os"
)

type ImageItem struct {
	Name    string
	Path    string
	Size    int64
	SizeStr string
	ModTime string
	Status  string
}

type JobProcessor struct {
	resize *Resize
	input  chan ImageItem
	output chan ImageItem
}

func NewJobProcessor(resize *Resize) *JobProcessor {
	JobProcessor := &JobProcessor{
		resize: resize,
		input:  make(chan ImageItem, 0),
		output: make(chan ImageItem, 0),
	}
	JobProcessor.Run()
	return JobProcessor
}

func (r JobProcessor) AddJob(items ...ImageItem) {
	for _, item := range items {
		r.input <- item
	}
}

func (r JobProcessor) Output() chan ImageItem {
	return r.output
}

func (r JobProcessor) Run() {
	for i := 0; i < 3; i++ {
		go func() {
			for {
				item := <-r.input
				outputFile, err := r.resize.Resize(item.Path)
				if err != nil {
					log.Printf("resize %s error: %s", item.Name, err)
					item.Status = "缩放错误"
					r.output <- item
					continue
				}
				stat, err := os.Stat(outputFile)
				if err != nil {
					item.Status = "未知错误"
					r.output <- item
					continue
				}
				item.Status = fmt.Sprintf("已完成[%s]", humanize.Bytes(uint64(stat.Size())))
				r.output <- item
			}
		}()
	}
}
