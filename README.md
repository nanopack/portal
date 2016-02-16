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
json:
```json
{
  "host": "127.0.0.1",
  "port": 1234,
  "type": "tcp",
  "scheduler": "wlc",
  "persistence": 300,
  "netmask": "255.255.255.0",
  "servers": []
}
```

json with a server:
```json
{
  "host": "127.0.0.1",
  "port": 8080,
  "type": "tcp",
  "scheduler": "wlc",
  "persistence": 300,
  "netmask": "",
  "servers": [
    {
      "host": "172.28.128.4",
      "port": 8081,
      "forwarder": "m",
      "weight": 1,
      "UpperThreshold": 0,
      "LowerThreshold": 0
    }
  ]
}
```

Fields:
- **host**: IP of the host the service is bound to.
- **port**: Port that the service listens to.
- **type**: Type of service. Either tcp or udp.
- **scheduler**: How to pick downstream server. On of the following: rr, wrr, lc, wlc, lblc, lblcr, dh, sh, sed, nq
- **persistence**: Timeout for keeping requests from the same client going to the same server
- **netmask**: How to group clients with persistence to servers
- **servers**: Array of server objects associated to the service

### Server:
json:
```json
{
  "host": "127.0.0.1",
  "port": 1234,
  "forwarder": "m",
  "weight": 1,
  "upper_threshold": 0,
  "lower_threshold": 0
}
```

Fields:
- **host**: IP of the host the service is bound to.
- **port**: Port that the service listens to.
- **forwarder**: Method to use to forward traffic to this server. One of the following: g (gatewaying), i (ipip), m (masquerading)
- **weight**: Weight to perfer this server. Set to 0 if no traffic should go to this server.
- **upper_threshold**: Stop sending connections to this server when this number is reached. 0 is no limit.
- **lower_threshold**: Restart sending connections when drains down to this number. 0 is not set.

### Error:
json:
```json
{
  "error": "exit status 2: unexpected argument \n"
}
```

Fields:
 - *error**: Error message

### Message:
json:
```json
{
  "msg": "Success"
}
```

Fields:
 - **msg**: Success message

## Routes:

| Route | Description | payload | output |
| --- | --- | --- | --- |
| **Get** /services | List all services | nil | json array of service objects |
| **Post** /services | Add a service | json service object | json service object |
| **Put** /services | Reset the list of services | json array of service objects | json array of service objects |
| **Put** /services/:service_id | Reset the specified service | nil | json service object |
| **Get** /services/:service_id | Get information about a service | nil | json service object |
| **Delete** /services/:service_id | Delete a service | nil | success message or an error |
| **Get** /services/:service_id/servers | List all servers on a service | nil | json array of server objects |
| **Post** /services/:service_id/servers | Add new server to a service | json server object | json server object |
| **Put** /services/:service_id/servers | Reset the list of servers on a service | json array of server objects | json array of server objects |
| **Get** /services/:service_id/servers/:server_id | Get information about a server on a service | nil | json server object |
| **Delete** /services/:service_id/servers/:server_id | Delete a server from a service | nil | success message or an error |

## Usage:

See [Api](api/README.md)
See [Cli](commands/README.md)  

[![portal logo](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
