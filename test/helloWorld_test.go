package test

import (
	"context"
	"fmt"
	go_document "github.com/chenxifun/go-document"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/rpc"
	"github.com/chenxifun/jsonrpc/server"
	"github.com/clearcodecn/swaggos"
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

func TestDoc(t *testing.T) {
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

	doc := &go_document.Doc{}

	doc.SetBaseDir("D:\\GoPath\\src\\").AddPkgPath("github.com\\chenxifun\\jsonrpc\\test") //.AddPkgPath("github.com\\chenxifun\\jsonrpc\\rpc")

	doc.Build()

	srv.BuildDoc(doc)
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

func TestSwagg(t *testing.T) {

	y := swaggos.NewSwaggo()
	v := NewHello()
	y.Define(v)
	data, _ := y.Build()
	fmt.Println(string(data))
}
