package cp

import (
	"os"
	"path/filepath"
	"time"
)

// func ModifyTime(from, to string) (time.Time, error) {
// 	err := filepath.Walk(from, func(srcFile string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			handleError(fmt.Errorf("failure accessing path %q: %v", srcFile, err), result)
// 			return nil
// 		}
// 		if info.IsDir() {
// 			return nil
// 		}

// 		if !isMedium(srcFile) || regIgnore.MatchString(srcFile) {
// 			return nil
// 		}
// 	})
// }

func CheckModificationTimes(dir string) ([]string, error) {
	var mismatchedFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dirName := filepath.Base(filepath.Dir(path))
		fileModTime := info.ModTime()
		dirTime, err := time.Parse("2006-01-02", dirName)
		// fmt.Printf("dirTime: %s %v %s\n", fileModTime, dirTime, dirName)
		if err != nil {
			return nil // Skip directories that don't match the date format
		}
		if !sameDay(fileModTime, dirTime) {
			mismatchedFiles = append(mismatchedFiles, path)
		}
		return nil
	})
	for _, file := range mismatchedFiles {
		println(file)
	}
	return mismatchedFiles, err
}

func sameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
