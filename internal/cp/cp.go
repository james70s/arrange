package cp

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	// 从文件名中取出日期格式字符串
	// regDate = regexp.MustCompile(`(20[0-2][0-9])[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})[-|_|\s]?([0-9]{2})`)
	// Medium
	regMedium = regexp.MustCompile(`^\.(jpeg|jpg|png|bmp|gif|tiff|tif|pcx|svg|psd|raw|raf|heic|mp4|mov|mkv|rmvb|ts|avi|m4v)$`)
	regIgnore = regexp.MustCompile(`(\.DS_Store|@eaDir)`) // .DS_Store @eaDir
)

type Info struct {
	DirCount int // 目录总数
	Total    int // 文件总数
	Success  int // 拷贝成功数
	Skip     int // 跳过文件数, 已经存在，MD5相同
	Chrcksum int // 检查数, 文件已经存在，但MD5不同
	Failure  int // 失败数
}

func XCopy(from, to string, c bool) error {
	result := &Info{
		DirCount: 0,
		Total:    0,
		Success:  0,
		Skip:     0,
		Chrcksum: 0,
		Failure:  0,
	}
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

		result.Total++
		// 如果文件名中包含时间信息，是否根据该时间信息重置文件的修改时间
		// if *t {
		// 	setModifyTime(srcFile)
		// }

		// 目标文件,etc: /Users/James/app/tools/arrange/t/2019/06/2019-06-13/181338.jpg
		destFile := getDestAbsPath(to, srcFile)
		destDir := filepath.Dir(destFile)

		if exists, _ := dirExists(destDir); !exists {
			if err = os.MkdirAll(destDir, 0755); err != nil {
				fmt.Printf("创建目录失败: %s. %s\n", destDir, err.Error())
				return err
			}
			fmt.Println("创建目录: ", strings.TrimLeft(destDir, "./"))
			result.DirCount++
		}

		srcMD5, _ := calculateMD5(srcFile) // 计算文件的MD5
		if exists, _ := fileExists(destFile); exists {
			destMD5, _ := calculateMD5(destFile) // 计算文件的MD5
			if srcMD5 == destMD5 {
				fmt.Printf("跳过文件: %s -> %s\n", srcFile, destFile)
				result.Skip++
				return nil
			}
			// 文件已经存在，但MD5不同, 说明是不同的文件，那么需要重命名
			destFile = fmt.Sprintf("%s_%s%s", strings.TrimSuffix(destFile, path.Ext(destFile)), srcMD5, path.Ext(destFile))
		}

		// 拷贝文件
		if c {
			if _, err = copyFile(srcFile, destFile); err != nil {
				fmt.Printf("拷贝文件失败：%s. %v\n", srcFile, err)
				result.Failure++
				return nil
			}
			fmt.Printf("拷贝文件: %s -> %s\n", srcFile, destFile)
		} else { // 移动文件
			if err = os.Rename(srcFile, destFile); err != nil {
				fmt.Printf("移动文件失败: %s. %s\n", srcFile, err.Error())
				result.Failure++
				return nil
			}
			fmt.Printf("移动文件: %s -> %s\n", srcFile, destFile)
		}

		checkMD5, _ := calculateMD5(destFile) // 计算文件的MD5
		if srcMD5 != checkMD5 {
			fmt.Printf("MD5校验失败: %s %s -> %s\n", srcFile, srcMD5, checkMD5)
			result.Chrcksum++
			return nil
		}

		result.Success++
		return nil
	})
	if err != nil {
		log.Fatal("xcopy failed: ", err)
	}

	fmt.Printf("目录总数: %d, 文件总数: %d, 成功数: %d, 跳过数: %d, 失败数: %d, MD5校验失败: %d\n", result.DirCount, result.Total, result.Success, result.Skip, result.Failure, result.Chrcksum)
	return nil
}

// ----------------------------------------------------------------

// 根据照片拍摄日期决定存储目录
// "2016-01-02 15:04:05" -> "2016/01/2016-01-02"
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

// func mkdir(path string) error {
// 	if err := os.MkdirAll(path, 0755); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // 创建目录
// func createPath(destFile string) (err error) {
// 	destDir := filepath.Dir(destFile)

// 	if _, err = os.Stat(destDir); os.IsNotExist(err) {
// 		if err = os.MkdirAll(destDir, 0755); err != nil {
// 			fmt.Printf("创建目录失败: %s. %s\n", destDir, err.Error())
// 			return err
// 		}
// 		fmt.Println("创建目录: ", strings.TrimLeft(destDir, "./"))
// 	}
// 	return nil
// }

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

// func Volumes() {
// 	cmd := exec.Command("ls", "/Volumes")
// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	err := cmd.Run()
// 	if err != nil {
// 		fmt.Println("Error executing command:", err)
// 		return
// 	}
// 	fmt.Println("Diskutil output:\n", out.String())
// }
