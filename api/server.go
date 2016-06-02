package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

func parseReqServer(req *http.Request) (*core.Server, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
		return nil, BodyReadFail
	}

	var srv core.Server

	if err := json.Unmarshal(b, &srv); err != nil {
		return nil, BadJson
	}

	srv.GenId()
	if srv.Id == "-0" {
		return nil, NoServerError
	}

	config.Log.Trace("SERVER: %+v", srv)
	return &srv, nil
}

// Get information about a backend server
func getServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers/{srvId}
	svcId := req.URL.Query().Get(":svcId")
	srvId := req.URL.Query().Get(":srvId")

	server, err := common.GetServer(svcId, srvId)
	if err != nil {
		writeError(rw, req, err, http.StatusNotFound)
		return
	}
	writeBody(rw, req, server, http.StatusOK)
}

// Create a backend server
func postServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers
	svcId := req.URL.Query().Get(":svcId")

	server, err := parseReqServer(req)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// idempotent additions (don't update server on post)
	if srv, _ := common.GetServer(svcId, server.Id); srv != nil {
		writeBody(rw, req, server, http.StatusOK)
		return
	}

	// localhost doesn't work properly, use service.Host
	if server.Host == "127.0.0.1" {
		server.GenHost(svcId)
	}

	// save to cluster
	err = cluster.SetServer(svcId, server)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// todo: or service (which would include server)
	writeBody(rw, req, server, http.StatusOK)
}

// Delete a backend server
func deleteServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers/{srvId}
	svcId := req.URL.Query().Get(":svcId")
	srvId := req.URL.Query().Get(":srvId")

	// remove from cluster
	err := cluster.DeleteServer(svcId, srvId)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// Get information about a backend server
func getServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers
	svcId := req.URL.Query().Get(":svcId")
	service, err := common.GetService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusNotFound)
		return
	}
	if service.Servers == nil {
		service.Servers = make([]core.Server, 0, 0)
	}
	// writeBody(rw, req, service, http.StatusOK)
	writeBody(rw, req, service.Servers, http.StatusOK)
}

// Create a backend server
func putServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers
	svcId := req.URL.Query().Get(":svcId")

	servers := []core.Server{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&servers); err != nil {
		writeError(rw, req, BadJson, http.StatusBadRequest)
		return
	}

	for i := range servers {
		servers[i].GenId()

		// localhost doesn't work properly, use service.Host
		if servers[i].Host == "127.0.0.1" {
			servers[i].GenHost(svcId)
		}
	}

	// add to cluster
	err := cluster.SetServers(svcId, servers)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, servers, http.StatusOK)
}
