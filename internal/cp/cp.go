package cp

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	regMedium = regexp.MustCompile(`^\.(jpeg|jpg|png|bmp|gif|tiff|tif|pcx|svg|psd|raw|raf|heic|mp4|mov|mkv|rmvb|ts|avi|m4v)$`)
	regIgnore = regexp.MustCompile(`(\.DS_Store|@eaDir)`)
)

type Info struct {
	DirCount int
	Total    int
	Success  int
	Skip     int
	Chrcksum int
	Failure  int
	Ignored  int
	Errors   []error
}

type Channels struct {
	FileChan chan string
	ErrChan  chan error
	DoneChan chan struct{}
	MD5Chan  chan string
}

func XCopy(from, to string, c bool) error {
	startTime := time.Now() // 记录开始时间

	result := &Info{}
	channels := &Channels{
		FileChan: make(chan string),
		ErrChan:  make(chan error, 10), // 增加缓冲区以防止阻塞
		DoneChan: make(chan struct{}),
		MD5Chan:  make(chan string),
	}

	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		go worker(channels, to, c, result)
	}

	go md5Verifier(channels, result)

	err := filepath.Walk(from, func(srcFile string, info os.FileInfo, err error) error {
		if err != nil {
			handleError(fmt.Errorf("failure accessing path %q: %v", srcFile, err), result)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		result.Total++
		if !isMedium(srcFile) || regIgnore.MatchString(srcFile) {
			result.Ignored++
			return nil
		}

		channels.FileChan <- srcFile
		return nil
	})
	close(channels.FileChan)

	if err != nil {
		fmt.Println("xcopy failed: ", err)
	}

	for i := 0; i < numWorkers; i++ {
		<-channels.DoneChan
	}

	close(channels.MD5Chan)
	close(channels.ErrChan)
	for err := range channels.ErrChan {
		result.Errors = append(result.Errors, err)
	}

	endTime := time.Now()                 // 记录结束时间
	elapsedTime := endTime.Sub(startTime) // 计算总共耗时
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Println("加载时区失败:", err)
		return err
	}
	startTimeInBeijing := startTime.In(location)
	endTimeInBeijing := endTime.In(location)

	fmt.Printf("\n✨ 拷贝完成 开始时间: %s 结束时间: %s 总共耗时: %s\n", startTimeInBeijing.Format("2006-01-02 15:04:05"), endTimeInBeijing.Format("2006-01-02 15:04:05"), elapsedTime)

	printSummary(result)
	return nil
}

func worker(channels *Channels, to string, c bool, result *Info) {
	for srcFile := range channels.FileChan {
		if err := processFile(srcFile, to, c, result, channels.MD5Chan); err != nil {
			channels.ErrChan <- err
		}
	}
	channels.DoneChan <- struct{}{}
}

func processFile(srcFile, to string, c bool, result *Info, md5Chan chan<- string) error {
	destFile := getDestAbsPath(to, srcFile)
	destDir := filepath.Dir(destFile)

	if exists, _ := dirExists(destDir); !exists {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return handleError(fmt.Errorf("创建目录失败: %s. %s", destDir, err.Error()), result)
		}
		fmt.Println("创建目录: ", strings.TrimLeft(destDir, "./"))
		result.DirCount++
	}

	if exists, _ := fileExists(destFile); exists {
		srcMD5, _ := calculateMD5(srcFile)
		destMD5, _ := calculateMD5(destFile)
		if srcMD5 == destMD5 {
			if err := modificationTime(srcFile, destFile); err != nil {
				fmt.Println(err)
			}
			fmt.Printf("跳过文件: %s -> %s\n", srcFile, destFile)
			result.Skip++
			return nil
		}
		destFile = fmt.Sprintf("%s_%s%s", strings.TrimSuffix(destFile, path.Ext(destFile)), srcMD5, path.Ext(destFile))
	}

	if c {
		if _, err := copyFile(srcFile, destFile); err != nil {
			return handleError(fmt.Errorf("拷贝文件失败：%s. %v", srcFile, err), result)
		}
		fmt.Printf("拷贝文件: %s -> %s\n", srcFile, destFile)
	} else {
		if err := os.Rename(srcFile, destFile); err != nil {
			return handleError(fmt.Errorf("移动文件失败: %s. %s", srcFile, err.Error()), result)
		}
		fmt.Printf("移动文件: %s -> %s\n", srcFile, destFile)
	}

	md5Chan <- fmt.Sprintf("%s|%s", srcFile, destFile)
	result.Success++
	return nil
}

func md5Verifier(channels *Channels, result *Info) {
	for files := range channels.MD5Chan {
		parts := strings.Split(files, "|")
		srcFile, destFile := parts[0], parts[1]

		srcMD5, _ := calculateMD5(srcFile)
		destMD5, _ := calculateMD5(destFile)
		if srcMD5 != destMD5 {
			fmt.Printf("MD5校验失败: %s %s -> %s\n", srcFile, srcMD5, destMD5)
			result.Chrcksum++
			channels.ErrChan <- fmt.Errorf("MD5校验失败: %s %s -> %s", srcFile, srcMD5, destMD5)
		}
	}
}

func handleError(err error, result *Info) error {
	fmt.Println(err)
	result.Errors = append(result.Errors, err)
	result.Failure++
	return err
}

func printSummary(result *Info) {
	fmt.Printf("目录总数: %d, 文件总数: %d, 成功: %d, 跳过: %d, 失败: %d, 忽略: %d, MD5校验失败: %d\n", result.DirCount, result.Total, result.Success, result.Skip, result.Failure, result.Ignored, result.Chrcksum)
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Printf("%v\n", err)
		}
	}
}

func getPlacePath(tm time.Time) string {
	return fmt.Sprintf("%d/%02d/%d-%02d-%02d", tm.Year(), tm.Month(), tm.Year(), tm.Month(), tm.Day())
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func getDestAbsPath(dest string, srcFile string) string {
	mt := getModifyTime(srcFile)
	path := getPlacePath(mt)

	destPath := filepath.Join(dest, path, filepath.Base(srcFile))
	absPath, _ := filepath.Abs(destPath)
	if absPath == "" {
		return destPath
	}
	return absPath
}

func copyFile(src, des string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	fi, _ := srcFile.Stat()
	perm := fi.Mode()

	desFile, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	written, err = io.Copy(desFile, srcFile)
	if err != nil {
		return 0, err
	}

	// Preserve the modification and access times
	modTime := fi.ModTime()
	err = os.Chtimes(des, modTime, modTime)
	if err != nil {
		return 0, err
	}

	return written, nil
}

func modificationTime(src, des string) error {
	srcT := getModifyTime(src)
	desT := getModifyTime(des)
	if srcT.Equal(desT) {
		return nil
	}

	fmt.Printf("修改时间: %s -> %s\n", des, srcT)
	return os.Chtimes(des, srcT, srcT)
}

func isMedium(fileName string) bool {
	return regMedium.MatchString(strings.ToLower(path.Ext(fileName)))
}

func getModifyTime(file string) time.Time {
	if fi, err := os.Stat(file); err == nil {
		return fi.ModTime()
	}
	return time.Now()
}

func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
