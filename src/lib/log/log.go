package lib_log

import (
	"app/src/constant"
	"fmt"
	"os"
	"path"
	"time"
)

const LogPermission = 0755

func setupDir() error {
	var err error = nil
	if _, err = os.Stat(constant.LOG_DIR); err == nil {
		return nil
	}
	err = os.MkdirAll(constant.LOG_DIR, LogPermission)
	return err
}

func write(filepath, msg string) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, LogPermission)
	if err != nil {
		return
	}
	defer file.Close()
	fmt.Fprintln(file, msg)
}

func log(name, prefix, msg string, args ...interface{}) {
	err := setupDir()
	if err != nil {
		return
	} else if 0 < len(args) {
		msg = fmt.Sprintf(msg, args...)
	}
	filepath := path.Join(constant.LOG_DIR, fmt.Sprintf("%s_%s.log", constant.APP_CODENAME, name))
	timeString := time.Now().Format("2006-01-02 15:04:05")
	write(filepath, fmt.Sprintf("%s %s%s", timeString, prefix, msg))
}

func Debug(msg string, args ...interface{}) {
	if !constant.IsDebug() {
		return
	}
	log("debug", "[DEBUG]", msg, args...)
}

func Info(msg string, args ...interface{}) {
	log("info", "[INFO]", msg, args...)
}

func Error(msg string, args ...interface{}) {
	log("error", "[ERROR]", msg, args...)
}

func Log(filename, msg string, args ...interface{}) {
	if !constant.IsDebug() {
		return
	}
	log(filename, "", msg, args...)
}

type Logger struct {
	Prefix string
}

func NewLogger(prefix string) *Logger {
	return &Logger{
		Prefix: prefix,
	}
}
func (logger *Logger) Debug(msg string, args ...interface{}) {
	Debug("%s %s", logger.Prefix, fmt.Sprintf(msg, args...))
}
func (logger *Logger) Error(msg string, args ...interface{}) {
	Error("%s %s", logger.Prefix, fmt.Sprintf(msg, args...))
}
