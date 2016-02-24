package cluster

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

var (
	self string
	ttl  = 20 // time until a member is deemed "dead"
	beat = time.Duration(ttl/2) * time.Second
)

type (
	Redis struct {
		pubconn redis.Conn
		subconn redis.PubSubConn
	}
)

func (r *Redis) Init() error {
	hostname, _ := os.Hostname()
	self = fmt.Sprintf("%v:%v", hostname, config.ApiPort)
	p, err := redis.DialURL(config.ClusterConnection, redis.DialConnectTimeout(30*time.Second), redis.DialWriteTimeout(10*time.Second), redis.DialReadTimeout(10*time.Second), redis.DialPassword(config.ClusterToken))
	if err != nil {
		return fmt.Errorf("Failed to reach redis for pubconn - %v", err)
	}
	r.pubconn = p

	// don't set read timeout on subscriber - it dies if no 'updates' within that time
	s, err := redis.DialURL(config.ClusterConnection, redis.DialPassword(config.ClusterToken))
	if err != nil {
		return fmt.Errorf("Failed to reach redis for subconn - %v", err)
	}
	r.subconn = redis.PubSubConn{s}
	r.subconn.Subscribe("portal")

	services, err := r.GetServices()
	if err != nil {
		return fmt.Errorf("Failed to get services - %v", err)
	}
	// write services
	if services != nil {
		config.Log.Trace("[cluster] - Setting services...")
		err = common.SetServices(services)
		if err != nil {
			return fmt.Errorf("Failed to set services - %v", err)
		}
	}

	r.pubconn.Do("SET", self, "alive", "EX", ttl)
	_, err = r.pubconn.Do("SADD", "members", self)
	if err != nil {
		return fmt.Errorf("Failed to add myself to list of members - %v", err)
	}

	go r.subscribe()
	go r.heartbeat()
	go r.cleanup()

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SERVICES
////////////////////////////////////////////////////////////////////////////////

// SetServices tells all members to replace the services in their database with a new set.
// rolls back on failure
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

// SetService tells all members to add the service to their database.
// rolls back on failure
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

// DeleteService tells all members to remove the service from their database.
// rolls back on failure
func (r *Redis) DeleteService(id string) error {
	oldService, err := common.GetService(id)
	// this should not return nil to ensure the service is gone from entire cluster
	if err != nil && !strings.Contains(err.Error(), "No Service Found") {
		return err
	}

	// publishService to others
	err = r.publishString("delete-service", id)
	if err != nil {
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

// SetServers tells all members to replace a service's servers with a new set.
// rolls back on failure
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

// SetServer tells all members to add the server to their database.
// rolls back on failure
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

// DeleteServer tells all members to remove the server from their database.
// rolls back on failure
func (r *Redis) DeleteServer(svcId, srvId string) error {
	oldServer, err := common.GetServer(svcId, srvId)
	// mustn't return nil here to ensure cluster removes the server
	if err != nil && !strings.Contains(err.Error(), "No Server Found") {
		return err
	}

	// publishServer to others
	// todo: swap srv/svc ids to match backender interface for better readability
	err = r.publishString("delete-server", srvId, svcId)
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

// GetServices gets a list of services from the database, or another cluster member.
func (r *Redis) GetServices() ([]core.Service, error) {
	// get known members(other than me) to 'poll' for services
	members, _ := redis.Strings(r.pubconn.Do("SMEMBERS", "members"))
	if len(members) == 0 {
		// should only happen on new cluster
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

	c, err := redis.DialURL(config.ClusterConnection, redis.DialConnectTimeout(15*time.Second), redis.DialPassword(config.ClusterToken))
	if err != nil {
		return nil, fmt.Errorf("Failed to reach redis for services subscriber - %v", err)
	}
	defer c.Close()

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

				// ask a member for its services
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
		}
	}
}

// GetServer - likely will never be called
func (r *Redis) GetServer(svcId, srvId string) (*core.Server, error) {
	return common.GetServer(svcId, srvId)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE
////////////////////////////////////////////////////////////////////////////////

// cleanup cleans up members not present after ttl seconds
func (r Redis) cleanup() {
	// cycle every second to check for dead members
	tick := time.Tick(time.Second)
	conn, err := redis.DialURL((config.ClusterConnection), redis.DialPassword(config.ClusterToken))
	if err != nil {
		config.Log.Error("[cluster] - Failed to reach redis for cleanup - %v", err)
		// clear balancer rules ("stop balancing if we are 'dead'")
		balance.SetServices(make([]core.Service, 0, 0))
		os.Exit(1)
	}

	for _ = range tick {
		// get list of members that should be alive
		members, _ := redis.Strings(conn.Do("SMEMBERS", "members"))
		for _, member := range members {
			// if the member timed out, remove the member from the member set
			exist, _ := redis.Int(conn.Do("EXISTS", member))
			if exist == 0 {
				conn.Do("SREM", "members", member)
				config.Log.Info("[cluster] - Member '%v' assumed dead. Removed.", member)
			}
		}
	}
	conn.Close()
}

// heartbeat records that the member is still alive
func (r Redis) heartbeat() {
	tick := time.Tick(beat)
	// set write timeout so each 'beat' ensures we can talk to redis (network partition) (rather than create new connection)
	conn, err := redis.DialURL((config.ClusterConnection), redis.DialWriteTimeout(5*time.Second), redis.DialReadTimeout(5*time.Second), redis.DialPassword(config.ClusterToken))
	if err != nil {
		config.Log.Error("[cluster] - Failed to reach redis for heartbeat - %v", err)
		// clear balancer rules ("stop balancing if we are 'dead'")
		balance.SetServices(make([]core.Service, 0, 0))
		os.Exit(1)
	}

	for _ = range tick {
		config.Log.Trace("[cluster] - Heartbeat...")
		_, err = conn.Do("SET", self, "alive", "EX", ttl)
		if err != nil {
			conn.Close()
			config.Log.Error("[cluster] - Failed to heartbeat - %v", err)
			// clear balancer rules ("stop balancing if we are 'dead'")
			balance.SetServices(make([]core.Service, 0, 0))
			os.Exit(1)
		}
		// re-add ourself to member list (just in case)
		_, err = conn.Do("SADD", "members", self)
		if err != nil {
			conn.Close()
			config.Log.Error("[cluster] - Failed to add myself to list of members - %v", err)
			// clear balancer rules ("stop balancing if we are 'dead'")
			balance.SetServices(make([]core.Service, 0, 0))
			os.Exit(1)
		}
	}
}

// subscribe listens on the portal channel and acts based on messages received
func (r Redis) subscribe() {
	config.Log.Info("[cluster] - Redis subscribing on %s...", config.ClusterConnection)

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
				config.Log.Debug("[cluster] - set-services successful")
			case "set-service":
				if len(pdata) != 2 {
					// shouldn't happen unless redis is not secure and someone manually `publishService`es
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
				config.Log.Trace("[cluster] - set-service hash - %v", actionHash)
				r.pubconn.Do("SADD", actionHash, self)
				config.Log.Debug("[cluster] - set-service successful")
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
				config.Log.Trace("[cluster] - delete-service hash - %v", actionHash)
				r.pubconn.Do("SADD", actionHash, self)
				config.Log.Debug("[cluster] - delete-service successful")
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
				config.Log.Trace("[cluster] - set-servers hash - %v", actionHash)
				r.pubconn.Do("SADD", actionHash, self)
				config.Log.Debug("[cluster] - set-servers successful")
			case "set-server":
				if len(pdata) != 3 {
					// shouldn't happen unless redis is not secure and someone manually publishServicees
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
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("set-server %s %s", *server, svcId))))
				config.Log.Trace("[cluster] - set-server hash - %v", actionHash)
				r.pubconn.Do("SADD", actionHash, self)
				config.Log.Debug("[cluster] - set-server successful")
			case "delete-server":
				if len(pdata) != 3 {
					config.Log.Error("[cluster] - service id not passed in message")
					break
				}
				srvId := pdata[1]
				svcId := pdata[2]
				err := common.DeleteServer(svcId, srvId)
				if err != nil {
					config.Log.Error("[cluster] - Failed to delete server - %v", err.Error())
					break
				}
				actionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("delete-server %s %s", srvId, svcId))))
				config.Log.Trace("[cluster] - delete-server hash - %v", actionHash)
				r.pubconn.Do("SADD", actionHash, self)
				config.Log.Debug("[cluster] - delete-server successful")
			default:
				config.Log.Error("[cluster] - Recieved unknown data on %v: %v", v.Channel, string(v.Data))
			}
		case error:
			config.Log.Error("[cluster] - Subscriber failed to receive - %v", v.Error())
			if strings.Contains(v.Error(), "closed network connection") {
				// clear balancer rules ("stop balancing if we are 'dead'")
				balance.SetServices(make([]core.Service, 0, 0))
				// exit so we don't get spammed with logs
				os.Exit(1)
			}
			continue
		}
	}
}

