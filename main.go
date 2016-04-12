// -*- mode: go; tab-width: 2; indent-tabs-mode: 1; st-rulers: [70] -*-
// vim: ts=4 sw=4 ft=lua noet
//--------------------------------------------------------------------
// @author Daniel Barney <daniel@nanobox.io>
// @author Greg Linton <lintong@nanobox.io>
// @copyright 2016, Pagoda Box Inc.
// @doc
//
// @end
// Created :    7 August 2015 by Daniel Barney <daniel@nanobox.io>
// Refactored : March 2016    by Greg Linton <lintong@nanobox.io>
//--------------------------------------------------------------------

// Portal is an api-driven, in-kernel layer 2/3 load balancer with
// http/s proxy cababilities.
//
// Usage
//
// To run as a server, using the defaults, starting portal is as simple as
//  portal -s
// For more specific usage information, refer to the help doc (portal -h):
//
//  Usage:
//    portal [flags]
//    portal [command]
//
//  Available Commands:
//    add-service    Add service
//    remove-service Remove service
//    show-service   Show service
//    show-services  Show all services
//    set-services   Set service list
//    set-service    Set service
//    add-server     Add server to a service
//    remove-server  Remove server from a service
//    show-server    Show server on a service
//    show-servers   Show all servers on a service
//    set-servers    Set server list on a service
//    add-route      Add route
//    set-routes     Set route list
//    show-routes    Show all routes
//    remove-route   Remove route
//    add-cert       Add cert
//    set-certs      Set cert list
//    show-certs     Show all certs
//    remove-cert    Remove cert
//
//  Flags:
//    -C, --api-cert="": SSL cert for the api
//    -H, --api-host="127.0.0.1": Listen address for the API
//    -k, --api-key="": SSL key for the api
//    -p, --api-key-password="": Password for the SSL key
//    -P, --api-port="8443": Listen address for the API
//    -t, --api-token="": Token for API Access
//    -r, --cluster-connection="none://": Cluster connection string (redis://127.0.0.1:6379)
//    -T, --cluster-token="": Cluster security token
//    -c, --conf="": Configuration file to load
//    -d, --db-connection="scribble:///var/db/portal": Database connection string
//    -i, --insecure[=false]: Disable tls key checking (client) and listen on http (server)
//    -j, --just-proxy[=false]: Proxy only (no tcp/udp load balancing)
//    -L, --log-file="": Log file to write to
//    -l, --log-level="INFO": Log level to output
//    -x, --proxy-http="0.0.0.0:80": Address to listen on for proxying http
//    -X, --proxy-tls="0.0.0.0:443": Address to listen on for proxying https
//    -s, --server[=false]: Run in server mode
//
//  Use "portal [command] --help" for more information about a command.
//
//
// Build Specs
//
// It is built with clustering at it's core
// to ensure syncronization between nodes. It utilizes a multi-master
// replication system allowing any node to accept requests to update
// load balancing or proxy rules. The high-level workflow is as follows:
//
//  // every call starts by hitting the api
//  API - setRoute
//  - calls cluster.SetRoute
//
//    // in order to ensure syncronization comes first, cluster starts the work
//    CLUSTER - SetRoute
//    - calls publish "set-route"
//
//      // the redis clusterer utilizes the pub/sub functionality for syncronization
//      // when it recieves a message, it calls on "common" to implement the changes
//      SUBSCRIBER - on "set-route"
//      - calls common.SetRoute
//
//        // common contains all the logic to perform an action, as well as roll back
//        // other "systems" upon failure, effectively "undoing" the action
//        COMMON - SetRoute
//        - calls proxymgr.SetRoute & database.SetRoute
//        - rolls back proxymgr if database fails
//
//      SUBSCRIBER
//      - if common.SetRoute was successful, write success to redis for self, otherwise
//        rollback self
//
//    // the cluster member that received the request ensures all members got the update
//    CLUSTER
//    - returns err if not all members have set route
//    - rolls back `common` (via publish) if not all members can set route
//
//  API
//  - if error, return 500 response, otherwise respond as fits
//
// Portal is also extremely "pluggable", meaning that it is easy to code in
// another tool to be a "Balancer" or "Proxy" by matching the interfaces.
// Individual object (database|balancer|proxy) functions are never called
// directly by "common". "Common" calls package level functions such as
//  database.SetService
// which calls/sreturns the selected "Backend's" SetService function.
// The type of backend (scribble/redis) is determined from the connection
// configuration option, as seen in the package's Init() function:
//	url, err := url.Parse(config.ClusterConnection)
//	if err != nil {
//		return fmt.Errorf("Failed to parse db connection - %v", err)
//	}
//
//	switch url.Scheme {
//	case "redis":
//		Clusterer = &Redis{}
//	case "none":
//		Clusterer = &None{}
//	default:
//		Clusterer = &None{}
//	}
//
//	return Clusterer.Init()
package main

import (
	"github.com/nanopack/portal/commands"
)

func main() {
	commands.Portal.Execute()
}
