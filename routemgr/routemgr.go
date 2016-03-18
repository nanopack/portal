// routemgr handles the adding of 'routes' (subdomain.domain/path sets)
// and their 'targets' or 'page', creating a reverse proxy router
package routemgr

import (
	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/config"
)

type Routable interface {
	SetRoute(route router.Route) error
	SetRoutes(routes []router.Route) error
	DeleteRoute(route router.Route) error
	GetRoutes() ([]router.Route, error)
}

var (
	Router Routable
)

// todo: for improved pluggable-ness, maybe define Route here?

// start http server
func Init() error {
	Router = &Nanobox{}
	return router.StartHTTP(config.RouteHttp)
}

func SetRoute(route router.Route) error {
	return Router.SetRoute(route)
}

func DeleteRoute(route router.Route) error {
	return Router.DeleteRoute(route)
}

func SetRoutes(routes []router.Route) error {
	return Router.SetRoutes(routes)
}

func GetRoutes() ([]router.Route, error) {
	return Router.GetRoutes()
}
