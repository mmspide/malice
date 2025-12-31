package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

// versionMatcher defines a variable matcher to be parsed by the router
// when a request is about to be served.
const versionMatcher = "/v{version:[0-9.]+}"

// Config provides the configuration for the API server
type Config struct {
	Logging     bool
	EnableCors  bool
	CorsHeaders string
	Version     string
	SocketGroup string
	TLSConfig   *tls.Config
}

// Server contains instance details for the server
type Server struct {
	cfg              *Config
	servers          []*HTTPServer
	middlewares      []http.Handler
	routerSwapper    *routerSwapper
	shutdownOnce     *sync.Once
	ctx              context.Context
	cancel           context.CancelFunc
}

// New returns a new instance of the server based on the specified configuration.
// It allocates resources which will be needed for ServeAPI(ports, unix-sockets).
func New(cfg *Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		cfg:          cfg,
		shutdownOnce: &sync.Once{},
		ctx:          ctx,
		cancel:       cancel,
	}
}

// UseMiddleware appends a new middleware to the request chain.
// This needs to be called before the API routes are configured.
func (s *Server) UseMiddleware(m http.Handler) {
	s.middlewares = append(s.middlewares, m)
}

// Accept sets a listener the server accepts connections into.
func (s *Server) Accept(addr string, listeners ...net.Listener) {
	for _, listener := range listeners {
		httpServer := &HTTPServer{
			srv: &http.Server{
				Addr:        addr,
				BaseContext: func(net.Listener) context.Context { return s.ctx },
			},
			l: listener,
		}
		s.servers = append(s.servers, httpServer)
	}
}

// Close closes servers and thus stop receiving requests
func (s *Server) Close() {
	s.shutdownOnce.Do(func() {
		logrus.Debug("closing API server")
		s.cancel() // Signal all goroutines to shut down

		for _, srv := range s.servers {
			if err := srv.Close(); err != nil {
				logrus.WithError(err).Warn("error closing server")
			}
		}
	})
}

// serveAPI loops through all initialized servers and spawns goroutine
// with Serve method for each. It sets createMux() as Handler also.
func (s *Server) serveAPI() error {
	if len(s.servers) == 0 {
		return fmt.Errorf("no servers configured")
	}

	chErrors := make(chan error, len(s.servers))
	defer close(chErrors)

	for _, srv := range s.servers {
		srv.srv.Handler = s.routerSwapper
		go func(httpSrv *HTTPServer) {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("server goroutine panic: %v", r)
					chErrors <- fmt.Errorf("server panic: %v", r)
				}
			}()

			logrus.Infof("API listen on %s", httpSrv.l.Addr())
			err := httpSrv.Serve()
			// Ignore closed connection error during shutdown
			if err != nil && err != http.ErrServerClosed && !strings.Contains(err.Error(), "use of closed network connection") {
				logrus.WithError(err).Error("serve error")
				chErrors <- err
			} else if err == http.ErrServerClosed {
				logrus.Debug("server closed")
			}
		}(srv)
	}

	// Collect errors from all servers
	var errs []error
	for i := 0; i < len(s.servers); i++ {
		if err := <-chErrors; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("server errors: %v", errs)
	}

	return nil
}
			return err
		}
	}

	return nil
}

// HTTPServer contains an instance of http server and the listener.
// srv *http.Server, contains configuration to create an http server and a mux router with all api end points.
// l   net.Listener, is a TCP or Socket listener that dispatches incoming request to the router.
type HTTPServer struct {
	srv *http.Server
	l   net.Listener
}

// Serve starts listening for inbound requests.
func (s *HTTPServer) Serve() error {
	return s.srv.Serve(s.l)
}

// Close closes the HTTPServer from listening for the inbound requests.
func (s *HTTPServer) Close() error {
	return s.l.Close()
}

func (s *Server) makeHTTPHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Define the context that we'll pass around to share info
		// like the docker-request-id.
		ctx := context.WithValue(r.Context(), "User-Agent", r.Header.Get("User-Agent"))
		vars := mux.Vars(r)
		if vars == nil {
			vars = make(map[string]string)
		}
		
		// Call the handler with context
		handler(w, r.WithContext(ctx))
	}
}

// InitRouter initializes the list of routers for the server.
func (s *Server) InitRouter(routes map[string]http.HandlerFunc) {
	m := s.createMux(routes)
	s.routerSwapper = &routerSwapper{
		router: m,
	}
}

type routerSwapper struct {
	router *mux.Router
}

func (rs *routerSwapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rs.router.ServeHTTP(w, r)
}

// createMux initializes the main router the server uses.
func (s *Server) createMux(routes map[string]http.HandlerFunc) *mux.Router {
	m := mux.NewRouter()

	logrus.Debug("Registering routes")
	for path, handler := range routes {
		f := s.makeHTTPHandler(handler)
		logrus.Debugf("Registering GET %s", path)
		m.HandleFunc(path, f).Methods("GET", "POST", "PUT", "DELETE")
	}

	// 404 handler
	m.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("page not found"))
	})

	return m
}

// Wait blocks the server goroutine until it exits.
// It sends an error message if there is any error during
// the API execution.
func (s *Server) Wait(waitChan chan error) {
	if err := s.serveAPI(); err != nil {
		logrus.Errorf("ServeAPI error: %v", err)
		waitChan <- err
		return
	}
	waitChan <- nil
}
