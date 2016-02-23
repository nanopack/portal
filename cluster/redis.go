package cluster

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

var (
	self string
	uri  *url.URL
	ttl  = 120
	beat = ttl / 2
)

type (
	Redis struct {
		pubconn redis.Conn
		subconn redis.PubSubConn
	}
)

func (r *Redis) Init() error {
	var err error
	uri, err = url.Parse(config.ClusterConnection)
	if err != nil {
		return err
	}
	hostname, _ := os.Hostname()
	self = fmt.Sprintf("%v:%v", hostname, config.ApiPort)
	c, err := redis.DialURL(fmt.Sprintf("%s", uri))
	aerr := auth(c)
	if err != nil || aerr != nil {
		c.Close()
		return fmt.Errorf("Failed to reach redis for pubconn - %v %v", err, aerr)
	}
	r.pubconn = c

	c, err = redis.DialURL(fmt.Sprintf("%s", uri))
	aerr = auth(c)
	if err != nil || aerr != nil {
		c.Close()
		return fmt.Errorf("Failed to reach redis for subconn - %v %v", err, aerr)
	}
	r.subconn = redis.PubSubConn{c}
	r.subconn.Subscribe("portal")
	// go subscribe(r)

	// todo: get list of cluster members to fetch services from
	// todo: loop attempts for x minutes
	services, err := r.GetServices()
	if err != nil {
		return fmt.Errorf("Failed to get services - %v", err)
	}
	if services != nil && len(services) != 0 {
		config.Log.Trace("[cluster] - Setting services...")
		err = common.SetServices(services)
		if err != nil {
			return fmt.Errorf("Failed to set services - %v", err)
		}
	}

	// common.SetServices(services)
	// todo: if at least one member is not up, wait for x mins, retrying then fail (cluster bad state)

	_, err = r.pubconn.Do("SADD", "members", self)
	if err != nil {
		return fmt.Errorf("Failed to add myself to list of members - %v", err)
	}
	r.pubconn.Do("SET", self, "alive", "EX", ttl)

	go subscribe(r)
	go heartbeat()
	go cleanup()

	return nil
}

// authenticate
func auth(conn redis.Conn) error {
	if config.ApiToken == "" {
		return nil
	}
	_, err := conn.Do("AUTH", config.ApiToken)
	if err != nil && strings.Contains(err.Error(), "no password is set") {
		err = nil
	}
	return err
}

// record member still alive
func heartbeat() {
	tick := time.Tick(time.Duration(beat) * time.Second)
	r, err := redis.DialURL((fmt.Sprintf("%s", uri)))
	aerr := auth(r)
	if err != nil || aerr != nil {
		r.Close()
		config.Log.Error("[cluster] - Failed to reach redis for heartbeat - %v %v", err, aerr)
		os.Exit(1)
	}

	for _ = range tick {
		r.Do("SET", self, "alive", "EX", ttl)
	}
}

// cleans up members not present after ttl seconds
func cleanup() {
	// todo: need to clean up faster, and shorten ttl, if member's dead, its dead
	//   - if i die before anyone cleans me up, and a change happens, i'm out of sync. // isn't possible
	//   - if someone comes online, they may ask me for services, because i'm always subscribing, they can get my out of date info..
	//     - need to subscribe early otherwise impossible to recover from blackout without manually removing all bad 'members' from db(redis)
	//       - change blackout recovery method.. if i'm in list, nobody got an update because they block, just use my own info?
	//                  // always subscribing allows anyone to ask me for my stuff
	// cycle every second to check for dead members
	tick := time.Tick(time.Second)
	r, err := redis.DialURL((fmt.Sprintf("%s", uri)))
	aerr := auth(r)
	if err != nil || aerr != nil {
		r.Close()
		config.Log.Error("[cluster] - Failed to reach redis for cleanup - %v %v", err, aerr)
		os.Exit(1)
	}

	for _ = range tick {
		// get list of members that should be alive
		members, _ := redis.Strings(r.Do("SMEMBERS", "members"))
		for _, member := range members {
			// if the member timed out, remove the member from the member set
			exist, _ := redis.Int(r.Do("EXISTS", member))
			if exist == 0 {
				r.Do("SREM", "members", member)
				config.Log.Info("[cluster] - Member '%v' assumed dead. Removed.", member)
			}
		}
		// members, _ := redis.Strings(r.Do("SMEMBERS", "members"))
		// config.Log.Trace("[cluster] - Members: ", members)
	}
}

