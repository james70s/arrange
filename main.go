package main

// https://blog.csdn.net/whatday/article/details/109287416

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/james70s/arrange/internal/gui"
	"github.com/james70s/arrange/internal/ver"
)

// 编译的时候通过 -ldflags "-X main.Version=0.0.1 -X main.Build=7c033ce" 传入
var Version = "0.0.1"
var Build = "7c033ce"

// 实际中应该用更好的变量名
var (
	h = flag.Bool("h", false, "This `help`")
	c = flag.Bool("c", false, "是拷贝还是移动文件.默认为移动文件.")
	// t = flag.Bool("t", false, "如果文件名中包含时间信息，是否根据该时间信息重置文件的修改时间.")

	// 从文件名中取出日期格式字符串
	// regDate = regexp.MustCompile(`(20[0-2][0-9])[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})`)
	// Medium
	regMedium = regexp.MustCompile(`^\.(jpeg|jpg|png|bmp|gif|tiff|tif|pcx|svg|psd|raw|raf|heic|mp4|mov|mkv|rmvb|ts|avi|m4v)$`)
	regIgnore = regexp.MustCompile(`(\.DS_Store|@eaDir)`) // .DS_Store @eaDir
)

func usage() {
	ver.Info()

	fmt.Fprintf(os.Stderr, `
Usage: main [-hct] [from] [to] 

Etc: main -t /Volumes/Untitled\ 1/DCIM /Volumes/home/Photos/PhotoLibrary

Options:
`)
	flag.PrintDefaults()
}

func init() {
	ver.Build = Build
	ver.Version = Version

	flag.Usage = usage // 改变默认的 Usage
	flag.Parse()       // 接受命令行参数
}

func main() {

	// if *h || flag.NArg() != 2 { // 该应用的命令行参数必须要有2个
	// 	flag.Usage()
	// 	return
	// }
	// ver.Info()

	// workPath(flag.Args()[0], flag.Args()[1]) // 运行主程序

	gui.Run()
}

func workPath(from, to string) {
	// 遍历输入文件夹
	err := filepath.Walk(from, func(srcFile string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", srcFile, err)
			return err
		}
		if info.IsDir() { // 忽略子目录
			return nil
		}
		if !isMedium(srcFile) || regIgnore.MatchString(srcFile) { // is not pic or movie, ignore it
			return nil
		}

		// 如果文件名中包含时间信息，是否根据该时间信息重置文件的修改时间
		// if *t {
		// 	setModifyTime(srcFile)
		// }

		// 目标文件,etc: /Users/James/app/tools/arrange/t/2019/06/2019-06-13/181338.jpg
		destFile := getDestAbsPath(to, srcFile)
		if err = createPath(destFile); err != nil {
			return nil
		}

		// 拷贝文件
		if *c {
			if _, err = copyFile(srcFile, destFile); err != nil {
				fmt.Printf("拷贝文件失败：%s. %v\n", srcFile, err)
				return nil
			}
			fmt.Printf("拷贝文件: %s -> %s\n", srcFile, destFile)
			// os.Remove(path)  // 删除源文件
		} else { // 移动文件
			if err = os.Rename(srcFile, destFile); err != nil {
				fmt.Printf("移动文件失败: %s. %s\n", srcFile, err.Error())
				return nil
			}
			fmt.Printf("移动文件: %s -> %s\n", srcFile, destFile)
		}

		return nil
	})
	if err != nil {
		log.Fatal("filepath.Walk failed; detail: ", err)
	}
}

// ----------------------------------------------------------------

// 根据照片拍摄日期决定存储目录
// "2016-01-02 15:04:05" -> "2016/01/2016-01-02"
func getPlacePath(tm time.Time) string {
	return fmt.Sprintf("%d/%02d/%d-%02d-%02d", tm.Year(), tm.Month(), tm.Year(), tm.Month(), tm.Day())
}

// 创建目录
func createPath(destFile string) (err error) {
	destDir := filepath.Dir(destFile)

	if _, err = os.Stat(destDir); os.IsNotExist(err) {
		if err = os.MkdirAll(destDir, 0755); err != nil {
			fmt.Printf("创建目录失败: %s. %s\n", destDir, err.Error())
			return err
		}
		fmt.Println("创建目录: ", strings.TrimLeft(destDir, "./"))
	}
	return nil
}

// 根据文件的修改时间，获取文件将要存放的目录
// dest: ./t 目标路径
// srcFile: test/2019-06-13 181338.jpg 原文件
func getDestAbsPath(dest string, srcFile string) string {
	mt := getModifyTime(srcFile) // 文件修改日期
	path := getPlacePath(mt)     // 生成存放路径

	destPath := filepath.Join(dest, path, filepath.Base(srcFile))
	absPath, _ := filepath.Abs(destPath)
	if absPath == "" {
		return destPath
	}
	return absPath
}

// 拷贝文件
func copyFile(src, des string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	//获取源文件的权限
	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	desFile, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm) // 复制源文件的所有权限
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}

// 是否是媒体文件
func isMedium(fileName string) bool {
	return regMedium.MatchString(strings.ToLower(path.Ext(fileName)))
}

// Time 字符串 -> 时间
// func toTime(s string) (time.Time, error) {
// 	return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
// }

// 获取文件的修改时间
func getModifyTime(file string) time.Time {
	if fi, err := os.Stat(file); err == nil {
		return fi.ModTime()
	}
	return time.Now()
}

// 如果文件名中包含时间信息，那么根据该时间信息重置文件的修改时间，设置正确的modifyTime
// func setModifyTime(srcFile string) {
// 	fileName := filepath.Base(srcFile)             // 文件名
// 	matchs := regDate.FindStringSubmatch(fileName) // 测试文件名是否包含日期信息
// 	if matchs != nil {
// 		omt := getModifyTime(srcFile) // 老的信息

// 		date := fmt.Sprintf("%s-%s-%s %s:%s:%s", matchs[1], matchs[2], matchs[3], matchs[4], matchs[5], matchs[6])
// 		mt, _ := toTime(date)
// 		// fmt.Printf("%s %s %s\n", fileName, date, mt)
// 		// 设置正确的modifyTime
// 		if err := os.Chtimes(srcFile, time.Now(), mt); err == nil {
// 			fmt.Printf("重设修改时间：%s %s -> %s\n", fileName, omt, mt)
// 		} else {
// 			log.Fatal(err)
// 		}
// 	}
// }

func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
