// routemgr handles the adding of 'routes' (subdomain.domain/path sets)
// and their 'targets' or 'page', creating a reverse proxy router
package routemgr

import (
	"fmt"

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

// Start both http and tls servers
func Init() error {
	Router = &Nanobox{}
	return router.Start(fmt.Sprintf("0.0.0.0:%v", config.RoutePortHttp), fmt.Sprintf("0.0.0.0:%v", config.RoutePortTls))
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
