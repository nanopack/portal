// certmgr handles the adding of 'certs' (key/cert sets)
// for a secure reverse proxy router
package certmgr

import (
	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/config"
)

type Keyable interface {
	SetCerts(certs []router.KeyPair) error
	SetCert(cert router.KeyPair) error
	DeleteCert(cert router.KeyPair) error
	GetCerts() ([]router.KeyPair, error)
}

var (
	Cert Keyable
)

// start tls server
func Init() error {
	Cert = &Nanobox{}
	return router.StartTLS(config.RouteTls)
}

func SetCerts(certs []router.KeyPair) error {
	return Cert.SetCerts(certs)
}

func SetCert(cert router.KeyPair) error {
	return Cert.SetCert(cert)
}

func DeleteCert(cert router.KeyPair) error {
	return Cert.DeleteCert(cert)
}

func GetCerts() ([]router.KeyPair, error) {
	return Cert.GetCerts()
}
