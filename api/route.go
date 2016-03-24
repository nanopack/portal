package api

import (
	"net/http"

	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

func postRoute(rw http.ResponseWriter, req *http.Request) {
	var route core.Route
	err := parseBody(req, &route)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// save to cluster
	err = cluster.SetRoute(route)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, route, http.StatusOK)
}

func deleteRoute(rw http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	var route core.Route
	err := parseBody(req, &route)
	if err != nil {
		// check and use query parameters if payload fails
		if len(query) == 0 {
			writeError(rw, req, err, http.StatusBadRequest)
			return
		}
		route = core.Route{SubDomain: query.Get("subdomain"), Domain: query.Get("domain"), Path: query.Get("path")}
	}

	// save to cluster
	err = cluster.DeleteRoute(route)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

func putRoutes(rw http.ResponseWriter, req *http.Request) {
	var routes []core.Route
	err := parseBody(req, &routes)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// save to cluster
	err = cluster.SetRoutes(routes)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, routes, http.StatusOK)
}

// List the routes registered in my system
func getRoutes(rw http.ResponseWriter, req *http.Request) {
	routes, err := common.GetRoutes()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}
	writeBody(rw, req, routes, http.StatusOK)
}
