package logger

import (
	"log"
	"os"
)

type Loggers struct {
	logInfo *log.Logger
	logWarn *log.Logger
	logErr  *log.Logger
}

func New() *Loggers {
	return &Loggers{
		logInfo: log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		logWarn: log.New(os.Stderr, "WARN\t", log.Ldate|log.Ltime|log.Lshortfile),
		logErr:  log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Loggers) Info(v ...interface{}) {
	l.logInfo.Println(v...)
}

func (l *Loggers) Warn(v ...interface{}) {
	l.logWarn.Println(v...)
}

func (l *Loggers) Error(v ...interface{}) {
	l.logErr.Println(v...)
}
