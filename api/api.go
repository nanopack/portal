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

	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
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

	server, err := database.GetServer(svcId, srvId)
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
	if srv, _ := database.GetServer(svcId, server.Id); srv != nil {
		writeBody(rw, req, server, http.StatusOK)
		return
	}

	// apply to balancer
	err = balance.SetServer(svcId, server)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// save to backend
	err = database.SetServer(svcId, server)
	if err != nil {
		// undo balance action
		if uerr := balance.DeleteServer(svcId, server.Id); uerr != nil {
			err = uerr
		}
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

	// get server in case of failure
	srv, err := database.GetServer(svcId, srvId)
	if err != nil {
		if strings.Contains(err.Error(), "No Server Found") {
			writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
			return
		}
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// delete rule from balancer
	if err = balance.DeleteServer(svcId, srvId); err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// remove from backend
	if err = database.DeleteServer(svcId, srvId); err != nil {
		// undo balance action
		if uerr := balance.SetServer(svcId, srv); uerr != nil {
			err = uerr
		}
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// Get information about a backend server
func getServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}/servers
	svcId := req.URL.Query().Get(":svcId")
	service, err := database.GetService(svcId)
	// service, err := balance.GetService(svcId)
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

	// in case of failure
	oldService, err := database.GetService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}
	oldServers := oldService.Servers

	// implement in balancer
	err = balance.SetServers(svcId, servers)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// add to backend
	err = database.SetServers(svcId, servers)
	if err != nil {
		// undo balance action
		if uerr := balance.SetServers(svcId, oldServers); uerr != nil {
			err = uerr
		}
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	svc, _ := balance.GetService(svcId)
	writeBody(rw, req, svc.Servers, http.StatusOK)
}

// Get information about a service
func getService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	service, err := database.GetService(svcId)
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

	// in case of failure
	oldServices, err := database.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// apply services to balancer
	err = balance.SetServices(services)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// save to backend
	err = database.SetServices(services)
	if err != nil {
		// undo balance action
		if uerr := balance.SetServices(oldServices); uerr != nil {
			err = uerr
		}
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
	if svc, _ := database.GetService(service.Id); svc != nil {
		writeBody(rw, req, service, http.StatusOK)
		return
	}

	// in case of failure
	oldServices, err := database.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// apply to balancer
	err = balance.SetService(service)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// save to backend
	err = database.SetService(service)
	if err != nil {
		// undo balance action
		if uerr := balance.SetServices(oldServices); uerr != nil {
			err = uerr
		}
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

	services, err := database.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// in case of failure
	oldServices := services

	// update service by id
	for i := range services {
		if services[i].Id == svcId {
			services[i] = *service
			break
		}
	}

	// apply services to balancer
	err = balance.SetServices(services)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// save to backend
	if err := database.SetServices(services); err != nil {
		// undo balance action
		if uerr := balance.SetServices(oldServices); uerr != nil {
			err = uerr
		}
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, service, http.StatusOK)
}

// Delete a service
func deleteService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	// in case of failure
	oldService, err := database.GetService(svcId)
	if err != nil {
		if strings.Contains(err.Error(), "No Service Found") {
			writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
			return
		}
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// delete backend rule
	err = balance.DeleteService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// remove from backend
	err = database.DeleteService(svcId)
	if err != nil {
		// undo balance action
		if uerr := balance.SetService(oldService); uerr != nil {
			err = uerr
		}
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// List all services
func getServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	svcs, err := database.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}
	writeBody(rw, req, svcs, http.StatusOK)
}

// Sync portal's database from running ipvsadm system
func getSync(rw http.ResponseWriter, req *http.Request) {
	// /sync
	// should go first if keeping (todo: get rid of balance.Sync altogether and just keep write to backend)
	// get (hopefully already applied?) rules from ipvsadm and register with lvs.DefaultIpvs.Services
	// found: if ipvsadm binary is faked, no rules would be applied, thus making balance.Sync 'useful'(?)
	err := balance.Sync()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// write lvs.DefaultIpvs.Services to backend
	// get services from applied balancer rules
	services, _ := balance.GetServices()

	err = database.SetServices(services)
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

	err = balance.SetServices(services)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}
