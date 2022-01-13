package rpc

import (
	"github.com/chenxifun/jsonrpc/doc/types"
)

func NewMethod(name string, cb *callback) *types.Method {

	m := &types.Method{
		Name:        name,
		IsSubscribe: cb.isSubscribe,
	}
	//fntype := cb.fn.Type()
	//
	//if fntype.NumIn() >2 {
	//	for i:=2;i<fntype.NumIn();i++ {
	//
	//		inTy :=fntype.In(i)
	//		kd :=inTy.Kind()
	//
	//		in :=types.Parameter{
	//			ID:inTy.Name(),
	//			Type: kd.String(),
	//		}
	//		m.Input = append(m.Input,in)
	//	}
	//}
	//
	//if fntype.NumOut() ==2 {
	//	outTy :=fntype.Out(1)
	//	kd :=outTy.Kind()
	//
	//	out :=types.Parameter{
	//		ID:outTy.Name(),
	//		Type: kd.String(),
	//	}
	//	m.OutPut = append(m.Input,out)
	//}

	return m

}
