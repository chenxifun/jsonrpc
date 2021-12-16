package server

import (
	"fmt"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/log"
	"github.com/chenxifun/jsonrpc/rpc"
	"os"
	"os/signal"
)

func NewServer(conf config.Config) *Server {
	s := &Server{
		conf: &conf,
		log:  log.DefLogger(),
	}

	s.server = newNETServer(log.DefLogger(), rpc.DefaultHTTPTimeouts)

	return s
}

type Server struct {
	conf   *config.Config
	server *netServer
	log    log.Logger
}

func (s *Server) RegisterService(api rpc.API) error {

	apis := []rpc.API{api}

	err := s.server.enableWS(apis, wsConfig{
		Origins: []string{"*"},
		Modules: []string{"test"},
	})

	err = s.server.enableRPC(apis, httpConfig{
		Modules:            []string{"test"},
		Vhosts:             s.conf.Vhosts,
		CorsAllowedOrigins: s.conf.Cors,
	})

	return err
}

func (s *Server) RegisterServices(apis []rpc.API, modules []string) error {

	err := s.server.enableWS(apis, wsConfig{
		Origins: []string{"*"},
		Modules: modules,
	})

	err = s.server.enableRPC(apis, httpConfig{
		Modules:            modules,
		Vhosts:             s.conf.Vhosts,
		CorsAllowedOrigins: s.conf.Cors,
	})

	return err
}

func (s *Server) Start() error {

	err := s.server.setListenAddr(s.conf.Hosts, s.conf.Port)
	if err != nil {
		return err
	}

	err = s.server.start()
	if err != nil {
		return err
	}

	defer func() {
		s.server.stop()

	}()

	abortChan := make(chan os.Signal, 1)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan
	fmt.Println("Exiting...", "signal", sig)

	return nil
}