func (r *Redis) UnInit() error {
	// remove myself from cluster(called on clean quit)
	_, err := r.pubconn.Do("SREM", "members", self)
	_, err = r.pubconn.Do("DEL", self)
	return err
}

func subscribe(r *Redis) {
	config.Log.Info("[cluster] - Redis subscribing on %s...", uri)

	for {
		switch v := r.subconn.Receive().(type) {
		case redis.Message:
			switch pdata := strings.Split(string(v.Data), " "); pdata[0] {
			// SERVICES ///////////////////////////////////////////////////////////////////////////////////////////////
			case "get-services":
				if len(pdata) != 2 {
					config.Log.Error("[cluster] - services not passed in message")
					break
				}
				member := pdata[1]

				if member == self {
					svcs, err := common.GetServices()
					if err != nil {
						config.Log.Error("[cluster] - Failed to get services - %v", err.Error())
						break
					}
					services, err := json.Marshal(svcs)
					if err != nil {
						config.Log.Error("[cluster] - Failed to marshal services - %v", err.Error())
						break
					}
					config.Log.Debug("[cluster] - get-services requested, publishing my services")
					r.pubconn.Do("PUBLISH", "services", fmt.Sprintf("%s", services))
				}
			case "set-services":
				if len(pdata) != 2 {
					config.Log.Error("[cluster] - services not passed in message")
					break
				}
				services, err := marshalSvcs([]byte(pdata[1]))
				if err != nil {
					config.Log.Error("[cluster] - Failed to marshal services - %v", err.Error())
					break
				}
				err = common.SetServices(*services)
				if err != nil {
					config.Log.Error("[cluster] - Failed to set services - %v", err.Error())
					break
				}
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-services %s", *services))))
				config.Log.Trace("[cluster] - set-services hash - %v", actionHash)
				r.pubconn.Do("SADD", actionHash, self)
			case "set-service":
				if len(pdata) != 2 {
					// shouldn't happen unless redis is not secure and someone manually `publishService`es
					// todo: secure with config.Token?
					config.Log.Error("[cluster] - service not passed in message")
					break
				}
				svc, err := marshalSvc([]byte(pdata[1]))
				if err != nil {
					config.Log.Error("[cluster] - Failed to marshal service - %v", err.Error())
					break
				}
				err = common.SetService(svc)
				if err != nil {
					config.Log.Error("[cluster] - Failed to set service - %v", err.Error())
					break
				}
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-service %s", *svc))))
				r.pubconn.Do("SADD", actionHash, self)
			case "delete-service":
				if len(pdata) != 2 {
					config.Log.Error("[cluster] - service id not passed in message")
					break
				}
				svcId := pdata[1]
				err := common.DeleteService(svcId)
				if err != nil {
					config.Log.Error("[cluster] - Failed to delete service - %v", err.Error())
					break
				}
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("delete-service %s", svcId))))
				r.pubconn.Do("SADD", actionHash, self)
				// SERVERS ///////////////////////////////////////////////////////////////////////////////////////////////
			case "set-servers":
				if len(pdata) != 3 {
					config.Log.Error("[cluster] - service id not passed in message")
					break
				}
				svcId := pdata[2]
				servers, err := marshalSrvs([]byte(pdata[1]))
				if err != nil {
					config.Log.Error("[cluster] - Failed to marshal server - %v", err.Error())
					break
				}
				err = common.SetServers(svcId, *servers)
				if err != nil {
					config.Log.Error("[cluster] - Failed to set servers - %v", err.Error())
					break
				}
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-servers %s %s", *servers, svcId))))
				r.pubconn.Do("SADD", actionHash, self)
			case "set-server":
				if len(pdata) != 3 {
					// shouldn't happen unless redis is not secure and someone manually publishServicees
					// todo: secure with config.Token?
					config.Log.Error("[cluster] - service id not passed in message")
					break
				}
				svcId := pdata[2]
				server, err := marshalSrv([]byte(pdata[1]))
				if err != nil {
					config.Log.Error("[cluster] - Failed to marshal server - %v", err.Error())
					break
				}
				err = common.SetServer(svcId, server)
				if err != nil {
					config.Log.Error("[cluster] - Failed to set server - %v", err.Error())
					break
				}
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-server %s", *server, svcId))))
				r.pubconn.Do("SADD", actionHash, self)
			case "delete-server":
				if len(pdata) != 3 {
					config.Log.Error("[cluster] - service id id not passed in message")
					break
				}
				srvId := pdata[1]
				svcId := pdata[2]
				err := common.DeleteServer(svcId, srvId)
				if err != nil {
					config.Log.Error("[cluster] - Failed to delete server - %v", err.Error())
					break
				}
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("delete-server %s", srvId, svcId))))
				r.pubconn.Do("SADD", actionHash, self)

			default:
				config.Log.Error("[cluster] - Recieved data on %v: %v", v.Channel, string(v.Data))
			}
		case error:
			config.Log.Error("[cluster] - Subscriber failed to receive - %v", v.Error())
			if strings.Contains(v.Error(), "closed network connection") {
				// todo: is it ok to exit if we can't communicate with db?
				// exit so we don't get spammed with logs
				os.Exit(1)
			}
			continue
		}
	}
}

