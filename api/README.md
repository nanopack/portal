[![portal logo](http://nano-assets.gopagoda.io/readme-headers/portal.png)](http://nanobox.io/open-source#portal)  
[![Build Status](https://travis-ci.org/nanopack/portal.svg)](https://travis-ci.org/nanopack/portal)

# Portal

An api-driven, in-kernel layer 2/3 load balancer.

## Routes:

| Route | Description | payload | output |
| --- | --- | --- | --- |
| **Get** /services | List all services | nil | json array of service objects |
| **Post** /services | Add a service | json service object | json service object |
| **Put** /services | Reset the list of services | json array of service objects | json array of service objects |
| **Get** /services/:service_id | Get information about a service | nil | json service object |
| **Delete** /services/:service_id | Delete a service | nil | success message or an error |
| **Get** /services/:service_id/servers | List all servers on a service | nil | json array of server objects |
| **Post** /services/:service_id/servers | Add new server to a service | json server object | json server object |
| **Put** /services/:service_id/servers | Reset the list of servers on a service | json array of server objects | json array of server objects |
| **Get** /services/:service_id/servers/:server_id | Get information about a server on a service | nil | json server object |
| **Delete** /services/:service_id/servers/:server_id | Delete a server from a service | nil | success message or an error |

## Usage Example:

add service
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services -d \
'{"host":"127.0.0.3", "port":1234, "type":"tcp", "scheduler": "rr", "persistence":0, "netmask":""}'
{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":""}
```

list services
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services
[{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":""}]
```

get service
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234
{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":""}
```

add server
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234/servers \
-d '{"host":"192.168.0.1", "port":8080, "forwarder": "m", "weight": 5, "upper_threshold": 10, "lower_threshold": 1}'
{"id":"192_168_0_1-8080","host":"192.168.0.1","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}
```

list servers
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234/servers
[{"id":"192_168_0_1-8080","host":"192.168.0.1","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]
```

get server
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234/servers/192_168_0_1-8080
{"id":"192_168_0_1-8080","host":"192.168.0.1","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}
```

delete server
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234/servers/192_168_0_1-8080 -X DELETE
{"msg":"Success"}
```

list servers
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234/servers
[]
```

delete service
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234 -X DELETE
{"msg":"Success"}
```

list services
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services
[]
```
reset services
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services -d \
'[{"host":"127.0.0.3", "port":1234, "type":"tcp", "scheduler": "rr", "persistence":0, "netmask":"", "servers":[
{"host":"192.168.0.3", "port":8080, "forwarder": "m", "weight": 5, "upper_threshold": 10, "lower_threshold": 1},
{"host":"192.168.0.4", "port":8080, "forwarder": "m", "weight": 4, "upper_threshold": 10, "lower_threshold": 1}]}]' \
-X PUT
[{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":"","servers":[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"id":"192_168_0_4-8080","host":"192.168.0.4","port":8080,"forwarder":"m","weight":4,"upper_threshold":10,"lower_threshold":1}]}]
```

reset servers
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services/tcp-127_0_0_3-1234/servers -d \
'[{"host":"192.168.0.5", "port":8080, "forwarder": "m", "weight": 5, "upper_threshold": 10, "lower_threshold": 1},
{"host":"192.168.0.6", "port":8080, "forwarder": "m", "weight": 4, "upper_threshold": 10, "lower_threshold": 1}]' \
-X PUT
[{"id":"192_168_0_5-8080","host":"192.168.0.5","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"id":"192_168_0_6-8080","host":"192.168.0.6","port":8080,"forwarder":"m","weight":4,"upper_threshold":10,"lower_threshold":1}]
```

list services
```
$ curl -k -H "X-NANOBOX-TOKEN:" https://127.0.0.1:8443/services
[{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":"","servers":[{"id":"192_168_0_5-8080","host":"192.168.0.5","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"id":"192_168_0_6-8080","host":"192.168.0.6","port":8080,"forwarder":"m","weight":4,"upper_threshold":10,"lower_threshold":1}]}]
```

[![portal logo](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)