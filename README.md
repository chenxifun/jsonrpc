# jsonrpc

从ETH拆出来的一个 `json-rpc`框架,支持`http`以及`ws`
使用方式
```go
srv := server.NewServer(config.DefaultConfig())
apis := []rpc.API{rpc.API{
Namespace: "test",
Public:    true,
Service:   NewHello(),
Version:   "1.2",
},
}

if err := srv.RegisterServices(apis, []string{"test"});err != nil {
t.Fatal(err)
}

if err := srv.Start();err != nil {
t.Fatal(err)
}
```

调用示例
```go
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
```
