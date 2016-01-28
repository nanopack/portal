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
	Balancer       balance.Lvs // should init
	NoServerError  = errors.New("No Server Found")
	NoServiceError = errors.New("No Service Found")
	Backend        = database.Backend //&database.Backend
)

type (
	apiError struct {
		ErrorString string `json:"error"`
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
	router.Delete("/services/{svc_id}/servers/{srv_id}", handleRequest(deleteServer))
	router.Get("/services/{svc_id}/servers/{srv_id}", handleRequest(getServer))
	// router.Post("/services/{svc_id}/servers", handleRequest(postServers))
	router.Post("/services/{svc_id}/servers", handleRequest(postServer))
	router.Get("/services/{svc_id}/servers", handleRequest(getServers))
	router.Delete("/services/{svc_id}", handleRequest(deleteService))
	router.Get("/services/{svc_id}", handleRequest(getService))
	router.Post("/services", handleRequest(postService))
	router.Put("/services", handleRequest(putServices))
	router.Get("/services", handleRequest(getServices))
	router.Post("/sync", handleRequest(postSync))
	router.Get("/sync", handleRequest(getSync))
	return router
}

func handleRequest(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		fn(rw, req)
	}
}

func writeBody(rw http.ResponseWriter, req *http.Request, v interface{}, status int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	config.Log.Trace("%s %d %s %s", req.RemoteAddr, status, req.Method, req.RequestURI)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(b)

	return nil
}

func writeError(rw http.ResponseWriter, req *http.Request, err error) error {
	config.Log.Error("%s %s %s %s", req.RemoteAddr, req.Method, req.RequestURI, err.Error())
	return writeBody(rw, req, apiError{ErrorString: err.Error()}, http.StatusInternalServerError)
}

func parseReqService(req *http.Request) (database.Service, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
	}

	var svc database.Service

	if err := json.Unmarshal(b, &svc); err != nil {
		return database.Service{}, err
	}

	svc.GenId()

	config.Log.Trace("SERVICE: %+v", svc)
	return svc, nil
}

func parseReqServer(req *http.Request) (database.Server, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
	}

	var srv database.Server

	if err := json.Unmarshal(b, &srv); err != nil {
		return database.Server{}, err
	}

	srv.GenId()

	config.Log.Trace("%+v", srv)
	return srv, nil
}

// Get information about a backend server
func getServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{svc_id}/servers/{srv_id}
	svc_id := req.URL.Query().Get(":svc_id")
	srv_id := req.URL.Query().Get(":srv_id")

	// getting from balancer ensures rule actually exists,
	// getting from backend ensures its written, which to use?
	// if write only on successful existance, db is good
	real_server := Balancer.GetServer(svc_id, srv_id)
	if real_server != nil {
		writeBody(rw, req, real_server, http.StatusOK)
		return
	}
	writeError(rw, req, NoServerError)
}

// Create a backend server
func postServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	server, err := parseReqServer(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	// Parse body for extra info:
	// Forwarder, Weight, UpperThreshold, LowerThreshold
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&server)
	err = Balancer.SetServer(service, server)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	writeBody(rw, req, nil, http.StatusOK)
}

// Delete a backend server
func deleteServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	server, err := parseReqServer(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	err = Balancer.DeleteServer(service, server)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	writeBody(rw, req, nil, http.StatusOK)
}

// Get information about a backend server
func getServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{svc_id}/servers
	svc_id := req.URL.Query().Get(":svc_id")
	real_service := Balancer.GetService(svc_id)
	if real_service == nil {
		writeError(rw, req, NoServiceError)
		return
	}
	writeBody(rw, req, real_service.Servers, http.StatusOK)
}

// Create a backend server
func postServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	// // todo: will this decode into servers properly
	// servers, err := parseReqServer(req)
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	servers := []database.Server{}
	// Servers?
	// - Host, Port, Forwarder, Weight, UpperThreshold, LowerThreshold
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&servers)
	for _, srv := range servers {
		srv.GenId()
	}
	err = Balancer.SetServers(service, servers)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	// return full service (database.x with id) rather than nil
	writeBody(rw, req, nil, http.StatusOK)
}

// Get information about a service
func getService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svc_id}
	svc_id := req.URL.Query().Get(":svc_id")

	service := Balancer.GetService(svc_id)
	if service == nil {
		writeError(rw, req, NoServiceError)
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
	}

	// apply services to balancer
	err := Balancer.SetServices(services)
	if err != nil {
		writeError(rw, req, err)
		return
	}

	// save to backend
	if Backend != nil {
		if err := Backend.SetServices(services); err != nil {
			writeError(rw, req, err)
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
		writeError(rw, req, err)
		return
	}

	// apply to balancer
	err = Balancer.SetService(service)
	if err != nil {
		writeError(rw, req, err)
		return
	}

	// save to backend
	if Backend != nil {
		err := Backend.SetService(service)
		if err != nil {
			writeError(rw, req, err)
			return
		}
	}

	writeBody(rw, req, service, http.StatusOK)
}

// Delete a service
func deleteService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svc_id}
	svc_id := req.URL.Query().Get(":svc_id")

	// delete backend rule
	err := Balancer.DeleteService(svc_id)
	if err != nil {
		writeError(rw, req, err)
		return
	}

	// remove from backend
	if Backend != nil {
		// in backend, on error, roll back 'insert'
		err := Backend.DeleteService(svc_id)
		if err != nil {
			writeError(rw, req, err)
			return
		}
	}

	// what to return here instead of nil?
	writeBody(rw, req, nil, http.StatusOK)
}

// List all services
func getServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	services := Balancer.GetServices()
	writeBody(rw, req, services, http.StatusOK)
}

// Sync portal's database from running system
func getSync(rw http.ResponseWriter, req *http.Request) {
	// /sync
	err := Balancer.SyncToPortal()
	if err != nil {
		writeError(rw, req, err)
		return
	}
	writeBody(rw, req, nil, http.StatusOK)
}

// Sync portal's database to running system
func postSync(rw http.ResponseWriter, req *http.Request) {
	// /sync
	err := Balancer.SyncToLvs()
	if err != nil {
		writeError(rw, req, err)
		return
	}
	writeBody(rw, req, nil, http.StatusOK)
}
