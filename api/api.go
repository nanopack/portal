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
package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/pat"
)

type (
	api struct {
		router *pat.Router
	}

	routerResponse interface{}
	response       struct {
		sucess bool
	}
)

var (
	defaultSuccess = &response{true}
	defaultApi     = &api{pat.New()}
)

func init() {
	defaultApi.router.Get("/ping", traceRequest(pong))
}

// pong to a ping.
func pong(res http.ResponseWriter, req *http.Request) (routerResponse, error) {
	return "pong", nil
}

// read and parse the entire body
func parseBody(req *http.Request, output interface{}) error {
	body, err := ioutil.ReadAll(req.Body)

	if err == nil {
		err = json.Unmarshal(body, output)
		req.Body.Close()
	}

	return err
}

// Start up the api and begin responding to requests. Blocking.
func Start(address string) error {
	return http.ListenAndServe(address, defaultApi.router)
}

// Traces all routes going through the api.
func traceRequest(fn func(http.ResponseWriter, *http.Request) (routerResponse, error)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// I need to do something to trace this request.
		value, err := fn(res, req)

		var bytes []byte
		if err == nil {
			if value == nil {
				value = defaultSuccess
			}
			bytes, err = json.Marshal(value)
		}

		if err != nil {
			res.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}", err)))
			return
		}
		res.Write(bytes)
	}
}
