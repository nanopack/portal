[![portal logo](http://nano-assets.gopagoda.io/readme-headers/portal.png)](http://nanobox.io/open-source#portal)  
[![Build Status](https://travis-ci.org/nanopack/portal.svg)](https://travis-ci.org/nanopack/portal)

# Portal

An api-driven, in-kernel layer 2/3 load balancer.

## CLI Commands:

```
Usage:
   [flags]
   [command]

Available Commands:
  add-service    Add service
  remove-service Remove service
  show-service   Show service
  show-services  Show all services
  set-services   Set service list
  set-service    Set service
  add-server     Add server to a service
  remove-server  Remove server from a service
  show-server    Show server on a service
  show-servers   Show all servers on a service
  set-servers    Set server list on a service

Flags:
  -C, --api-crt="": SSL cert for the api
  -H, --api-host="127.0.0.1": Listen address for the API
  -k, --api-key="": SSL key for the api
  -p, --api-key-password="": Password for the SSL key
  -P, --api-port="8443": Listen address for the API
  -t, --api-token="": Token for API Access
  -r, --cluster-connection="redis://127.0.0.1:6379": Database connection string
  -c, --conf="": Configuration file to load
  -d, --db-connection="scribble:///var/db/portal": Database connection string
  -i, --insecure[=true]: Disable tls key checking
  -l, --log-file="": Log file to write to
  -L, --log-level="INFO": Log level to output
  -s, --server[=false]: Run in server mode

Use " [command] --help" for more information about a command.
```

## Usage Example:

add service
```
$ ./portal add-service -O "127.0.0.3" -R 1234 -T "tcp" -s "rr" -e 0 -n ""
{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":""}
$ ./portal add-service -j '{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":"","servers":[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"id":"192_168_0_4-8080","host":"192.168.0.4","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]}'
{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":"","servers":[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"id":"192_168_0_4-8080","host":"192.168.0.4","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]}
```

show services
```
$ ./portal show-services
[{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":""}]
```

show service
```
$ ./portal show-service -I "tcp-127_0_0_3-1234"
{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":""}
```

add server
```
$ ./portal add-server -I "tcp-127_0_0_3-1234" -o "192.168.0.3" -p 8080 -f "m" -w 5 -u 10 -l 1
{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}
$ ./portal add-server -I "tcp-127_0_0_3-1234" -j '{"host":"192.168.0.3", "port":8080, "forwarder": "m", "weight": 5, "upper_threshold": 10, "lower_threshold": 1}'
{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}
```

show servers
```
$ ./portal show-servers -O "127.0.0.3" -R "1234"
[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]
```

show server
```
$ ./portal show-server -I "tcp-127_0_0_3-1234" -S "192_168_0_3-8080"
{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}
```

remove server
```
$ ./portal remove-server -I "tcp-127_0_0_3-1234" -S "192_168_0_3-8080"
{"msg":"Success"}
```

show servers
```
$ ./portal show-servers -O "127.0.0.3" -R "1234"
[]
```

remove service
```
$ ./portal remove-service -I "tcp-127_0_0_3-1234"
{"msg":"Success"}
```

show services
```
$ ./portal show-services
[]
```

reset services
```
$ ./portal set-services -j '[{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":"","servers":[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"id":"192_168_0_4-8080","host":"192.168.0.4","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]}]'
[{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":"","servers":[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1},{"id":"192_168_0_4-8080","host":"192.168.0.4","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]}]
```

reset servers
```
$ ./portal set-servers -I "tcp-127_0_0_3-1234" -j '[{"host":"192.168.0.3", "port":8080, "forwarder": "m", "weight": 5, "upper_threshold": 10, "lower_threshold": 1}]'
[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]
```

show services
```
$ ./portal show-services
[{"id":"tcp-127_0_0_3-1234","host":"127.0.0.3","port":1234,"type":"tcp","scheduler":"rr","persistence":0,"netmask":"","servers":[{"id":"192_168_0_3-8080","host":"192.168.0.3","port":8080,"forwarder":"m","weight":5,"upper_threshold":10,"lower_threshold":1}]}]
```

[![portal logo](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
