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

	"github.com/gorilla/pat"
	"github.com/nanobox-io/golang-nanoauth"

	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/database"
)

var (
	auth           nanoauth.Auth
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
	router.Put("/services/{svcId}", postService)
	router.Get("/services/{svcId}", getService)
	router.Post("/services", postService)
	router.Put("/services", putServices)
	router.Get("/services", getServices)
	router.Post("/sync", postSync)
	router.Get("/sync", getSync)
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

func parseReqService(req *http.Request) (*database.Service, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
		// return?
	}

	var svc database.Service

	if err := json.Unmarshal(b, &svc); err != nil {
		return nil, fmt.Errorf("Bad JSON syntax received in body")
	}

	svc.GenId()
	if svc.Id == "--0" {
		return nil, NoServiceError
	}

	for i, _ := range svc.Servers {
		svc.Servers[i].GenId()
	}

	config.Log.Trace("SERVICE: %+v", svc)
	return &svc, nil
}

func parseReqServer(req *http.Request) (*database.Server, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
	}

	var srv database.Server

	if err := json.Unmarshal(b, &srv); err != nil {
		return nil, err
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

	// getting from balancer ensures rule actually exists,
	// getting from backend ensures its written, which to use?
	// if write only on successful existance, db is good
	server, err := balance.GetServer(svcId, srvId)
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

	// apply to balancer
	err = balance.SetServer(svcId, server)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// save to backend
	if database.Backend != nil {
		err = database.SetServer(svcId, server)
		if err != nil {
			writeError(rw, req, err, http.StatusInternalServerError)
			return
		}
	}

	// or service (which would include server)
	writeBody(rw, req, server, http.StatusOK)
}

// Delete a backend server
func deleteServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers/{srvId}
	svcId := req.URL.Query().Get(":svcId")
	srvId := req.URL.Query().Get(":srvId")

	// delete rule from balancer
	if err := balance.DeleteServer(svcId, srvId); err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// remove from backend
	if database.Backend != nil {
		if err := database.DeleteServer(svcId, srvId); err != nil {
			writeError(rw, req, err, http.StatusInternalServerError)
			return
		}
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// Get information about a backend server
func getServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers
	svcId := req.URL.Query().Get(":svcId")
	service, err := balance.GetService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusNotFound)
		return
	}
	writeBody(rw, req, service.Servers, http.StatusOK)
}

// Create a backend server
func putServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers
	svcId := req.URL.Query().Get(":svcId")

	servers := []database.Server{}
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&servers)

	for _, srv := range servers {
		srv.GenId()
	}
	// implement in balancer
	err := balance.SetServers(svcId, servers)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// add to backend
	if database.Backend != nil {
		err = database.SetServers(svcId, servers)
		if err != nil {
			writeError(rw, req, err, http.StatusInternalServerError)
			return
		}
	}

	svc, _ := balance.GetService(svcId)
	writeBody(rw, req, svc.Servers, http.StatusOK)
}

// Get information about a service
func getService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	service, err := balance.GetService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusNotFound)
		return
	}
	writeBody(rw, req, service, http.StatusOK)
}

// Reset all services
// /services
func putServices(rw http.ResponseWriter, req *http.Request) {
	services := []database.Service{}
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&services)

	for _, svc := range services {
		svc.GenId()
		if svc.Id == "--0" {
			writeError(rw, req, NoServiceError, http.StatusBadRequest)
			return
		}
	}

	// apply services to balancer
	err := balance.SetServices(services)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// save to backend
	if database.Backend != nil {
		if err := database.SetServices(services); err != nil {
			writeError(rw, req, err, http.StatusInternalServerError)
			return
		}
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

	// apply to balancer
	err = balance.SetService(service)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// save to backend
	if database.Backend != nil {
		err := database.SetService(service)
		if err != nil {
			writeError(rw, req, err, http.StatusInternalServerError)
			return
		}
	}

	writeBody(rw, req, service, http.StatusOK)
}

// Delete a service
func deleteService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	// delete backend rule
	err := balance.DeleteService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// remove from backend
	if database.Backend != nil {
		// in backend, on error, roll back 'insert'
		err := database.DeleteService(svcId)
		if err != nil {
			writeError(rw, req, err, http.StatusInternalServerError)
			return
		}
	}

	// what to return here instead of nil?
	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// List all services
func getServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	writeBody(rw, req, balance.GetServices(), http.StatusOK)
}

// Sync portal's database from running system
func getSync(rw http.ResponseWriter, req *http.Request) {
	// /sync
	// write to backend
	if database.Backend != nil {
		// get services from applied balancer rules
		services := balance.GetServices()

		err := database.SetServices(services)
		if err != nil {
			writeError(rw, req, err, http.StatusInternalServerError)
			return
		}
	}

	// apply (todo: already applied?) rules
	err := balance.SyncToPortal()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// Sync portal's database to running system
func postSync(rw http.ResponseWriter, req *http.Request) {
	// /sync
	// get recorded services to sync to backend
	services, err := database.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	err = balance.SyncToBalancer(services)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}
