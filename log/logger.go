package log

type Logger interface {
	Info(format string, a ...interface{})

	Fatal(format string, a ...interface{})

	Error(format string, a ...interface{})

	Debug(format string, a ...interface{})

	Warn(format string, a ...interface{})

	Trace(format string, a ...interface{})
}
