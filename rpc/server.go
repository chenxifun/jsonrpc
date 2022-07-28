package rpc

import (
	"context"
	"github.com/chenxifun/jsonrpc/doc/types"
	"github.com/chenxifun/jsonrpc/log"
	mapset "github.com/deckarep/golang-set"
	"io"
	"sync/atomic"
)

const (
	MetadataApi     = "rpc"
	MetadataVersion = "1.0"
)

type CodecOption int

type Server struct {
	services serviceRegistry
	idgen    func() ID
	run      int32
	codecs   mapset.Set

	headerKeys map[string]bool
}

// NewServer creates a new server instance with no registered handlers.
func NewServer() *Server {
	server := &Server{codecs: mapset.NewSet(), idgen: randomIDGenerator(), run: 1, headerKeys: make(map[string]bool)}

	server.AddHeaderKeys("Origin", "User-Agent")

	// Register the default service providing meta information about the RPC service such
	// as the services and methods it offers.
	rpcService := &RPCService{server}
	server.RegisterName(MetadataApi, MetadataVersion, rpcService)
	return server
}

func (s *Server) AddHeaderKeys(keys ...string) {
	for _, key := range keys {
		if !s.headerKeys[key] {
			s.headerKeys[key] = true
		}
	}
}

func (s *Server) ModsInfo() []*types.Module {
	return s.services.docInfo
}

// RegisterName creates a service for the given receiver type under the given name. When no
// methods on the given receiver match the criteria to be either a RPC method or a
// subscription an error is returned. Otherwise a new service is created and added to the
// service collection this server provides to clients.
func (s *Server) RegisterName(name, version string, receiver interface{}) error {
	return s.services.registerName(name, version, receiver)
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed or the
// server is stopped. In either case the codec is closed.
//
// Note that codec options are no longer supported.
func (s *Server) ServeCodec(codec ServerCodec, options CodecOption) {
	defer codec.close()

	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}

	// Add the codec to the set so it can be closed by Stop.
	s.codecs.Add(codec)
	defer s.codecs.Remove(codec)

	c := initClient(codec, s.idgen, &s.services)
	<-codec.closed()
	c.Close()
}

// serveSingleRequest reads and processes a single RPC request from the given codec. This
// is used to serve HTTP connections. Subscriptions and reverse calls are not allowed in
// this mode.
func (s *Server) serveSingleRequest(ctx context.Context, codec ServerCodec) {
	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}

	h := newHandler(ctx, codec, s.idgen, &s.services)
	h.allowSubscribe = false
	defer h.close(io.EOF, nil)

	reqs, batch, err := codec.readBatch()
	if err != nil {
		if err != io.EOF {
			codec.writeJSON(ctx, errorMessage(&invalidMessageError{"parse error"}))
		}
		return
	}
	if batch {
		h.handleBatch(reqs)
	} else {
		h.handleMsg(reqs[0])
	}
}

// Stop stops reading new requests, waits for stopPendingRequestTimeout to allow pending
// requests to finish, then closes all codecs which will cancel pending requests and
// subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log.DefLogger().Debug("RPC server shutting down")
		s.codecs.Each(func(c interface{}) bool {
			c.(ServerCodec).close()
			return true
		})
	}
}

// RPCService gives meta information about the server.
// e.g. gives information about the loaded modules.
type RPCService struct {
	server *Server
}

// Modules returns the list of RPC services with their version number
func (s *RPCService) Modules() map[string]string {
	s.server.services.mu.Lock()
	defer s.server.services.mu.Unlock()

	modules := make(map[string]string)
	for name, s := range s.server.services.services {
		modules[name] = s.version
	}
	return modules
}
