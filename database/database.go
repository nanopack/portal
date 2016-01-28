package database

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/coreos/go-iptables/iptables"

	"github.com/nanopack/portal/config"
)

type (
	Backender interface {
		Init() error
		GetServices() ([]Service, error)
		// arg could be service.Id (== "service.Type"+"-"+"service.Host"+"-"+"service.Port")
		// GetService(id string) (Service, error)
		GetService(service Service) (Service, error)
		SetServices(services []Service) error
		SetService(service Service) error
		// DeleteService(id string) error
		DeleteService(service Service) error

		// implement servers here?

	}

	Server struct {
		// sanitize id
		Id             string `json:"id,omitempty"`
		Host           string `json:"host"`
		Port           int    `json:"port"`
		Forwarder      string `json:"forwarder"`
		Weight         int    `json:"weight"`
		UpperThreshold int    `json:upper_threshold`
		LowerThreshold int    `json:lower_threshold`
	}
	Service struct {
		// sanitize id
		Id          string   `json:"id,omitempty"`
		Host        string   `json:"host"`
		Port        int      `json:"port"`
		Type        string   `json:"type"`
		Scheduler   string   `json:"scheduler"`
		Persistence int      `json:"persistence"`
		Netmask     string   `json:"netmask"`
		Servers     []Server `json:"servers,omitempty"`
	}
)

var (
	Backend Backender
	Tab     *iptables.IPTables
)

func (s *Service) GenId() {
	s.Id = fmt.Sprintf("%v-%v-%d", s.Type, strings.Replace(s.Host, ".", "_", -1), s.Port)
}

func (s *Server) GenId() {
	s.Id = fmt.Sprintf("%v-%d", strings.Replace(s.Host, ".", "_", -1), s.Port)
}

func Init() error {
	var err error
	var u *url.URL
	u, err = url.Parse(config.DatabaseConnection)
	if err != nil {
		return fmt.Errorf("Failed to parse db connection - %v", err)
	}
	switch u.Scheme {
	case "scribble":
		Backend = &ScribbleDatabase{}
	default:
		Backend = nil
	}
	if Backend != nil {
		err = Backend.Init()
		if err != nil {
			Backend = nil
		}
	}
	Tab, err = iptables.New()
	if err != nil {
		Tab = nil
	}
	if Tab != nil {
		Tab.Delete("filter", "INPUT", "-j", "portal")
		Tab.ClearChain("filter", "portal")
		Tab.DeleteChain("filter", "portal")
		err = Tab.NewChain("filter", "portal")
		if err != nil {
			return fmt.Errorf("Failed to create new chain - %v", err)
		}
		err = Tab.AppendUnique("filter", "portal", "-j", "RETURN")
		if err != nil {
			return fmt.Errorf("Failed to append to portal chain - %v", err)
		}
		err = Tab.AppendUnique("filter", "INPUT", "-j", "portal")
		if err != nil {
			return fmt.Errorf("Failed to append to INPUT chain - %v", err)
		}
	}
	return nil
}
