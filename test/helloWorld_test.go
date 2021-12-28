package test

import (
	"context"
	"fmt"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/rpc"
	"github.com/chenxifun/jsonrpc/server"
	"testing"
)

func TestRPC(t *testing.T) {
	srv := server.NewServer(config.DefaultConfig())
	apis := []rpc.API{rpc.API{
		Namespace: "test",
		Public:    true,
		Service:   NewHello(),
		Version:   "1.2",
	},
	}

	if err := srv.RegisterServices(apis, []string{"test"}); err != nil {
		t.Fatal(err)
	}

	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
}

func TestCall(t *testing.T) {

	cli, err := rpc.Dial("ws://127.0.0.1:8003")
	if err != nil {
		t.Fatal(err)
	}

	var res string
	err = cli.Call(&res, "test_hello", "yaya")
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	fmt.Println(res)

}

func TestSub(t *testing.T) {

	cli, err := rpc.Dial("ws://127.0.0.1:8003")
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan string)

	sub, err := cli.Subscribe(context.Background(), "test", ch, "subTx", "txData")
	if err != nil {
		t.Fatal(err)
	}

	defer sub.Unsubscribe()

	for {
		select {
		case res := <-ch:
			{
				fmt.Println(res)
			}
		}
	}
}
