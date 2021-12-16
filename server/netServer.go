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
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// httpConfig is the JSON-RPC/HTTP configuration.
type httpConfig struct {
	Modules            []string
	CorsAllowedOrigins []string
	Vhosts             []string
}

// wsConfig is the JSON-RPC/Websocket configuration
type wsConfig struct {
	Origins []string
	Modules []string
}

type rpcHandler struct {
	http.Handler
	server *rpc.Server
}

type netServer struct {
	log      log.Logger
	timeouts rpc.HTTPTimeouts
	mux      http.ServeMux // registered handlers go here

	mu       sync.Mutex
	server   *http.Server
	listener net.Listener // non-nil when server is running

	// HTTP RPC handler things.
	httpConfig  httpConfig
	httpHandler atomic.Value // *rpcHandler

	// WebSocket handler things.
	wsConfig  wsConfig
	wsHandler atomic.Value // *rpcHandler

	// These are set by setListenAddr.
	endpoint string
	host     string
	port     int

	handlerNames map[string]string
}

func newNETServer(log log.Logger, timeouts rpc.HTTPTimeouts) *netServer {
	h := &netServer{log: log, timeouts: timeouts, handlerNames: make(map[string]string)}
	h.httpHandler.Store((*rpcHandler)(nil))
	h.wsHandler.Store((*rpcHandler)(nil))
	return h
}

// setListenAddr configures the listening address of the server.
// The address can only be set while the server isn't running.
func (h *netServer) setListenAddr(host string, port int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.listener != nil && (host != h.host || port != h.port) {
		return fmt.Errorf("HTTP server already running on %s", h.endpoint)
	}

	h.host, h.port = host, port
	h.endpoint = fmt.Sprintf("%s:%d", host, port)
	return nil
}

// listenAddr returns the listening address of the server.
func (h *netServer) listenAddr() string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.listener != nil {
		return h.listener.Addr().String()
	}
	return h.endpoint
}

// start starts the HTTP server if it is enabled and not already running.
func (h *netServer) start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.endpoint == "" || h.listener != nil {
		return nil // already running or not configured
	}

	// Initialize the server.
	h.server = &http.Server{Handler: h}
	if h.timeouts != (rpc.HTTPTimeouts{}) {
		CheckTimeouts(&h.timeouts)
		h.server.ReadTimeout = h.timeouts.ReadTimeout
		h.server.WriteTimeout = h.timeouts.WriteTimeout
		h.server.IdleTimeout = h.timeouts.IdleTimeout
	}

	// Start the server.
	listener, err := net.Listen("tcp", h.endpoint)
	if err != nil {
		// If the server fails to start, we need to clear out the RPC and WS
		// configuration so they can be configured another time.
		h.disableRPC()
		h.disableWS()
		return err
	}
	h.listener = listener
	go h.server.Serve(listener)

	// if server is websocket only, return after logging
	if h.wsAllowed() && !h.rpcAllowed() {
		h.log.Info("WebSocket enabled", "url", fmt.Sprintf("ws://%v", listener.Addr()))
		return nil
	}
	// Log http endpoint.
	h.log.Info("HTTP server started",
		"endpoint", listener.Addr(),
		"cors", strings.Join(h.httpConfig.CorsAllowedOrigins, ","),
		"vhosts", strings.Join(h.httpConfig.Vhosts, ","),
	)

	// Log all handlers mounted on server.
	var paths []string
	for path := range h.handlerNames {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	logged := make(map[string]bool, len(paths))
	for _, path := range paths {
		name := h.handlerNames[path]
		if !logged[name] {
			log.DefLogger().Info(name+" enabled", "url", "http://"+listener.Addr().String()+path)
			logged[name] = true
		}
	}
	return nil
}

func (h *netServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rpc := h.httpHandler.Load().(*rpcHandler)
	if r.RequestURI == "/" {
		// Serve JSON-RPC on the root path.
		ws := h.wsHandler.Load().(*rpcHandler)
		if ws != nil && isWebsocket(r) {
			ws.ServeHTTP(w, r)
			return
		}
		if rpc != nil {
			rpc.ServeHTTP(w, r)
			return
		}
	} else if rpc != nil {
		// Requests to a path below root are handled by the mux,
		// which has all the handlers registered via Node.RegisterHandler.
		// These are made available when RPC is enabled.
		h.mux.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(404)
}

// stop shuts down the HTTP server.
func (h *netServer) stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.doStop()
}

