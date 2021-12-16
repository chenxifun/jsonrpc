package test

import (
	"context"
	"fmt"
	"github.com/chenxifun/jsonrpc/rpc"
	"strings"
	"time"
)

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

type HelloWorld struct {
	sub chan *subdata
}

func (h *HelloWorld) Hello(ctx context.Context, name string) (string, error) {

	fmt.Println(ctx.Value("User-Agent"))
	return "hello," + name, nil
}

func (h *HelloWorld) Say(ctx context.Context, name, what string) (string, error) {

	time.Sleep(time.Second)
	fmt.Println(ctx.Value("User-Agent"))
	return fmt.Sprintf("%s : %s", name, what), nil
}

func (h *HelloWorld) How(ctx context.Context, what string) (string, error) {

	fmt.Println(ctx.Value("User-Agent"))

	ss := strings.SplitN(what, ":", 2)

	return ss[0], nil
}

func (h *HelloWorld) SubTx(ctx context.Context, txData string) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
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
				time.Sleep(time.Second)
			}

		}

	}

}
