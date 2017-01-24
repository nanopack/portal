package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/nanopack/portal/core"
)

// add-route
// remove-route
// show-routes
// set-routes

var (
	routeAddCmd = &cobra.Command{
		Use:   "add-route",
		Short: "Add route",
		Long:  ``,

		Run: routeAdd,
	}
	routeRemoveCmd = &cobra.Command{
		Use:   "remove-route",
		Short: "Remove route",
		Long:  ``,

		Run: routeRemove,
	}
	routesShowCmd = &cobra.Command{
		Use:   "show-routes",
		Short: "Show all routes",
		Long:  ``,

		Run: routesShow,
	}
	routesSetCmd = &cobra.Command{
		Use:   "set-routes",
		Short: "Set route list",
		Long:  ``,

		Run: routesSet,
	}
	routeJsonString string
	route           core.Route
)

func init() {
	routeSimpleFlags(routeAddCmd) // must use json string because Targets is a []string
	routeSimpleFlags(routesSetCmd)
	routeComplexFlags(routeRemoveCmd)

}

func routeSimpleFlags(ccmd *cobra.Command) {
	ccmd.Flags().StringVarP(&routeJsonString, "json", "j", "", "Json encoded data for route(s)")
}

func routeComplexFlags(ccmd *cobra.Command) {
	ccmd.Flags().StringVarP(&route.SubDomain, "subdomain", "s", "", "Subdomain to match by")
	ccmd.Flags().StringVarP(&route.Domain, "domain", "d", "", "Domain to match by")
	ccmd.Flags().StringVarP(&route.Path, "path", "p", "", "Path to match by")
}

func routeAdd(ccmd *cobra.Command, args []string) {
	rte := core.Route{}

	err := json.Unmarshal([]byte(routeJsonString), &rte)
	if err != nil {
		fail("Bad JSON syntax")
	}

	jsonBytes, err := json.Marshal(rte)
	if err != nil {
		fail("Bad values for route")
	}
	res, err := rest("routes", "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func routeRemove(ccmd *cobra.Command, args []string) {
	if routeJsonString != "" {
		err := json.Unmarshal([]byte(routeJsonString), &route)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}

	jsonBytes, err := json.Marshal(route)
	if err != nil {
		fail("Bad values for route")
	}
	res, err := rest("routes", "DELETE", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func routesShow(ccmd *cobra.Command, args []string) {
	res, err := rest("routes", "GET", nil)
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func routesSet(ccmd *cobra.Command, args []string) {
	routes := []core.Route{}

	err := json.Unmarshal([]byte(routeJsonString), &routes)
	if err != nil {
		fail("Bad JSON syntax")
	}
	jsonBytes, err := json.Marshal(routes)
	if err != nil {
		fail("Bad values for route")
	}
	res, err := rest("routes", "PUT", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}
