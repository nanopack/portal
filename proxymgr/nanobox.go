// nanobox.go contains logic to use nanobox-router as a proxy

package proxymgr

import (
	"github.com/jcelliott/lumber"
	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

type Nanobox struct{}

func (self Nanobox) Init() error {
	// we want to see nanobox-router logs
	lumber.Level(lumber.LvlInt(config.LogLevel))

	// start http proxy
	err := router.StartHTTP(config.RouteHttp)
	if err != nil {
		return err
	}
	config.Log.Info("Proxy listening at http://%s...", config.RouteHttp)

	// start https proxy
	err = router.StartTLS(config.RouteTls)
	if err != nil {
		return err
	}
	config.Log.Info("Proxy listening at https://%s...", config.RouteTls)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////

func (self Nanobox) SetRoute(route core.Route) error {
	routes := router.Routes()
	// for idempotency
	for i := range routes {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			return nil
		}
	}

	routes = append(routes, self.rToRoute(route))
	return self.SetRoutes(self.routesToR(routes))
}

func (self Nanobox) DeleteRoute(route core.Route) error {
	routes := router.Routes()
	for i := range routes {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			routes = append(routes[:i], routes[i+1:]...)
			break
		}
	}
	return self.SetRoutes(self.routesToR(routes))
}

func (self Nanobox) SetRoutes(routes []core.Route) error {
	router.UpdateRoutes(self.rToRoutes(routes))
	return nil
}

func (self Nanobox) GetRoutes() ([]core.Route, error) {
	return self.routesToR(router.Routes()), nil
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////

func (self Nanobox) SetCert(cert core.CertBundle) error {
	certs := router.Keys()
	// for idempotency
	for i := range certs {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			return nil
		}
	}

	certs = append(certs, self.cToKey(cert))
	return self.SetCerts(self.keysToC(certs))
}

func (self Nanobox) DeleteCert(cert core.CertBundle) error {
	certs := router.Keys()
	for i := range certs {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			certs = append(certs[:i], certs[i+1:]...)
			break
		}
	}

	return self.SetCerts(self.keysToC(certs))
}

func (self Nanobox) SetCerts(certs []core.CertBundle) error {
	router.UpdateCerts(self.cToKeys(certs))
	return nil
}

func (self Nanobox) GetCerts() ([]core.CertBundle, error) {
	return self.keysToC(router.Keys()), nil
}

func (self Nanobox) cToKeys(certs []core.CertBundle) []router.KeyPair {
	var keyPairs []router.KeyPair
	for i := range certs {
		keyPairs = append(keyPairs, router.KeyPair{Key: certs[i].Key, Cert: certs[i].Cert})
	}
	return keyPairs
}

func (self Nanobox) keysToC(certs []router.KeyPair) []core.CertBundle {
	var keyPairs []core.CertBundle
	for i := range certs {
		keyPairs = append(keyPairs, core.CertBundle{Key: certs[i].Key, Cert: certs[i].Cert})
	}
	return keyPairs
}

func (self Nanobox) rToRoutes(routes []core.Route) []router.Route {
	var rts []router.Route
	for i := range routes {
		rts = append(rts, router.Route{SubDomain: routes[i].SubDomain, Domain: routes[i].Domain, Path: routes[i].Path, Targets: routes[i].Targets, FwdPath: routes[i].FwdPath, Page: routes[i].Page})
	}
	return rts
}

func (self Nanobox) routesToR(routes []router.Route) []core.Route {
	var rts []core.Route
	for i := range routes {
		rts = append(rts, core.Route{SubDomain: routes[i].SubDomain, Domain: routes[i].Domain, Path: routes[i].Path, Targets: routes[i].Targets, FwdPath: routes[i].FwdPath, Page: routes[i].Page})
	}
	return rts
}

func (self Nanobox) cToKey(cert core.CertBundle) router.KeyPair {
	return router.KeyPair{Key: cert.Key, Cert: cert.Cert}
}

func (self Nanobox) rToRoute(route core.Route) router.Route {
	return router.Route{SubDomain: route.SubDomain, Domain: route.Domain, Path: route.Path, Targets: route.Targets, FwdPath: route.FwdPath, Page: route.Page}
}
