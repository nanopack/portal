package balance

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync"
	"text/template"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

var (
	nginxLock = &sync.RWMutex{}
)

type (
	Nginx struct {
		Services       []core.Service
		configFile     string
		originalConfig string
	}
)

func (n *Nginx) Init() error {
	n.Services = make([]core.Service, 0)

	// ensure config location exists
	err := os.MkdirAll(config.WorkDir, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create working directory - %s", err)
	}
	n.configFile = path.Join(config.WorkDir, "portal-nginx.conf")
	primerConfig := path.Join(config.WorkDir, "portal-nginx-primer.conf")

	// read primer config file (if any)
	cfg, err := ioutil.ReadFile(primerConfig)
	if err != nil {
		// read config file and save as primer (first run generally)
		cfg, err = ioutil.ReadFile(n.configFile)
		if err != nil {
			return fmt.Errorf("Failed to read a config file - %s", err)
		}

		// persist primer config
		err = ioutil.WriteFile(primerConfig, cfg, 0644)
		if err != nil {
			return fmt.Errorf("Failed to write primer config - %s", err)
		}
	}

	// store primer config
	n.originalConfig = string(cfg)

	// reload nginx - don't return (in the event nginx is not running before portal)
	n.regenerloadConfig()

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SERVICES
////////////////////////////////////////////////////////////////////////////////

func (n *Nginx) GetServices() ([]core.Service, error) {
	return n.Services, nil
}

func (n Nginx) GetService(id string) (*core.Service, error) {
	// break up ids
	svc, err := parseSvc(id)
	if err != nil {
		return nil, err
	}

	// to prevent issues if a delete is happening
	nginxLock.Lock()
	defer nginxLock.Unlock()

	for i := range n.Services {
		if n.Services[i].Type == svc.Type && n.Services[i].Host == svc.Host && n.Services[i].Port == svc.Port {
			return &n.Services[i], nil
		}
	}

	return nil, NoServiceError
}

func (n *Nginx) SetServices(services []core.Service) error {
	nginxLock.Lock()
	n.Services = services
	defer nginxLock.Unlock()

	return n.regenerloadConfig()
}

// SetService updates or adds the service
func (n *Nginx) SetService(service *core.Service) error {
	nginxLock.Lock()
	defer nginxLock.Unlock()

	updated := false

	// update service if found
	for i := range n.Services {
		if n.Services[i].Type == service.Type && n.Services[i].Host == service.Host && n.Services[i].Port == service.Port {
			n.Services[i] = *service
			updated = true
			break
		}
	}

	if !updated {
		n.Services = append(n.Services, *service)
	}

	return n.regenerloadConfig()
}

func (n *Nginx) DeleteService(id string) error {
	service, err := parseSvc(id)
	if err != nil {
		return err
	}

	nginxLock.Lock()
	defer nginxLock.Unlock()

	// delete service if found
	for i := range n.Services {
		if n.Services[i].Type == service.Type && n.Services[i].Host == service.Host && n.Services[i].Port == service.Port {
			n.Services = append(n.Services[:i], n.Services[i+1:]...)
			break
		}
	}

	return n.regenerloadConfig()
}

////////////////////////////////////////////////////////////////////////////////
// SERVERS
////////////////////////////////////////////////////////////////////////////////

// GetServer gets server from the service.
func (n Nginx) GetServer(svcId, srvId string) (*core.Server, error) {
	// get service
	svc, err := n.GetService(svcId)
	if err != nil {
		return nil, err
	}

	// break up ids
	srv, err := parseSrv(srvId)
	if err != nil {
		return nil, err
	}

	// to prevent issues if a delete is happening
	nginxLock.Lock()
	defer nginxLock.Unlock()

	for i := range svc.Servers {
		if svc.Servers[i].Host == srv.Host && svc.Servers[i].Port == srv.Port {
			return &svc.Servers[i], nil
		}
	}

	return nil, NoServerError
}

// SetServers updates the servers for the service in nginx.
func (n *Nginx) SetServers(svcId string, servers []core.Server) error {
	svc, err := n.GetService(svcId)
	if err != nil {
		return err
	}

	nginxLock.Lock()
	svc.Servers = servers
	nginxLock.Unlock()

	return n.regenerloadConfig()
}

// SetServer adds or updates the server for the service in nginx.
func (n *Nginx) SetServer(svcId string, server *core.Server) error {
	// get service
	svc, err := n.GetService(svcId)
	if err != nil {
		return err
	}
	update := false

	nginxLock.Lock()
	defer nginxLock.Unlock()

	for i := range svc.Servers {
		if svc.Servers[i].Host == server.Host && svc.Servers[i].Port == server.Port {
			svc.Servers[i] = *server
			update = true
		}
	}

	if !update {
		svc.Servers = append(svc.Servers, *server)
	}

	// regenerate and reload nginx
	return n.regenerloadConfig()
}

// DeleteServer deletes the server from the service in nginx.
func (n *Nginx) DeleteServer(svcId, srvId string) error {
	// break up ids
	srv, err := parseSrv(srvId)
	if err != nil {
		return err
	}

	svc, err := n.GetService(svcId)
	if err != nil {
		return err
	}

	nginxLock.Lock()
	defer nginxLock.Unlock()

	// delete service if found
	for i := range svc.Servers {
		if svc.Servers[i].Host == srv.Host && svc.Servers[i].Port == srv.Port {
			svc.Servers = append(svc.Servers[:i], svc.Servers[i+1:]...)
			break
		}
	}

	return n.regenerloadConfig()
}

// regenerloadConfig regenerates the nginx config and performs a live reload on
// nginx to re-read in the config.
// todo: use http proxying if http detected
func (n *Nginx) regenerloadConfig() error {
	nginxConfig := fmt.Sprintf(`%s

# CONFIG GENERATED BY PORTAL
stream {
{{- range . }}
	upstream {{.Id }} {
	{{- if (eq .Scheduler "lc") or (eq .Scheduler "wlc") or (eq .Scheduler "lblc") or (eq .Scheduler "lblcr") -}}
	least_con
	{{- end -}}
	{{if eq .Scheduler "sh" -}}
	hash $remote_addr
	{{- end -}}
	{{if eq .Scheduler "dh" -}}
	ip_hash # hash $remote_addr # todo: double check what _addr to use
	{{- end -}}
	{{if eq .Scheduler "sed" -}}
	least_time connect
	{{- end -}}
	{{if eq .Scheduler "nq" -}}
	least_time first_byte # (or "header") todo: double check
	{{- end -}}
	{{- with .Servers -}}
		{{range .}}
		server {{.Host}}:{{.Port}} {{if .Weight}}weight={{.Weight}}{{end}} {{- if .UpperThreshold}} max_conns={{.UpperThreshold}}{{end}};
		{{- end -}}
	{{end}}
	}

	server {
		listen        {{.Host}}:{{.Port}}{{if ne .Type "tcp"}} udp{{end}};
		proxy_pass    {{.Id}};
		{{if ne .Persistence 0 -}}
		proxy_timeout {{.Persistence}}s;
		{{- end }}
		proxy_connect_timeout 1s;
	}
{{end}}
}
`, n.originalConfig)

	// create a new template and parse the config into it.
	t := template.Must(template.New("nginxConfig").Parse(nginxConfig))

	cfgFile, err := os.Create(n.configFile)
	defer cfgFile.Close()
	if err != nil {
		return fmt.Errorf("Failed to create config file - %s", err)
	}

	config.Log.Trace("Regenerating nginx config - %+q", n.Services)

	// execute the template
	err = t.ExecuteTemplate(cfgFile, "nginxConfig", n.Services)
	if err != nil {
		return fmt.Errorf("Failed to generate config file - %s", err)
	}

	// reload nginx
	out, err := exec.Command("nginx", "-s", "reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to reload nginx - %s", out)
	}

	return nil
}
