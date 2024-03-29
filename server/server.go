package server

import (
	"fmt"
	go_document "github.com/chenxifun/go-document"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/doc"
	"github.com/chenxifun/jsonrpc/log"
	"github.com/chenxifun/jsonrpc/rpc"
	"os"
	"os/signal"
)

type ServerOption func(server *Server) error

func WithServerLogger(logger log.Logger) ServerOption {
	return func(server *Server) error {
		server.log = logger
		return nil
	}
}

func NewServer(conf config.Config, options ...ServerOption) (*Server, error) {
	s := &Server{
		conf: &conf,
	}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.log == nil {
		s.log = log.DefLogger()
	}

	s.server = newNETServer(s.log, GetTimeouts(&conf), conf.HeaderKey)

	return s, nil
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

	if err := s.DoStart(); err != nil {
		return err
	}

	defer func() {
		s.Stop()

	}()

	abortChan := make(chan os.Signal, 1)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan
	fmt.Println("Exiting...", "signal", sig)

	return nil
}

func (s *Server) DoStart() error {
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

	return nil
}

func (s *Server) Stop() {
	s.server.stop()
}

func (s *Server) BuildDoc(d *go_document.Doc) {

	doc.BuildDoc(d, s.server.docInfo)

	//srv := types.Server{Modules: s.server.docInfo}
	//
	//for i, m := range srv.Modules {
	//	data, ok := d.Packages[m.PkgPath]
	//	if ok {
	//
	//		parseModels(srv.Modules[i], data)
	//
	//	}
	//}

}
