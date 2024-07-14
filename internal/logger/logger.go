package logger

import (
	"URLChecker/internal/config"
	"encoding/json"
	"io"
	"log"
	"os"
)

type Loggers struct {
	logInfo    *log.Logger
	logWarn    *log.Logger
	logErr     *log.Logger
	target     io.WriteCloser
	fileTarget bool
}

func New(cfg *config.Config) *Loggers {
	var target io.WriteCloser
	var fileTarget bool
	if cfg.LogsTarget == "stdout" {
		target = os.Stdout
		fileTarget = false
	} else {
		f, err := os.OpenFile(cfg.LogsTarget, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		target = f
		fileTarget = true
	}
	return &Loggers{
		logInfo:    log.New(target, "INFO\t", log.Ldate|log.Ltime),
		logWarn:    log.New(target, "WARN\t", log.Ldate|log.Ltime|log.Lshortfile),
		logErr:     log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		target:     target,
		fileTarget: fileTarget,
	}
}

func (l *Loggers) Info(v ...interface{}) {
	if !l.fileTarget {
		l.logInfo.Println(v...)
		return
	}
	msg, err := json.Marshal(v)
	if err != nil {
		l.logErr.Println("JSON marshal error:", err)
	}
	l.logInfo.Println(string(msg))
}

func (l *Loggers) Warn(v ...interface{}) {
	if !l.fileTarget {
		l.logWarn.Println(v...)
		return
	}
	msg, err := json.Marshal(v)
	if err != nil {
		l.logErr.Println("JSON marshal error:", err)
	}
	l.logWarn.Println(string(msg))
}

func (l *Loggers) Error(v ...interface{}) {
	l.logErr.Println(v...)
}

func (l *Loggers) CloseTarget() {
	if !l.fileTarget {
		return
	}
	err := l.target.Close()
	if err != nil {
		log.Fatal(err)
	}
}
