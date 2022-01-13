package test

import (
	"context"
	"fmt"
	r "github.com/chenxifun/jsonrpc/rpc"
	"strings"
	tt "time"
)

type Td struct {
	// 名称
	Name string `json:"name"`

	//年龄
	Age int `json:"age"`

	Sub subdata `json:"sub"`

	t tt.Time `json:"t"`
}

type subdata struct {
	data string
	tx   chan string
	stop chan bool
}

func (s *subdata) Stop() {
	fmt.Println("s stop")
	s.stop <- true

}

func NewHello() *HelloWorld {

	h := &HelloWorld{}
	h.sub = make(chan *subdata)
	go h.handle()

	return h
}

type HelloSay interface {
	Hello(ctx context.Context, name string) (string, error)
	How(ctx context.Context, what string) (string, error)
	Say(ctx context.Context, name, what string) (string, error)
}

type At string
type Bt func(string) string

// HelloWorld this is Hello Server
type HelloWorld struct {

	// sub chan data
	sub chan *subdata
	// Name 名称
	Name string
	// T 时间
	T tt.Time
}

// Hello 获取一个字符串
func (h *HelloWorld) Hello(ctx context.Context, name string) (string, error) {

	fmt.Println(ctx.Value("User-Agent"))
	return "hello," + name, nil
}

// Say say hello
func (h *HelloWorld) Say(ctx context.Context,
	name,
	what string) (string, error) {
	tt.Sleep(tt.Second)
	fmt.Println(ctx.Value("User-Agent"))
	return fmt.Sprintf("%s : %s", name, what), nil
}

// @title HowWhat 函数名称
// @description   函数的详细描述
// @param ctx 参数注释
// @param what 参数注释
// @param t 参数注释t
// @return 返回值注释
func (h *HelloWorld) HowWhat(ctx context.Context,
	what Td,
	t tt.Time,
	sts []string, stt []Td, ks map[string]Td) (string, error) {

	fmt.Println(ctx.Value("User-Agent"))

	ss := strings.SplitN(what.Name, ":", 2)

	return ss[0], nil
}

func (h *HelloWorld) How(ctx context.Context, what string) (string, error) {

	fmt.Println(ctx.Value("User-Agent"))

	ss := strings.SplitN(what, ":", 2)

	return ss[0], nil
}

func (h *HelloWorld) SubTx(ctx context.Context, txData string) (*r.Subscription, error) {
	notifier, supported := r.NotifierFromContext(ctx)
	if !supported {
		return &r.Subscription{}, r.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		tx := make(chan string)
		s := h.subTx(tx, txData)
		defer s.Stop()
		for {
			select {
			case t := <-tx:
				{
					notifier.Notify(rpcSub.ID, t)
				}
			case <-rpcSub.Err():
				return
			case <-notifier.Closed():
				return
			}

		}
	}()

	return rpcSub, nil
}

func (h *HelloWorld) subTx(tx chan string, data string) *subdata {
	s := &subdata{tx: tx, data: data}
	h.sub <- s
	return s
}

func (h *HelloWorld) handle() {
	for {
		select {
		case s, ok := <-h.sub:
			{
				if ok {
					go h.handleSub(s)
				}
			}
		}
	}
}

func (h *HelloWorld) handleSub(sub *subdata) {

	i := 0
	for {
		select {
		case s := <-sub.stop:
			{

				if s {
					fmt.Println("stop")
					return
				} else {
					fmt.Println("go on")
				}

			}
		default:
			{
				sub.tx <- fmt.Sprintf("%s:%d", sub.data, i)
				fmt.Println("sub to ", sub.data, i)
				i++
				tt.Sleep(tt.Second)
			}

		}

	}

}
