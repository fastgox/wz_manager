package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/luoxk/wzlib"
)

// SoundPlayer 音频播放器
type SoundPlayer struct {
	content      *fyne.Container
	infoLabel    *widget.Label
	playButton   *widget.Button
	stopButton   *widget.Button
	saveButton   *widget.Button
	volumeSlider *widget.Slider
	currentSound *wzlib.WzSound
	isPlaying    bool
}

// NewSoundPlayer 创建新的音频播放器
func NewSoundPlayer() *SoundPlayer {
	sp := &SoundPlayer{
		infoLabel: widget.NewLabel("No audio selected"),
		isPlaying: false,
	}

	sp.createContent()
	return sp
}

// createContent 创建音频播放器内容
func (sp *SoundPlayer) createContent() {
	// 控制按钮
	sp.playButton = widget.NewButton("Play", sp.playSound)
	sp.stopButton = widget.NewButton("Stop", sp.stopSound)
	sp.saveButton = widget.NewButton("Save Audio", sp.saveSound)

	// 初始状态禁用按钮
	sp.playButton.Disable()
	sp.stopButton.Disable()
	sp.saveButton.Disable()

	// 音量控制
	sp.volumeSlider = widget.NewSlider(0.0, 1.0)
	sp.volumeSlider.Value = 0.5
	sp.volumeSlider.OnChanged = sp.onVolumeChanged

	// 控制面板
	controlPanel := container.NewVBox(
		container.NewHBox(sp.playButton, sp.stopButton, sp.saveButton),
		widget.NewLabel("Volume:"),
		sp.volumeSlider,
	)

	// 组装内容
	sp.content = container.NewVBox(
		widget.NewLabel("Audio Player"),
		widget.NewSeparator(),
		sp.infoLabel,
		widget.NewSeparator(),
		controlPanel,
		widget.NewSeparator(),
		widget.NewLabel("Note: Audio playback requires additional audio library support"),
	)
}

// LoadSound 加载音频
func (sp *SoundPlayer) LoadSound(nodeValue interface{}) {
	if wzSound, ok := nodeValue.(*wzlib.WzSound); ok {
		sp.currentSound = wzSound
		sp.loadSoundInfo(wzSound)
	} else {
		sp.clearSound()
		sp.infoLabel.SetText("选中的节点不是音频")
	}
}

// loadSoundInfo 加载音频信息
func (sp *SoundPlayer) loadSoundInfo(wzSound *wzlib.WzSound) {
	// 更新信息显示
	sp.infoLabel.SetText(fmt.Sprintf("音频文件\n长度: %d 字节",
		wzSound.DataLength))

	// 启用控制按钮
	sp.playButton.Enable()
	sp.saveButton.Enable()
}

// clearSound 清空音频
func (sp *SoundPlayer) clearSound() {
	sp.currentSound = nil
	sp.isPlaying = false

	// 禁用控制按钮
	sp.playButton.Disable()
	sp.stopButton.Disable()
	sp.saveButton.Disable()
}

// playSound 播放音频
func (sp *SoundPlayer) playSound() {
	if sp.currentSound == nil {
		return
	}

	// TODO: 实现实际的音频播放功能
	// 这里需要集成音频播放库，如 beep 或其他 Go 音频库

	sp.isPlaying = true
	sp.playButton.Disable()
	sp.stopButton.Enable()

	// 模拟播放状态
	dialog.ShowInformation("播放", "音频播放功能需要额外的音频库支持\n当前为模拟播放状态",
		fyne.CurrentApp().Driver().AllWindows()[0])

	// 模拟播放结束
	sp.stopSound()
}

// stopSound 停止播放
func (sp *SoundPlayer) stopSound() {
	sp.isPlaying = false
	sp.playButton.Enable()
	sp.stopButton.Disable()
}

// onVolumeChanged 音量变化事件
func (sp *SoundPlayer) onVolumeChanged(value float64) {
	// TODO: 实现音量控制
}

// saveSound 保存音频
func (sp *SoundPlayer) saveSound() {
	if sp.currentSound == nil {
		return
	}

	// 创建文件保存对话框
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		// 获取音频数据
		soundData, extractErr := sp.currentSound.ExtractSound()
		if extractErr != nil {
			dialog.ShowError(fmt.Errorf("获取音频数据失败: %v", extractErr),
				fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		// 写入文件
		_, writeErr := writer.Write(soundData)
		if writeErr != nil {
			dialog.ShowError(fmt.Errorf("文件保存失败: %v", writeErr),
				fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		dialog.ShowInformation("保存成功", "音频文件已成功保存",
			fyne.CurrentApp().Driver().AllWindows()[0])

	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// 设置默认文件名和过滤器
	saveDialog.SetFileName("sound.mp3")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp3", ".wav"}))
	saveDialog.Show()
}

// GetContent 获取音频播放器内容
func (sp *SoundPlayer) GetContent() fyne.CanvasObject {
	return sp.content
}
