package log

import (
	"bufio"
	"fmt"
	"github.com/sencydai/gameworld/base"
	"os"
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
	Opt    = Logger.Opt
	Optf   = Logger.Optf
)

var (
	skipLevel = 4
	levelText = map[int]string{DEBUG_N: sDEBUG, INFO_N: sINFO, WARN_N: sWARN, ERROR_N: sERROR, FATAL_N: sFATAL}
)

type logger struct {
	level LogLevel

	print  *loggerPrint
	record *loggerFile
	opt    *loggerFile
	fatal  *loggerFile

	close chan bool
}

func newLogger() *logger {
	print := &loggerPrint{}
	print.file, print.writer = os.Stdout, bufio.NewWriterSize(os.Stdout, defaultBufSize)

	logger := &logger{
		level: DEBUG_N,
		print: print,
		close: make(chan bool, 1),
	}

	print.run()

	return logger
}

func (l *logger) SetFileName(fileName string) error {
	l.record = &loggerFile{
		fileName: fileName,
		suffix:   "log",
		lastTime: time.Now(),
	}
	l.opt = &loggerFile{
		fileName: fileName,
		suffix:   "opt",
		lastTime: time.Now(),
	}
	l.fatal = &loggerFile{
		fileName: fileName,
		suffix:   "error",
		lastTime: time.Now(),
	}

	l.record.run()
	l.opt.run()
	l.fatal.run()

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
	l.print.close()

	if l.record != nil {
		l.record.close()
	}

	if l.fatal != nil {
		l.fatal.close()
	}

	l.close <- true
}

func (l *logger) writeBufferf(level LogLevel, format string, data ...interface{}) {
	now := time.Now()
	fileLine := base.FileLine(skipLevel)
	text := fmt.Sprintf("%s %s [%s] - %s\n", base.FormatDateTime(now), levelText[level], fileLine, fmt.Sprintf(format, data...))

	l.print.flush(text)

	if level >= l.level {
		if l.record != nil {
			l.record.flush(now, text)
		}

		if level == FATAL_N && l.fatal != nil {
			l.fatal.flush(now, text)
		}
	}
}

func (l *logger) writeBuffer(level LogLevel, data ...interface{}) {
	now := time.Now()
	fileLine := base.FileLine(skipLevel)
	text := fmt.Sprintf("%s %s [%s] - %s\n", base.FormatDateTime(now), levelText[level], fileLine, fmt.Sprint(data...))

	l.print.flush(text)
	
	if level >= l.level {
		if l.record != nil {
			l.record.flush(now, text)
		}

		if level == FATAL_N && l.fatal != nil {
			l.fatal.flush(now, text)
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

func (l *logger) Opt(data ...interface{}) {
	if l.opt != nil {
		now := time.Now()
		l.opt.flush(now, fmt.Sprintf("%s - %s\n", base.FormatDateTime(now), fmt.Sprint(data...)))
	}
}

func (l *logger) Optf(format string, data ...interface{}) {
	if l.opt != nil {
		now := time.Now()
		l.opt.flush(now, fmt.Sprintf("%s - %s\n", base.FormatDateTime(now), fmt.Sprintf(format, data...)))
	}
}
