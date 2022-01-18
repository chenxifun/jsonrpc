package types

import (
	"fmt"
	"time"
)

type Server struct {
	ServiceName string       `json:"service_name"`
	Links       []string     `json:"links"`
	Modules     []*Module    `json:"modules"`
	ParamDatas  []*ParamData `json:"param_datas"`
	ErrorCodes  []*ErrorDes  `json:"error_codes"`
}

type ErrorDes struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Describe string `json:"describe"`
}

type ParamData struct {
	Name     string        `json:"name"`
	Describe string        `json:"describe"`
	PkgPath  string        `json:"pkg_path"`
	Fields   []*ParamField `json:"fields"`
}

type Module struct {
	NameSpace string    `json:"name_space"`
	Version   string    `json:"version"`
	PkgPath   string    `json:"pkg_path"`
	Name      string    `json:"name"`
	Methods   []*Method `json:"methods"`
	Describe  string    `json:"describe"`
}

type Method struct {
	Title        string        `json:"title"`
	Name         string        `json:"name"`
	Describe     string        `json:"describe"`
	IsSubscribe  bool          `json:"is_Subscribe"`
	Input        []*ParamField `json:"input"`
	OutPut       []*ParamField `json:"out_put"`
	RequestBody  string        `json:"request_body"`
	ResponseBody string        `json:"response_body"`
}

type ParamField struct {
	Name     string `json:"name"`
	Obj      bool   `json:"obj"`
	Type     string `json:"type"`
	DataType string `json:"data_type"`
	Example  string `json:"example"`
	Describe string `json:"describe"`
}

func (p *ParamField) SetParamField() {
	d := NewExampleData(p.DataType)
	switch p.Type {
	case "Obj":
		p.Example = d
	case "Array":
		p.Example = fmt.Sprintf("[%s,%s]", d, d)
	case "Map":
		p.Example = fmt.Sprintf("{\"key\":%s}", d)
	default:
		p.Example = d
	}
}

func NewExampleData(kind string) string {

	switch kind {
	case "bool":
		return "true"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "123"
	case "float32", "float64":
		return "1.23"
	case "string":
		return "\"string 123\""
	case "time.Time":
		return "\"" + time.Now().String() + "\""
	default:
		return kind
	}

}
