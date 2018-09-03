package log

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sencydai/gameworld/base"
)

type loggerFile struct {
	fileName string
	suffix   string
	lastTime time.Time
	file     *os.File
	writer   *bufio.Writer
	lock     sync.Mutex
}

func (l *loggerFile) close() {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.file == nil {
		return
	}
	l.writer.Flush()
	l.file.Sync()
}

func (l *loggerFile) sync() {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.file == nil {
		return
	}
	l.writer.Flush()
	l.file.Sync()
}

func (l *loggerFile) flush(now time.Time, buffer string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.file == nil || !base.IsSameDay(l.lastTime, now) {
		if l.file != nil {
			l.writer.Flush()
			l.file.Sync()
		}

		fileName := fmt.Sprintf("%s_%s.%s", l.fileName, now.Format("20060102"), l.suffix)
		if file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err == nil {
			l.file.Close()
			l.lastTime, l.file, l.writer = now, file, bufio.NewWriterSize(file, defaultBufSize)
		}
	}

	l.writer.WriteString(buffer)
}

func (l *loggerFile) run() {
	go func() {
		t := time.NewTimer(time.Second)
		for {
			select {
			case <-t.C:
				l.sync()
				t.Reset(time.Second)
			}
		}
	}()
}

type loggerPrint struct {
	file   *os.File
	writer *bufio.Writer
	lock   sync.Mutex
}

func (l *loggerPrint) close() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.writer.Flush()
	l.file.Sync()
}

func (l *loggerPrint) sync() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.writer.Flush()
	l.file.Sync()
}

func (l *loggerPrint) flush(buffer string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.writer.WriteString(buffer)
}

func (l *loggerPrint) run() {
	go func() {
		delay := time.Millisecond * 100
		t := time.NewTimer(delay)
		for {
			select {
			case <-t.C:
				l.sync()
				t.Reset(delay)
			}
		}
	}()
}