// func publishService(r *Redis) {
func (r Redis) publishService(action string, v interface{}) error {
	// func (r Redis) publishService(message string) error {
	s, err := json.Marshal(v)
	if err != nil {
		return BadJson
	}

	_, err = r.pubconn.Do("PUBLISH", "portal", fmt.Sprintf("%s %s", action, s))
	// SADD-ing in subscriber
	// if err != nil {
	// 	actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s %s", action, s))))
	// 	_, err2 := r.pubconn.Do("SADD", actionHash, self)
	// 	if err2 != nil {
	// 		err = fmt.Errorf("%v - %v", err, err2)
	// 	}
	// }
	return err
}

// func publishServer(r *Redis) {
func (r Redis) publishServer(action, svcId string, v interface{}) error {
	s, err := json.Marshal(v)
	if err != nil {
		return BadJson
	}

	_, err = r.pubconn.Do("PUBLISH", "portal", fmt.Sprintf("%s %s", action, s, svcId))
	return err
}

func (r Redis) waitForMembers(actionHash string) error {
	config.Log.Trace("[cluster] - Waiting for updates to %s", actionHash)
	// todo: make timeout configurable
	timeout := time.After(30 * time.Second)
	tick := time.Tick(500 * time.Millisecond)
	var list []string
	var err error
	for {
		select {
		case <-tick:
			// compare who we know about, to who performed the update
			list, err = redis.Strings(r.pubconn.Do("SDIFF", "members", actionHash))
			if err != nil {
				return err
			}
			if len(list) == 0 {
				// if all members respond, all is well
				return nil
			}
		// if members don't respond in time, return error
		case <-timeout:
			return fmt.Errorf("Member(s) '%v' failed to set-service", list)
		}
	}
}

// SetServices
func (r *Redis) SetServices(services []core.Service) error {
	oldServices, err := common.GetServices()
	if err != nil {
		return err
	}

	// publishService to others
	err = r.publishService("set-services", services)
	if err != nil {
		// if i failed to publishService, request should fail
		return err
	}

	actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-services %s", services))))

	// ensure all members applied action
	err = r.waitForMembers(actionHash)
	if err != nil {
		uActionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-services %s", oldServices))))
		// cleanup rollback cruft. clear actionHash ensures no mistakes on re-submit
		defer r.pubconn.Do("DEL", uActionHash, actionHash)
		// attempt rollback - no need to waitForMembers here
		uerr := r.publishService("set-services", oldServices)
		if uerr != nil {
			err = fmt.Errorf("%v - %v", err, uerr)
		}
		return err
	}

	// clear cruft
	_, err = r.pubconn.Do("DEL", actionHash)
	if err != nil {
		return err
	}

	return nil
}

