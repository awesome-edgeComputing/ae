package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// 版本信息，由构建脚本通过 -ldflags 设置
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// FileHeader 定义与打包工具相同的文件头结构
type FileHeader struct {
	NameLength uint32
	Size       uint64
	Mode       uint32
	Magic      [4]byte
}

const (
	magicNumber = "\x00AEB"
	fileMarker  = "\x00AEPKG"
)

func findFileMarker(f *os.File) (int64, error) {
	// 读取文件内容
	content, err := io.ReadAll(f)
	if err != nil {
		return 0, err
	}

	// 查找文件标记
	marker := []byte(fileMarker)
	if idx := bytes.Index(content, marker); idx != -1 {
		return int64(idx + len(marker)), nil
	}

	return 0, fmt.Errorf("file marker not found")
}

func getCommandName(baseCmd string) []string {
	// 基本命令名
	cmdNames := []string{baseCmd}

	// 添加带后缀的命令名
	if runtime.GOOS == "windows" {
		cmdNames = append(cmdNames, baseCmd+".exe")
	}

	// 添加带平台后缀的命令名
	platformSuffix := fmt.Sprintf("_%s_%s", runtime.GOOS, runtime.GOARCH)
	cmdNames = append(cmdNames,
		baseCmd+platformSuffix,
		"dist/"+baseCmd+platformSuffix,
		baseCmd)
	if runtime.GOOS == "windows" {
		cmdNames = append(cmdNames, baseCmd+platformSuffix+".exe")
	}

	return cmdNames
}

func extractAndRun(command string, args []string) error {
	// 获取当前可执行文件
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}

	// 打开自身文件
	f, err := os.Open(exe)
	if err != nil {
		return fmt.Errorf("error opening self: %v", err)
	}
	defer f.Close()

	// 查找文件标记
	offset, err := findFileMarker(f)
	if err != nil {
		return fmt.Errorf("error finding file marker: %v", err)
	}

	// 移动到标记后的位置
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("error seeking to data: %v", err)
	}

	// 读取文件数量
	var numFiles uint32
	if err := binary.Read(f, binary.LittleEndian, &numFiles); err != nil {
		return fmt.Errorf("error reading file count: %v", err)
	}

	// 获取可能的命令名称
	cmdNames := getCommandName(command)

	// 查找并提取目标命令
	for i := uint32(0); i < numFiles; i++ {
		var header FileHeader
		if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
			return fmt.Errorf("error reading header: %v", err)
		}

		// 验证魔数
		if string(header.Magic[:]) != magicNumber {
			return fmt.Errorf("invalid magic number")
		}

		// 读取文件名
		nameBytes := make([]byte, header.NameLength)
		if _, err := io.ReadFull(f, nameBytes); err != nil {
			return fmt.Errorf("error reading filename: %v", err)
		}
		fileName := string(nameBytes)

		// 检查是否是目标命令
		found := false
		for _, cmdName := range cmdNames {
			baseName := filepath.Base(fileName)
			if strings.HasSuffix(baseName, cmdName) {
				found = true
				break
			}
		}

		if found {
			// 创建临时目录
			tmpDir := filepath.Join(os.TempDir(), "ae-"+command)
			os.MkdirAll(tmpDir, 0755)

			// 创建临时文件
			tmpFile := filepath.Join(tmpDir, filepath.Base(fileName))
			out, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("error creating temp file: %v", err)
			}

			// 复制文件内容
			if _, err := io.CopyN(out, f, int64(header.Size)); err != nil {
				out.Close()
				return fmt.Errorf("error copying file content: %v", err)
			}
			out.Close()

			// 执行命令
			cmd := exec.Command(tmpFile, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Env = append(os.Environ(), fmt.Sprintf("RUST_BACKTRACE=1"))

			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
				}
				return fmt.Errorf("error running command: %v", err)
			}

			// 清理临时文件
			os.RemoveAll(tmpDir)

			return nil
		}

		// 跳过文件内容
		if _, err := f.Seek(int64(header.Size), io.SeekCurrent); err != nil {
			return fmt.Errorf("error seeking file: %v", err)
		}
	}

	return fmt.Errorf("command %s not found", command)
}

func printVersion() {
	fmt.Printf("AE CLI %s (%s)\n", Version, GitCommit)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ae <command> [args...]")
		fmt.Println("Available commands:")
		fmt.Println("  esk  - Kubernetes cluster management tools")
		fmt.Println("  ei2  - Edge AI inference infrastructure")
		fmt.Println("  sys  - System utilities")
		fmt.Println()
		fmt.Println("Global flags:")
		fmt.Println("  -v, --version  Print version information")
		os.Exit(1)
	}

	// 检查版本标志
	if os.Args[1] == "-v" || os.Args[1] == "--version" {
		printVersion()
		return
	}

	// 获取命令和参数
	cmd := os.Args[1]
	args := os.Args[2:]

	// 执行命令
	if err := extractAndRun(cmd, args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
