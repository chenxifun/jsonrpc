package doc

import (
	go_document "github.com/chenxifun/go-document"
	docty "github.com/chenxifun/go-document/types"
	"github.com/chenxifun/jsonrpc/doc/types"
	"unicode"
)

func BuildDoc(d *go_document.Doc, mods []*types.Module) {

	srv := types.Server{Modules: mods}

	for i, m := range srv.Modules {
		data, ok := d.Packages[m.PkgPath]
		if ok {
			parseModels(srv.Modules[i], data)
		}
	}

}

func parseModels(mod *types.Module, pd *docty.PkgData) {
	sd := pd.FindStruct(mod.Name)
	if sd == nil {
		return
	}
	mod.Describe = sd.Doc

	for i, f := range mod.Methods {
		fd := pd.FindFunc(mod.Name, formatName(f.Name))
		parseMrthod(mod.Methods[i], fd)

	}
}

func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToUpper(ret[0])
	}
	return string(ret)
}

func parseMrthod(f *types.Method, fd *docty.FuncData) []string {
	var pds []string
	if fd == nil {
		return pds
	}
	f.Describe = fd.Description
	f.Title = fd.Title
	for _, p := range fd.Params {

		if p.Field.ID() == "context.Context" {
			continue
		}

		pf := ToParamField(p)

		if pf.Obj {
			pds = append(pds, p.Field.ID())
		}

		f.Input = append(f.Input, pf)
	}

	for _, r := range fd.Results {

		if r.Field.ID() == "error" {
			continue
		}

		pf := ToParamField(r)
		if pf.Obj {
			pds = append(pds, r.Field.ID())
		}

		f.OutPut = append(f.OutPut, pf)
	}

	return pds

}

func ToParamField(p *docty.DeclField) *types.ParamField {
	pf := &types.ParamField{
		Name:     p.Name,
		Obj:      p.Field.IsObj(),
		Describe: p.Doc,
	}
	pf.Type = p.Field.Type()
	pf.DataType = p.Field.ID()

	pf.SetParamField()

	return pf
}
