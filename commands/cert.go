package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/nanopack/portal/core"
)

// add-cert
// remove-cert
// show-certs
// set-certs

var (
	certAddCmd = &cobra.Command{
		Use:   "add-cert",
		Short: "Add cert",
		Long:  ``,

		Run: certAdd,
	}
	certRemoveCmd = &cobra.Command{
		Use:   "remove-cert",
		Short: "Remove cert",
		Long:  ``,

		Run: certRemove,
	}
	certsShowCmd = &cobra.Command{
		Use:   "show-certs",
		Short: "Show all certs",
		Long:  ``,

		Run: certsShow,
	}
	certsSetCmd = &cobra.Command{
		Use:   "set-certs",
		Short: "Set cert list",
		Long:  ``,

		Run: certsSet,
	}
	certJsonString string
	cert           core.CertBundle
)

func init() {
	certFlags(certAddCmd) // must use json string because Targets is a []string
	certFlags(certsSetCmd)
	certFlags(certRemoveCmd)

}

func certFlags(ccmd *cobra.Command) {
	ccmd.Flags().StringVarP(&certJsonString, "json", "j", "", "Json encoded data for cert(s)")
	ccmd.Flags().StringVarP(&cert.Key, "key", "k", "", "PEM style key")
	ccmd.Flags().StringVarP(&cert.Cert, "cert", "C", "", "PEM style certificate")
}

func certAdd(ccmd *cobra.Command, args []string) {
	if certJsonString != "" {
		err := json.Unmarshal([]byte(certJsonString), &cert)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}

	jsonBytes, err := json.Marshal(cert)
	if err != nil {
		fail("Bad values for cert")
	}
	res, err := rest("certs", "POST", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func certRemove(ccmd *cobra.Command, args []string) {
	if certJsonString != "" {
		err := json.Unmarshal([]byte(certJsonString), &cert)
		if err != nil {
			fail("Bad JSON syntax")
		}
	}

	jsonBytes, err := json.Marshal(cert)
	if err != nil {
		fail("Bad values for cert")
	}
	res, err := rest("certs", "DELETE", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func certsShow(ccmd *cobra.Command, args []string) {
	res, err := rest("certs", "GET", nil)
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}

func certsSet(ccmd *cobra.Command, args []string) {
	certs := []core.CertBundle{}

	err := json.Unmarshal([]byte(certJsonString), &certs)
	if err != nil {
		fail("Bad JSON syntax")
	}
	jsonBytes, err := json.Marshal(certs)
	if err != nil {
		fail("Bad values for cert")
	}
	res, err := rest("certs", "PUT", bytes.NewBuffer(jsonBytes))
	if err != nil {
		fail("Could not contact portal - %v", err)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fail("Could not read portal's response - %v", err)
	}
	fmt.Print(string(b))
}
