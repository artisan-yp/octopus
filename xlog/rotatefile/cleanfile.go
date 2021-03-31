package rotatefile

import (
	"io/fs"
	"log"
	stdlog "log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type File struct {
	path    string
	modTime time.Time
}

type FileSlice []*File

func (fs FileSlice) Len() int {
	return len(fs)
}

func (fs FileSlice) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

// The file latest modified is on the first index.
func (fs FileSlice) Less(i, j int) bool {
	return fs[i].modTime.After(fs[j].modTime)
}

func cleanFile(dir, filePrefix string,
	t time.Time, maxAge time.Duration, maxCount int) {
	fileSlice := collectFile(dir, filePrefix)
	sort.Sort(fileSlice)

	if maxCount > 0 {
		fileSlice = cleanByCount(fileSlice, maxCount)
	}

	if maxAge > 0 {
		fileSlice = CleanByAge(fileSlice, t, maxAge)
	}

}

func collectFile(dir, filePrefix string) FileSlice {
	pattern := filepath.Join(dir, filePrefix) + "*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		stdlog.Println(err)
	}

	fileSlice := make(FileSlice, 0)
	for _, file := range matches {
		fileinfo, err := os.Lstat(file)
		// ignore error
		if err != nil ||
			// ignore director
			fileinfo.IsDir() ||
			// ignore symlink
			(fileinfo.Mode()&fs.ModeSymlink) == 0 {
			continue
		}

		fileSlice = append(fileSlice, &File{
			path:    file,
			modTime: fileinfo.ModTime(),
		})
	}

	return fileSlice
}

func cleanByCount(fileSlice FileSlice, maxCount int) FileSlice {
	if maxCount > 0 && len(fileSlice) > maxCount {
		for i := maxCount; i < len(fileSlice); i++ {
			os.Remove(fileSlice[i].path)
			log.Println("remove file:", fileSlice[i].path)
		}
		fileSlice = fileSlice[:maxCount]
	}

	return fileSlice
}

func CleanByAge(fileSlice FileSlice, t time.Time, maxAge time.Duration) FileSlice {
	limit := t.Add(-maxAge)
	i := len(fileSlice) - 1
	for ; i > 0; i-- {
		if fileSlice[i].modTime.After(limit) {
			break
		}

		os.Remove(fileSlice[i].path)
		log.Println("remove file:", fileSlice[i].path)
	}

	return fileSlice[:i+1]
}
