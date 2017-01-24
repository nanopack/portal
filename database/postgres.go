package database

import (
	"database/sql"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

type (
	PostgresDb struct {
		pg *sql.DB
	}
)

// todo: prepare statements

func (p *PostgresDb) connect() error {
	// todo: example: config.DatabaseConnection = "postgres://postgres@127.0.0.1?sslmode=disable"
	db, err := sql.Open("postgres", config.DatabaseConnection)
	if err != nil {
		return fmt.Errorf("Failed to connect to postgres - %s", err)
	}
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("Failed to ping postgres on connect - %s", err)
	}

	p.pg = db
	return nil
}

func (p PostgresDb) createTables() error {
	// create services table
	_, err := p.pg.Exec(`
CREATE TABLE IF NOT EXISTS services (
	serviceId      SERIAL PRIMARY KEY NOT NULL,
	id             TEXT NOT NULL UNIQUE,
	host           TEXT NOT NULL,
	interface      TEXT,
	port           INTEGER NOT NULL,
	type           TEXT,
	scheduler      TEXT,
	persistence    INTEGER,
	netmask        TEXT
)`)
	if err != nil {
		return fmt.Errorf("Failed to create services table - %s", err)
	}

	// create servers table
	_, err = p.pg.Exec(`
CREATE TABLE IF NOT EXISTS servers (
	serverId       SERIAL PRIMARY KEY NOT NULL,
	serviceId      TEXT REFERENCES services (id) ON DELETE CASCADE,
	id             TEXT NOT NULL UNIQUE,
	host           TEXT NOT NULL,
	port           INTEGER NOT NULL,
	forwarder      TEXT,
	weight         TEXT,
	upperThreshold TEXT,
	lowerThreshold TEXT
)`)
	if err != nil {
		return fmt.Errorf("Failed to create servers table - %s", err)
	}

	// create routes table
	_, err = p.pg.Exec(`
CREATE TABLE IF NOT EXISTS routes (
	routeId   SERIAL PRIMARY KEY NOT NULL,
	subdomain TEXT,
	domain    TEXT,
	path      TEXT,
	targets   TEXT,
	fwdPath   TEXT,
	page      TEXT
)`)
	if err != nil {
		return fmt.Errorf("Failed to create routes table - %s", err)
	}

	// create certs table
	_, err = p.pg.Exec(`
CREATE TABLE IF NOT EXISTS certs (
	certId SERIAL PRIMARY KEY NOT NULL,
	cert   TEXT NOT NULL,
	key    TEXT NOT NULL
)`)
	if err != nil {
		return fmt.Errorf("Failed to create cert table - %s", err)
	}

	// create vips table
	_, err = p.pg.Exec(`
CREATE TABLE IF NOT EXISTS vips (
	vipId     SERIAL PRIMARY KEY NOT NULL,
	ip        TEXT,
	interface TEXT,
	alias     TEXT
)`)
	if err != nil {
		return fmt.Errorf("Failed to create vips table - %s", err)
	}

	return nil
}

func (p *PostgresDb) Init() error {
	err := p.connect()
	if err != nil {
		return fmt.Errorf("Failed to create new connection - %s", err)
	}

	// create tables
	err = p.createTables()
	if err != nil {
		return fmt.Errorf("Failed to create tables - %s", err)
	}

	return nil
}

func (p PostgresDb) GetServices() ([]core.Service, error) {
	// read from services table
	rows, err := p.pg.Query("SELECT id, host, interface, port, type, scheduler, persistence, netmask FROM services")
	if err != nil {
		return nil, fmt.Errorf("Failed to select from services table - %s", err)
	}
	defer rows.Close()

	services := make([]core.Service, 0, 0)

	// get data
	for rows.Next() {
		svc := core.Service{}
		err = rows.Scan(&svc.Id, &svc.Host, &svc.Interface, &svc.Port, &svc.Type, &svc.Scheduler, &svc.Persistence, &svc.Netmask)
		if err != nil {
			return nil, fmt.Errorf("Failed to save results into service - %s", err)
		}

		// get service's servers
		servers, err := p.GetServers(svc.Id)
		if err != nil {
			return nil, err
		}
		svc.Servers = servers

		services = append(services, svc)
	}

	// check for errors
	if err = rows.Err(); err != nil {
		return services, fmt.Errorf("Error with results - %s", err)
	}
	return services, nil
}

