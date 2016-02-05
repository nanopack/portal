package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/nanopack/portal/core"
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
	serviceSetCmd = &cobra.Command{
		Use:   "set-service",
		Short: "Set service",
		Long:  ``,

		Run: serviceSet,
	}
	serviceJsonString string
	service           core.Service
)

func init() {
	serviceComplexFlags(serviceAddCmd)
	serviceComplexFlags(serviceSetCmd)
	serviceSimpleFlags(serviceRemoveCmd)
	serviceSimpleFlags(serviceShowCmd)
	servicesSetCmd.Flags().StringVarP(&serviceJsonString, "json", "j", "", "Json encoded data for services")
	serviceSetCmd.Flags().StringVarP(&serviceJsonString, "json", "j", "", "Json encoded data for services")
	serviceAddCmd.Flags().StringVarP(&serviceJsonString, "json", "j", "", "Json encoded data for services")
}

func serviceSimpleFlags(ccmd *cobra.Command) {
	ccmd.Flags().StringVarP(&service.Id, "service-id", "I", "",
		"Id of down-stream service")

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
	if serviceJsonString != "" {
		err := json.Unmarshal([]byte(serviceJsonString), &service)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}

	jsonBytes, err := json.Marshal(service)
	if err != nil {
		fail("Bad values for service")
	}
	res, err := rest("services", "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func serviceRemove(ccmd *cobra.Command, args []string) {
	svcValidate(&service)

	path := fmt.Sprintf("services/%s", service.Id)
	res, err := rest(path, "DELETE", nil)
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func serviceShow(ccmd *cobra.Command, args []string) {
	svcValidate(&service)

	path := fmt.Sprintf("services/%s", service.Id)
	res, err := rest(path, "GET", nil)
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func servicesShow(ccmd *cobra.Command, args []string) {
	res, err := rest("services", "GET", nil)
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func servicesSet(ccmd *cobra.Command, args []string) {
	services := []core.Service{}

	err := json.Unmarshal([]byte(serviceJsonString), &services)
	if err != nil {
		fail("Bad JSON syntax")
	}
	jsonBytes, err := json.Marshal(services)
	if err != nil {
		fail("Bad values for service")
	}
	res, err := rest("services", "PUT", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func serviceSet(ccmd *cobra.Command, args []string) {
	svcValidate(&service)
	// set path here in case they set the id in their payload
	path := fmt.Sprintf("services/%v", service.Id)

	if serviceJsonString != "" {
		err := json.Unmarshal([]byte(serviceJsonString), &service)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}
	jsonBytes, err := json.Marshal(service)
	if err != nil {
		fail("Bad values for service")
	}

	res, err := rest(path, "PUT", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func svcValidate(service *core.Service) {
	if service.Id == "" {
		service.GenId()
		if service.Host == "" || int(service.Port) == 0 {
			fail("Must enter host and port combo or id of service")
		}
	}
}
