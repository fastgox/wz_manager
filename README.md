# WZ Manager

A WZ file manager developed with Go and Fyne GUI framework for viewing and managing MapleStory WZ files.

## Features

### File Management
- Load multiple WZ files
- File list management
- Support for .wz file format

### Tree Structure Viewer
- Display complete directory structure of WZ files
- Support for node expand/collapse
- Search functionality
- Node type identification

### Image Viewer
- Display image resources from WZ files
- Zoom functionality
- Save images as PNG format
- Scroll view for large images

### Audio Player
- Audio file information display
- Volume control
- Audio file export
- Support for MP3 and WAV formats

### Data Exporter
- Support multiple export formats: JSON, XML, CSV
- Batch export image files
- Batch export audio files
- Path filtering functionality

## 项目结构

```
wz_manager/
├── main.go              # 主程序入口
├── go.mod              # Go模块文件
├── ui/                 # UI组件目录
│   ├── main_window.go  # 主窗口
│   ├── file_manager.go # 文件管理器
│   ├── tree_viewer.go  # 树形视图
│   ├── image_viewer.go # 图像查看器
│   ├── sound_player.go # 音频播放器
│   └── data_exporter.go# 数据导出器
└── pkg/wzlib/          # WZ文件解析库
```

## 编译和运行

### 前置要求
- Go 1.23.1 或更高版本
- Fyne GUI框架依赖

### 编译
```bash
go mod tidy
go build -o wz_manager.exe .
```

### 运行
```bash
./wz_manager.exe
```

## Usage Instructions

1. **Load WZ Files**
   - Click "Load WZ File" button
   - Select .wz file
   - File will appear in the file list

2. **Browse File Structure**
   - Select a loaded file from the file list
   - Left tree view will display the file structure
   - Click nodes to expand/collapse subdirectories

3. **View Images**
   - Select Canvas type nodes in the tree view
   - Images will be displayed in the right image viewer
   - Use zoom slider to adjust image size
   - Click "Save Image" to export PNG files

4. **Play Audio**
   - Select Sound type nodes
   - View audio information in the audio player tab
   - Click "Save Audio" to export audio files

5. **Export Data**
   - Select export type in the data exporter tab
   - Set path filter conditions
   - Choose export directory
   - Click "Start Export"

## 技术栈

- **Go语言**: 主要编程语言
- **Fyne**: 跨平台GUI框架
- **wzlib**: WZ文件解析库

## 注意事项

- 音频播放功能需要额外的音频库支持
- 大型WZ文件可能需要较长的加载时间
- 建议在加载大量文件时注意内存使用

## 开发计划

- [ ] 完善音频播放功能
- [ ] 添加更多导出格式支持
- [ ] 优化大文件加载性能
- [ ] 添加文件搜索功能
- [ ] 支持WZ文件编辑功能
