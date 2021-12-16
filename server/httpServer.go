package server

import (
	"context"
	"fmt"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/log"
	"github.com/chenxifun/jsonrpc/node"
	"github.com/chenxifun/jsonrpc/rpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func NewHTTPServer(conf config.Config) *httpServer {

	return &httpServer{
		conf: &conf,
		srv:  rpc.NewServer(),
		log:  log.DefLogger(),
	}
}

type httpServer struct {
	srv  *rpc.Server
	conf *config.Config
	log  log.Logger
}

func (rpc *httpServer) RegisterService(api rpc.API) error {
	return rpc.srv.RegisterName(api.Namespace, api.Service)
}

func (rpc *httpServer) RegisterServices(apis []rpc.API, modules []string, exposeAll bool) error {
	if bad, available := checkModuleAvailability(modules, apis); len(bad) > 0 {
		rpc.log.Error("Unavailable modules in HTTP API list", "unavailable", bad, "available", available)
	}

	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}

	for _, api := range apis {
		if exposeAll || whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := rpc.srv.RegisterName(api.Namespace, api.Service); err != nil {
				return err
			}
		}
	}

	return nil
}

func (rpc *httpServer) Start() error {
	httpEndpoint := fmt.Sprintf("%s:%d", rpc.conf.Hosts, rpc.conf.Port)

	handler := node.NewHTTPHandlerStack(rpc.srv, rpc.conf.Cors, rpc.conf.Vhosts)

	httpServer, _, err := rpc.startHTTPEndpoint(httpEndpoint, handler)
	if err != nil {
		return err
	}

	defer func() {
		// Don't bother imposing a timeout here.
		httpServer.Shutdown(context.Background())

	}()

	abortChan := make(chan os.Signal, 1)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan
	fmt.Println("Exiting...", "signal", sig)

	return nil
}

func (rpc *httpServer) startHTTPEndpoint(endpoint string, handler http.Handler) (*http.Server, net.Addr, error) {
	// start the HTTP listener
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return nil, nil, err
	}
	// make sure timeout values are meaningful
	CheckConfTimeouts(rpc.conf)
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
func CheckConfTimeouts(conf *config.Config) {
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