func (p PostgresDb) GetService(id string) (*core.Service, error) {
	// read from services table
	rows, err := p.pg.Query(fmt.Sprintf("SELECT id, host, interface, port, type, scheduler, persistence, netmask FROM services WHERE id = '%s'", template.HTMLEscapeString(id)))
	if err != nil {
		return nil, fmt.Errorf("Failed to select from services table - %s", err)
	}
	defer rows.Close()

	services := make([]core.Service, 0, 0)

	// get data
	for rows.Next() {
		svc := core.Service{}
		err = rows.Scan(&svc.Id, &svc.Host, &svc.Interface, &svc.Port, &svc.Type, &svc.Scheduler, &svc.Persistence, &svc.Netmask)
		if err != nil {
			return nil, fmt.Errorf("Failed to save results into service - %s", err)
		}

		// get service's servers
		servers, err := p.GetServers(svc.Id)
		if err != nil {
			return nil, err
		}
		svc.Servers = servers

		services = append(services, svc)
	}

	// check for errors
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Error with results - %s", err)
	}

	if len(services) == 0 {
		return nil, NoServiceError
	}

	return &services[0], nil
}

func (p PostgresDb) SetServices(services []core.Service) error {
	// truncate services table
	_, err := p.pg.Exec("TRUNCATE services CASCADE")
	if err != nil {
		return fmt.Errorf("Failed to truncate services table - %s", err)
	}
	for i := range services {
		err = p.SetService(&services[i]) // prevents duplicates
		if err != nil {
			return err
		}
	}
	return nil
}

// todo: not pointer
func (p PostgresDb) SetService(service *core.Service) error {
	services, err := p.GetServices()
	if err != nil {
		return err
	}
	// for idempotency
	for i := range services {
		// update services table
		if services[i].Id == service.Id {
			_, err = p.pg.Exec(fmt.Sprintf(`
UPDATE services SET host = '%s', interface = '%s', port = '%s', type = '%s', scheduler = '%s', persistence = '%s', netmask = '%s'
WHERE id = '%s'`,
				template.HTMLEscapeString(service.Host), template.HTMLEscapeString(service.Interface), template.HTMLEscapeString(strconv.Itoa(service.Port)),
				template.HTMLEscapeString(service.Type), template.HTMLEscapeString(service.Scheduler), template.HTMLEscapeString(strconv.Itoa(service.Persistence)),
				template.HTMLEscapeString(service.Netmask), template.HTMLEscapeString(service.Id)))
			if err != nil {
				return fmt.Errorf("Failed to update services table - %s", err)
			}

			// reset servers
			err = p.SetServers(service.Id, service.Servers)
			if err != nil {
				return err
			}

			return nil
		}
	}

	// insert into services table
	_, err = p.pg.Exec(fmt.Sprintf(`
INSERT INTO services(id, host, interface, port, type, scheduler, persistence, netmask)
VALUES('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')`,
		template.HTMLEscapeString(service.Id), template.HTMLEscapeString(service.Host), template.HTMLEscapeString(service.Interface), template.HTMLEscapeString(strconv.Itoa(service.Port)),
		template.HTMLEscapeString(service.Type), template.HTMLEscapeString(service.Scheduler), template.HTMLEscapeString(strconv.Itoa(service.Persistence)), template.HTMLEscapeString(service.Netmask)))
	if err != nil {
		return fmt.Errorf("Failed to insert into services table - %s", err)
	}

	// reset servers
	err = p.SetServers(service.Id, service.Servers)
	if err != nil {
		return err
	}

	return nil
}

func (p PostgresDb) DeleteService(id string) error {
	// delete from services table
	_, err := p.pg.Exec(fmt.Sprintf(`DELETE FROM services WHERE id = '%s'`, template.HTMLEscapeString(id)))
	if err != nil {
		return fmt.Errorf("Failed to delete from services table - %s", err)
	}
	return nil
}

