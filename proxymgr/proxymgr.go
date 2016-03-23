// proxymgr handles the adding of 'routes' (subdomain.domain/path sets)
// and their 'targets' or 'page', creating a reverse proxy router. It also
// handles the adding of 'certs' (key/cert sets) for a secure reverse proxy router
package proxymgr

import "github.com/nanopack/portal/core"

type proxyable interface {
	Init() error
	core.Proxyable
}

var (
	Proxy proxyable
)

// todo: for improved pluggable-ness, maybe define Route here? // is proxies stored?

// start http server
func Init() error {
	Proxy = &Nanobox{}
	return Proxy.Init()
}

func SetRoute(route core.Route) error {
	return Proxy.SetRoute(route)
}

func DeleteRoute(route core.Route) error {
	return Proxy.DeleteRoute(route)
}

func SetRoutes(routes []core.Route) error {
	return Proxy.SetRoutes(routes)
}

func GetRoutes() ([]core.Route, error) {
	return Proxy.GetRoutes()
}

func SetCerts(certs []core.CertBundle) error {
	return Proxy.SetCerts(certs)
}

func SetCert(cert core.CertBundle) error {
	return Proxy.SetCert(cert)
}

func DeleteCert(cert core.CertBundle) error {
	return Proxy.DeleteCert(cert)
}

func GetCerts() ([]core.CertBundle, error) {
	return Proxy.GetCerts()
}
