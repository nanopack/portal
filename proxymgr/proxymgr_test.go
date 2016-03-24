package proxymgr_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/proxymgr"
)

var (
	key       = "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCbZCr7ZOKrTVOd\nXpwatbrGLl8v4SRdJd17ys+Gm/VToe9oF4FMTGq02agtmeiPeaQHpnaXW5cr8oEB\n39q40SVzPEZIkT6n5vbWj+SMpGH1/ppo07H/aNEU0BBwdgvSR+2sU+ypeX04RQuS\n6pdTuXhy4kNoNaKIWyO7OqK9M0Hn8/x2yLyZhKts4gn835IEQiLWSSE0hK/Q0yqX\npx2HxkH14F+rdFDxMn7StEZoyoTaVpGQRpIO6etejg9BoDMXhSB7VgiO5hvLwp43\n36tbWpVadeqkS6YW+V5cs95LE3+oYK6l5Yyz572BqZM0Caitfh//710lbthe6p5+\nSzsNb/Q5AgMBAAECggEATL78O41oJhLa6S6BCvAWfysH+C3KN/crnKheNq1wTQ39\nn/t78KMNUKTvWxZYtgPt75lXmQmzcBElhjd5Xy5swK1USSLzPxnjb7VBu/S0LTrC\nKGPl1a9/FDhu5hxnWkQMLsCEcm9+WPxA6x7R/pfr1VHK2P0keRQKYb5kAe3+7v/c\n7jTMRMmlcY48SBIIObbPClPrpQEhOPIv5Eig0P+1Pmer7HkMVuNtyMropRQ6v5gt\n+nc0ytmwWylZMMbhiF8XHTAKY2xEyUc56zlKjzRCL80iwtaH/Vr4h01zLSwGUH1w\n84oFuwEYyxhm4GZAFwXRX3gf+FD5gV4+mj+4H5wSwQKBgQDMYEaQd/S6EEUbNaHq\n6JDZNSb2Re96mknh7YEyB/oCaID3MsCbuNQMX5uFtDI1mc3vJly17oR1v+et5zhP\nMHl8OZ5wEyArrHcoTE/r8K96jZleeUX9Cz8ujV0ZD/CGoBLL6OlptKt5FHcoga7H\n0ZdE024CHT+DI8PPqpZpu1n0rwKBgQDCpFn5kF5iBkfJBKDciy+i9gWFd6gDhx5I\nnQvwGvAC02BWuPKH6uzmRJYFSvRvfaG1oKqX5xVlAQZksJUMZxqT9j1riGABKXMr\nnnhq8bNyFYDorCaaVfxSt+GB0z/siDYVeZOJlcUIOKviVqH+HMXC9kTfJCTQuF6d\nR+M9pfOvlwKBgBDYlrhtytRTZv7ZKuGMDfR5dx6xoQ3ADfr7crzG/4qXRpoZqtqr\nH39tmgopUkIszVa7GMU+RdjW2qfw+Sk926Wrsi2Wxf4TlzbRI31VN4Gojk3FPUmg\nVbLmoBfiwna2VxZLuoGmDMRMNY43MkryMb/Qla7C7mtG1WsWqpNIiB+tAoGASWoS\nIcZpQxHZW6GqRuUct5uR45CJR6NcMclCanLOmlI94RfrKobaidPOvfpSjgbVyprq\nHVdkw28KiUntPftZk/tpmTib9XQ743TnOHcn1tzzfU8JVGcgP9bpcL1MPBv4QktT\n8a4S3hH6CungOeeCVBHtUjjgxfT0guBNfsAsVMsCgYEAmNVIr1uTRaIAOnSl3H9u\nrCMz2IhsvPHxS2R0VPHiJCjCRld16O8cLjdkf8F1DGVJVbjLgUR8YDmgaGsFrc1d\nKuWr0SEvUEpwWMEhBeBzVrfWUNgfHo4nTP6WmGAj2S4++mk6F44RuPnky1R8Ea/i\nq01TKnEAgdm+zV2a1ydiSpc=\n-----END PRIVATE KEY-----"
	cert      = "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAL/FFFuKTjwRMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV\nBAYTAlVTMQswCQYDVQQIDAJJRDETMBEGA1UECgwKbmFub2JveC5pbzEUMBIGA1UE\nAwwLcG9ydGFsLnRlc3QwHhcNMTYwMzIzMTQ1NjMzWhcNMTcwMzIzMTQ1NjMzWjBF\nMQswCQYDVQQGEwJVUzELMAkGA1UECAwCSUQxEzARBgNVBAoMCm5hbm9ib3guaW8x\nFDASBgNVBAMMC3BvcnRhbC50ZXN0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAm2Qq+2Tiq01TnV6cGrW6xi5fL+EkXSXde8rPhpv1U6HvaBeBTExqtNmo\nLZnoj3mkB6Z2l1uXK/KBAd/auNElczxGSJE+p+b21o/kjKRh9f6aaNOx/2jRFNAQ\ncHYL0kftrFPsqXl9OEULkuqXU7l4cuJDaDWiiFsjuzqivTNB5/P8dsi8mYSrbOIJ\n/N+SBEIi1kkhNISv0NMql6cdh8ZB9eBfq3RQ8TJ+0rRGaMqE2laRkEaSDunrXo4P\nQaAzF4Uge1YIjuYby8KeN9+rW1qVWnXqpEumFvleXLPeSxN/qGCupeWMs+e9gamT\nNAmorX4f/+9dJW7YXuqefks7DW/0OQIDAQABo1AwTjAdBgNVHQ4EFgQU66LzKbHE\nyE9LCnaqkcEwOeVQ3fgwHwYDVR0jBBgwFoAU66LzKbHEyE9LCnaqkcEwOeVQ3fgw\nDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAQ2lAzHHyJfyONWfcao6C\nOz5k8Il4eJ3d55qqYvyVBBWp/sFIh9aLGDazbaX7sO55cur/uWp0SiiMw/tt+2nG\n6Yn08l1FeSBDXwvrFOJXScSMEb7Ttl3y2qfJ3z6/rPx6eIBU0c/uzAH+sHiIQNJ1\n7FXD7CvGSIzxU0UU1LEsgM0o5HrOLPubsHmKruM8hcKxHkj9pXKIgY4SJe4BOhwm\nbVh43+VrCDNJf79/KmWrwFXFMg2QvsGS673ps1uGEafGj5vzX4n9S0aCV71ser5P\nmVX2N3jj2WgiYIXI5SmH3BlfR5aGWq4Fq124gi9dxkZljFTolTc6aYyQu0i40B0X\nzQ==\n-----END CERTIFICATE-----"
	testCert  = core.CertBundle{Key: key, Cert: cert}
	testRoute = core.Route{Domain: "portal.test", Page: "routing works\n"}
)

