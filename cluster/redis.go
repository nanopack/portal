package cluster

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

var (
	self string
)

type (
	Redis struct {
		pubconn redis.Conn
		subconn redis.PubSubConn
	}

	members []string
)

func (r *Redis) Init() error {
	url, err := url.Parse(config.ClusterConnection)
	if err != nil {
		return err
	}
	self = fmt.Sprintf("%v:%v", config.ApiHost, config.ApiPort)
	config.Log.Info("Redis subscribing on %s...", url)
	c, err := redis.DialURL(fmt.Sprintf("%s", url))
	if err != nil {
		fmt.Println("Failed to reach redis for pubconn", err)
		return err
	}
	r.pubconn = c

	c, err = redis.DialURL(fmt.Sprintf("%s", url))
	if err != nil {
		fmt.Println("Failed to reach redis for subconn", err)
		return err
	}
	r.subconn = redis.PubSubConn{c}
	r.subconn.Subscribe("portal")
	r.pubconn.Do("SADD", "members", self)

	go subscribe(r)

	return nil
}

func (r *Redis) UnInit() error {
	// remove myself from cluster(called on clean quit)
	_, err := r.pubconn.Do("SREM", "members", self)
	return err
}

func subscribe(r *Redis) {
	for {
		switch v := r.subconn.Receive().(type) {
		case redis.Message:
			switch pdata := strings.Split(string(v.Data), " "); pdata[0] {
			// SERVICES ///////////////////////////////////////////////////////////////////////////////////////////////
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
				config.Log.Trace("set-services hash - %v", actionHash)
				r.pubconn.Do("SADD", actionHash, self)
			case "set-service":
				if len(pdata) != 2 {
					// shouldn't happen unless redis is not secure and someone manually publishServicees
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
				fmt.Printf("Recieved data on %v: %v\n", v.Channel, string(v.Data))
			}
		case error:
			config.Log.Error("[cluster] Subscriber failed to receive - %v", v.Error())
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
	config.Log.Trace("Waiting for updates to %s", actionHash)
	// make timeout configurable
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
		// nothing to rollback yet (nobody recieved)
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
// GetService
func (r *Redis) GetService(id string) (*core.Service, error) {
	return nil, nil
}

// GetServices
// doesn't need to be a pointer method because it doesn't modify original object
func (r *Redis) GetServices() ([]core.Service, error) {
	return nil, nil
}

// GetServer
func (r *Redis) GetServer(svcId, srvId string) (*core.Server, error) {
	return nil, nil
}
