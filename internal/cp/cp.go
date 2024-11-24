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

func XCopy(from, to string, c bool) error {
	result := &Info{}
	fileChan := make(chan string)
	errChan := make(chan error)
	doneChan := make(chan struct{})

	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		go worker(fileChan, errChan, doneChan, to, c, result)
	}

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

		fileChan <- srcFile
		return nil
	})
	close(fileChan)

	if err != nil {
		fmt.Println("xcopy failed: ", err)
	}

	for i := 0; i < numWorkers; i++ {
		<-doneChan
	}

	close(errChan)
	for err := range errChan {
		result.Errors = append(result.Errors, err)
	}

	printSummary(result)
	return nil
}

func worker(fileChan <-chan string, errChan chan<- error, doneChan chan<- struct{}, to string, c bool, result *Info) {
	for srcFile := range fileChan {
		if err := processFile(srcFile, to, c, result); err != nil {
			errChan <- err
		}
	}
	doneChan <- struct{}{}
}

func processFile(srcFile, to string, c bool, result *Info) error {
	destFile := getDestAbsPath(to, srcFile)
	destDir := filepath.Dir(destFile)

	if exists, _ := dirExists(destDir); !exists {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return handleError(fmt.Errorf("创建目录失败: %s. %s", destDir, err.Error()), result)
		}
		fmt.Println("创建目录: ", strings.TrimLeft(destDir, "./"))
		result.DirCount++
	}

	srcMD5, _ := calculateMD5(srcFile)
	if exists, _ := fileExists(destFile); exists {
		destMD5, _ := calculateMD5(destFile)
		if srcMD5 == destMD5 {
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

	checkMD5, _ := calculateMD5(destFile)
	if srcMD5 != checkMD5 {
		fmt.Printf("MD5校验失败: %s %s -> %s\n", srcFile, srcMD5, checkMD5)
		result.Chrcksum++
		return nil
	}

	result.Success++
	return nil
}

func handleError(err error, result *Info) error {
	fmt.Println(err)
	result.Errors = append(result.Errors, err)
	result.Failure++
	return err
}

func printSummary(result *Info) {
	fmt.Printf("目录总数: %d, 文件总数: %d, 成功数: %d, 跳过数: %d, 失败数: %d, 忽略数: %d, MD5校验失败: %d\n", result.DirCount, result.Total, result.Success, result.Skip, result.Failure, result.Ignored, result.Chrcksum)
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			fmt.Println(err)
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

	return io.Copy(desFile, srcFile)
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
