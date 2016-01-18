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

// service-add
// service-remove
// service-show
// services-show
// services-set

var (
	serviceAddCmd = &cobra.Command{
		Use:   "add-service",
		Short: "Add service",
		Long:  ``,

		Run: serviceAdd,
	}
	serviceRemoveCmd = &cobra.Command{
		Use:   "remove-service",
		Short: "Remove service",
		Long:  ``,

		Run: serviceRemove,
	}
	serviceShowCmd = &cobra.Command{
		Use:   "show-service",
		Short: "Show service",
		Long:  ``,

		Run: serviceShow,
	}
	servicesShowCmd = &cobra.Command{
		Use:   "show-services",
		Short: "Show all services",
		Long:  ``,

		Run: servicesShow,
	}
	servicesSetCmd = &cobra.Command{
		Use:   "set-services",
		Short: "Set service list",
		Long:  ``,

		Run: servicesSet,
	}
	serviceJsonString string
	service           lvs.Service
)

func init() {
	serviceComplexFlags(serviceAddCmd)
	serviceSimpleFlags(serviceRemoveCmd)
	serviceSimpleFlags(serviceShowCmd)
	servicesSetCmd.Flags().StringVarP(&serviceJsonString, "json", "j", "", "Json encoded data for services")
}

func serviceSimpleFlags(ccmd *cobra.Command) {
	ccmd.Flags().StringVarP(&service.Host, "service-host", "O", "",
		"Host of down-stream service")
	ccmd.Flags().IntVarP(&service.Port, "service-port", "R", 0,
		"Port of down-stream service")
	ccmd.Flags().StringVarP(&service.Type, "service-type", "T", "tcp",
		"Type of service [tcp udp fwmark]")
}

func serviceComplexFlags(ccmd *cobra.Command) {
	serviceSimpleFlags(ccmd)
	ccmd.Flags().StringVarP(&service.Scheduler, "service-scheduler", "s", "wlc", "Scheduler method [rr wrr lc wlc lblc lblcr dh sh sed nq]")
	ccmd.Flags().IntVarP(&service.Persistence, "service-persistence", "e", 0, "keep connections persistent to the same down stream server")
	ccmd.Flags().StringVarP(&service.Netmask, "service-netmask", "n", "", "Netmask to group by")
}

func serviceAdd(ccmd *cobra.Command, args []string) {
	jsonBytes, err := json.Marshal(service)
	if err != nil {
		panic(err)
	}
	path := fmt.Sprintf("services/%s/%s/%d", service.Type, service.Host, service.Port)
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

func serviceRemove(ccmd *cobra.Command, args []string) {
	path := fmt.Sprintf("services/%s/%s/%d", service.Type, service.Host, service.Port)
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

func serviceShow(ccmd *cobra.Command, args []string) {
	path := fmt.Sprintf("services/%s/%s/%d", service.Type, service.Host, service.Port)
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

func servicesShow(ccmd *cobra.Command, args []string) {
	res, err := rest("services", "GET", nil)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func servicesSet(ccmd *cobra.Command, args []string) {
	services := []lvs.Service{}

	err := json.Unmarshal([]byte(serviceJsonString), &services)
	if err != nil {
		panic(err)
	}
	jsonBytes, err := json.Marshal(services)
	if err != nil {
		panic(err)
	}
	res, err := rest("services", "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