func TestMain(m *testing.M) {
	initialize()

	os.Exit(m.Run())
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////
func TestSetRoute(t *testing.T) {
	if err := proxymgr.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
		t.FailNow()
	}

	// test idempotency
	if err := proxymgr.SetRoute(testRoute); err != nil {
		t.Errorf("Failed to SET route - %v", err)
		t.FailNow()
	}

	routes, err := proxymgr.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) == 1 && routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestSetRoutes(t *testing.T) {
	if err := proxymgr.SetRoutes([]core.Route{testRoute}); err != nil {
		t.Errorf("Failed to SET routes - %v", err)
		t.FailNow()
	}

	routes, err := proxymgr.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) == 1 && routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestGetRoutes(t *testing.T) {
	routes, err := proxymgr.GetRoutes()
	if err != nil {
		t.Errorf("Failed to GET routes - %v", err)
		t.FailNow()
	}

	if routes[0].Domain != testRoute.Domain {
		t.Errorf("Read route differs from written route")
	}
}

func TestDeleteRoute(t *testing.T) {
	if err := proxymgr.DeleteRoute(testRoute); err != nil {
		t.Errorf("Failed to DELETE route - %v", err)
	}

	routes, err := proxymgr.GetRoutes()
	if err != nil {
		t.Error(err)
	}

	if len(routes) != 0 {
		t.Errorf("Failed to DELETE route")
	}
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////
func TestSetCert(t *testing.T) {
	if err := proxymgr.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
		t.FailNow()
	}

	// test idempotency
	if err := proxymgr.SetCert(testCert); err != nil {
		t.Errorf("Failed to SET cert - %v", err)
		t.FailNow()
	}

	certs, err := proxymgr.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) == 1 && certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestSetCerts(t *testing.T) {
	if err := proxymgr.SetCerts([]core.CertBundle{testCert}); err != nil {
		t.Errorf("Failed to SET certs - %v", err)
		t.FailNow()
	}

	certs, err := proxymgr.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) == 1 && certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}

	// test bad tls start (certs must be in place)
	config.RouteHttp = "0.0.0.0:9084"
	config.RouteTls = "!@#$%^&*"
	err = proxymgr.Init()
	if err == nil {
		fmt.Printf("Proxymgr init succeeded when it should have failed\n")
		os.Exit(1)
	}
}

func TestGetCerts(t *testing.T) {
	certs, err := proxymgr.GetCerts()
	if err != nil {
		t.Errorf("Failed to GET certs - %v", err)
		t.FailNow()
	}

	if certs[0].Cert != testCert.Cert {
		t.Errorf("Read cert differs from written cert")
	}
}

func TestDeleteCert(t *testing.T) {
	if err := proxymgr.DeleteCert(testCert); err != nil {
		t.Errorf("Failed to DELETE cert - %v", err)
	}

	certs, err := proxymgr.GetCerts()
	if err != nil {
		t.Error(err)
	}

	if len(certs) != 0 {
		t.Errorf("Failed to DELETE cert")
	}
}

// initialize proxymgr
func initialize() {
	config.LogLevel = "FATAL"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// bad initialize proxymgr
	config.RouteHttp = "!@#$%^&*"
	err := proxymgr.Init()
	if err == nil {
		fmt.Printf("Proxymgr init succeeded when it should have failed\n")
		os.Exit(1)
	}

	// initialize proxymgr
	config.RouteHttp = "0.0.0.0:9083"
	config.RouteTls = "0.0.0.0:9446"
	err = proxymgr.Init()
	if err != nil {
		fmt.Printf("Proxymgr init failed - %v\n", err)
		os.Exit(1)
	}
}
