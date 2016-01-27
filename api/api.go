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
)

type (
	apiError struct {
		ErrorString string `json:"error"`
	}

	// // probably won't use
	// service struct {
	// 	id    string `json:"id"` //:omit_empty`
	// 	proto string `json:"proto"`
	// 	ip    string `json:"ip"`
	// 	port  uint32 `json:"port"`
	// }

	// // var svc []service
	// // [{"port":"32"}, {"port":"33"}]
	// if err := json.Unmarshal(b, &svc); err != nil {
	// 	return lvs.Service{}, err
	// }
	// if req.URL.Query().Get(":id") != ""; {
	// 	svc.id = req.URL.Query().Get(":id")
	// }

	// fmt.Printf("%+v\n", svc)
	// proto             := svc.proto
	// service_ip        := svc.ip
	// service_port, err := strconv.Atoi(svc.port)

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
	// router.Get("/services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}", handleRequest(getServer))
	// router.Post("/services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}", handleRequest(postServer))
	// router.Delete("/services/{proto}/{service_ip}/{service_port}/servers/{server_ip}/{server_port}", handleRequest(deleteServer))
	// router.Get("/services/{proto}/{service_ip}/{service_port}/servers", handleRequest(getServers))
	// router.Post("/services/{proto}/{service_ip}/{service_port}/servers", handleRequest(postServers))
	// router.Get("/services/{proto}/{service_ip}/{service_port}", handleRequest(getService))
	// router.Post("/services/{proto}/{service_ip}/{service_port}", handleRequest(postService))
	// router.Delete("/services/{proto}/{service_ip}/{service_port}", handleRequest(deleteService))
	// router.Get("/services", handleRequest(getServices))
	// router.Post("/services", handleRequest(postServices))
	// router.Get("/sync", handleRequest(getSync))
	// router.Post("/sync", handleRequest(postSync))

	router.Delete("/services/{svc_id}/servers/{srv_id}", handleRequest(deleteServer))
	router.Get("/services/{svc_id}/servers/{srv_id}", handleRequest(getServer))
	// router.Post("/services/{svc_id}/servers", handleRequest(postServers))
	router.Post("/services/{svc_id}/servers", handleRequest(postServer))
	router.Get("/services/{svc_id}/servers", handleRequest(getServers))
	router.Delete("/services/{svc_id}", handleRequest(deleteService))
	router.Get("/services/{svc_id}", handleRequest(getService))
	// router.Post("/services", handleRequest(postServices))
	router.Post("/services", handleRequest(postService))
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
	if req.URL.Query().Get(":svc_id") != "" {
		svc.Id = req.URL.Query().Get(":svc_id")
	}

	config.Log.Trace("%+v", svc)
	if err != nil {
		return database.Service{}, err
	}
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
	if req.URL.Query().Get(":srv_id") != "" {
		srv.Id = req.URL.Query().Get(":srv_id")
	}

	config.Log.Trace("%+v", srv)
	if err != nil {
		return database.Server{}, err
	}
	return srv, nil
}

// Get information about a backend server
func getServer(rw http.ResponseWriter, req *http.Request) {
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
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	// err = server.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	real_server := Balancer.GetServer(service, server)
	if real_server != nil {
		writeBody(rw, req, real_server, http.StatusOK)
		return
	}
	writeError(rw, req, NoServerError)
}

// Create a backend server
func postServer(rw http.ResponseWriter, req *http.Request) {
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
	// validate elsewhere
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	// err = server.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
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
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	// err = server.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	err = Balancer.DeleteServer(service, server)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	writeBody(rw, req, nil, http.StatusOK)
}

// Get information about a backend server
func getServers(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}/servers
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	real_service := Balancer.GetService(service)
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
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	servers := []database.Server{}
	// Servers?
	// - Host, Port, Forwarder, Weight, UpperThreshold, LowerThreshold
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&servers)
	// for _, server := range servers {
	// 	err = server.Validate()
	// 	if err != nil {
	// 		writeError(rw, req, err)
	// 		return
	// 	}
	// }
	err = Balancer.SetServers(service, servers)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	writeBody(rw, req, nil, http.StatusOK)
}

// Get information about a service
func getService(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	real_service := Balancer.GetService(service)
	if real_service == nil {
		writeError(rw, req, NoServiceError)
		return
	}
	writeBody(rw, req, real_service, http.StatusOK)
}

// Reset all services
func postServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	services := []database.Service{}

	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&services)

	// for _, service := range services {
	// err := service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	// }

	// err := Balancer.SetServices(services)
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	writeBody(rw, req, nil, http.StatusOK)
}

// Create a service
func postService(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }

	// Scheduler, Persistence, Netmask
	// Servers?
	// - Host, Port, Forwarder, Weight, UpperThreshold, LowerThreshold
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&service)
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	err = Balancer.SetService(service)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	writeBody(rw, req, nil, http.StatusOK)
}

// Delete a service
func deleteService(rw http.ResponseWriter, req *http.Request) {
	// /services/{proto}/{service_ip}/{service_port}
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err)
		return
	}
	// err = service.Validate()
	// if err != nil {
	// 	writeError(rw, req, err)
	// 	return
	// }
	err = Balancer.DeleteService(service)
	if err != nil {
		writeError(rw, req, err)
		return
	}
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
