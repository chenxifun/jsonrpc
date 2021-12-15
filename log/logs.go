package log

import (
	"encoding/json"
	"github.com/chenxifun/logger"
	"time"
)

const (
	timeFormat     = "2006-01-02 15:04:05.000"
	log_source_out = 1
	log_source_in  = 2
)

var cityCode = ""

func SetLogger(param, serviceName, orgCode string) {
	var params []string
	if param != "" {
		params = append(params, param)
	}
	cityCode = orgCode
	logger.SetLogger(params...)
	logger.GetlocalLogger().SetFileLogConvert(BsnLogConvert)
	logger.GetlocalLogger().SetTimeFormat(timeFormat)

	logger.GetlocalLogger().SetCallDepth(3)
	logger.GetlocalLogger().SetAppName(serviceName)
}

func SetServiceInfo(serviceName, cityCode string) {
	logger.GetlocalLogger().SetAppName(serviceName)
}

func BsnLogConvert(t time.Time, msg logger.LogContent) string {

	d := &bsnLogData{
		LogTime:      msg.GetTime(),
		LogLevel:     convLevelName(msg.GetLevel()),
		ServiceName:  msg.GetName(),
		CodeLocation: msg.GetPath(),
		Message:      msg.GetContent(),
		CityCode:     cityCode,
		HostIP:       "",
	}

	e, ok := msg.GetExtendProp().(*passData)

	if ok {
		d.RequestId = e.RequestId
		d.LogSource = e.LogSource
	}

	jb, _ := json.Marshal(d)
	return string(jb)
}

func convLevelName(l string) string {
	//ERROR、WARN、INFO、DEBUG
	//EMER ALRT CRIT EROR WARN INFO DEBG TRAC
	switch l {
	case "EROR":
		return "ERROR"
	case "DEBG":
		return "DEBUG"
	default:
		return l
	}
}

type bsnLogData struct {
	RequestId    string `json:"request_id"`
	LogTime      string `json:"log_time"`
	ServiceName  string `json:"service_name"`
	LogLevel     string `json:"log_level"`
	CodeLocation string `json:"code_location"`
	Message      string `json:"message"`
	LogSource    int    `json:"log_source"`
	CityCode     string `json:"city_code"`
	HostIP       string `json:"host_ip"`
}

type passData struct {
	RequestId string
	LogSource int
}

func defaultLogConvert(t time.Time, msg logger.LogContent) string {

	msgStr := msg.GetTime() + " [" + msg.GetLevel() + "] " + "[" + msg.GetPath() + "] " + msg.GetContent()
	return msgStr
}
