// nanobox.go contains logic to use nanobox-router as a certificate manager

package certmgr

import (
	"github.com/nanobox-io/nanobox-router"
)

type Nanobox struct{}

func (self Nanobox) SetCert(cert router.KeyPair) error {
	certs := router.Keys()
	// for idempotency
	for i := 0; i < len(certs); i++ {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			return nil
		}
	}

	certs = append(certs, cert)
	return self.SetCerts(certs)
}

func (self Nanobox) DeleteCert(cert router.KeyPair) error {
	certs := router.Keys()
	for i := 0; i < len(certs); i++ {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			certs = append(certs[:i], certs[i+1:]...)
			break
		}
	}

	return self.SetCerts(certs)
}

func (self Nanobox) SetCerts(certs []router.KeyPair) error {
	router.UpdateCerts(certs)
	return nil
}

func (self Nanobox) GetCerts() ([]router.KeyPair, error) {
	return router.Keys(), nil
}
