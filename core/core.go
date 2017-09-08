// core centralizes the commonly used interfaces and structs.
package core

import (
	"fmt"
	"strings"
)

type (
	Backender interface {
		Init() error
		// services
		GetServices() ([]Service, error)
		GetService(id string) (*Service, error)
		SetServices(services []Service) error
		SetService(service *Service) error
		DeleteService(id string) error
		// servers
		SetServers(svcId string, servers []Server) error
		SetServer(svcId string, server *Server) error
		DeleteServer(svcId, srvId string) error
		GetServer(svcId, srvId string) (*Server, error)
	}

	Proxyable interface {
		// routes
		SetRoute(route Route) error
		SetRoutes(routes []Route) error
		DeleteRoute(route Route) error
		GetRoutes() ([]Route, error)
		// certs
		SetCerts(certs []CertBundle) error
		SetCert(cert CertBundle) error
		DeleteCert(cert CertBundle) error
		GetCerts() ([]CertBundle, error)
	}

	Vipable interface {
		// vips
		SetVip(vip Vip) error
		SetVips(vips []Vip) error
		DeleteVip(vip Vip) error
		GetVips() ([]Vip, error)
	}

	Server struct {
		// todo: change "Id" to "name" (for clarity)
		Id             string `json:"id,omitempty"`
		Host           string `json:"host"`
		Port           int    `json:"port"`
		Forwarder      string `json:"forwarder"`
		Weight         int    `json:"weight"`
		UpperThreshold int    `json:"upper_threshold"`
		LowerThreshold int    `json:"lower_threshold"`
	}
	Service struct {
		Id          string   `json:"id,omitempty"`
		Host        string   `json:"host"`
		Interface   string   `json:"interface,omitempty"`
		Port        int      `json:"port"`
		Type        string   `json:"type"`
		Scheduler   string   `json:"scheduler"`
		Persistence int      `json:"persistence"`
		Netmask     string   `json:"netmask"`
		Servers     []Server `json:"servers,omitempty"`
	}

	Route struct {
		// defines match characteristics
		SubDomain string `json:"subdomain"` // subdomain to match on - "admin"
		Domain    string `json:"domain"`    // domain to match on - "myapp.com"
		Path      string `json:"path"`      // route to match on - "/admin"
		// defines actions
		Targets []string `json:"targets"` // ips of servers - ["http://127.0.0.1:8080/app1","http://127.0.0.2"] (optional)
		FwdPath string   `json:"fwdpath"` // path to forward to targets - "/goadmin" incoming req: test.com/admin -> 127.0.0.1/goadmin (optional)
		Page    string   `json:"page"`    // page to serve instead of routing to targets - "<HTML>We are fixing it</HTML>" (optional)
		// defines health check
		Endpoint       string `json:"endpoint"`        // url path to check for health (todo: what to do when fwdpath is set) (non blank enables health checks)
		ExpectedCode   int    `json:"expected_code"`   // expected http response code (default 200)
		ExpectedBody   string `json:"expected_body"`   // expected body
		ExpectedHeader string `json:"expected_header"` // expected http header (field:value)
		Host           string `json:"host"`            // 'host' header to use when performing health check
		Timeout        int    `json:"timeout"`         // milliseconds before connection times out (default 3000 (3s))
		Attempts       int    `json:"attempts"`        // number of times to try before marking dead
	}

	CertBundle struct {
		Cert string `json:"cert"`
		Key  string `json:"key"`
	}

	Vip struct {
		Ip        string `json:"ip"`        // ip/cidr
		Interface string `json:"interface"` // interface to bind to
		Alias     string `json:"alias"`     // label for ip
	}
)

func (s *Service) GenId() {
	if s.Type == "" { // default to tcp
		s.Type = "tcp"
	}
	s.Id = fmt.Sprintf("%s-%s-%d", s.Type, strings.Replace(s.Host, ".", "_", -1), s.Port)
}

func (s *Server) GenId() {
	s.Id = fmt.Sprintf("%s-%d", strings.Replace(s.Host, ".", "_", -1), s.Port)
}

// GenHost resets the server's Host it's service's Host if "127.0.0.1" was detected
func (s *Server) GenHost(svcId string) {
	if s.Host != "127.0.0.1" {
		return
	}

	host := strings.Split(strings.Replace(svcId, "_", ".", -1), "-")

	if len(host) != 3 {
		return
	}

	s.Host = host[1]
}
