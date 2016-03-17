// nanobox.go contains logic to use nanobox-router as a router

package routemgr

import (
	"github.com/nanobox-io/nanobox-router"
)

type Nanobox struct{}

func (self Nanobox) SetRoute(route router.Route) error {
	routes := router.Routes()
	// for idempotency
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			return nil
		}
	}

	routes = append(routes, route)
	return self.SetRoutes(routes)
}

func (self Nanobox) DeleteRoute(route router.Route) error {
	routes := router.Routes()
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			routes = append(routes[:i], routes[i+1:]...)
			break
		}
	}
	return self.SetRoutes(routes)
}

func (self Nanobox) SetRoutes(routes []router.Route) error {
	router.UpdateRoutes(routes)
	return nil
}

func (self Nanobox) GetRoutes() ([]router.Route, error) {
	return router.Routes(), nil
}
