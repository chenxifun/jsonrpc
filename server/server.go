package server

import (
	"fmt"
	go_document "github.com/chenxifun/go-document"
	docty "github.com/chenxifun/go-document/types"
	"github.com/chenxifun/jsonrpc/config"
	"github.com/chenxifun/jsonrpc/doc/types"
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

func (s *Server) BuildDoc(d *go_document.Doc) {

	srv := types.Server{Modules: s.server.docInfo}

	for i, m := range srv.Modules {
		data, ok := d.Packages[m.PkgPath]
		if ok {

			parseModels(srv.Modules[i], data)

		}
	}

}

func parseModels(mod *types.Module, pd *docty.PkgData) {
	sd := pd.FindStruct(mod.Name)
	if sd == nil {
		return
	}
	mod.Describe = sd.Doc

	for i, f := range mod.Methods {
		fd := pd.FindFunc(mod.Name, f.Name)
		parseMrthod(mod.Methods[i], fd)

	}
}

func parseMrthod(f *types.Method, fd *docty.FuncData) {
	if fd == nil {
		return
	}

	f.Describe = fd.Description
	f.Title = fd.Title

}
