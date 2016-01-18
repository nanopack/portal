package commands

import (
	"fmt"
	"io/ioutil"

	// "github.com/nanobox-io/golang-lvs"
	"github.com/spf13/cobra"

	// "github.com/nanopack/portal/config"
)

// sync-lvs
// sync-portal

var (
	syncLvsCmd = &cobra.Command{
		Use:   "sync-lvs",
		Short: "Sync to LVS from portal",
		Long:  ``,

		Run: syncLvs,
	}
	syncPortalCmd = &cobra.Command{
		Use:   "sync-portal",
		Short: "Sync to portal from LVS",
		Long:  ``,

		Run: syncPortal,
	}
)

func syncLvs(ccmd *cobra.Command, args []string) {
	res, err := rest("sync", "POST", nil)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func syncPortal(ccmd *cobra.Command, args []string) {
	res, err := rest("sync", "GET", nil)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
