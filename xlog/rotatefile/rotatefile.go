package rotatefile

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	program = filepath.Base(os.Args[0])
	pid     = strconv.Itoa(os.Getpid())
)

const (
	DEFAULT_DIR = "./log/"

	DEFAULT_TIME_LAYOUT = "20060102-030405"

	DEFAULT_MAX_SIZE  = 1024 * 1024 * 256
	DEFAULT_MAX_COUNT = 15
	DEFAULT_MAX_AGE   = time.Hour * 24 * 15
	DEFAULT_PEROID    = 24 * time.Hour
)

func New(biz, severity string, options ...Option) (*RotateFile, error) {
	rf := &RotateFile{
		dir:        DEFAULT_DIR,
		biz:        biz,
		severity:   severity,
		clock:      Local,
		peroid:     DEFAULT_PEROID,
		maxSize:    DEFAULT_MAX_SIZE,
		maxCount:   DEFAULT_MAX_COUNT,
		timeLayout: DEFAULT_TIME_LAYOUT,
	}

	rf.withOption(options...)
	rf.cachefixedPart()

	if err := rf.rotate(rf.clock.Now()); err != nil {
		return nil, err
	} else {
		return rf, nil
	}
}

type RotateFile struct {
	// implements io.WriteCloser and improves performance.
	*bufio.Writer
	*os.File

	// current written size.
	curSize int64
	// next rotate time.
	nextRotateTime time.Time

	// filename property.
	dir        string
	biz        string
	severity   string
	timeLayout string

	// time standard
	clock Clock

	// rotate rule.
	maxSize int64
	peroid  time.Duration

	// clean rule.
	maxCount int
	maxAge   time.Duration

	// cache of filename fixed part and symlink
	fixedPart string
	symlink   string
}

func (rf *RotateFile) Write(p []byte) (int, error) {
	now := rf.clock.Now()
	if rf.curSize+int64(len(p)) > rf.maxSize || now.After(rf.nextRotateTime) {
		if err := rf.rotate(now); err != nil {
			log.Println(err)
			return 0, err
		}
	}

	n, err := rf.Writer.Write(p)
	if err == nil {
		rf.curSize += int64(n)
	}
	return n, err
}

func (rf *RotateFile) rotate(t time.Time) error {
	truncTime := t.Truncate(rf.peroid)
	rf.nextRotateTime = truncTime.Add(rf.peroid)

	err := rf.createFile(truncTime)
	if err == nil {
		go cleanFile(rf.dir, rf.fixedPart, t, rf.maxAge, rf.maxCount)
	}

	return err
}

// createFile creates a new log file.
func (rf *RotateFile) createFile(t time.Time) error {
	var filename string
	var fileSize int64
	for i := 0; ; i++ {
		filename = filepath.Join(rf.dir, rf.filename(t, i))

		fileinfo, err := os.Stat(filename)
		if os.IsNotExist(err) {
			fileSize = 0
			break
		} else if err == nil && fileinfo.Size() < int64(rf.maxSize) {
			fileSize = fileinfo.Size()
			break
		}
	}

	err := os.MkdirAll(rf.dir, 0755)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	//symlink := filepath.Join(rf.dir, rf.symlink)
	os.Remove(rf.symlink)
	os.Symlink(filename, rf.symlink)

	if rf.Writer != nil {
		rf.Writer.Flush()
	}
	if rf.File != nil {
		rf.File.Sync()
		rf.File.Close()
	}

	rf.File = file
	rf.curSize = fileSize
	rf.Writer = bufio.NewWriter(rf.File)

	return nil
}

// filename returns a new log file name.
func (rf *RotateFile) filename(t time.Time, index int) string {
	if rf.timeLayout == "" {
		rf.timeLayout = rf.layoutTimeByPeroid()
	}

	timePart := t.Format(rf.timeLayout)
	name := rf.fixedPart + "." + timePart + "." + pid + "." + strconv.Itoa(index)

	return name
}

func (rf *RotateFile) layoutTimeByPeroid() string {
	switch {
	case rf.peroid >= 24*time.Hour:
		return "20060102"
	case rf.peroid >= time.Hour:
		return "20060102-03"
	case rf.peroid >= time.Minute:
		return "20060102-0304"
	default:
		return DEFAULT_TIME_LAYOUT
	}
}

func (rf *RotateFile) cachefixedPart() {
	if program != "" {
		rf.fixedPart = program
	}
	if rf.biz != "" {
		rf.fixedPart += "." + rf.biz
	}
	if rf.severity != "" {
		rf.fixedPart += "." + rf.severity
	}

	rf.fixedPart += ".log"
	rf.symlink = rf.fixedPart
}
