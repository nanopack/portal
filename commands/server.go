package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/nanopack/portal/core"
)

// server-add
// server-remove
// server-show
// servers-show
// servers-set

var (
	serverAddCmd = &cobra.Command{
		Use:   "add-server",
		Short: "Add server to a service",
		Long:  ``,

		Run: serverAdd,
	}
	serverRemoveCmd = &cobra.Command{
		Use:   "remove-server",
		Short: "Remove server from a service",
		Long:  ``,

		Run: serverRemove,
	}
	serverShowCmd = &cobra.Command{
		Use:   "show-server",
		Short: "Show server on a service",
		Long:  ``,

		Run: serverShow,
	}
	serversShowCmd = &cobra.Command{
		Use:   "show-servers",
		Short: "Show all servers on a service",
		Long:  ``,

		Run: serversShow,
	}
	serversSetCmd = &cobra.Command{
		Use:   "set-servers",
		Short: "Set server list on a service",
		Long:  ``,

		Run: serversSet,
	}
	server           core.Server
	serverJsonString string
)

func init() {
	serviceSimpleFlags(serverAddCmd)
	serverComplexFlags(serverAddCmd)

	serviceSimpleFlags(serverRemoveCmd)
	serverSimpleFlags(serverRemoveCmd)

	serviceSimpleFlags(serverShowCmd)
	serverSimpleFlags(serverShowCmd)

	serviceSimpleFlags(serversShowCmd)

	serviceSimpleFlags(serversSetCmd)
	serversSetCmd.Flags().StringVarP(&serverJsonString, "json", "j", "", "Json encoded data for servers")
	serverAddCmd.Flags().StringVarP(&serverJsonString, "json", "j", "", "Json encoded data for servers")
}

func serverSimpleFlags(ccmd *cobra.Command) {
	ccmd.Flags().StringVarP(&server.Id, "server-id", "S", "",
		"Id of down-stream server")

	ccmd.Flags().StringVarP(&server.Host, "server-host", "o", "",
		"Host of down-stream server")
	ccmd.Flags().IntVarP(&server.Port, "server-port", "p", 0,
		"Port of down-stream server")
}

func serverComplexFlags(ccmd *cobra.Command) {
	serverSimpleFlags(ccmd)
	ccmd.Flags().StringVarP(&server.Forwarder, "server-forwarder", "f", "g", "Forwarder method [g i m]")
	ccmd.Flags().IntVarP(&server.Weight, "server-weight", "w", 1, "weight of down-stream server")
	ccmd.Flags().IntVarP(&server.UpperThreshold, "server-upper-threshold", "u", 0, "Upper threshold of down-stream server")
	ccmd.Flags().IntVarP(&server.LowerThreshold, "server-lower-threshold", "l", 0, "Lower threshold of down-stream server")
}

func serverAdd(ccmd *cobra.Command, args []string) {
	svcValidate(&service)

	if serverJsonString != "" {
		err := json.Unmarshal([]byte(serverJsonString), &server)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}
	jsonBytes, err := json.Marshal(server)
	if err != nil {
		fail("Bad values for server")
	}
	path := fmt.Sprintf("services/%s/servers", service.Id)
	res, err := rest(path, "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func serverRemove(ccmd *cobra.Command, args []string) {
	svcValidate(&service)
	srvValidate(&server)

	path := fmt.Sprintf("services/%s/servers/%s", service.Id, server.Id)
	res, err := rest(path, "DELETE", nil)
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func serverShow(ccmd *cobra.Command, args []string) {
	svcValidate(&service)
	srvValidate(&server)

	path := fmt.Sprintf("services/%s/servers/%s", service.Id, server.Id)
	res, err := rest(path, "GET", nil)
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func serversShow(ccmd *cobra.Command, args []string) {
	svcValidate(&service)

	path := fmt.Sprintf("services/%s/servers", service.Id)
	res, err := rest(path, "GET", nil)
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func serversSet(ccmd *cobra.Command, args []string) {
	svcValidate(&service)

	var servers []core.Server
	err := json.Unmarshal([]byte(serverJsonString), &servers)
	if err != nil {
		fail("Bad JSON syntax")
	}
	jsonBytes, err := json.Marshal(servers)
	if err != nil {
		fail("Bad values for server")
	}
	path := fmt.Sprintf("services/%s/servers", service.Id)
	res, err := rest(path, "PUT", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %s", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %s", err)
	}
	fmt.Print(string(b))
}

func srvValidate(server *core.Server) {
	if server.Id == "" {
		server.GenId()
		if server.Host == "" || int(server.Port) == 0 {
			fail("Must enter host and port combo or id of server")
		}
	}
}
