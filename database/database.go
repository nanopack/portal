package database

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nanopack/portal/config"
)

type (
	Backender interface {
		Init() error
		GetServices() ([]Service, error)
		GetService(id string) (*Service, error)
		SetServices(services []Service) error
		SetService(service *Service) error
		DeleteService(id string) error

		SetServers(svcId string, servers []Server) error
		SetServer(svcId string, server *Server) error
		DeleteServer(svcId, srvId string) error
		GetServer(svcId, srvId string) (*Server, error)
	}

	Server struct {
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
		Port        int      `json:"port"`
		Type        string   `json:"type"`
		Scheduler   string   `json:"scheduler"`
		Persistence int      `json:"persistence"`
		Netmask     string   `json:"netmask"`
		Servers     []Server `json:"servers,omitempty"`
	}
)

var (
	Backend        Backender
	NoServiceError = errors.New("No Service Found")
	NoServerError  = errors.New("No Server Found")
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
	return nil
}

func GetServices() ([]Service, error) {
	return Backend.GetServices()
}

func GetService(id string) (*Service, error) {
	return Backend.GetService(id)
}

func SetServices(services []Service) error {
	return Backend.SetServices(services)
}

func SetService(service *Service) error {
	return Backend.SetService(service)
}

func DeleteService(id string) error {
	return Backend.DeleteService(id)
}

func SetServers(svcId string, servers []Server) error {
	return Backend.SetServers(svcId, servers)
}

func SetServer(svcId string, server *Server) error {
	return Backend.SetServer(svcId, server)
}

func DeleteServer(svcId, srvId string) error {
	return Backend.DeleteServer(svcId, srvId)
}

func GetServer(svcId, srvId string) (*Server, error) {
	return Backend.GetServer(svcId, srvId)
}