func (h *netServer) doStop() {
	if h.listener == nil {
		return // not running
	}

	// Shut down the server.
	httpHandler := h.httpHandler.Load().(*rpcHandler)
	wsHandler := h.httpHandler.Load().(*rpcHandler)
	if httpHandler != nil {
		h.httpHandler.Store((*rpcHandler)(nil))
		httpHandler.server.Stop()
	}
	if wsHandler != nil {
		h.wsHandler.Store((*rpcHandler)(nil))
		wsHandler.server.Stop()
	}
	h.server.Shutdown(context.Background())
	h.listener.Close()
	h.log.Info("HTTP server stopped", "endpoint", h.listener.Addr())

	// Clear out everything to allow re-configuring it later.
	h.host, h.port, h.endpoint = "", 0, ""
	h.server, h.listener = nil, nil
}

// enableRPC turns on JSON-RPC over HTTP on the server.
func (h *netServer) enableRPC(apis []rpc.API, config httpConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rpcAllowed() {
		return fmt.Errorf("JSON-RPC over HTTP is already enabled")
	}

	// Create RPC server and handler.
	srv := rpc.NewServer()
	if err := RegisterApisFromWhitelist(apis, config.Modules, srv, false); err != nil {
		return err
	}
	h.httpConfig = config
	h.httpHandler.Store(&rpcHandler{
		Handler: node.NewHTTPHandlerStack(srv, config.CorsAllowedOrigins, config.Vhosts),
		server:  srv,
	})
	return nil
}

// disableRPC stops the HTTP RPC handler. This is internal, the caller must hold h.mu.
func (h *netServer) disableRPC() bool {
	handler := h.httpHandler.Load().(*rpcHandler)
	if handler != nil {
		h.httpHandler.Store((*rpcHandler)(nil))
		handler.server.Stop()
	}
	return handler != nil
}

// enableWS turns on JSON-RPC over WebSocket on the server.
func (h *netServer) enableWS(apis []rpc.API, config wsConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.wsAllowed() {
		return fmt.Errorf("JSON-RPC over WebSocket is already enabled")
	}

	// Create RPC server and handler.
	srv := rpc.NewServer()
	if err := RegisterApisFromWhitelist(apis, config.Modules, srv, false); err != nil {
		return err
	}
	h.wsConfig = config
	h.wsHandler.Store(&rpcHandler{
		Handler: srv.WebsocketHandler(config.Origins),
		server:  srv,
	})
	return nil
}

// stopWS disables JSON-RPC over WebSocket and also stops the server if it only serves WebSocket.
func (h *netServer) stopWS() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.disableWS() {
		if !h.rpcAllowed() {
			h.doStop()
		}
	}
}

// disableWS disables the WebSocket handler. This is internal, the caller must hold h.mu.
func (h *netServer) disableWS() bool {
	ws := h.wsHandler.Load().(*rpcHandler)
	if ws != nil {
		h.wsHandler.Store((*rpcHandler)(nil))
		ws.server.Stop()
	}
	return ws != nil
}

// rpcAllowed returns true when JSON-RPC over HTTP is enabled.
func (h *netServer) rpcAllowed() bool {
	return h.httpHandler.Load().(*rpcHandler) != nil
}

// wsAllowed returns true when JSON-RPC over WebSocket is enabled.
func (h *netServer) wsAllowed() bool {
	return h.wsHandler.Load().(*rpcHandler) != nil
}

