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

	s.server = newNETServer(log.DefLogger(), GetTimeouts(&conf))

	return s
}

type Server struct {
	conf   *config.Config
	server *netServer
	log    log.Logger
}

func (s *Server) RegisterService(api rpc.API) error {

	apis := []rpc.API{api}

	if s.conf.EnableWS {
		err := s.server.enableWS(apis, wsConfig{
			Origins: s.conf.Origins,
			Modules: []string{api.Namespace},
		})
		if err != nil {
			return err
		}
	}
	if s.conf.EnableRPC {
		err := s.server.enableRPC(apis, httpConfig{
			Modules:            []string{api.Namespace},
			Vhosts:             s.conf.Vhosts,
			CorsAllowedOrigins: s.conf.Cors,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) RegisterServices(apis []rpc.API, modules []string) error {

	if s.conf.EnableWS {
		err := s.server.enableWS(apis, wsConfig{
			Origins: s.conf.Origins,
			Modules: modules,
		})

		if err != nil {
			return err
		}
	}

	if s.conf.EnableRPC {
		err := s.server.enableRPC(apis, httpConfig{
			Modules:            modules,
			Vhosts:             s.conf.Vhosts,
			CorsAllowedOrigins: s.conf.Cors,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Start() error {

	if !s.conf.EnableRPC && !s.conf.EnableWS {
		return fmt.Errorf("WS and RPC must have one open")
	}

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
