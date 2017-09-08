package api

import (
	"net/http"

	"github.com/nanobox-io/nanobox-router"
)

// todo: update router so it uses map[string]string|byte for errors
// todo: need to overload reverseproxy.ServeHTTP and catch/handle 502 errors there for actual 502
// todo: need to save errors for reboots

func setNoRoutes(errorBody string) {
	router.ErrNoRoutes = []byte(errorBody)
}

func setNoHealthy(errorBody string) {
	router.ErrNoHealthy = []byte(errorBody)
}

// Create, update, or delete errors
func postErrors(rw http.ResponseWriter, req *http.Request) {
	var errors map[string]string
	err := parseBody(req, &errors)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	for key, val := range errors {
		switch key {
		case "no-routes":
			setNoRoutes(val)
		case "no-healthy":
			setNoHealthy(val)
		}
	}

	writeBody(rw, req, errors, http.StatusOK)
}

// List the errors registered in my system
func getErrors(rw http.ResponseWriter, req *http.Request) {
	errors := map[string]string{}
	errors["no-routes"] = string(router.ErrNoRoutes)
	errors["no-health"] = string(router.ErrNoHealthy)

	writeBody(rw, req, errors, http.StatusOK)
}
