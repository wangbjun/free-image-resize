package app

import (
	"fmt"
	"freeImageResize/common"
	"freeImageResize/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type App struct {
	fyne         fyne.App
	window       fyne.Window
	inputWidth   *widget.Entry
	inputHeight  *widget.Entry
	inputLabel   *widget.Label
	inputQuality *widget.Slider

	imageList   *widget.Table
	chooseInput *widget.Button
	inputOpt    *widget.SelectEntry
	inputDir    string

	chooseOutput *widget.Button
	outputDir    *widget.Entry
	outputOpt    *widget.Check

	inputRotate *widget.SelectEntry

	submitButton *widget.Button
	clearButton  *widget.Button

	imageData    []common.ImageItem
	resize       *common.Resize
	jobProcessor *common.JobProcessor
}

var imgExtMap = map[string]int{"png": 1, "jpg": 1, "jpeg": 1, "gif": 1}

func NewApp() *App {
	resize := app.NewWithID("fyneResize")
	resize.Settings().SetTheme(&theme.MyTheme{})
	n := common.New()
	application := &App{
		fyne:         resize,
		window:       resize.NewWindow("Free永久免费图片压缩工具---支持jpg、png、gif格式"),
		resize:       n,
		jobProcessor: common.NewJobProcessor(n),
	}
	return application
}

func (app *App) Run() {
	app.setUp()
	app.window.Resize(fyne.NewSize(855, 555))
	app.window.CenterOnScreen()
	app.window.ShowAndRun()
}

func (app *App) setUp() {
	app.inputWidth = widget.NewEntry()
	app.inputWidth.PlaceHolder = "长度"
	app.inputWidth.OnChanged = func(s string) {
		i, _ := strconv.Atoi(s)
		app.resize.SetWidth(uint(i))
	}

	app.inputHeight = widget.NewEntry()
	app.inputHeight.PlaceHolder = "宽度"
	app.inputHeight.OnChanged = func(s string) {
		i, _ := strconv.Atoi(s)
		app.resize.SetHeight(uint(i))
	}

	f := 90.0
	data := binding.BindFloat(&f)
	app.inputLabel = widget.NewLabelWithData(binding.FloatToStringWithFormat(data, "压缩率： %.0f%%"))
	app.inputQuality = widget.NewSliderWithData(1, 100, data)
	app.inputQuality.OnChanged = func(f float64) {
		app.resize.SetQuality(int(f))
		data.Set(f)
	}

	app.setupImageList()

	app.chooseInput = widget.NewButton("选择文件夹", app.chooseInputSubmit)

	outputDir := ""
	outputData := binding.BindString(&outputDir)
	app.outputDir = widget.NewEntryWithData(outputData)
	app.outputDir.OnChanged = func(s string) {
		app.resize.SetOutputDir(s)
	}
	app.outputDir.PlaceHolder = "默认为图片输入文件夹"
	app.chooseOutput = widget.NewButton("选择输出", app.chooseOutputSubmit)
	app.outputOpt = widget.NewCheck("是否覆盖", app.outputOptSubmit)

	app.submitButton = widget.NewButton("开始处理", app.submit)
	app.clearButton = widget.NewButton("清除列表", func() {
		app.imageData = []common.ImageItem{}
		app.imageList.Hide()
	})

	app.inputRotate = widget.NewSelectEntry([]string{"0", "90", "180", "270"})
	app.inputRotate.SetText("0")
	app.inputRotate.OnChanged = func(s string) {
		i, _ := strconv.Atoi(s)
		app.resize.SetRotate(i)
	}

	app.inputOpt = widget.NewSelectEntry([]string{"1MB", "3MB", "5MB", "10MB"})
	app.inputOpt.SetText("1MB")
	app.inputOpt.OnChanged = func(s string) {
		gt, _ := strconv.Atoi(strings.ReplaceAll(s, "MB", ""))
		var tmp []common.ImageItem
		for _, img := range app.imageData {
			if img.Size >= int64(gt*1024*1024) {
				tmp = append(tmp, img)
			}
		}
		app.imageData = tmp
		app.imageList.Length = func() (int, int) {
			return len(app.imageData), 6
		}
		app.imageList.Refresh()
	}

	app.window.SetContent(
		container.NewBorder(
			container.NewVBox(
				container.NewAdaptiveGrid(3, container.NewAdaptiveGrid(3, app.chooseInput, widget.NewLabel("筛选大于："), app.inputOpt), container.NewAdaptiveGrid(2, app.chooseOutput, app.outputOpt), app.outputDir),
				container.NewAdaptiveGrid(7, widget.NewLabel("压缩配置："), app.inputWidth, app.inputHeight, app.inputLabel, app.inputQuality, widget.NewLabel("旋转角度："), app.inputRotate),
				container.NewAdaptiveGrid(4, layout.NewSpacer(), app.submitButton, app.clearButton, layout.NewSpacer()),
				widget.NewSeparator(),
			),
			nil, nil, nil, app.imageList,
		))
}

