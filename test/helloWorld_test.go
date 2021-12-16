package test

import (
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/rpc"
	"github.com/chenxifun/jsonrpc/server"
	"testing"
)

func TestHTTP(t *testing.T) {
	conf := config.Config{
		Vhosts: []string{"*"},
		Hosts:  "",
		Port:   8002,
	}

	srv := server.NewHTTPServer(conf)

	api := rpc.API{
		Namespace: "test",
		Public:    true,
		Service:   NewHello(),
		Version:   "1.0",
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

func TestRPC(t *testing.T) {
	conf := config.Config{
		Vhosts: []string{"*"},
		Cors:   []string{"*"},
		Hosts:  "localhost",
		Port:   8003,
	}

	srv := server.NewServer(conf)

	api := rpc.API{
		Namespace: "test",
		Public:    true,
		Service:   NewHello(),
		Version:   "1.0",
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
