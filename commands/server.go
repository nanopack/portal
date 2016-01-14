package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/nanobox-io/golang-lvs"
	"github.com/spf13/cobra"

	// "github.com/nanopack/portal/config"
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
)

func serverSimpleFlags(ccmd *cobra.Command, server *lvs.Server) {
	ccmd.Flags().StringVarP(&server.Host, "host", "h", "",
		"Host of down-stream server")
	ccmd.Flags().IntVarP(&server.Port, "port", "p", 0,
		"Port of down-stream server")
}

func serverComplexFlags(ccmd *cobra.Command, server *lvs.Server) {
	serverSimpleFlags(ccmd, server)
	ccmd.Flags().StringVarP(&server.Forwarder, "forwarder", "f", "g", "Forwarder method [g i m]")
	ccmd.Flags().IntVarP(&server.Weight, "weight", "w", 1, "weight of down-stream server")
	ccmd.Flags().IntVarP(&server.UpperThreshold, "upper-threshold", "u", 0, "Upper threshold of down-stream server")
	ccmd.Flags().IntVarP(&server.LowerThreshold, "lower-threshold", "l", 0, "Lower threshold of down-stream server")
}

func serverAdd(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	server := lvs.Server{}
	serviceSimpleFlags(ccmd, &service)
	serverComplexFlags(ccmd, &server)

	jsonBytes, err := json.Marshal(server)
	if err != nil {
		panic(err)
	}
	path := fmt.Sprintf("services/%s/%s/%d/servers/%s/%d", service.Type, service.Host, service.Port, server.Host, server.Port)
	res, err := rest(path, "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

}

func serverRemove(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	server := lvs.Server{}
	serviceSimpleFlags(ccmd, &service)
	serverSimpleFlags(ccmd, &server)

	path := fmt.Sprintf("services/%s/%s/%d/servers/%s/%d", service.Type, service.Host, service.Port, server.Host, server.Port)
	res, err := rest(path, "DELETE", nil)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func serverShow(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	server := lvs.Server{}
	serviceSimpleFlags(ccmd, &service)
	serverSimpleFlags(ccmd, &server)

	path := fmt.Sprintf("services/%s/%s/%d/servers/%s/%d", service.Type, service.Host, service.Port, server.Host, server.Port)
	res, err := rest(path, "GET", nil)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func serversShow(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	serviceSimpleFlags(ccmd, &service)

	path := fmt.Sprintf("services/%s/%s/%d/servers", service.Type, service.Host, service.Port)
	res, err := rest(path, "GET", nil)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func serversSet(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	servers := []lvs.Server{}
	jsonString := ""
	serviceSimpleFlags(ccmd, &service)
	ccmd.Flags().StringVarP(&jsonString, "json", "j", "", "Json encoded data for servers")

	err := json.Unmarshal([]byte(jsonString), &servers)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.Marshal(servers)
	if err != nil {
		panic(err)
	}
	path := fmt.Sprintf("services/%s/%s/%d/servers", service.Type, service.Host, service.Port)
	res, err := rest(path, "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
