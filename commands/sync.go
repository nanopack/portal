package commands

import (
	// "github.com/nanobox-io/golang-lvs"
	"github.com/spf13/cobra"

	// "github.com/nanopack/portal/config"
)

// sync-lvs
// sync-portal

var (
	syncLvsCmd = &cobra.Command{
		Use:   "sync-lvs",
		Short: "Add server to a service",
		Long:  ``,

		Run: syncLvs,
	}
	syncPortalCmd = &cobra.Command{
		Use:   "sync-portal",
		Short: "Remove server from a service",
		Long:  ``,

		Run: syncPortal,
	}
)

func syncLvs(ccmd *cobra.Command, args []string) {

}

func syncPortal(ccmd *cobra.Command, args []string) {

}
