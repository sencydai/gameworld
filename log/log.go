package log

import (
	"bufio"
	"fmt"
	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/utils"
	"os"
	"path"
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

	syncPeriod     = time.Millisecond * 100
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
	file     *os.File
	writer   *bufio.Writer

	buffers chan string
	close   chan bool
}

type loggerBuffer struct {
	tm      time.Time
	level   LogLevel
	content string
}

func newLogger() *logger {
	logger := &logger{level: DEBUG_N, buffers: make(chan string, 10000), close: make(chan bool, 1)}
	go func() {
		var buffer string
		write := os.Stdout.WriteString
		var logCount int
		for {
			select {
			case buffer = <-logger.buffers:
				if len(buffer) == 0 {
					if logger.writer != nil {
						logger.writer.Flush()
						logger.file.Sync()
					}

					logger.close <- true
					return
				}

				write(buffer)
				if logger.writer != nil {
					logger.writer.WriteString(buffer)
					logCount++
					if logCount > 1000 {
						logger.writer.Flush()
						logger.file.Sync()
						logCount = 0
					}
				}
			case <-time.After(syncPeriod):
				if logCount > 0 {
					logger.writer.Flush()
					logger.file.Sync()
					logCount = 0
				}

				if logger.writer != nil {
					now := time.Now()
					if !base.IsSameDay(logger.lastTime, now) {
						fileName := fmt.Sprintf("%s_%s.log", logger.fileName, now.Format("20060102"))
						if file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err == nil {
							logger.file.Close()
							logger.lastTime, logger.file, logger.writer = now, file, bufio.NewWriterSize(file, defaultBufSize)
						}
					}
				}
			}
		}
	}()

	return logger
}

func (l *logger) SetFileName(fileName string) error {
	l.fileName = fileName
	l.lastTime = time.Now()
	fileName = fmt.Sprintf("%s_%s.log", fileName, l.lastTime.Format("20060102"))
	os.MkdirAll(path.Dir(fileName), os.ModeDir)
	if file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return err
	} else {
		l.file, l.writer = file, bufio.NewWriterSize(file, defaultBufSize)
		return nil
	}
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
		fileLine := utils.FileLine(skipLevel)
		text := levelText[level]
		l.buffers <- fmt.Sprintf("%s %s [%s] - %s\n", now, text, fileLine, fmt.Sprintf(format, data...))
		if level == FATAL_N {
			//l.buffers <- string(debug.Stack())
		}
	}
}

func (l *logger) writeBuffer(level LogLevel, data ...interface{}) {
	if level >= l.level {
		now := base.FormatDateTime(time.Now())
		fileLine := utils.FileLine(skipLevel)
		text := levelText[level]
		l.buffers <- fmt.Sprintf("%s %s [%s] - %s\n", now, text, fileLine, fmt.Sprint(data...))
		if level == FATAL_N {
			//l.buffers <- string(debug.Stack())
		}
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
