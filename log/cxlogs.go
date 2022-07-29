package log

import (
	"fmt"
	"github.com/chenxifun/logger"
	"strings"
)

var defLog Logger

func DefLogger() Logger {
	if defLog == nil {
		defLog = NewLogs()
	}
	logger.SetLogger(`{
		"Console": {
			"level": "EROR",
			"color": true
		}}`)
	return defLog
}

func NewLogs() Logger {
	bsn := &CXLogs{}
	bsn.pd = &passData{
		RequestId: "",
		LogSource: log_source_in,
	}
	return bsn
}

func NewRequestLogs(requestId string) Logger {

	bsn := &CXLogs{
		RequestId: requestId,
	}

	bsn.pd = &passData{
		RequestId: requestId,
		LogSource: log_source_out,
	}

	return bsn
}

type CXLogs struct {
	RequestId string

	pd *passData
}

func (b *CXLogs) Info(format string, a ...interface{}) {
	msg := b.formatLog(format, a...)
	logger.GetlocalLogger().Info(b.pd, msg)
}
func (b *CXLogs) Fatal(format string, a ...interface{}) {
	msg := b.formatLog(format, a...)
	logger.GetlocalLogger().Fatal(b.pd, msg)
}

func (b *CXLogs) Error(format string, a ...interface{}) {
	msg := b.formatLog(format, a...)
	logger.GetlocalLogger().Error(b.pd, msg)
}

func (b *CXLogs) Debug(format string, a ...interface{}) {
	msg := b.formatLog(format, a...)
	logger.GetlocalLogger().Debug(b.pd, msg)
}

func (b *CXLogs) Warn(format string, a ...interface{}) {
	msg := b.formatLog(format, a...)
	logger.GetlocalLogger().Warn(b.pd, msg)
}

func (b *CXLogs) Trace(format string, a ...interface{}) {
	msg := b.formatLog(format, a...)
	logger.GetlocalLogger().Trace(b.pd, msg)
}

func (b *CXLogs) formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
