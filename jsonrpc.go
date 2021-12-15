package jsonrpc

import (
	"context"
	"fmt"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/node"
	"github.com/chenxifun/jsonrpc/rpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func NewServer(conf config.Config) *rpcServer {

	return &rpcServer{
		conf: &conf,
		srv:  rpc.NewServer(),
	}
}

type rpcServer struct {
	srv  *rpc.Server
	conf *config.Config
}

func (rpc *rpcServer) RegisterServices(api rpc.API) error {
	return rpc.srv.RegisterName(api.Namespace, api.Service)
}

func (rpc *rpcServer) Start() error {
	httpEndpoint := fmt.Sprintf("%s:%d", rpc.conf.HTTPListenAddr, rpc.conf.RPCPort)

	handler := node.NewHTTPHandlerStack(rpc.srv, rpc.conf.Cors, rpc.conf.Vhosts)

	httpServer, _, err := rpc.startHTTPEndpoint(httpEndpoint, handler)
	if err != nil {
		return err
	}

	defer func() {
		// Don't bother imposing a timeout here.
		httpServer.Shutdown(context.Background())

	}()

	abortChan := make(chan os.Signal, 11)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan
	fmt.Println("Exiting...", "signal", sig)

	return nil
}

func (rpc *rpcServer) startHTTPEndpoint(endpoint string, handler http.Handler) (*http.Server, net.Addr, error) {
	// start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return nil, nil, err
	}
	// make sure timeout values are meaningful
	CheckTimeouts(rpc.conf)
	// Bundle and start the HTTP server
	httpSrv := &http.Server{
		Handler:      handler,
		ReadTimeout:  rpc.conf.ReadTimeout,
		WriteTimeout: rpc.conf.WriteTimeout,
		IdleTimeout:  rpc.conf.IdleTimeout,
	}
	go httpSrv.Serve(listener)
	return httpSrv, listener.Addr(), err
}

// CheckTimeouts ensures that timeout values are meaningful
func CheckTimeouts(conf *config.Config) {
	if conf.ReadTimeout < time.Second {
		//log.Warn("Sanitizing invalid HTTP read timeout", "provided", timeouts.ReadTimeout, "updated", rpc.DefaultHTTPTimeouts.ReadTimeout)
		conf.ReadTimeout = rpc.DefaultHTTPTimeouts.ReadTimeout
	}
	if conf.WriteTimeout < time.Second {
		//log.Warn("Sanitizing invalid HTTP write timeout", "provided", timeouts.WriteTimeout, "updated", rpc.DefaultHTTPTimeouts.WriteTimeout)
		conf.WriteTimeout = rpc.DefaultHTTPTimeouts.WriteTimeout
	}
	if conf.IdleTimeout < time.Second {
		//log.Warn("Sanitizing invalid HTTP idle timeout", "provided", timeouts.IdleTimeout, "updated", rpc.DefaultHTTPTimeouts.IdleTimeout)
		conf.IdleTimeout = rpc.DefaultHTTPTimeouts.IdleTimeout
	}
}
