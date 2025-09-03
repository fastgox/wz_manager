package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// ImageViewer 图像查看器
type ImageViewer struct {
	content       *fyne.Container
	imageDisplay  *canvas.Image
	scrollContent *container.Scroll
	infoLabel     *widget.Label
	currentImage  image.Image
	currentPng    *wzlib.WzPng
	zoomSlider    *widget.Slider
	saveButton    *widget.Button
	resetButton   *widget.Button
}

// NewImageViewer 创建新的图像查看器
func NewImageViewer() *ImageViewer {
	iv := &ImageViewer{
		infoLabel: widget.NewLabel("No image selected"),
	}

	iv.createContent()
	return iv
}

// createContent 创建图像查看器内容
func (iv *ImageViewer) createContent() {
	// 图像显示区域
	iv.imageDisplay = canvas.NewImageFromImage(nil)
	iv.imageDisplay.FillMode = canvas.ImageFillOriginal
	iv.imageDisplay.ScaleMode = canvas.ImageScaleSmooth

	// 滚动容器
	iv.scrollContent = container.NewScroll(iv.imageDisplay)
	iv.scrollContent.SetMinSize(fyne.NewSize(400, 300))

	// 缩放滑块
	iv.zoomSlider = widget.NewSlider(0.1, 3.0)
	iv.zoomSlider.Value = 1.0
	iv.zoomSlider.Step = 0.1
	iv.zoomSlider.OnChanged = iv.onZoomChanged

	// 控制按钮
	iv.saveButton = widget.NewButtonWithIcon("💾 保存图像", theme.DocumentSaveIcon(), iv.saveImage)
	iv.saveButton.Importance = widget.HighImportance
	iv.saveButton.Disable()

	iv.resetButton = widget.NewButtonWithIcon("🔄 重置缩放", theme.ViewRefreshIcon(), iv.resetZoom)
	iv.resetButton.Importance = widget.MediumImportance

	// 控制面板
	controlPanel := container.NewVBox(
		widget.NewLabel("Zoom:"),
		iv.zoomSlider,
		container.NewHBox(iv.saveButton, iv.resetButton),
	)

	// 组装内容
	iv.content = container.NewVBox(
		widget.NewLabel("Image Viewer"),
		widget.NewSeparator(),
		iv.infoLabel,
		controlPanel,
		widget.NewSeparator(),
		iv.scrollContent,
	)
}

// ShowImage 显示图像
func (iv *ImageViewer) ShowImage(nodeValue interface{}) {
	if wzPng, ok := nodeValue.(*wzlib.WzPng); ok {
		iv.currentPng = wzPng
		iv.loadPngImage(wzPng)
	} else {
		iv.clearImage()
		iv.infoLabel.SetText("Selected node is not an image")
	}
}

// loadPngImage 加载PNG图像
func (iv *ImageViewer) loadPngImage(wzPng *wzlib.WzPng) {
	img, err := wzPng.ExtractImage()
	if err != nil {
		iv.clearImage()
		iv.infoLabel.SetText(fmt.Sprintf("图像提取失败: %v", err))
		return
	}

	if img == nil {
		iv.clearImage()
		iv.infoLabel.SetText("图像为空")
		return
	}

	iv.currentImage = img
	iv.imageDisplay.Image = img
	iv.imageDisplay.Refresh()

	// 更新信息
	bounds := img.Bounds()
	iv.infoLabel.SetText(fmt.Sprintf("图像尺寸: %dx%d, 格式: %d",
		bounds.Dx(), bounds.Dy(), wzPng.Form))

	// 启用保存按钮
	iv.saveButton.Enable()

	// 重置缩放
	iv.resetZoom()
}

// clearImage 清空图像显示
func (iv *ImageViewer) clearImage() {
	iv.currentImage = nil
	iv.currentPng = nil
	iv.imageDisplay.Image = nil
	iv.imageDisplay.Refresh()
	iv.saveButton.Disable()
}

// onZoomChanged 缩放变化事件
func (iv *ImageViewer) onZoomChanged(value float64) {
	if iv.currentImage == nil {
		return
	}

	bounds := iv.currentImage.Bounds()
	newWidth := float32(bounds.Dx()) * float32(value)
	newHeight := float32(bounds.Dy()) * float32(value)

	iv.imageDisplay.Resize(fyne.NewSize(newWidth, newHeight))
	iv.scrollContent.Refresh()
}

// resetZoom 重置缩放
func (iv *ImageViewer) resetZoom() {
	iv.zoomSlider.SetValue(1.0)
}

// saveImage 保存图像
func (iv *ImageViewer) saveImage() {
	if iv.currentImage == nil {
		return
	}

	// 创建文件保存对话框
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		// 将图像编码为PNG格式
		var buf bytes.Buffer
		encodeErr := png.Encode(&buf, iv.currentImage)
		if encodeErr != nil {
			dialog.ShowError(fmt.Errorf("图像编码失败: %v", encodeErr),
				fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 写入文件
		_, writeErr := writer.Write(buf.Bytes())
		if writeErr != nil {
			dialog.ShowError(fmt.Errorf("文件保存失败: %v", writeErr),
				fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		dialog.ShowInformation("保存成功", "图像已成功保存",
			fyne.CurrentApp().Driver().AllWindows()[0])

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// 设置默认文件名和过滤器
	saveDialog.SetFileName("image.png")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
	saveDialog.Show()
}

// GetContent 获取图像查看器内容
func (iv *ImageViewer) GetContent() fyne.CanvasObject {
	return iv.content
}
