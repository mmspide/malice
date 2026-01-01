package router

import (
	"context"
	"net/http"

	"github.com/docker/docker/api/server/httputils"
)

// Router defines an interface to specify a group of routes to add to the docker server.
type Router interface {
	// Routes returns the list of routes to add to the docker server.
	Routes() []Route
}

// Route defines an individual API route in the docker server.
type Route interface {
	// Handler returns the raw function to create the http handler.
	Handler() httputils.APIFunc
	// Method returns the http method that the route responds to.
	Method() string
	// Path returns the subpath where the route responds to.
	Path() string
}

// GetRoutes returns all API routes as a map for the server
func GetRoutes() (map[string]http.HandlerFunc, error) {
	routes := make(map[string]http.HandlerFunc)

	// Health check endpoint
	routes["/health"] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"0.4.0"}`))
	}

	// Info endpoint
	routes["/info"] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"Malice","version":"0.4.0","description":"Open Source Malware Analysis Framework"}`))
	}

	// Scan endpoint (placeholder)
	routes["/scan"] = func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"status":"submitted","scan_id":"pending"}`))
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"scans":[]}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

	// Results endpoint (placeholder)
	routes["/results"] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[]}`))
	}

	return routes, nil
}
