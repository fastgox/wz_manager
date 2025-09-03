# WZ Manager Usage Guide

## Overview
WZ Manager is a GUI application for viewing and managing MapleStory WZ files. The interface is organized into several panels:

- **Left Panel**: File Manager and Tree Viewer
- **Right Panel**: Tabbed interface with Image Viewer, Audio Player, and Data Exporter

## Getting Started

### 1. Loading WZ Files
1. Click the "Load WZ File" button in the File Manager panel
2. Select a .wz file from your system
3. The file will appear in the file list
4. Click on a file in the list to load its structure

### 2. Browsing File Structure
- Once a file is selected, the Tree Viewer will display the file structure
- Use "Expand All" and "Collapse All" buttons to control the tree view
- Use the search box to find specific nodes
- Click on nodes to explore the hierarchy

### 3. Viewing Images
1. Navigate to a Canvas type node in the tree
2. Click on the node to select it
3. Switch to the "Image Viewer" tab
4. The image will be displayed with zoom controls
5. Use the zoom slider to adjust the image size
6. Click "Save Image" to export as PNG

### 4. Audio Files
1. Navigate to a Sound type node in the tree
2. Click on the node to select it
3. Switch to the "Audio Player" tab
4. View audio file information
5. Click "Save Audio" to export the audio file

### 5. Data Export
1. Switch to the "Data Exporter" tab
2. Select the export type (JSON, XML, CSV, Image Files, Audio Files)
3. Set path filters if needed
4. Choose an export directory
5. Click "Start Export" to begin the export process

## File Management
- **Load WZ File**: Add new WZ files to the manager
- **Remove File**: Remove selected files from the list
- **Clear List**: Remove all files from the list

## Tips
- The application supports multiple WZ files loaded simultaneously
- Use the search function in the tree viewer to quickly find specific nodes
- Image zoom can be reset using the "Reset Zoom" button
- Export functions support batch operations for multiple files

## Troubleshooting
- If images don't display, ensure the node is of type "Canvas"
- Audio playback requires additional audio library support
- Large WZ files may take time to load completely
- Export operations may take time depending on file size and content

## Supported Formats
- **Input**: .wz files (MapleStory data files)
- **Image Export**: PNG format
- **Audio Export**: Original format (MP3, WAV, etc.)
- **Data Export**: JSON, XML, CSV formats
