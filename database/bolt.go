package database

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	bolt "github.com/coreos/bbolt"
	"github.com/twinj/uuid"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

type (
	BoltDb struct {
		path string
	}
)

func (s *BoltDb) boltConnect(f func(*bolt.DB) error) error {
	db, err := bolt.Open(s.path, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = f(db)
	if err != nil {
		return err
	}

	return nil
}

func (s *BoltDb) Init() error {
	u, err := url.Parse(config.DatabaseConnection)
	if err != nil {
		return err
	}
	s.path = u.Path

	err = s.boltConnect(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			buckets := []string{
				"Services",
				"Routes",
				"Certs",
				"VIPs",
			}

			for _, bucket := range buckets {
				_, err := tx.CreateBucketIfNotExists([]byte(bucket))
				if err != nil {
					return fmt.Errorf("create %s bucket: %v", bucket, err)
				}
			}

			return nil
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func (s BoltDb) GetServices() ([]core.Service, error) {
	services := make([]core.Service, 0, 0)

	err := s.boltConnect(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Services"))

			b.ForEach(func(k, v []byte) error {
				var service core.Service

				if err := json.Unmarshal([]byte(v), &service); err != nil {
					return fmt.Errorf("Bad JSON syntax stored in db")
				}

				services = append(services, service)

				return nil
			})

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return services, nil
}

func (s BoltDb) GetService(id string) (*core.Service, error) {
	service := core.Service{}

	err := s.boltConnect(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Services"))
			v := b.Get([]byte(id))
			if err := json.Unmarshal([]byte(v), &service); err != nil {
				return fmt.Errorf("Bad JSON syntax stored in db")
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (s BoltDb) SetServices(services []core.Service) error {
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			err := tx.DeleteBucket([]byte("Services"))
			if err != nil {
				return err
			}

			b, err := tx.CreateBucket([]byte("Services"))
			if err != nil {
				return err
			}

			for i := range services {
				value, err := json.Marshal(services[i])
				if err != nil {
					return err
				}
				err = b.Put([]byte(services[i].Id), value)
				if err != nil {
					return err
				}
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func (s BoltDb) SetService(service *core.Service) error {
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Services"))
			value, err := json.Marshal(service)
			if err != nil {
				return err
			}
			err = b.Put([]byte(service.Id), value)
			if err != nil {
				return err
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func (s BoltDb) DeleteService(id string) error {
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Services"))
			err := b.Delete([]byte(id))
			if err != nil {
				return err
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func (s BoltDb) SetServer(svcId string, server *core.Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}
	for _, srv := range service.Servers {
		if srv.Id == server.Id {
			// if server already exists, don't duplicate it
			return nil
		}
	}
	service.Servers = append(service.Servers, *server)

	return s.SetService(service)
}

func (s BoltDb) SetServers(svcId string, servers []core.Server) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}

	// pretty simple, reset all servers
	service.Servers = servers

	return s.SetService(service)
}

func (s BoltDb) DeleteServer(svcId, srvId string) error {
	service, err := s.GetService(svcId)
	if err != nil {
		return err
	}
	config.Log.Trace("Deleting %s from %s", srvId, svcId)
checkRemove:
	for i, srv := range service.Servers {
		if srv.Id == srvId {
			service.Servers = append(service.Servers[:i], service.Servers[i+1:]...)
			goto checkRemove // prevents 'slice bounds out of range' panic
		}
	}

	return s.SetService(service)
}

func (s BoltDb) GetServer(svcId, srvId string) (*core.Server, error) {
	service, err := s.GetService(svcId)
	if err != nil {
		return nil, err
	}

	for _, srv := range service.Servers {
		if srv.Id == srvId {
			return &srv, nil
		}
	}

	return nil, NoServerError
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////

func (s BoltDb) GetRoutes() ([]core.Route, error) {
	routes := make([]core.Route, 0, 0)
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Routes"))

			b.ForEach(func(k, v []byte) error {
				var route core.Route

				if err := json.Unmarshal([]byte(v), &route); err != nil {
					return fmt.Errorf("Bad JSON syntax stored in db")
				}

				routes = append(routes, route)

				return nil
			})

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return routes, nil
}

func (s BoltDb) SetRoutes(routes []core.Route) error {
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			err := tx.DeleteBucket([]byte("Routes"))
			if err != nil {
				return err
			}

			b, err := tx.CreateBucket([]byte("Routes"))
			if err != nil {
				return err
			}

			for i := range routes {
				// unique key to store route by
				value, err := json.Marshal(routes[i])
				if err != nil {
					return err
				}
				ukey := uuid.NewV4().String()
				err = b.Put([]byte(ukey), value)
				if err != nil {
					return err
				}
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func (s BoltDb) SetRoute(route core.Route) error {
	routes, err := s.GetRoutes()
	if err != nil {
		return err
	}
	// for idempotency
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			return nil
		}
	}

	routes = append(routes, route)
	return s.SetRoutes(routes)
}

func (s BoltDb) DeleteRoute(route core.Route) error {
	routes, err := s.GetRoutes()
	// todo: need to ensure db doesn't get cleared with new bolt (shared mutex) (previously if err finding .tmp file, db could be dropped)
	if err != nil {
		return err
	}
	for i := 0; i < len(routes); i++ {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			routes = append(routes[:i], routes[i+1:]...)
			break
		}
	}
	return s.SetRoutes(routes)
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////

func (s BoltDb) GetCerts() ([]core.CertBundle, error) {
	certs := make([]core.CertBundle, 0, 0)
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("Certs"))

			b.ForEach(func(k, v []byte) error {
				var cert core.CertBundle

				if err := json.Unmarshal([]byte(v), &cert); err != nil {
					return fmt.Errorf("Bad JSON syntax stored in db")
				}

				certs = append(certs, cert)

				return nil
			})

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return certs, nil
}

func (s BoltDb) SetCerts(certs []core.CertBundle) error {
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			err := tx.DeleteBucket([]byte("Certs"))
			if err != nil {
				return err
			}

			b, err := tx.CreateBucket([]byte("Certs"))
			if err != nil {
				return err
			}

			for i := range certs {
				// unique key to store cert by
				value, err := json.Marshal(certs[i])
				if err != nil {
					return err
				}
				ukey := uuid.NewV4().String()
				err = b.Put([]byte(ukey), value)
				if err != nil {
					return err
				}
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func (s BoltDb) SetCert(cert core.CertBundle) error {
	certs, err := s.GetCerts()
	if err != nil {
		return err
	}

	// for idempotency
	for i := 0; i < len(certs); i++ {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			return nil
		}
		// update if key is different
		if certs[i].Cert == cert.Cert {
			certs[i].Key = cert.Key
		}
	}

	certs = append(certs, cert)
	return s.SetCerts(certs)
}

func (s BoltDb) DeleteCert(cert core.CertBundle) error {
	certs, err := s.GetCerts()
	if err != nil {
		return err
	}
	for i := 0; i < len(certs); i++ {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			certs = append(certs[:i], certs[i+1:]...)
			break
		}
	}
	return s.SetCerts(certs)
}

////////////////////////////////////////////////////////////////////////////////
// VIPS
////////////////////////////////////////////////////////////////////////////////

func (s BoltDb) GetVips() ([]core.Vip, error) {
	vips := make([]core.Vip, 0, 0)
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("VIPs"))

			b.ForEach(func(k, v []byte) error {
				var vip core.Vip

				if err := json.Unmarshal([]byte(v), &vip); err != nil {
					return fmt.Errorf("Bad JSON syntax stored in db")
				}

				vips = append(vips, vip)

				return nil
			})

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return vips, nil
}

func (s BoltDb) SetVips(vips []core.Vip) error {
	err := s.boltConnect(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			err := tx.DeleteBucket([]byte("VIPs"))
			if err != nil {
				return err
			}

			b, err := tx.CreateBucket([]byte("VIPs"))
			if err != nil {
				return err
			}

			for i := range vips {
				// unique (as much as what we keep) key to store vip by
				value, err := json.Marshal(vips[i])
				if err != nil {
					return err
				}
				ukey := fmt.Sprintf("%s-%s", strings.Replace(vips[i].Ip, ".", "_", -1), vips[i].Interface)
				err = b.Put([]byte(ukey), value)
				if err != nil {
					return err
				}
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	return nil
}

func (s BoltDb) SetVip(vip core.Vip) error {
	vips, err := s.GetVips()
	if err != nil {
		return err
	}
	// for idempotency
	for i := 0; i < len(vips); i++ {
		if vips[i].Ip == vip.Ip && vips[i].Interface == vip.Interface {
			return nil
		}
	}

	vips = append(vips, vip)
	return s.SetVips(vips)
}

func (s BoltDb) DeleteVip(vip core.Vip) error {
	vips, err := s.GetVips()
	if err != nil {
		return err
	}
	for i := 0; i < len(vips); i++ {
		if vips[i].Ip == vip.Ip && vips[i].Interface == vip.Interface {
			vips = append(vips[:i], vips[i+1:]...)
			break
		}
	}
	return s.SetVips(vips)
}
