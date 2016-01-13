package commands

import (
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
)

func serviceSimpleFlags(ccmd *cobra.Command, service *lvs.Service) {
	ccmd.Flags().StringVarP(&service.Host, "host", "h", "",
		"Host of down-stream service")
	ccmd.Flags().IntVarP(&service.Port, "port", "p", 0,
		"Port of down-stream service")
	ccmd.Flags().StringVarP(&service.Type, "type", "T", "tcp",
		"Type of service [tcp udp fwmark]")
}

func serviceComplexFlags(ccmd *cobra.Command, service *lvs.Service) {
	serviceSimpleFlags(ccmd, service)
	ccmd.Flags().StringVarP(&service.Scheduler, "scheduler", "s", "wlc", "Scheduler method [rr wrr lc wlc lblc lblcr dh sh sed nq]")
	ccmd.Flags().IntVarP(&service.Persistance, "persistance", "e", 0, "keep connections persistent to the same down stream server")
	ccmd.Flags().StringVarP(&service.Netmask, "netmask", "n", "", "Netmask to group by")
}

func serviceAdd(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	serviceComplexFlags(ccmd, &service)

}

func serviceRemove(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	serviceSimpleFlags(ccmd, &service)
}

func serviceShow(ccmd *cobra.Command, args []string) {
	service := lvs.Service{}
	serviceSimpleFlags(ccmd, &service)
}

func servicesShow(ccmd *cobra.Command, args []string) {

}

func servicesSet(ccmd *cobra.Command, args []string) {

}
