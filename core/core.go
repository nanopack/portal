package core

import (
	"fmt"
	"strings"
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
		Port        int      `json:"port"`
		Type        string   `json:"type"`
		Scheduler   string   `json:"scheduler"`
		Persistence int      `json:"persistence"`
		Netmask     string   `json:"netmask"`
		Servers     []Server `json:"servers,omitempty"`
	}
)

func (s *Service) GenId() {
	s.Id = fmt.Sprintf("%v-%v-%d", s.Type, strings.Replace(s.Host, ".", "_", -1), s.Port)
}

func (s *Server) GenId() {
	s.Id = fmt.Sprintf("%v-%d", strings.Replace(s.Host, ".", "_", -1), s.Port)
}
