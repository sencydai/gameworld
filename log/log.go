package log

import (
	"bufio"
	"fmt"
	"github.com/sencydai/gameworld/base"
	"os"
	"path"
	"sync"
	"time"
)

type LogLevel = int

//日志等级
const (
	DEBUG_N LogLevel = iota
	INFO_N
	WARN_N
	ERROR_N
	FATAL_N
	TRASH
	CLOSED_N
)

const (
	sDEBUG = "DEBUG"
	sINFO  = "INFO"
	sWARN  = "WARN"
	sERROR = "ERROR"
	sFATAL = "FATAL"

	syncPeriod     = time.Second
	defaultBufSize = 1024 * 1024
)

type ILogger interface {
	SetFileName(fileName string) error
	SetLevel(level LogLevel) bool
	Close()

	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

var (
	Logger = newLogger()

	SetFileName = Logger.SetFileName
	SetLevel    = Logger.SetLevel
	Close       = Logger.Close

	Debug  = Logger.Debug
	Debugf = Logger.Debugf
	Info   = Logger.Info
	Infof  = Logger.Infof
	Warn   = Logger.Warn
	Warnf  = Logger.Warnf
	Error  = Logger.Error
	Errorf = Logger.Errorf
	Fatal  = Logger.Fatal
	Fatalf = Logger.Fatalf
)

var (
	skipLevel = 4
	levelText = map[int]string{DEBUG_N: sDEBUG, INFO_N: sINFO, WARN_N: sWARN, ERROR_N: sERROR, FATAL_N: sFATAL}
)

type logger struct {
	level    LogLevel
	fileName string
	lastTime time.Time

	file   *os.File
	writer *bufio.Writer
	lock   sync.Mutex

	buffers      chan string
	fatalBuffers chan string

	close chan bool
}

type loggerBuffer struct {
	tm      time.Time
	level   LogLevel
	content string
}

func (l *logger) closed() {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.writer != nil {
		l.writer.Flush()
		l.file.Sync()
	}
	l.close <- true
}

func (l *logger) flush(buffer string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.writer == nil {
		return
	}

	now := time.Now()
	if !base.IsSameDay(l.lastTime, now) {
		l.writer.Flush()
		l.file.Sync()

		fileName := fmt.Sprintf("%s_%s.log", l.fileName, now.Format("20060102"))
		if file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err == nil {
			l.file.Close()
			l.lastTime, l.file, l.writer = now, file, bufio.NewWriterSize(file, defaultBufSize)
		}
	}

	l.writer.WriteString(buffer)
}

func (l *logger) sync() {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.writer == nil {
		return
	}

	l.writer.Flush()
	l.file.Sync()
}

func newLogger() *logger {
	logger := &logger{
		level:        DEBUG_N,
		buffers:      make(chan string, 10000),
		fatalBuffers: make(chan string, 1000),
		close:        make(chan bool, 1),
	}
	chError := make(chan bool, 1)
	go func() {
		write := os.Stdout.WriteString
		for buffer := range logger.buffers {
			if len(buffer) == 0 {
				logger.fatalBuffers <- ""
				<-chError
				logger.closed()
				return
			}
			logger.flush(buffer)
			write(buffer)
		}
	}()

	go func() {
		lastTime := time.Now()
		var file *os.File
		for buffer := range logger.fatalBuffers {
			if len(buffer) == 0 {
				chError <- true
				return
			}

			if len(logger.fileName) == 0 {
				continue
			}

			now := time.Now()
			if file == nil || !base.IsSameDay(lastTime, now) {
				if file != nil {
					file.Close()
				}
				
				lastTime = now

				fileName := fmt.Sprintf("%s_%s.error", logger.fileName, lastTime.Format("20060102"))
				os.MkdirAll(path.Dir(fileName), os.ModeDir)
				file, _ = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			}

			file.WriteString(buffer)
			file.Sync()
		}
	}()

	return logger
}

func (l *logger) SetFileName(fileName string) error {
	l.fileName = fileName
	l.lastTime = time.Now()
	fileName = fmt.Sprintf("%s_%s.log", fileName, l.lastTime.Format("20060102"))
	os.MkdirAll(path.Dir(fileName), os.ModeDir)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	l.file, l.writer = file, bufio.NewWriterSize(file, defaultBufSize)

	go func() {
		for {
			select {
			case <-time.After(syncPeriod):
				l.sync()
			}
		}
	}()

	return nil
}

func (l *logger) SetLevel(level LogLevel) bool {
	if level < DEBUG_N || level >= TRASH {
		return false
	}

	l.level = level
	return true
}

func (l *logger) Close() {
	l.buffers <- ""
	<-l.close
}

func (l *logger) writeBufferf(level LogLevel, format string, data ...interface{}) {
	if level >= l.level {
		now := base.FormatDateTime(time.Now())
		fileLine := base.FileLine(skipLevel)
		text := fmt.Sprintf("%s %s [%s] - %s\n", now, levelText[level], fileLine, fmt.Sprintf(format, data...))
		if level == FATAL_N {
			l.fatalBuffers <- text
		}

		l.buffers <- text
	}
}

func (l *logger) writeBuffer(level LogLevel, data ...interface{}) {
	if level >= l.level {
		now := base.FormatDateTime(time.Now())
		fileLine := base.FileLine(skipLevel)
		text := fmt.Sprintf("%s %s [%s] - %s\n", now, levelText[level], fileLine, fmt.Sprint(data...))
		if level == FATAL_N {
			l.fatalBuffers <- text
		}
		l.buffers <- text
	}
}

func (l *logger) Debug(data ...interface{}) {
	l.writeBuffer(DEBUG_N, data...)
}

func (l *logger) Debugf(format string, data ...interface{}) {
	l.writeBufferf(DEBUG_N, format, data...)
}

func (l *logger) Info(data ...interface{}) {
	l.writeBuffer(INFO_N, data...)
}

func (l *logger) Infof(format string, data ...interface{}) {
	l.writeBufferf(INFO_N, format, data...)
}

func (l *logger) Warn(data ...interface{}) {
	l.writeBuffer(WARN_N, data...)
}

func (l *logger) Warnf(format string, data ...interface{}) {
	l.writeBufferf(WARN_N, format, data...)
}

func (l *logger) Error(data ...interface{}) {
	l.writeBuffer(ERROR_N, data...)
}

func (l *logger) Errorf(format string, data ...interface{}) {
	l.writeBufferf(ERROR_N, format, data...)
}

func (l *logger) Fatal(data ...interface{}) {
	l.writeBuffer(FATAL_N, data...)
}

func (l *logger) Fatalf(format string, data ...interface{}) {
	l.writeBufferf(FATAL_N, format, data...)
}