// publishService publishes service[s] to the "portal" channel
func (r Redis) publishService(action string, v interface{}) error {
	s, err := json.Marshal(v)
	if err != nil {
		return BadJson
	}

	// todo: should create new connection(or use pool) - single connection limits concurrency
	_, err = r.pubconn.Do("PUBLISH", "portal", fmt.Sprintf("%s %s", action, s))
	return err
}

// publishServer publishes server[s] to the "portal" channel
func (r Redis) publishServer(action, svcId string, v interface{}) error {
	s, err := json.Marshal(v)
	if err != nil {
		return BadJson
	}

	// todo: should create new connection(or use pool) - single connection limits concurrency
	_, err = r.pubconn.Do("PUBLISH", "portal", fmt.Sprintf("%s %s %s", action, s, svcId))
	return err
}

// publishString publishes string[s] to the "portal" channel
func (r Redis) publishString(action string, s ...string) error {
	// todo: should create new connection(or use pool) - single connection limits concurrency
	_, err := r.pubconn.Do("PUBLISH", "portal", fmt.Sprintf("%s %s", action, strings.Join(s, " ")))
	return err
}

// waitForMembers waits for all members to apply the action
func (r Redis) waitForMembers(actionHash string) error {
	config.Log.Trace("[cluster] - Waiting for updates to %s", actionHash)
	// todo: make timeout configurable
	// timeout is the amount of time to wait for members to apply the action
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