// SetService would be called following common.SetService, no need to do it again
func (r *Redis) SetService(service *core.Service) error {
	// publishService to others
	err := r.publishService("set-service", service)
	if err != nil {
		// nothing to rollback yet (nobody received)
		return err
	}

	actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-service %s", *service))))

	// ensure all members applied action
	err = r.waitForMembers(actionHash)
	if err != nil {
		uActionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("delete-service %s", service.Id))))
		// cleanup rollback cruft. clear actionHash ensures no mistakes on re-submit
		defer r.pubconn.Do("DEL", uActionHash, actionHash)
		// attempt rollback - no need to waitForMembers here
		uerr := r.publishService("delete-service", service.Id)
		if uerr != nil {
			err = fmt.Errorf("%v - %v", err, uerr)
		}
		return err
	}

	// clear cruft
	_, err = r.pubconn.Do("DEL", actionHash)
	if err != nil {
		return err
	}

	return nil
}

// DeleteService
func (r *Redis) DeleteService(id string) error {
	oldService, err := common.GetService(id)
	if err != nil { // todo: this should not return nil (ensure gone from cluster)
		// if invalid service, 'delete' succesful
		return nil // common.DeleteService as example/extended functionality
	}

	// publishService to others
	err = r.publishService("delete-service", id)
	if err != nil {
		// api will rollback my things with common.SetService
		// if i failed to publishService, request should fail
		return err
	}

	actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("delete-service %s", id))))

	// ensure all members applied action
	err = r.waitForMembers(actionHash)
	if err != nil {
		uActionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("add-service %s", oldService))))
		// cleanup rollback cruft. clear actionHash ensures no mistakes on re-submit
		defer r.pubconn.Do("DEL", uActionHash, actionHash)
		// attempt rollback - no need to waitForMembers here
		uerr := r.publishService("add-service", oldService)
		if uerr != nil {
			err = fmt.Errorf("%v - %v", err, uerr)
		}
		return err
	}

	// clear cruft
	_, err = r.pubconn.Do("DEL", actionHash)
	if err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SERVERS
////////////////////////////////////////////////////////////////////////////////
// SetServers
func (r *Redis) SetServers(svcId string, servers []core.Server) error {
	service, err := common.GetService(svcId)
	if err != nil {
		return NoServiceError
	}
	oldServers := service.Servers

	// publishServer to others
	err = r.publishServer("set-servers", svcId, servers)
	if err != nil {
		return err
	}

	actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-servers %s %s", servers, svcId))))

	// ensure all members applied action
	err = r.waitForMembers(actionHash)
	if err != nil {
		uActionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-servers %s %s", oldServers, svcId))))
		// cleanup rollback cruft. clear actionHash ensures no mistakes on re-submit
		defer r.pubconn.Do("DEL", uActionHash, actionHash)
		// attempt rollback - no need to waitForMembers here
		uerr := r.publishServer("set-servers", svcId, oldServers)
		if uerr != nil {
			err = fmt.Errorf("%v - %v", err, uerr)
		}
		return err
	}

	return nil
}

// SetServer
func (r *Redis) SetServer(svcId string, server *core.Server) error {
	// publishServer to others
	err := r.publishServer("set-server", svcId, server)
	if err != nil {
		return err
	}

	actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-server %s %s", *server, svcId))))

	// ensure all members applied action
	err = r.waitForMembers(actionHash)
	if err != nil {
		uActionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("delete-server %s %s", server.Id, svcId))))
		// cleanup rollback cruft. clear actionHash ensures no mistakes on re-submit
		defer r.pubconn.Do("DEL", uActionHash, actionHash)
		// attempt rollback - no need to waitForMembers here
		uerr := r.publishServer("delete-server", server.Id, svcId)
		if uerr != nil {
			err = fmt.Errorf("%v - %v", err, uerr)
		}
		return err
	}

	return nil
}

