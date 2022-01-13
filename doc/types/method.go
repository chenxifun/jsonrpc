package types

type Module struct {
	NameSpace string `json:"name_space"`
	Version   string `json:"version"`
	PkgPath   string
	Name      string
	Methods   []*Method `json:"methods"`
	Describe  string    `json:"describe"`
}

type Method struct {
	Name        string      `json:"name"`
	IsSubscribe bool        `json:"is_Subscribe"`
	Input       []Parameter `json:"input"`
	OutPut      []Parameter `json:"out_put"`
	Describe    string      `json:"describe"`
	Title       string      `json:"title"`
}

type Parameter struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	//Name string `json:"name"`
	Describe    string `json:"describe"`
	ExampleData string `json:"example_data"`
}
