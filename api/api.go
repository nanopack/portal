package api

// Things this api needs to support
// - Add services
// - Remove services
// - Add server to service
// - Remove server from service
// - Reset entire list

// lvs likes to identify services with a combination of protocol, ip, and port
// /services/:proto-:service_ip-:service_port
// /services/:proto-:service_ip-:service_port/servers/:server_ip-:server_port

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/pat"
	"github.com/nanobox-io/golang-nanoauth"

	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

var (
	auth           nanoauth.Auth
	BadJson        = errors.New("Bad JSON syntax received in body")
	BodyReadFail   = errors.New("Body Read Failed")
	NoServerError  = errors.New("No Server Found")
	NoServiceError = errors.New("No Service Found")
)

type (
	apiError struct {
		ErrorString string `json:"error"`
	}
	apiMsg struct {
		MsgString string `json:"msg"`
	}
)

func StartApi() error {
	var cert *tls.Certificate
	var err error
	if config.ApiCert == "" {
		cert, err = nanoauth.Generate("portal.nanobox.io")
	} else {
		cert, err = nanoauth.Load(config.ApiCert, config.ApiKey, config.ApiKeyPassword)
	}
	if err != nil {
		return err
	}
	auth.Certificate = cert
	auth.Header = "X-NANOBOX-TOKEN"

	config.Log.Info("Api listening at %s:%s...", config.ApiHost, config.ApiPort)
	return auth.ListenAndServeTLS(fmt.Sprintf("%s:%s", config.ApiHost, config.ApiPort), config.ApiToken, routes())
}

func routes() *pat.Router {
	router := pat.New()
	router.Delete("/services/{svcId}/servers/{srvId}", deleteServer)
	router.Get("/services/{svcId}/servers/{srvId}", getServer)
	router.Put("/services/{svcId}/servers", putServers)
	router.Post("/services/{svcId}/servers", postServer)
	router.Get("/services/{svcId}/servers", getServers)
	router.Delete("/services/{svcId}", deleteService)
	router.Put("/services/{svcId}", putService)
	router.Get("/services/{svcId}", getService)
	router.Post("/services", postService)
	router.Put("/services", putServices)
	router.Get("/services", getServices)
	return router
}

func writeBody(rw http.ResponseWriter, req *http.Request, v interface{}, status int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// print the error only if there is one
	var msg map[string]string
	json.Unmarshal(b, &msg)

	var errMsg string
	if msg["error"] != "" {
		errMsg = msg["error"]
	}

	config.Log.Debug("%s %d %s %s %s", req.RemoteAddr, status, req.Method, req.RequestURI, errMsg)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(append(b, byte('\n')))

	return nil
}

func writeError(rw http.ResponseWriter, req *http.Request, err error, status int) error {
	return writeBody(rw, req, apiError{ErrorString: err.Error()}, status)
}

func parseReqService(req *http.Request) (*core.Service, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
		return nil, BodyReadFail
	}

	var svc core.Service

	if err := json.Unmarshal(b, &svc); err != nil {
		return nil, BadJson
	}

	svc.GenId()
	if svc.Id == "--0" {
		return nil, NoServiceError
	}

	for i := range svc.Servers {
		svc.Servers[i].GenId()
	}

	config.Log.Trace("SERVICE: %+v", svc)
	return &svc, nil
}

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
	}

	// add to cluster
	err := cluster.SetServers(svcId, servers)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, servers, http.StatusOK)
}

// Get information about a service
func getService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	service, err := common.GetService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusNotFound)
		return
	}
	writeBody(rw, req, service, http.StatusOK)
}

// Reset all services
// /services
func putServices(rw http.ResponseWriter, req *http.Request) {
	services := []core.Service{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&services); err != nil {
		writeError(rw, req, BadJson, http.StatusBadRequest)
		return
	}

	for i := range services {
		services[i].GenId()
		if services[i].Id == "--0" {
			writeError(rw, req, NoServiceError, http.StatusBadRequest)
			return
		}
		for j := range services[i].Servers {
			services[i].Servers[j].GenId()
			if services[i].Servers[j].Id == "-0" {
				writeError(rw, req, NoServerError, http.StatusBadRequest)
				return
			}
		}
	}

	// save to cluster
	err := cluster.SetServices(services)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, services, http.StatusOK)
}

// Create a service
func postService(rw http.ResponseWriter, req *http.Request) {
	// /services
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// idempotent additions (don't update service on post)
	if svc, _ := common.GetService(service.Id); svc != nil {
		writeBody(rw, req, service, http.StatusOK)
		return
	}

	// save to cluster
	err = cluster.SetService(service)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, service, http.StatusOK)
}

// Replace a service (by replacing all services)
func putService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")
	// rough sanitization
	if len(strings.Split(svcId, "-")) != 3 {
		writeError(rw, req, NoServiceError, http.StatusBadRequest)
		return
	}

	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	services, err := common.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// update service by id
	for i := range services {
		if services[i].Id == svcId {
			services[i] = *service
			break
		}
	}

	// save to cluster
	if err := cluster.SetServices(services); err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, service, http.StatusOK)
}

// Delete a service
func deleteService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	// remove from cluster
	err := cluster.DeleteService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// List all services
func getServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	svcs, err := common.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}
	writeBody(rw, req, svcs, http.StatusOK)
}
