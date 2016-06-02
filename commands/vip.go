package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/nanopack/portal/core"
)

// add-vip
// remove-vip
// show-vips
// set-vips

var (
	vipAddCmd = &cobra.Command{
		Use:   "add-vip",
		Short: "Add vip",
		Long:  ``,

		Run: vipAdd,
	}
	vipRemoveCmd = &cobra.Command{
		Use:   "remove-vip",
		Short: "Remove vip",
		Long:  ``,

		Run: vipRemove,
	}
	vipsShowCmd = &cobra.Command{
		Use:   "show-vips",
		Short: "Show all vips",
		Long:  ``,

		Run: vipsShow,
	}
	vipsSetCmd = &cobra.Command{
		Use:   "set-vips",
		Short: "Set vip list",
		Long:  ``,

		Run: vipsSet,
	}
	vipJsonString string
	vip           core.Vip
)

func init() {
	vipFlags(vipAddCmd)
	vipFlags(vipsSetCmd)
	vipFlags(vipRemoveCmd)

}

func vipFlags(ccmd *cobra.Command) {
	ccmd.Flags().StringVarP(&vipJsonString, "json", "j", "", "Json encoded data for vip(s)")
	ccmd.Flags().StringVarP(&vip.Ip, "ip", "I", "", "Ip to add (ip/cidr)")
	ccmd.Flags().StringVarP(&vip.Interface, "interface", "F", "", "Interface to add ip on")
	ccmd.Flags().StringVarP(&vip.Alias, "alias", "A", "", "Alias for ip (used as service interface)")
}

func vipAdd(ccmd *cobra.Command, args []string) {
	if vipJsonString != "" {
		err := json.Unmarshal([]byte(vipJsonString), &vip)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}

	jsonBytes, err := json.Marshal(vip)
	if err != nil {
		fail("Bad values for vip")
	}
	res, err := rest("vips", "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func vipRemove(ccmd *cobra.Command, args []string) {
	if vipJsonString != "" {
		err := json.Unmarshal([]byte(vipJsonString), &vip)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}

	jsonBytes, err := json.Marshal(vip)
	if err != nil {
		fail("Bad values for vip")
	}
	res, err := rest("vips", "DELETE", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func vipsShow(ccmd *cobra.Command, args []string) {
	res, err := rest("vips", "GET", nil)
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func vipsSet(ccmd *cobra.Command, args []string) {
	vips := []core.Vip{}

	err := json.Unmarshal([]byte(vipJsonString), &vips)
	if err != nil {
		fail("Bad JSON syntax")
	}
	jsonBytes, err := json.Marshal(vips)
	if err != nil {
		fail("Bad values for vip")
	}
	res, err := rest("vips", "PUT", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}
