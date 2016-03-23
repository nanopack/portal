package api

import (
	"net/http"

	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

func postCert(rw http.ResponseWriter, req *http.Request) {
	var cert core.CertBundle
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
	var cert core.CertBundle
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

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

func putCerts(rw http.ResponseWriter, req *http.Request) {
	var certs []core.CertBundle
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
