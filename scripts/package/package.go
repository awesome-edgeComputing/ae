package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// 文件头部结构
type FileHeader struct {
	NameLength uint32  // 文件名长度
	Size       uint64  // 文件大小
	Mode       uint32  // 文件权限
	Magic      [4]byte // 魔数，用于验证
}

const (
	magicNumber = "\x00AEB"
	fileMarker  = "\x00AEPKG"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: package <o> <input1> [input2...]")
		os.Exit(1)
	}

	output := os.Args[1]
	inputs := os.Args[2:]

	// 读取主程序
	mainProgram, err := os.ReadFile(inputs[0])
	if err != nil {
		fmt.Printf("Error reading main program: %v\n", err)
		os.Exit(1)
	}

	// 创建输出文件
	out, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	// 写入主程序
	if _, err := out.Write(mainProgram); err != nil {
		fmt.Printf("Error writing main program: %v\n", err)
		os.Exit(1)
	}

	// 写入文件标记
	if _, err := out.Write([]byte(fileMarker)); err != nil {
		fmt.Printf("Error writing file marker: %v\n", err)
		os.Exit(1)
	}

	// 写入文件数量（不包括主程序）
	binary.Write(out, binary.LittleEndian, uint32(len(inputs)-1))

	// 处理其他输入文件
	for _, input := range inputs[1:] {
		in, err := os.Open(input)
		if err != nil {
			fmt.Printf("Error opening input file %s: %v\n", input, err)
			os.Exit(1)
		}

		stat, err := in.Stat()
		if err != nil {
			in.Close()
			fmt.Printf("Error getting file info for %s: %v\n", input, err)
			os.Exit(1)
		}

		// 获取基本文件名（不包含路径）
		baseName := filepath.Base(input)
		// 去掉 .tmp 后缀
		baseName = baseName[:len(baseName)-4]

		// 写入文件头
		header := FileHeader{
			NameLength: uint32(len(baseName)),
			Size:       uint64(stat.Size()),
			Mode:       uint32(stat.Mode()),
		}
		copy(header.Magic[:], magicNumber)

		if err := binary.Write(out, binary.LittleEndian, header); err != nil {
			in.Close()
			fmt.Printf("Error writing header for %s: %v\n", input, err)
			os.Exit(1)
		}

		// 写入文件名（只使用基本文件名）
		if _, err := out.Write([]byte(baseName)); err != nil {
			in.Close()
			fmt.Printf("Error writing filename for %s: %v\n", input, err)
			os.Exit(1)
		}

		// 复制文件内容
		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			fmt.Printf("Error copying content for %s: %v\n", input, err)
			os.Exit(1)
		}

		in.Close()
	}

	fmt.Printf("Successfully packaged %d files into %s\n", len(inputs)-1, output)
}
