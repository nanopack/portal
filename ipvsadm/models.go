// -*- mode: go; tab-width: 2; indent-tabs-mode: 1; st-rulers: [70] -*-
// vim: ts=4 sw=4 ft=lua noet
//--------------------------------------------------------------------
// @author Daniel Barney <daniel@nanobox.io>
// @copyright 2015, Pagoda Box Inc.
// @doc
//
// @end
// Created :   7 August 2015 by Daniel Barney <daniel@nanobox.io>
//--------------------------------------------------------------------
package ipvsadm

import (
	"encoding/json"
	"fmt"
)

type (
	ToJson interface {
		ToJson() ([]byte, error)
	}
	FromJson interface {
		FromJson([]byte) error
	}

	ider interface {
		getId() string
	}

	Server struct {
		Host                string `json:"host"`
		Port                int    `json:"port"`
		Forwarder           string `json:"forwarder"`
		Weight              int    `json:"weight"`
		InactiveConnections int    `json:"innactive_connections"`
		ActiveConnections   int    `json:"active_connections"`
	}

	Vip struct {
		Host        string   `json:"host"`
		Port        int      `json:"port"`
		Schedular   string   `json:"schedular"`
		Persistance int      `json:"persistance"`
		Servers     []Server `json:"servers"`
	}
)

func (s *Server) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}
func (s Server) ToJson() ([]byte, error) {
	return json.Marshal(s)
}
func (s Server) getId() string {
	return fmt.Sprintf("%v:%v", s.Host, s.Port)
}

func (v *Vip) FromJson(bytes []byte) error {
	return json.Unmarshal(bytes, v)
}
func (v Vip) ToJson() ([]byte, error) {
	return json.Marshal(v)
}
func (v Vip) getId() string {
	return fmt.Sprintf("%v:%v", v.Host, v.Port)
}