// isWebsocket checks the header of an http request for a websocket upgrade request.
func isWebsocket(r *http.Request) bool {

	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// CheckTimeouts ensures that timeout values are meaningful
func CheckTimeouts(timeouts *rpc.HTTPTimeouts) {
	if timeouts.ReadTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP read timeout", "provided", timeouts.ReadTimeout, "updated", rpc.DefaultHTTPTimeouts.ReadTimeout)
		timeouts.ReadTimeout = rpc.DefaultHTTPTimeouts.ReadTimeout
	}
	if timeouts.WriteTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP write timeout", "provided", timeouts.WriteTimeout, "updated", rpc.DefaultHTTPTimeouts.WriteTimeout)
		timeouts.WriteTimeout = rpc.DefaultHTTPTimeouts.WriteTimeout
	}
	if timeouts.IdleTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP idle timeout", "provided", timeouts.IdleTimeout, "updated", rpc.DefaultHTTPTimeouts.IdleTimeout)
		timeouts.IdleTimeout = rpc.DefaultHTTPTimeouts.IdleTimeout
	}
}

// RegisterApisFromWhitelist checks the given modules' availability, generates a whitelist based on the allowed modules,
// and then registers all of the APIs exposed by the services.
func RegisterApisFromWhitelist(apis []rpc.API, modules []string, srv *rpc.Server, exposeAll bool) error {
	if bad, available := checkModuleAvailability(modules, apis); len(bad) > 0 {
		log.DefLogger().Error("Unavailable modules in HTTP API list", "unavailable", bad, "available", available)
	}
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	for _, api := range apis {
		if exposeAll || whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := srv.RegisterName(api.Namespace, api.Version, api.Service); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkModuleAvailability checks that all names given in modules are actually
// available API services. It assumes that the MetadataApi module ("rpc") is always available;
// the registration of this "rpc" module happens in NewServer() and is thus common to all endpoints.
func checkModuleAvailability(modules []string, apis []rpc.API) (bad, available []string) {
	availableSet := make(map[string]struct{})
	for _, api := range apis {
		if _, ok := availableSet[api.Namespace]; !ok {
			availableSet[api.Namespace] = struct{}{}
			available = append(available, api.Namespace)
		}
	}
	for _, name := range modules {
		if _, ok := availableSet[name]; !ok && name != rpc.MetadataApi {
			bad = append(bad, name)
		}
	}
	return bad, available
}

// CheckTimeouts ensures that timeout values are meaningful
func CheckConfTimeouts(conf *config.Config) {
	if conf.ReadTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP read timeout", "provided", conf.ReadTimeout, "updated", rpc.DefaultHTTPTimeouts.ReadTimeout)
		conf.ReadTimeout = rpc.DefaultHTTPTimeouts.ReadTimeout
	}
	if conf.WriteTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP write timeout", "provided", conf.WriteTimeout, "updated", rpc.DefaultHTTPTimeouts.WriteTimeout)
		conf.WriteTimeout = rpc.DefaultHTTPTimeouts.WriteTimeout
	}
	if conf.IdleTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP idle timeout", "provided", conf.IdleTimeout, "updated", rpc.DefaultHTTPTimeouts.IdleTimeout)
		conf.IdleTimeout = rpc.DefaultHTTPTimeouts.IdleTimeout
	}
}

// CheckTimeouts ensures that timeout values are meaningful
func GetTimeouts(conf *config.Config) rpc.HTTPTimeouts {

	if conf.ReadTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP read timeout", "provided", conf.ReadTimeout, "updated", rpc.DefaultHTTPTimeouts.ReadTimeout)
		conf.ReadTimeout = rpc.DefaultHTTPTimeouts.ReadTimeout
	}
	if conf.WriteTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP write timeout", "provided", conf.WriteTimeout, "updated", rpc.DefaultHTTPTimeouts.WriteTimeout)
		conf.WriteTimeout = rpc.DefaultHTTPTimeouts.WriteTimeout
	}
	if conf.IdleTimeout < time.Second {
		log.DefLogger().Warn("Sanitizing invalid HTTP idle timeout", "provided", conf.IdleTimeout, "updated", rpc.DefaultHTTPTimeouts.IdleTimeout)
		conf.IdleTimeout = rpc.DefaultHTTPTimeouts.IdleTimeout
	}

	return rpc.HTTPTimeouts{
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		IdleTimeout:  conf.IdleTimeout,
	}

}
