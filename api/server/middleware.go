package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// handlerWithGlobalMiddlewares wraps the handler function for a request with
// the server's global middlewares. The order of the middlewares is backwards,
// meaning that the first in the list will be evaluated last.
func (s *Server) handlerWithGlobalMiddlewares(handler http.HandlerFunc) http.HandlerFunc {
	next := handler

	for _, m := range s.middlewares {
		// Wrap handler with middleware
		next = func(h http.HandlerFunc, mw http.Handler) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				mw.ServeHTTP(w, r)
				h(w, r)
			}
		}(next, m)
	}

	if s.cfg.Logging && logrus.GetLevel() == logrus.DebugLevel {
		// Wrap with debug logging
		next = func(h http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				logrus.Debugf("DEBUG: %s %s", r.Method, r.URL.Path)
				h(w, r)
			}
		}(next)
	}

	return next
}
