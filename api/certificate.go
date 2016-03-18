package api

import (
	"net/http"

	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/core/common"
)

func postCert(rw http.ResponseWriter, req *http.Request) {
	var cert router.KeyPair
	err := parseBody(req, &cert)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// save to cluster
	err = cluster.SetCert(cert)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, cert, http.StatusOK)
}

func deleteCert(rw http.ResponseWriter, req *http.Request) {
	var cert router.KeyPair
	err := parseBody(req, &cert)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// save to cluster
	err = cluster.DeleteCert(cert)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, cert, http.StatusOK)
}

func putCerts(rw http.ResponseWriter, req *http.Request) {
	var certs []router.KeyPair
	err := parseBody(req, &certs)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// save to cluster
	err = cluster.SetCerts(certs)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, certs, http.StatusOK)
}

// List the certs registered in my system
func getCerts(rw http.ResponseWriter, req *http.Request) {
	certs, err := common.GetCerts()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}
	writeBody(rw, req, certs, http.StatusOK)
}
