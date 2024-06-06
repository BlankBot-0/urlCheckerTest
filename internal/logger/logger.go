package logger

import "log"

type Loggers struct {
	logInfo *log.Logger
	logWarn *log.Logger
	logErr  *log.Logger
}

func New(logInfo, logWarn, logError *log.Logger) *Loggers {
	return &Loggers{
		logInfo: logInfo,
		logWarn: logWarn,
		logErr:  logError,
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
