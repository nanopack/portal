package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core/common"
)

// parseBody parses the json body into v
func parseBody(req *http.Request, v interface{}) error {

	// read the body
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
		return BodyReadFail
	}
	defer req.Body.Close()

	// parse body and store in v
	err = json.Unmarshal(b, v)
	if err != nil {
		return BadJson
	}

	return nil
}

func postRoute(rw http.ResponseWriter, req *http.Request) {
	var route router.Route
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
	var route router.Route
	err := parseBody(req, &route)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// save to cluster
	err = cluster.DeleteRoute(route)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, route, http.StatusOK)
}

func putRoutes(rw http.ResponseWriter, req *http.Request) {
	var routes []router.Route
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