func (app *App) chooseInputSubmit() {
	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil {
			app.alert(err.Error())
			return
		}
		if dir == nil {
			log.Println("Cancelled")
			return
		}
		app.inputDir = dir.Path()
		app.outputDir.SetText(dir.Path())
		app.resize.SetOutputDir(dir.Path())
		readDir, err := os.ReadDir(dir.Path())
		if err != nil {
			app.alert(err.Error())
			return
		}
		app.imageData = []common.ImageItem{}
		for _, d := range readDir {
			if d.IsDir() {
				continue
			}
			index := strings.LastIndex(d.Name(), ".")
			if index == -1 {
				continue
			}
			//只处理jpg、png、gif
			if _, ok := imgExtMap[d.Name()[index+1:]]; !ok {
				continue
			}
			info, err := d.Info()
			if err != nil {
				continue
			}
			gt, _ := strconv.Atoi(strings.ReplaceAll(app.inputOpt.Text, "MB", ""))
			if info.Size() < int64(gt*1024*1024) {
				continue
			}
			app.imageData = append(app.imageData, common.ImageItem{
				Name:    d.Name(),
				Path:    dir.Path() + "/" + d.Name(),
				Size:    info.Size(),
				SizeStr: humanize.Bytes(uint64(info.Size())),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
				Status:  "待处理",
			})

			sort.Slice(app.imageData, func(i, j int) bool {
				return app.imageData[i].Size > app.imageData[j].Size
			})
		}

		if len(app.imageData) == 0 {
			app.imageList.Hide()
			app.alert("找不到符合要求的图片")
			return
		}
		app.imageList.Length = func() (int, int) {
			return len(app.imageData), 6
		}
		app.imageList.Refresh()
		app.imageList.ScrollToTop()
		app.imageList.Show()
	}, app.window)
}

func (app *App) chooseOutputSubmit() {
	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil {
			app.alert(err.Error())
			return
		}
		if dir == nil {
			return
		}
		app.outputDir.SetText(dir.Path())
		app.resize.SetOutputDir(dir.Path())
	}, app.window)
}

func (app *App) outputOptSubmit(b bool) {
	app.resize.SetIsOverwrite(b)
	if b == true {
		app.chooseOutput.Disable()
		app.outputDir.Disable()
		app.outputDir.SetText(app.inputDir)
		app.resize.SetOutputDir(app.inputDir)
	} else {
		app.chooseOutput.Enable()
		app.outputDir.Enable()
	}
}

func (app *App) setupImageList() {
	app.imageList = widget.NewTable(
		func() (int, int) { return 0, 0 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell 000, 000")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", id.Row+1))
			case 1:
				label.SetText(app.imageData[id.Row].Name)
			case 2:
				label.SetText(app.imageData[id.Row].SizeStr)
			case 3:
				label.SetText(app.imageData[id.Row].ModTime)
			case 4:
				label.SetText(app.imageData[id.Row].Status)
			case 5:
				label.SetText("移除")
			}
		})
	app.imageList.OnSelected = func(id widget.TableCellID) {
		if id.Col == 5 {
			app.imageData = append(app.imageData[:id.Row], app.imageData[id.Row+1:]...)
			app.imageList.Length = func() (int, int) {
				return len(app.imageData), 6
			}
			app.imageList.Refresh()
			app.imageList.UnselectAll()
		}
		if id.Col == 1 {
			item := app.imageData[id.Row]
			img := canvas.NewImageFromFile(item.Path)
			w := fyne.CurrentApp().NewWindow("查看图片")
			img.FillMode = canvas.ImageFillContain
			w.SetContent(img)
			w.Resize(fyne.Size{
				Width:  600,
				Height: 400,
			})
			w.CenterOnScreen()
			w.Show()
		}
	}
	app.imageList.SetColumnWidth(0, 50)
	app.imageList.SetColumnWidth(1, 370)
	app.imageList.SetColumnWidth(2, 100)
	app.imageList.SetColumnWidth(3, 190)
	app.imageList.SetColumnWidth(4, 120)
	app.imageList.SetColumnWidth(5, 90)
}

func (app *App) submit() {
	if len(app.imageData) == 0 {
		app.alert("请先选择图片")
		return
	}
	app.imageList.UnselectAll()
	app.imageList.ScrollToTop()
	if app.imageData[0].Status != "待处理" {
		for i := range app.imageData {
			app.imageData[i].Status = "待处理"
		}
		app.imageList.Refresh()
	}
	app.submitButton.SetText("正在处理中...")
	app.submitButton.Disable()
	app.clearButton.Disable()

	go app.jobProcessor.AddJob(app.imageData...)
	log.Printf("Add %d Job Success, quality: %f", len(app.imageData), app.inputQuality.Value)
	count := 0
	for {
		result := <-app.jobProcessor.Output()
		for i, image := range app.imageData {
			if image.Name == result.Name {
				app.imageData[i].Status = result.Status
			}
		}
		app.imageList.Refresh()
		count++
		if count == len(app.imageData) {
			break
		}
	}
	app.submitButton.SetText("开始处理")
	app.submitButton.Enable()
	app.clearButton.Enable()
	log.Println("All Job finished")
}

func (app *App) alert(msg string) {
	info := dialog.NewInformation("提示", msg, app.window)
	info.Show()
}
