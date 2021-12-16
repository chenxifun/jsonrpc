package test

import (
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/rpc"
	"github.com/chenxifun/jsonrpc/server"
	"testing"
)

func TestRPC(t *testing.T) {
	conf := config.Config{
		Origins:   []string{"*"},
		Vhosts:    []string{"*"},
		Cors:      []string{"*"},
		Hosts:     "localhost",
		Port:      8003,
		EnableRPC: true,
		EnableWS:  false,
	}

	srv := server.NewServer(conf)
	hello := NewHello()

	var sayApi HelloSay = hello
	api := rpc.API{
		Namespace: "test",
		Public:    true,
		Service:   sayApi,
		Version:   "1.2",
	}

	err := srv.RegisterService(api)

	if err != nil {
		t.Fatal(err)
	}

	err = srv.Start()
	if err != nil {
		t.Fatal(err)
	}
}
