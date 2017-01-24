package database

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nanobox-io/golang-scribble"
	"github.com/twinj/uuid"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

type (
	ScribbleDatabase struct {
		scribbleDb *scribble.Driver
	}
)

func (s *ScribbleDatabase) Init() error {
	u, err := url.Parse(config.DatabaseConnection)
	if err != nil {
		return err
	}
	dir := u.Path
	db, err := scribble.New(dir, nil)
	if err != nil {
		return err
	}

	s.scribbleDb = db
	return nil
}

func (s ScribbleDatabase) GetServices() ([]core.Service, error) {
	services := make([]core.Service, 0, 0)
	values, err := s.scribbleDb.ReadAll("services")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// if error is about a missing db, return empty array
			return services, nil
		}
		return nil, err
	}
	for i := range values {
		var service core.Service
		if err = json.Unmarshal([]byte(values[i]), &service); err != nil {
			return nil, fmt.Errorf("Bad JSON syntax received in body")
		}
		services = append(services, service)
	}
	return services, nil
}

func (s ScribbleDatabase) GetService(id string) (*core.Service, error) {
	service := core.Service{}
	err := s.scribbleDb.Read("services", id, &service)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			err = NoServiceError
		}
		return nil, err
	}
	return &service, nil
}

func (s ScribbleDatabase) SetServices(services []core.Service) error {
	s.scribbleDb.Delete("services", "")
	for i := range services {
		err := s.scribbleDb.Write("services", services[i].Id, services[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetService(service *core.Service) error {
	return s.scribbleDb.Write("services", service.Id, *service)
}

func (s ScribbleDatabase) DeleteService(id string) error {
	err := s.scribbleDb.Delete("services", id)
	if err != nil {
		if strings.Contains(err.Error(), "Unable to find") {
			err = nil
		}
		return err
	}
	return nil
}

func (s ScribbleDatabase) SetServer(svcId string, server *core.Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}
	for _, srv := range service.Servers {
		if srv.Id == server.Id {
			// if server already exists, don't duplicate it
			return nil
		}
	}
	service.Servers = append(service.Servers, *server)

	return s.scribbleDb.Write("services", service.Id, service)
}

func (s ScribbleDatabase) SetServers(svcId string, servers []core.Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}

	// pretty simple, reset all servers
	service.Servers = servers

	return s.scribbleDb.Write("services", service.Id, service)
}

func (s ScribbleDatabase) DeleteServer(svcId, srvId string) error {
	service, err := s.GetService(svcId)
	if err != nil {
		if strings.Contains(err.Error(), "Unable to find") {
			err = nil
		}
		return err
	}
	config.Log.Trace("Deleting %s from %s", srvId, svcId)
checkRemove:
	for i, srv := range service.Servers {
		if srv.Id == srvId {
			service.Servers = append(service.Servers[:i], service.Servers[i+1:]...)
			goto checkRemove // prevents 'slice bounds out of range' panic
		}
	}

	return s.scribbleDb.Write("services", service.Id, service)
}

func (s ScribbleDatabase) GetServer(svcId, srvId string) (*core.Server, error) {
	service := core.Service{}
	err := s.scribbleDb.Read("services", svcId, &service)
	if err != nil {
		return nil, err
	}

	for _, srv := range service.Servers {
		if srv.Id == srvId {
			return &srv, nil
		}
	}

	return nil, NoServerError
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////

func (s ScribbleDatabase) GetRoutes() ([]core.Route, error) {
	routes := make([]core.Route, 0, 0)
	values, err := s.scribbleDb.ReadAll("routes")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// if error is about a missing db, return empty array
			return routes, nil
		}
		return nil, err
	}
	for i := range values {
		var route core.Route
		if err = json.Unmarshal([]byte(values[i]), &route); err != nil {
			return nil, fmt.Errorf("Bad JSON syntax stored in db")
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func (s ScribbleDatabase) SetRoutes(routes []core.Route) error {
	s.scribbleDb.Delete("routes", "")
	for i := range routes {
		// unique key to store cert by
		ukey := uuid.NewV4().String()
		err := s.scribbleDb.Write("routes", ukey, routes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetRoute(route core.Route) error {
	routes, err := s.GetRoutes()
	if err != nil {
		return err
	}
	// for idempotency
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			return nil
		}
	}

	routes = append(routes, route)
	return s.SetRoutes(routes)
}

func (s ScribbleDatabase) DeleteRoute(route core.Route) error {
	routes, err := s.GetRoutes()
	if err != nil {
		return err
	}
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			routes = append(routes[:i], routes[i+1:]...)
			break
		}
	}
	return s.SetRoutes(routes)
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////

func (s ScribbleDatabase) GetCerts() ([]core.CertBundle, error) {
	certs := make([]core.CertBundle, 0, 0)
	values, err := s.scribbleDb.ReadAll("certs")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// if error is about a missing db, return empty array
			return certs, nil
		}
		return nil, err
	}
	for i := range values {
		var cert core.CertBundle
		if err = json.Unmarshal([]byte(values[i]), &cert); err != nil {
			return nil, fmt.Errorf("Bad JSON syntax stored in db")
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

func (s ScribbleDatabase) SetCerts(certs []core.CertBundle) error {
	s.scribbleDb.Delete("certs", "")
	for i := range certs {
		// unique key to store cert by
		ukey := uuid.NewV4().String()
		err := s.scribbleDb.Write("certs", ukey, certs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetCert(cert core.CertBundle) error {
	certs, err := s.GetCerts()
	if err != nil {
		return err
	}

	// for idempotency
	for i := 0; i < len(certs); i++ {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			return nil
		}
		// update if key is different
		if certs[i].Cert == cert.Cert {
			certs[i].Key = cert.Key
		}
	}

	certs = append(certs, cert)
	return s.SetCerts(certs)
}

func (s ScribbleDatabase) DeleteCert(cert core.CertBundle) error {
	certs, err := s.GetCerts()
	if err != nil {
		return err
	}
	for i := 0; i < len(certs); i++ {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			certs = append(certs[:i], certs[i+1:]...)
			break
		}
	}
	return s.SetCerts(certs)
}

////////////////////////////////////////////////////////////////////////////////
// VIPS
////////////////////////////////////////////////////////////////////////////////

func (s ScribbleDatabase) GetVips() ([]core.Vip, error) {
	vips := make([]core.Vip, 0, 0)
	values, err := s.scribbleDb.ReadAll("vips")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// if error is about a missing db, return empty array
			return vips, nil
		}
		return nil, err
	}
	for i := range values {
		var vip core.Vip
		if err = json.Unmarshal([]byte(values[i]), &vip); err != nil {
			return nil, fmt.Errorf("Bad JSON syntax stored in db")
		}
		vips = append(vips, vip)
	}
	return vips, nil
}

func (s ScribbleDatabase) SetVips(vips []core.Vip) error {
	s.scribbleDb.Delete("vips", "")
	for i := range vips {
		// unique (as much as what we keep) key to store vip by
		ukey := fmt.Sprintf("%s-%s", strings.Replace(vips[i].Ip, ".", "_", -1), vips[i].Interface)
		err := s.scribbleDb.Write("vips", ukey, vips[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s ScribbleDatabase) SetVip(vip core.Vip) error {
	vips, err := s.GetVips()
	if err != nil {
		return err
	}
	// for idempotency
	for i := 0; i < len(vips); i++ {
		if vips[i].Ip == vip.Ip && vips[i].Interface == vip.Interface {
			return nil
		}
	}

	vips = append(vips, vip)
	return s.SetVips(vips)
}

func (s ScribbleDatabase) DeleteVip(vip core.Vip) error {
	vips, err := s.GetVips()
	if err != nil {
		return err
	}
	for i := 0; i < len(vips); i++ {
		if vips[i].Ip == vip.Ip && vips[i].Interface == vip.Interface {
			vips = append(vips[:i], vips[i+1:]...)
			break
		}
	}
	return s.SetVips(vips)
}
