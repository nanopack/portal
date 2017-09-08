// api handles the api routes and pertaining funtionality.
//
//  | Action | Route                             | Description                                 | Payload                                     | Output                        |
//  |--------|-----------------------------------|---------------------------------------------|---------------------------------------------|-------------------------------|
//  | GET    | /services                         | List all services                           | nil                                         | json array of service objects |
//  | POST   | /services                         | Add a service                               | json service object                         | json service object           |
//  | PUT    | /services                         | Reset the list of services                  | json array of service objects               | json array of service objects |
//  | PUT    | /services/:svc_id                 | Reset the specified service                 | nil                                         | json service object           |
//  | GET    | /services/:svc_id                 | Get information about a service             | nil                                         | json service object           |
//  | DELETE | /services/:svc_id                 | Delete a service                            | nil                                         | success message or an error   |
//  | GET    | /services/:svc_id/servers         | List all servers on a service               | nil                                         | json array of server objects  |
//  | POST   | /services/:svc_id/servers         | Add new server to a service                 | json server object                          | json server object            |
//  | PUT    | /services/:svc_id/servers         | Reset the list of servers on a service      | json array of server objects                | json array of server objects  |
//  | GET    | /services/:svc_id/servers/:srv_id | Get information about a server on a service | nil                                         | json server object            |
//  | DELETE | /services/:svc_id/servers/:srv_id | Delete a server from a service              | nil                                         | success message or an error   |
//  | DELETE | /routes                           | Delete a route                              | subdomain, domain, and path (json or query) | success message or an error   |
//  | GET    | /routes                           | List all routes                             | nil                                         | json array of route objects   |
//  | POST   | /routes                           | Add new route                               | json route object                           | json route object             |
//  | PUT    | /routes                           | Reset the list of routes                    | json array of route objects                 | json array of route objects   |
//  | DELETE | /certs                            | Delete a cert                               | json cert object                            | success message or an error   |
//  | GET    | /certs                            | List all certs                              | nil                                         | json array of cert objects    |
//  | POST   | /certs                            | Add new cert                                | json cert object                            | json cert object              |
//  | PUT    | /certs                            | Reset the list of certs                     | json array of cert objects                  | json array of cert objects    |
//  | DELETE | /vips                             | Delete a vip                                | json vip object                             | success message or an error   |
//  | GET    | /vips                             | List all vips                               | nil                                         | json array of vip objects     |
//  | POST   | /vips                             | Add new vip                                 | json vip object                             | json vip object               |
//  | PUT    | /vips                             | Reset the list of vips                      | json array of vip objects                   | json array of vip objects     |
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

	"github.com/nanopack/portal/config"
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
	auth.Header = "X-AUTH-TOKEN"

	if config.Insecure {
		config.Log.Info("Api listening at http://%s:%s...", config.ApiHost, config.ApiPort)
		return auth.ListenAndServe(fmt.Sprintf("%s:%s", config.ApiHost, config.ApiPort), config.ApiToken, routes())
	}

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

	config.Log.Info("Api listening at https://%s:%s...", config.ApiHost, config.ApiPort)
	return auth.ListenAndServeTLS(fmt.Sprintf("%s:%s", config.ApiHost, config.ApiPort), config.ApiToken, routes())
}

func routes() *pat.Router {
	router := pat.New()
	// balancing
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

	// errors
	router.Post("/errors", postErrors)
	router.Put("/errors", postErrors)
	router.Get("/errors", getErrors)

	// routing
	router.Delete("/routes", deleteRoute)
	router.Put("/routes", putRoutes)
	router.Get("/routes", getRoutes)
	router.Post("/routes", postRoute)

	// certificates
	router.Delete("/certs", deleteCert)
	router.Put("/certs", putCerts)
	router.Get("/certs", getCerts)
	router.Post("/certs", postCert)

	// ips
	router.Delete("/vips", deleteVip)
	router.Put("/vips", putVips)
	router.Get("/vips", getVips)
	router.Post("/vips", postVip)

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

	remoteAddr := req.RemoteAddr
	if fwdFor := req.Header.Get("X-Forwarded-For"); len(fwdFor) > 0 {
		remoteAddr = fwdFor
	}

	config.Log.Debug("%s %d %s %s %s", remoteAddr, status, req.Method, req.RequestURI, errMsg)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(append(b, byte('\n')))

	return nil
}

func writeError(rw http.ResponseWriter, req *http.Request, err error, status int) error {
	return writeBody(rw, req, apiError{ErrorString: err.Error()}, status)
}

// parseBody parses the json body into v
func parseBody(req *http.Request, v interface{}) error {

	// read the body
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error("Failed to read body - %s", err)
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