// SetServers resets all servers for the service
func (p PostgresDb) SetServers(svcId string, servers []core.Server) error {
	// delete servers from service
	_, err := p.pg.Exec(fmt.Sprintf(`DELETE FROM servers WHERE serviceId = '%s'`, template.HTMLEscapeString(svcId)))
	if err != nil {
		return fmt.Errorf("Failed to remove old servers - %s", err)
	}

	for i := range servers {
		err = p.SetServer(svcId, &servers[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (p PostgresDb) SetServer(svcId string, server *core.Server) error {
	service, err := p.GetService(svcId)
	if err != nil {
		return err
	}

	// for idempotency
	for i := range service.Servers {
		// update servers table
		if service.Servers[i].Id == server.Id {
			_, err = p.pg.Exec(fmt.Sprintf(`
UPDATE servers SET host = '%s', port = '%s', forwarder = '%s',
weight = '%s', upperThreshold = '%s', lowerThreshold = '%s'
WHERE id = '%s' AND serviceId = '%s'`,
				template.HTMLEscapeString(server.Host), template.HTMLEscapeString(strconv.Itoa(server.Port)), template.HTMLEscapeString(server.Forwarder), template.HTMLEscapeString(strconv.Itoa(server.Weight)),
				template.HTMLEscapeString(strconv.Itoa(server.UpperThreshold)), template.HTMLEscapeString(strconv.Itoa(server.LowerThreshold)), template.HTMLEscapeString(server.Id), template.HTMLEscapeString(svcId)))
			if err != nil {
				return fmt.Errorf("Failed to update servers table - %s", err)
			}
			return nil
		}
	}

	// insert into servers table
	_, err = p.pg.Exec(fmt.Sprintf(`
INSERT INTO servers(serviceId, id, host, port, forwarder, weight, upperThreshold, lowerThreshold)
VALUES('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')`,
		template.HTMLEscapeString(svcId), template.HTMLEscapeString(server.Id), template.HTMLEscapeString(server.Host), template.HTMLEscapeString(strconv.Itoa(server.Port)), template.HTMLEscapeString(server.Forwarder),
		template.HTMLEscapeString(strconv.Itoa(server.Weight)), template.HTMLEscapeString(strconv.Itoa(server.UpperThreshold)), template.HTMLEscapeString(strconv.Itoa(server.LowerThreshold))))
	if err != nil {
		return fmt.Errorf("Failed to insert into servers table - %s", err)
	}
	return nil
}

func (p PostgresDb) DeleteServer(svcId, srvId string) error {
	// delete from servers table
	_, err := p.pg.Exec(fmt.Sprintf(`DELETE FROM servers WHERE id = '%s' AND serviceId = '%s'`, template.HTMLEscapeString(srvId), template.HTMLEscapeString(svcId)))
	if err != nil {
		return fmt.Errorf("Failed to delete from servers table - %s", err)
	}
	return nil
}

func (p PostgresDb) GetServer(svcId, srvId string) (*core.Server, error) {
	// read from servers table
	rows, err := p.pg.Query(fmt.Sprintf("SELECT id, host, port, forwarder, weight, upperThreshold, lowerThreshold FROM servers WHERE id = '%s' AND serviceId = '%s'", template.HTMLEscapeString(srvId), template.HTMLEscapeString(svcId)))
	if err != nil {
		return nil, fmt.Errorf("Failed to select from servers table - %s", err)
	}
	defer rows.Close()

	servers := make([]core.Server, 0, 0)

	// get data
	for rows.Next() {
		srv := core.Server{}
		err = rows.Scan(&srv.Id, &srv.Host, &srv.Port, &srv.Forwarder, &srv.Weight, &srv.UpperThreshold, &srv.LowerThreshold)
		if err != nil {
			return nil, fmt.Errorf("Failed to save results into server - %s", err)
		}

		servers = append(servers, srv)
	}

	// check for errors
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Error with results - %s", err)
	}

	if len(servers) == 0 {
		return nil, NoServerError
	}

	return &servers[0], nil
}

func (p PostgresDb) GetServers(svcId string) ([]core.Server, error) {
	// read from servers table
	rows, err := p.pg.Query(fmt.Sprintf("SELECT id, host, port, forwarder, weight, upperThreshold, lowerThreshold FROM servers WHERE serviceId = '%s'", template.HTMLEscapeString(svcId)))
	if err != nil {
		return nil, fmt.Errorf("Failed to select from servers table - %s", err)
	}
	defer rows.Close()

	servers := make([]core.Server, 0, 0)

	// get data
	for rows.Next() {
		srv := core.Server{}
		err = rows.Scan(&srv.Id, &srv.Host, &srv.Port, &srv.Forwarder, &srv.Weight, &srv.UpperThreshold, &srv.LowerThreshold)
		if err != nil {
			return nil, fmt.Errorf("Failed to save results into server - %s", err)
		}

		servers = append(servers, srv)
	}

	// check for errors
	if err = rows.Err(); err != nil {
		return servers, fmt.Errorf("Error with results - %s", err)
	}

	return servers, nil
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////

func (p PostgresDb) GetRoutes() ([]core.Route, error) {
	// read from routes table
	rows, err := p.pg.Query("SELECT subdomain, domain, path, targets, fwdPath, page FROM routes")
	if err != nil {
		return nil, fmt.Errorf("Failed to select from routes table - %s", err)
	}
	defer rows.Close()

	routes := make([]core.Route, 0, 0)

	// get data
	for rows.Next() {
		route := core.Route{}
		var tmpTargets string
		err = rows.Scan(&route.SubDomain, &route.Domain, &route.Path, &tmpTargets, &route.FwdPath, &route.Page)
		if err != nil {
			return nil, fmt.Errorf("Failed to save results into route - %s", err)
		}
		route.Targets = strings.Split(tmpTargets, ",")

		routes = append(routes, route)
	}

	// check for errors
	if err = rows.Err(); err != nil {
		return routes, fmt.Errorf("Error with results - %s", err)
	}
	return routes, nil
}

func (p PostgresDb) SetRoutes(routes []core.Route) error {
	// truncate routes table
	_, err := p.pg.Exec("TRUNCATE routes")
	if err != nil {
		return fmt.Errorf("Failed to truncate routes table - %s", err)
	}
	for i := range routes {
		err = p.SetRoute(routes[i]) // prevents duplicates
		if err != nil {
			return err
		}
	}
	return nil
}

func (p PostgresDb) SetRoute(route core.Route) error {
	routes, err := p.GetRoutes()
	if err != nil {
		return err
	}
	// for idempotency
	for i := range routes {
		// update routes table
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			_, err = p.pg.Exec(fmt.Sprintf(`
UPDATE routes SET targets = '%s', fwdPath = '%s', page = '%s'
WHERE subdomain = '%s' AND domain = '%s' AND path = '%s'`,
				template.HTMLEscapeString(strings.Join(route.Targets, ",")),
				template.HTMLEscapeString(route.FwdPath), template.HTMLEscapeString(route.Page), template.HTMLEscapeString(route.SubDomain), template.HTMLEscapeString(route.Domain),
				template.HTMLEscapeString(route.Path)))
			if err != nil {
				return fmt.Errorf("Failed to update routes table - %s", err)
			}
			return nil
		}
	}

	// insert into routes table
	_, err = p.pg.Exec(fmt.Sprintf(`
INSERT INTO routes(subdomain, domain, path, targets, fwdPath, page)
VALUES('%s', '%s', '%s', '%s', '%s', '%s')`,
		template.HTMLEscapeString(route.SubDomain), template.HTMLEscapeString(route.Domain), template.HTMLEscapeString(route.Path),
		template.HTMLEscapeString(strings.Join(route.Targets, ",")),
		template.HTMLEscapeString(route.FwdPath), template.HTMLEscapeString(route.Page)))
	if err != nil {
		return fmt.Errorf("Failed to insert into routes table - %s", err)
	}
	return nil
}

func (p PostgresDb) DeleteRoute(route core.Route) error {
	// delete from routes table
	_, err := p.pg.Exec(fmt.Sprintf(`DELETE FROM routes WHERE subdomain = '%s' AND domain = '%s' AND path = '%s'`,
		template.HTMLEscapeString(route.SubDomain), template.HTMLEscapeString(route.Domain), template.HTMLEscapeString(route.Path)))
	if err != nil {
		return fmt.Errorf("Failed to delete from routes table - %s", err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////

func (p PostgresDb) GetCerts() ([]core.CertBundle, error) {
	// read from certs table
	rows, err := p.pg.Query("SELECT cert, key FROM certs")
	if err != nil {
		return nil, fmt.Errorf("Failed to select from certs table - %s", err)
	}
	defer rows.Close()

	certs := make([]core.CertBundle, 0, 0)

	// get data
	for rows.Next() {
		cert := core.CertBundle{}
		err = rows.Scan(&cert.Cert, &cert.Key)
		if err != nil {
			return nil, fmt.Errorf("Failed to save results into certbundle - %s", err)
		}

		certs = append(certs, cert)
	}

	// check for errors
	if err = rows.Err(); err != nil {
		return certs, fmt.Errorf("Error with results - %s", err)
	}
	return certs, nil
}

func (p PostgresDb) SetCerts(certs []core.CertBundle) error {
	// truncate certs table
	_, err := p.pg.Exec("TRUNCATE certs")
	if err != nil {
		return fmt.Errorf("Failed to truncate certs table - %s", err)
	}
	for i := range certs {
		err = p.SetCert(certs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (p PostgresDb) SetCert(cert core.CertBundle) error {
	certs, err := p.GetCerts()
	if err != nil {
		return err
	}
	// for idempotency
	for i := 0; i < len(certs); i++ {
		// todo: can there be multiple keys for same cert?
		// if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {

		// update certs table
		if certs[i].Cert == cert.Cert {
			_, err = p.pg.Exec(fmt.Sprintf(`UPDATE certs SET key = '%s' WHERE cert = '%s'`, template.HTMLEscapeString(cert.Key), template.HTMLEscapeString(cert.Cert)))
			if err != nil {
				return fmt.Errorf("Failed to update certs table - %s", err)
			}
			return nil
		}
	}

	// insert into certs table
	_, err = p.pg.Exec(fmt.Sprintf(`INSERT INTO certs(cert, key) VALUES('%s', '%s')`, template.HTMLEscapeString(cert.Cert), template.HTMLEscapeString(cert.Key)))
	if err != nil {
		return fmt.Errorf("Failed to insert into certs table - %s", err)
	}
	return nil
}

func (p PostgresDb) DeleteCert(cert core.CertBundle) error {
	// todo: can there be multiple keys for same cert?
	// _, err := p.pg.Exec(fmt.Sprintf(`DELETE FROM certs WHERE cert = '%s' AND key = '%s'`, template.HTMLEscapeString(cert.Cert), template.HTMLEscapeString(cert.Key)))

	// delete from certs table
	_, err := p.pg.Exec(fmt.Sprintf(`DELETE FROM certs WHERE cert = '%s'`, template.HTMLEscapeString(cert.Cert)))
	if err != nil {
		return fmt.Errorf("Failed to delete from certs table - %s", err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// VIPS
////////////////////////////////////////////////////////////////////////////////

func (p PostgresDb) GetVips() ([]core.Vip, error) {
	// read from vips table
	rows, err := p.pg.Query("SELECT ip, interface, alias FROM vips")
	if err != nil {
		return nil, fmt.Errorf("Failed to select from vips table - %s", err)
	}
	defer rows.Close()

	vips := make([]core.Vip, 0, 0)

	// get data
	for rows.Next() {
		vip := core.Vip{}
		err = rows.Scan(&vip.Ip, &vip.Interface, &vip.Alias)
		if err != nil {
			return nil, fmt.Errorf("Failed to save results into vip - %s", err)
		}

		vips = append(vips, vip)
	}

	// check for errors
	if err = rows.Err(); err != nil {
		return vips, fmt.Errorf("Error with results - %s", err)
	}
	return vips, nil
}

func (p PostgresDb) SetVips(vips []core.Vip) error {
	// truncate vips table
	_, err := p.pg.Exec("TRUNCATE vips")
	if err != nil {
		return fmt.Errorf("Failed to truncate vips table - %s", err)
	}
	for i := range vips {
		err = p.SetVip(vips[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (p PostgresDb) SetVip(vip core.Vip) error {
	vips, err := p.GetVips()
	if err != nil {
		return err
	}
	// for idempotency
	for i := 0; i < len(vips); i++ {
		// update vips table
		if vips[i].Ip == vip.Ip {
			_, err = p.pg.Exec(fmt.Sprintf(`UPDATE vips SET interface = '%s', alias = '%s' WHERE ip = '%s'`, template.HTMLEscapeString(vip.Interface), template.HTMLEscapeString(vip.Alias), template.HTMLEscapeString(vip.Ip)))
			if err != nil {
				return fmt.Errorf("Failed to update vips table - %s", err)
			}
			return nil
		}
	}

	// insert into vips table
	_, err = p.pg.Exec(fmt.Sprintf(`INSERT INTO vips(ip, interface, alias) VALUES('%s', '%s', '%s')`, template.HTMLEscapeString(vip.Ip), template.HTMLEscapeString(vip.Interface), template.HTMLEscapeString(vip.Alias)))
	if err != nil {
		return fmt.Errorf("Failed to insert into vips table - %s", err)
	}
	return nil
}

func (p PostgresDb) DeleteVip(vip core.Vip) error {
	// delete from vips table
	_, err := p.pg.Exec(fmt.Sprintf(`DELETE FROM vips WHERE ip = '%s'`, template.HTMLEscapeString(vip.Ip)))
	if err != nil {
		return fmt.Errorf("Failed to delete from vips table - %s", err)
	}
	return nil
}
