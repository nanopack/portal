package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/certmgr"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/routemgr"
)

var (
	// Clusterer      core.Backender
	Clusterer      Clusterable
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
	BadJson        = errors.New("Bad JSON syntax received in body")
)

type Clusterable interface {
	core.Backender
	routemgr.Routable
	certmgr.Keyable
}

func Init() error {
	url, err := url.Parse(config.ClusterConnection)
	if err != nil {
		return fmt.Errorf("Failed to parse db connection - %v", err)
	}

	switch url.Scheme {
	case "redis":
		Clusterer = &Redis{}
	case "none":
		Clusterer = &None{}
	default:
		Clusterer = &None{}
		// Clusterer = &Redis{}
	}

	return Clusterer.Init()
}

func GetServices() ([]core.Service, error) {
	return Clusterer.GetServices()
}

func GetService(id string) (*core.Service, error) {
	return Clusterer.GetService(id)
}

func SetServices(services []core.Service) error {
	return Clusterer.SetServices(services)
}

func SetService(service *core.Service) error {
	return Clusterer.SetService(service)
}

func DeleteService(id string) error {
	return Clusterer.DeleteService(id)
}

func SetServers(svcId string, servers []core.Server) error {
	return Clusterer.SetServers(svcId, servers)
}

func SetServer(svcId string, server *core.Server) error {
	return Clusterer.SetServer(svcId, server)
}

func DeleteServer(svcId, srvId string) error {
	return Clusterer.DeleteServer(svcId, srvId)
}

func GetServer(svcId, srvId string) (*core.Server, error) {
	return Clusterer.GetServer(svcId, srvId)
}

func SetRoutes(routes []router.Route) error {
	return Clusterer.SetRoutes(routes)
}

func SetRoute(route router.Route) error {
	return Clusterer.SetRoute(route)
}

func DeleteRoute(route router.Route) error {
	return Clusterer.DeleteRoute(route)
}

func GetRoutes() ([]router.Route, error) {
	return Clusterer.GetRoutes()
}

func SetCerts(certs []router.KeyPair) error {
	return Clusterer.SetCerts(certs)
}

func SetCert(cert router.KeyPair) error {
	return Clusterer.SetCert(cert)
}

func DeleteCert(cert router.KeyPair) error {
	return Clusterer.DeleteCert(cert)
}

func GetCerts() ([]router.KeyPair, error) {
	return Clusterer.GetCerts()
}

func parseSvc(serviceId string) (*core.Service, error) {
	s := strings.Replace(serviceId, "_", ".", -1)
	svc := strings.Split(s, "-")
	if len(svc) != 3 {
		return nil, NoServiceError
	}
	p, _ := strconv.Atoi(svc[2])
	return &core.Service{Type: svc[0], Host: svc[1], Port: p}, nil
}

func parseSrv(serverId string) (*core.Server, error) {
	s := strings.Replace(serverId, "_", ".", -1)
	srv := strings.Split(s, "-")
	if len(srv) != 2 {
		return nil, NoServerError
	}
	p, _ := strconv.Atoi(srv[1])
	return &core.Server{Host: srv[0], Port: p}, nil
}

func marshalSvc(service []byte) (*core.Service, error) {
	var svc core.Service

	if err := json.Unmarshal(service, &svc); err != nil {
		return nil, BadJson
	}

	svc.GenId()
	if svc.Id == "--0" {
		return nil, NoServiceError
	}

	for i := range svc.Servers {
		svc.Servers[i].GenId()
		if svc.Servers[i].Id == "-0" {
			return nil, NoServerError
		}
	}
	return &svc, nil
}

func marshalSvcs(services []byte) (*[]core.Service, error) {
	var svcs []core.Service

	if err := json.Unmarshal(services, &svcs); err != nil {
		return nil, BadJson
	}

	for i := range svcs {
		svcs[i].GenId()
		if svcs[i].Id == "--0" {
			return nil, NoServiceError
		}
		for j := range svcs[i].Servers {
			svcs[i].Servers[j].GenId()
			if svcs[i].Servers[j].Id == "-0" {
				return nil, NoServerError
			}
		}
	}
	return &svcs, nil
}

func marshalSrv(server []byte) (*core.Server, error) {
	var srv core.Server

	if err := json.Unmarshal(server, &srv); err != nil {
		return nil, BadJson
	}

	srv.GenId()
	if srv.Id == "-0" {
		return nil, NoServerError
	}

	return &srv, nil
}

func marshalSrvs(servers []byte) (*[]core.Server, error) {
	var srvs []core.Server

	if err := json.Unmarshal(servers, &srvs); err != nil {
		return nil, BadJson
	}

	for i := range srvs {
		srvs[i].GenId()
		if srvs[i].Id == "-0" {
			return nil, NoServerError
		}
	}

	return &srvs, nil
}

func parseBody(body []byte, v interface{}) error {
	if err := json.Unmarshal(body, v); err != nil {
		return BadJson
	}
	return nil
}
