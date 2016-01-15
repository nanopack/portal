[![portal logo](http://nano-assets.gopagoda.io/readme-headers/portal.png)](http://nanobox.io/open-source#portal)
[![Build Status](https://travis-ci.org/nanopack/portal.svg)](https://travis-ci.org/nanopack/portal)

# Portal

An api-driven, in-kernel layer 2/3 load balancer.

### Status
Incomplete/Experimental/Unstable

## Memory Usage

Currently portal uses 900k of ram while idling.

## Data types:
### Service:
json: {"host": "127.0.0.1", "port": 1234, "type": "tcp", "scheduler": "wlc", "persistence": 300, "netmask": "255.255.255.0", "servers": []}

Fields:
- host: IP of the host the service is bound to.
- port: Port that the service listens to.
- type: Type of service. Either tcp or udp.
- scheduler: How to pick downstream server. On of the following: rr, wrr, lc, wlc, lblc, lblcr, dh, sh, sed, nq
- persistence: Timeout for keeping requests from the same client going to the same server
- netmask: How to group clients with persistence to servers
- servers: Array of server objects associated to the service

### Server:
json: {"host": "127.0.0.1", "port": 1234, "forwarder": "m", "weight": 1, "upper_threshold": 0, "lower_threshold": 0}

Fields:
- host: IP of the host the service is bound to.
- port: Port that the service listens to.
- forwarder: Method to use to forward traffic to this server. One of the following: g (gatewaying), i (ipip), m (masquerading)
- weight: Weight to perfer this server. Set to 0 if no traffic should go to this server.
- upper_threshold: Stop sending connections to this server when this number is reached. 0 is no limit.
- lower_threshold: Restart sending connections when drains down to this number. 0 is not set.

## Routes :

| Route | Description | payload | output |
| --- | --- | --- | --- |
| **Get** /services/:type/:service_ip/:service_port/servers/:server_ip/:server_port | Get information about a server on a service | nil | json server object |
| **Post** /services/:type/:service_ip/:service_port/servers/:server_ip/:server_port | Add new server to a service | service json string | nil or an error |
| **Delete** /services/:type/:service_ip/:service_port/servers/:server_ip/:server_port | Delete a server from a service | nil | nil or an error |
| **Get** /services/:type/:service_ip/:service_port/servers | List all servers on a service | nil | json array of server objects |
| **Post** /services/:type/:service_ip/:service_port/servers | Reset the list of servers on a service | json array of server objects | nil or an error |
| **Get** /services/:type/:service_ip/:service_port | Get information about a service | nil | json service object |
| **Post** /services/:type/:service_ip/:service_port | Add a service | json service object | nil or an error |
| **Delete** /services/:type/:service_ip/:service_port | Delete a service | nil | nil or an error |
| **Get** /services | List all services | nil | json array of server objects |
| **Post** /services | Reset the list of services | json array of server objects | nil or an error |
| **Get** /sync | Sync portal with data in LVS | nil | nil or an error |
| **Post** /sync | Sync LVS with data in portal | nil | nil or an error |

[![portal logo](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