// DeleteServer
func (r *Redis) DeleteServer(svcId, srvId string) error {
	oldServer, err := common.GetServer(svcId, srvId)
	if err != nil {
		// if service not valid, 'delete' successful
		return nil
	}

	// publishServer to others
	err = r.publishServer("delete-server", svcId, srvId)
	if err != nil {
		return err
	}

	actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("delete-server %s %s", srvId, svcId))))

	// ensure all members applied action
	err = r.waitForMembers(actionHash)
	if err != nil {
		uActionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-server %s %s", *oldServer, svcId))))
		// cleanup rollback cruft. clear actionHash ensures no mistakes on re-submit
		defer r.pubconn.Do("DEL", uActionHash, actionHash)
		// attempt rollback - no need to waitForMembers here
		uerr := r.publishServer("add-server", svcId, oldServer)
		if uerr != nil {
			err = fmt.Errorf("%v - %v", err, uerr)
		}
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// GETS
////////////////////////////////////////////////////////////////////////////////
// GetService - likely will never be called
func (r *Redis) GetService(id string) (*core.Service, error) {
	return common.GetService(id)
}

// GetServices
func (r *Redis) GetServices() ([]core.Service, error) {
	// get known members(other than me) to 'poll' for services
	members, _ := redis.Strings(r.pubconn.Do("SMEMBERS", "members"))
	if len(members) == 0 {
		// todo: this probably isn't ok to assume. update: probably ok because would only happen on new cluster
		// assume i'm ok to be master so don't reset imported services
		config.Log.Trace("[cluster] - Assuming OK to be master, using services from my database...")
		return nil, nil
	}
	for i := range members {
		if members[i] == self {
			// if i'm in the list of members, new requests should have failed while `waitForMembers`ing
			config.Log.Trace("[cluster] - Assuming I was in sync, using services from my database...")
			return nil, nil
		}
	}

	c, err := redis.DialURL(fmt.Sprintf("%s", uri))
	aerr := auth(c)
	if err != nil || aerr != nil {
		c.Close()
		config.Log.Error("[cluster] - Failed to reach redis for subconn - %v %v", err, aerr)
		os.Exit(1)
		return nil, err
	}
	message := make(chan interface{})
	subconn := redis.PubSubConn{c}

	// subscribe to channel that services will be published on
	subconn.Subscribe("services")
	defer subconn.Close()

	// listen always
	go func() {
		for {
			message <- subconn.Receive()
		}
	}()

	// todo: loop attempts for x minutes, allows last dead members to start back up
	//   - why not just use my db if i'm in the list of 'last dead members'?
	//     - ah, because if i go offline for less than ttl, and an update occurs, i'll be out of sync
	// todo: maybe use ttl?

	// timeout is how long to wait for the listed members to come back online
	timeout := time.After(time.Duration(20) * time.Second)

	// loop attempts for timeout, allows last dead members to start back up
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("Timed out waiting for services from %v", strings.Join(members, ", "))
		default:
			// request services from each member until successful
			for _, member := range members {
				// memberTimeout is how long to wait for a member to respond with list of services
				memberTimeout := time.After(3 * time.Second)

				config.Log.Trace("[cluster] - Attempting to request services from %v...", member)
				_, err = r.pubconn.Do("PUBLISH", "portal", fmt.Sprintf("get-services %s", member))
				if err != nil {
					return nil, err
				}

				// wait for member to respond
				for {
					select {
					case <-memberTimeout:
						config.Log.Debug("[cluster] - Timed out waiting for services from %v", member)
						goto nextMember
					case msg := <-message:
						switch v := msg.(type) {
						case redis.Message:
							config.Log.Trace("[cluster] - Received message on 'services' channel")
							services, err := marshalSvcs(v.Data)
							if err != nil {
								return nil, fmt.Errorf("Failed to marshal services - %v", err.Error())
							}
							config.Log.Trace("[cluster] - Services from cluster: %#v\n", *services)
							return *services, nil
						case error:
							return nil, fmt.Errorf("Subscriber failed to receive services - %v", v.Error())
						}
					}
				}
			nextMember:
			}
			// retry loop end
		}
	}
}

// GetServer - likely will never be called
func (r *Redis) GetServer(svcId, srvId string) (*core.Server, error) {
	return common.GetServer(svcId, srvId)
}
