package main

// https://blog.csdn.net/whatday/article/details/109287416

import (
	"flag"
	"fmt"
	"os"

	"github.com/james70s/arrange/internal/cp"
	"github.com/james70s/arrange/internal/ver"
)

// 编译的时候通过 -ldflags "-X main.Version=0.0.1 -X main.Build=7c033ce" 传入
var Version = "0.0.1"
var Build = "7c033ce"

// 实际中应该用更好的变量名
var (
	h     = flag.Bool("h", false, "This `help`")
	c     = flag.Bool("c", true, "是拷贝还是移动文件, 默认为拷贝文件.")
	check = flag.Bool("check", false, "检查文件的修改时间是否与目录名中的时间是同一天.")
	// t = flag.Bool("t", false, "如果文件名中包含时间信息，是否根据该时间信息重置文件的修改时间.")
)

func usage() {
	ver.Info()

	fmt.Fprintf(os.Stderr, `
Usage: main [-hc] [from] [to] 

Etc: main -c /Volumes/Untitled /Volumes/home/Photos/PhotoLibrary

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

	if *h { // 该应用的命令行参数必须要有2个
		flag.Usage()
		return
	}
	ver.Info()

	if *check {
		cp.CheckModificationTimes(flag.Args()[0])
		return
	}
	cp.XCopy(flag.Args()[0], flag.Args()[1], *c)
	// Volumes()
	// gui.Run()
	// form.Setup()
}
