package api

// Things this api needs to support
// - Add services
// - Remove services
// - Add server to service
// - Remove server from service
// - Reset entire list

// lvs likes to identify services with a combination of protocol, ip, and port
// /services/:proto/:service_ip/:service_port
// /services/:proto/:service_ip/:service_port/servers/:server_ip/:server_port

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/pat"
	"github.com/nanobox-io/golang-lvs"
	"github.com/nanobox-io/golang-nanoauth"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/database"
)

var (
	auth nanoauth.Auth
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

	return auth.ListenAndServeTLS(fmt.Sprintf("%s:%s", config.ApiHost, config.ApiPort), config.ApiToken, routes())
}

func routes() *pat.Router {
	router := pat.New()
	router.Get("/services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}", handleRequest(getServer))
	router.Post("/services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}", handleRequest(postServer))
	router.Delete("/services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}", handleRequest(deleteServer))
	router.Get("/services/{proto}/{service_ip}/{service_port}/servers", handleRequest(getServers))
	router.Post("/services/{proto}/{service_ip}/{service_port}/servers", handleRequest(postServers))
	router.Get("/services/{proto}/{service_ip}/{service_port}", handleRequest(getService))
	router.Post("/services/{proto}/{service_ip}/{service_port}", handleRequest(postService))
	router.Delete("/services/{proto}/{service_ip}/{service_port}", handleRequest(deleteService))
	router.Get("/services", handleRequest(getServices))
	router.Post("/services", handleRequest(postServices))
	router.Get("/sync", handleRequest(postServices))
	router.Post("/sync", handleRequest(postServices))
	return router
}

func handleRequest(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		fn(rw, req)
	}
}

func writeBody(v interface{}, rw http.ResponseWriter, status int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(b)

	return nil
}

func parseReqService(req *http.Request) (lvs.Service, error) {
	proto := req.URL.Query().Get(":proto")
	service_ip := req.URL.Query().Get(":service_ip")
	service_port, err := strconv.Atoi(req.URL.Query().Get(":service_port"))
	if err != nil {
		return lvs.Service{}, err
	}
	service := lvs.Service{Type: proto, Host: service_ip, Port: service_port}
	return service, nil
}

func parseReqServer(req *http.Request) (lvs.Server, error) {
	server_ip := req.URL.Query().Get(":server_ip")
	server_port, err := strconv.Atoi(req.URL.Query().Get(":server_port"))
	if err != nil {
		return lvs.Server{}, err
	}
	server := lvs.Server{Host: server_ip, Port: server_port}
	return server, nil
}

// Get information about a backend server
func getServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	server, err := parseReqServer(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	real_server := database.GetServer(service, server)
	if real_server != nil {
		writeBody(real_server, rw, http.StatusOK)
		return
	}
	writeBody(database.NoServerError, rw, http.StatusInternalServerError)
}

// Create a backend server
func postServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	server, err := parseReqServer(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	// Parse body for extra info:
	// Forwarder, Weight, UpperThreshold, LowerThreshold
	err = database.SetServer(service, server)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}

// Delete a backend server
func deleteServer(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	server, err := parseReqServer(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	err = database.DeleteServer(service, server)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}

// Get information about a backend server
func getServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	real_service := database.GetService(service)
	if real_service == nil {
		writeBody(database.NoServiceError, rw, http.StatusInternalServerError)
		return
	}
	writeBody(real_service.Servers, rw, http.StatusOK)
}

// Create a backend server
func postServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	servers := []lvs.Server{}
	// Servers?
	// - Host, Port, Forwarder, Weight, UpperThreshold, LowerThreshold
	err = database.SetServers(service, servers)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}

// Get information about a service
func getService(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	real_service := database.GetService(service)
	if real_service == nil {
		writeBody(database.NoServiceError, rw, http.StatusInternalServerError)
		return
	}
	writeBody(real_service, rw, http.StatusOK)
}

// Create a service
func postService(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	// Scheduler, Persistance, Netmask
	// Servers?
	// - Host, Port, Forwarder, Weight, UpperThreshold, LowerThreshold
	err = database.SetService(service)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}

// Delete a service
func deleteService(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}
	service, err := parseReqService(req)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	err = database.DeleteService(service)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}

// List all services
func getServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	services := database.GetServices()
	writeBody(services, rw, http.StatusOK)
}

// Reset all services
func postServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	services := []lvs.Service{}
	err := database.SetServices(services)
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}

// Sync portal's database from running system
func getSync(rw http.ResponseWriter, req *http.Request) {
	// /sync
	err := database.SyncToPortal()
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}

// Sync portal's database to running system
func postSync(rw http.ResponseWriter, req *http.Request) {
	// /sync
	err := database.SyncToLvs()
	if err != nil {
		writeBody(err, rw, http.StatusInternalServerError)
		return
	}
	writeBody(nil, rw, http.StatusOK)
}
