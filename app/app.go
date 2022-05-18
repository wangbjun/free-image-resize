package app

import (
	"fmt"
	"freeImageResize/common"
	"freeImageResize/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"log"
	"os"
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

	chooseOutput *widget.Button
	outputDir    *widget.Entry
	outputOpt    *widget.Check

	submitButton *widget.Button
	clearButton  *widget.Button

	imageData []*imageStatus
	resize    *common.Resize
}

type imageStatus struct {
	Name    string
	Path    string
	Size    string
	ModTime string
	Status  string
}

var imgExtMap = map[string]int{"png": 1, "jpg": 1, "jpeg": 1, "gif": 1}

func NewApp() *App {
	resize := app.NewWithID("fyneResize")
	resize.Settings().SetTheme(&theme.MyTheme{})
	application := &App{
		fyne:   resize,
		window: resize.NewWindow("Free永久免费图片压缩工具---支持jpg、png、gif格式"),
		resize: common.New(),
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
	app.inputWidth.PlaceHolder = "长度，默认保持不变"
	app.inputWidth.OnChanged = func(s string) {
		i, _ := strconv.Atoi(s)
		app.resize.SetWidth(uint(i))
	}

	app.inputHeight = widget.NewEntry()
	app.inputHeight.PlaceHolder = "宽度，默认保持不变"
	app.inputHeight.OnChanged = func(s string) {
		i, _ := strconv.Atoi(s)
		app.resize.SetHeight(uint(i))
	}

	f := 75.0
	data := binding.BindFloat(&f)
	app.inputLabel = widget.NewLabelWithData(binding.FloatToStringWithFormat(data, "压缩率（图片大小）: %.0f%%"))
	app.inputQuality = widget.NewSliderWithData(0, 100, data)
	app.inputQuality.OnChanged = func(f float64) {
		app.resize.SetQuality(int(f))
		data.Set(f)
	}

	app.setupImageList()

	app.chooseInput = widget.NewButton("选择图片文件夹", app.chooseInputSubmit)

	outputDir := ""
	outputData := binding.BindString(&outputDir)
	app.outputDir = widget.NewEntryWithData(outputData)
	app.outputDir.PlaceHolder = "默认为图片输入文件夹"
	app.chooseOutput = widget.NewButton("选择输出文件夹", app.chooseOutputSubmit)
	app.outputOpt = widget.NewCheck("覆盖原文件", app.outputOptSubmit)

	app.submitButton = widget.NewButton("开始处理", app.submit)
	app.clearButton = widget.NewButton("清除列表", func() {
		app.imageData = []*imageStatus{}
		app.imageList.Hide()
	})

	app.window.SetContent(
		container.NewBorder(
			container.NewVBox(
				container.NewAdaptiveGrid(2, container.NewAdaptiveGrid(3, app.chooseInput, app.chooseOutput, app.outputOpt), app.outputDir),
				container.NewAdaptiveGrid(3, widget.NewLabel("压缩长宽配置："), app.inputWidth, app.inputHeight),
				container.NewAdaptiveGrid(2, app.inputLabel, app.inputQuality),
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
		app.resize.SetOutputDir(dir.Path())
		readDir, err := os.ReadDir(dir.Path())
		if err != nil {
			app.alert(err.Error())
			return
		}
		app.imageData = []*imageStatus{}
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
			app.imageData = append(app.imageData, &imageStatus{
				Name:    d.Name(),
				Path:    dir.Path() + "/" + d.Name(),
				Size:    humanize.Bytes(uint64(info.Size())),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
				Status:  "待处理",
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
		app.outputDir.SetText("")
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
				label.SetText(app.imageData[id.Row].Size)
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
	}
	app.imageList.SetColumnWidth(0, 40)
	app.imageList.SetColumnWidth(1, 360)
	app.imageList.SetColumnWidth(2, 90)
	app.imageList.SetColumnWidth(3, 180)
	app.imageList.SetColumnWidth(4, 90)
	app.imageList.SetColumnWidth(5, 80)
}

func (app *App) submit() {
	if len(app.imageData) == 0 {
		app.alert("请先选择图片")
		return
	}
}

func (app *App) alert(msg string) {
	info := dialog.NewInformation("提示", msg, app.window)
	info.Show()
}
