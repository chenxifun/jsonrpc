package node

import (
	"compress/gzip"
	"fmt"
	"github.com/chenxifun/jsonrpc/rpc"

	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/cors"
)

// NewHTTPHandlerStack returns wrapped http-related handlers
func NewHTTPHandlerStack(srv http.Handler, cors []string, vhosts []string) http.Handler {
	// Wrap the CORS-handler within a host-handler
	handler := newCorsHandler(srv, cors)
	handler = newVHostHandler(vhosts, handler)
	return newGzipHandler(handler)
}

func newCorsHandler(srv http.Handler, allowedOrigins []string) http.Handler {
	// disable CORS support if user has not specified a custom CORS configuration
	if len(allowedOrigins) == 0 {
		return srv
	}
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{http.MethodPost, http.MethodGet},
		AllowedHeaders: []string{"*"},
		MaxAge:         600,
	})
	return c.Handler(srv)
}

type virtualHostHandler struct {
	vhosts map[string]struct{}
	next   http.Handler
}

func newVHostHandler(vhosts []string, next http.Handler) http.Handler {
	vhostMap := make(map[string]struct{})
	for _, allowedHost := range vhosts {
		vhostMap[strings.ToLower(allowedHost)] = struct{}{}
	}
	return &virtualHostHandler{vhostMap, next}
}

// ServeHTTP serves JSON-RPC requests over HTTP, implements http.Handler
func (h *virtualHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if r.Host is not set, we can continue serving since a browser would set the Host header
	if r.Host == "" {
		h.next.ServeHTTP(w, r)
		return
	}
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		// Either invalid (too many colons) or no port specified
		host = r.Host
	}
	if ipAddr := net.ParseIP(host); ipAddr != nil {
		// It's an IP address, we can serve that
		h.next.ServeHTTP(w, r)
		return

	}
	// Not an IP address, but a hostname. Need to validate
	if _, exist := h.vhosts["*"]; exist {
		h.next.ServeHTTP(w, r)
		return
	}
	if _, exist := h.vhosts[host]; exist {
		h.next.ServeHTTP(w, r)
		return
	}
	http.Error(w, "invalid host specified", http.StatusForbidden)
}

var gzPool = sync.Pool{
	New: func() interface{} {
		w := gzip.NewWriter(ioutil.Discard)
		return w
	},
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func newGzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gz := gzPool.Get().(*gzip.Writer)
		defer gzPool.Put(gz)

		gz.Reset(w)
		defer gz.Close()

		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
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

// RegisterApisFromWhitelist checks the given modules' availability, generates a whitelist based on the allowed modules,
// and then registers all of the APIs exposed by the services.
func RegisterApisFromWhitelist(apis []rpc.API, modules []string, srv *rpc.Server, exposeAll bool) error {
	if bad, available := checkModuleAvailability(modules, apis); len(bad) > 0 {
		fmt.Println("Unavailable modules in HTTP API list", "unavailable", bad, "available", available)
	}
	// Generate the whitelist based on the allowed modules
	whitelist := make(map[string]bool)
	for _, module := range modules {
		whitelist[module] = true
	}
	// Register all the APIs exposed by the services
	for _, api := range apis {
		if exposeAll || whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := srv.RegisterName(api.Namespace, api.Service); err != nil {
				return err
			}
		}
	}
	return nil
}
