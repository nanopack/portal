# Portal

An api-driven, in-kernel layer 2/3 load balancer.

### Status
Incomplete/Experimental/Unstable

## Memory Usage

Currently portal uses 900k of ram while idling.


## Routes :

| Route | Description | payload | output |
| --- | --- | --- | --- |
| **GET** /vips | list vips that are currently being load balanced | nil | `{"vips": [{"ip":"127.0.0.1"}]}` |
| **POST** /vips | create a new ip to be load balanced | `{"method":"rr" ,"ip":"127.0.0.1" ,"port":80}` | `{"sucess":"true"}` |
| **DELETE** /vips/:vip | remove an ip from the load balancer | nil | `{"sucess":"true"}` |
| **GET** /vips/:vip/servers | list servers that are being load balanced for the :vip | nil | `{"servers": [{"server":"127.0.0.1:1234" ,"weight":1000}]}` |
| **POST** /vips/:vip/servers | add a server to the load balancing group for :vip | `{"server":"127.0.0.1:1234" ,"weight":1000}` | `{"sucess":"true"}` |
| **PUT** /vips/:vip/servers/:server | enable or disable a server without removing it from the pool | `{"enabled":true}` | `{"sucess":"true"}` |
| **DELETE** /vips/:vip/servers/:server | remove :server from the load balancing group for :vip | nil | `{"sucess":"true"}` |

### Notes:

- Nothing is stored in an intermediary database, ipvsadm is considered the backend datastore for the api.
- There is a copy of the rules that are generated after every update stored at `/somewhere/ipvsadm.rules` that is used to initialize ipvsadm on boot.
