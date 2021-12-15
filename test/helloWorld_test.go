package test

import (
	"github.com/chenxifun/jsonrpc"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/rpc"
	"testing"
)

func TestRpc(t *testing.T) {
	conf := config.Config{
		Vhosts:         []string{"*"},
		HTTPListenAddr: "",
		RPCPort:        8002,
	}

	srv := jsonrpc.NewServer(conf)

	api := rpc.API{
		Namespace: "test",
		Public:    true,
		Service:   &HelloWorld{},
		Version:   "1.0",
	}

	err := srv.RegisterServices(api)

	if err != nil {
		t.Fatal(err)
	}

	err = srv.Start()
	if err != nil {
		t.Fatal(err)
	}

}
