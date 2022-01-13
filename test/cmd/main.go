package main

import (
	"fmt"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/rpc"
	"github.com/chenxifun/jsonrpc/server"
	"github.com/chenxifun/jsonrpc/test"
)

func main() {
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
	hello := test.NewHello()

	api := rpc.API{
		Namespace: "test",
		Public:    true,
		Service:   hello,
		Version:   "1.2",
	}

	err := srv.RegisterService(api)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = srv.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
}
