package log

type LogConfig struct {
	TimeFormat string         `json:"TimeFormat"`
	Console    *ConsoleLogger `json:"Console,omitempty"`
	File       *FileLogger    `json:"File,omitempty"`
}

type ConsoleLogger struct {
	Level    string `json:"level"`
	Colorful bool   `json:"color"`
	LogLevel int
}

type FileLogger struct {
	Filename   string `json:"filename"`
	Append     bool   `json:"append"`
	MaxLines   int    `json:"maxlines"`
	MaxSize    int    `json:"maxsize"`
	Daily      bool   `json:"daily"`
	MaxDays    int64  `json:"maxdays"`
	Level      string `json:"level"`
	PermitMask string `json:"permit"`
}
