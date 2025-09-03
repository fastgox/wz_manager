package test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/luoxk/wzlib"
	"image/png"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWzHeader(t *testing.T) {
	f, err := wzlib.NewWzFile("E:\\Games\\MapleRoyals\\Skill.wz")
	if err != nil {
		t.Error(err)
	}
	t.Log(f.Loaded)
}

func TestWzLoader(t *testing.T) {
	a := &wzlib.WzStructure{}
	err := a.LoadWzFile("F:\\GAME\\v117_2\\MapleStory\\Skill.wz")
	if err != nil {
		t.Fatal(err)
	}

	// 1. 收集所有 img 节点
	var imgNodes []*wzlib.WzNode
	for _, n := range a.WzNode.Nodes {
		if _, ok := n.Value.(*wzlib.WzImage); ok {
			imgNodes = append(imgNodes, n)
		}
	}
	t.Logf("共找到 img 节点: %d", len(imgNodes))

	// 2. 多线程 TryExtract 并记录耗时
	type imgResult struct {
		Node   *wzlib.WzNode
		Img    *wzlib.WzImage
		UseDur time.Duration
		Err    error
	}
	imgResults := make([]imgResult, len(imgNodes))
	var wg sync.WaitGroup
	for i, n := range imgNodes {
		wg.Add(1)
		go func(i int, n *wzlib.WzNode) {
			defer wg.Done()
			img := n.Value.(*wzlib.WzImage)
			start := time.Now()
			err := img.TryExtract()
			useDur := time.Since(start)
			imgResults[i] = imgResult{Node: n, Img: img, UseDur: useDur, Err: err}
		}(i, n)
	}
	wg.Wait()
	for _, r := range imgResults {
		if r.Err != nil {
			t.Logf("img %s 解压失败: %v", r.Node.Text, r.Err)
		} else {
			t.Logf("img %s 解压耗时: %v", r.Node.Text, r.UseDur)
		}
	}

	// 3. 遍历所有 img 下的 node，收集 Canvas 节点
	var canvasNodes []*wzlib.WzNode
	var collectCanvas func(n *wzlib.WzNode)
	collectCanvas = func(n *wzlib.WzNode) {
		if n.Type == "Canvas" {
			canvasNodes = append(canvasNodes, n)
			return
		}
		if n.Type == "Property" || n.Type == "" {
			for _, child := range n.Nodes {
				collectCanvas(child)
			}
		}
	}
	for _, r := range imgResults {
		if r.Err == nil {
			collectCanvas(r.Img.Node)
		}
	}
	t.Logf("共找到 Canvas 节点: %d", len(canvasNodes))
	if len(canvasNodes) < 10 {
		t.Fatalf("Canvas 节点不足10个")
	}

	// 4. 随机选10个 Canvas，多线程解码输出图片并记录耗时
	rand.Seed(time.Now().UnixNano())
	selected := rand.Perm(len(canvasNodes))[:10]
	var decodeWg sync.WaitGroup
	for i, idx := range selected {
		node := canvasNodes[idx]
		decodeWg.Add(1)
		go func(i int, node *wzlib.WzNode) {
			defer decodeWg.Done()
			start := time.Now()
			pngObj, ok := node.Value.(*wzlib.WzPng)
			if !ok {
				t.Logf("节点 %s 不是 WzPng", node.GetFullPath())
				return
			}
			img, err := pngObj.ExtractImage()
			useDur := time.Since(start)
			if err != nil {
				t.Logf("解码 %s 失败: %v", node.GetFullPath(), err)
				return
			}
			// 输出图片到文件
			outPath := filepath.Join("test_output", fmt.Sprintf("canvas_%d.png", i+1))
			os.MkdirAll("test_output", 0755)
			f, _ := os.Create(outPath)
			defer f.Close()
			png.Encode(f, img)
			t.Logf("解码 %s 用时: %v, 输出: %s", node.GetFullPath(), useDur, outPath)
		}(i, node)
	}
	decodeWg.Wait()
}

func TestOneWzLoader(t *testing.T) {
	a := &wzlib.WzStructure{}
	err := a.LoadWzFile("F:\\GAME\\v117_2\\MapleStory\\Skill.wz")
	if err != nil {
		t.Error(err)
	}
	img_node := a.WzNode.GetNode("130.img/skill/1300002/hit/0/0")
	if img_node != nil {
		if img_node.Type == "Canvas" {
			p, err := img_node.Value.(*wzlib.WzPng).ExtractImage()
			if err != nil {
				t.Error(err)
			}
			// 编码为PNG并转为base64
			var buf bytes.Buffer
			if err := png.Encode(&buf, p); err != nil {
				t.Error(err)
			}
			b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
			fmt.Println(len(b64))
			os.WriteFile("1.png", buf.Bytes(), 0755)
		}
		fmt.Println(img_node.GetFullPath())
	} else {
		t.Error("节点不存在")
	}
}

func TestConcurrentImgExport(t *testing.T) {
	a := &wzlib.WzStructure{}
	err := a.LoadWzFile("F:\\GAME\\v117_2\\冒险岛online\\Skill.wz")
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}

	outputDir := "F:\\GAME\\v117_2\\冒险岛online\\Exported"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		t.Fatalf("创建导出目录失败: %v", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	for _, node := range a.WzNode.Nodes {
		if img, ok := node.Value.(*wzlib.WzImage); ok {
			wg.Add(1)
			go func(n *wzlib.WzNode, image *wzlib.WzImage) {
				defer wg.Done()

				outPath := filepath.Join(outputDir, n.Text)
				f, err := os.Create(outPath)
				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("创建文件失败 [%s]: %v", outPath, err))
					mu.Unlock()
					return
				}
				defer f.Close()

				if _, err := image.Stream.Seek(0, io.SeekStart); err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("seek失败 [%s]: %v", n.Text, err))
					mu.Unlock()
					return
				}

				if _, err := io.Copy(f, image.Stream); err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("写入失败 [%s]: %v", outPath, err))
					mu.Unlock()
					return
				}

				t.Logf("成功导出: %s", outPath)
			}(node, img)
		}
	}

	wg.Wait()

	if len(errors) > 0 {
		for _, e := range errors {
			t.Error(e)
		}
	}
}
