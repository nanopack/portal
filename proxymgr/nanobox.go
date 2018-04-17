// nanobox.go contains logic to use nanobox-router as a proxy

package proxymgr

import (
	"fmt"

	"github.com/jcelliott/lumber"
	"github.com/nanobox-io/nanobox-router"

	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
)

type Nanobox struct{}

func (self Nanobox) Init() error {
	// we want to see nanobox-router logs
	lumber.Level(lumber.LvlInt(config.LogLevel))

	// configure upstream cert checks
	router.IgnoreUpstreamCerts = config.ProxyIgnore

	// start http proxy
	err := router.StartHTTP(config.RouteHttp)
	if err != nil {
		return err
	}
	config.Log.Info("Proxy listening at http://%s...", config.RouteHttp)

	// set a default cert (*.nanoapp.io) self signed
	err = self.SetDefaultCert(core.CertBundle{Cert: "-----BEGIN CERTIFICATE-----\nMIIDnzCCAoegAwIBAgIJAMoiK3cYcT01MA0GCSqGSIb3DQEBCwUAMGYxCzAJBgNV\nBAYTAlVTMQ4wDAYDVQQIDAVJZGFobzEQMA4GA1UEBwwHUmV4YnVyZzEQMA4GA1UE\nCgwHTmFub2JveDEMMAoGA1UECwwDT3JnMRUwEwYDVQQDDAwqLm5hbm9hcHAuaW8w\nHhcNMTgwNDExMjIxNjEzWhcNMjgwNDA4MjIxNjEzWjBmMQswCQYDVQQGEwJVUzEO\nMAwGA1UECAwFSWRhaG8xEDAOBgNVBAcMB1JleGJ1cmcxEDAOBgNVBAoMB05hbm9i\nb3gxDDAKBgNVBAsMA09yZzEVMBMGA1UEAwwMKi5uYW5vYXBwLmlvMIIBIjANBgkq\nhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvKzIzdVw25whbmrFYLWg+D/xT0Smzli8\nc8vwXwcpx6XXcBcNdagAoxXy4zxjrYHkin0TxLJWNELNlBPkXt7GHJoe0/pbn9Iq\nnpJw1asi0VXGDO4ne+J7q+aUbpshpWG2NE9KEi1LFNfYnNLwvVkAylrmBLYXbdhN\nuWT164PtjEcT1mx3RNR4l8Zey3RgrBXp3y61ePFwPrHTM3t/AQixTPzU5UzArYJI\neEgFJe2cltuwLezLaXisR3IS42m3oBP1toAb6xM+wgznwQnByjCJy75658QGccJ9\npseIZBDYeArYDBcNUjW2Gp//vdDBRyYCn7bAJ9MHAOyHB50eWGLmMwIDAQABo1Aw\nTjAdBgNVHQ4EFgQUyMdlMFwMjKcBxPWC8vUhuQZ9iUYwHwYDVR0jBBgwFoAUyMdl\nMFwMjKcBxPWC8vUhuQZ9iUYwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOC\nAQEALfJe6r2XD/93ZPNOhU6uyTvLkVKTIiE9E1NC5/bqhtHg8XVUR6kp2wwOH0Td\nMdop8XuhZDsZ5kGV9f8UA5s2Q2SFk9Nyu5gNmtVxVUrUb2yWwNmYVoxH+ntvthaK\nfH2lkgqXVQbFU5qYS+Vt4oxhb3ox0cxMkFJ5UsHzL2+8vZfzzL42g7WAolYOdXuo\nh+sLnKQfnqbYSdjyRGDu/uwyoZidrk7bZ2CQIIAKEF3kYDq29Y8zNRXKng83pS/3\nBx/nDgQW4+AAHm7YmfdbzcLLi4oW8lMb6fMW5DrOZ2pbfpSGGb6r1EVP8cHxLUso\n+nefsCWniPqS+FAxsXTxMvVLKA==\n-----END CERTIFICATE-----\n",
		Key: "-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQC8rMjN1XDbnCFu\nasVgtaD4P/FPRKbOWLxzy/BfBynHpddwFw11qACjFfLjPGOtgeSKfRPEslY0Qs2U\nE+Re3sYcmh7T+luf0iqeknDVqyLRVcYM7id74nur5pRumyGlYbY0T0oSLUsU19ic\n0vC9WQDKWuYEthdt2E25ZPXrg+2MRxPWbHdE1HiXxl7LdGCsFenfLrV48XA+sdMz\ne38BCLFM/NTlTMCtgkh4SAUl7ZyW27At7MtpeKxHchLjabegE/W2gBvrEz7CDOfB\nCcHKMInLvnrnxAZxwn2mx4hkENh4CtgMFw1SNbYan/+90MFHJgKftsAn0wcA7IcH\nnR5YYuYzAgMBAAECggEAOiI+6PUIBhKQVnY9hLPR+kuxbYwonVHIFyHSWWVaoTJf\nNCFWO1ddguKDaTK1P8PTCDzLt4J/fzDKKQMMDZM0laGDOCteydq22Q8kByHo43k7\nQcarkcdR9cBhIcdY0Z1Ox8VafElKZgyvqHpyRNVEohTp5K+6flT0ddg+0adfrSW9\n7ilafXKX1976BQWeS2j5AUsaw9qWTMz/QhjYDvvYHsEG6rZFnUbQn+k65ubLOqG8\nYHj5z9Y7EZP/GfsByMgdyTLSnynz5ifzt05yeCUIEEhqsN1NZEMh8wfhzMCAnqYo\nRWujn5hayFmdZhXMiiBTSCUq52HPucOw5IXFfjnzSQKBgQDdNicOtx/o2j0QgzHo\ngB2tyieao+nlleQyK06xgRz/xjiRyObRuRpTYB1f6lyWKjwn5gdCnsj2JyUO0UAr\n7Xpc8suESyFbKBJ8QlYfagB72QN5k2NlxNE8kXSDruSS5Cu8N3OX025sFTuwiOhB\nmV3LU5VWvjcEs2T8goaMmGCV/QKBgQDaWLo8yvM19sc8z/3D4eQFcCQFlbz0/qPR\nPm8hWhd+U5KIKCKi4b03RnedJ8SSIHGM5JNKbsXVzCIOLOKq+eNCabYxI4TfgVWw\nIOXKOg9EdUdpHjf0V0kVFvl6bKcjkv/k51bVqjzeoeE8/iAiEWRmTNjyFVt3NXzB\nuncih2kL7wKBgQCIcUlf+zXUYx/9Gl7jQHqN4j3RVT8EnBKXmzy7oZ6oaLQlv5wX\nSaviN0uHCMA44y4dkfVycwwTQAvMeuaw8ZZi1GMRY2Hcnvff6u7CC5jmyvEowO8z\nK1W/nRwXyP01WUVcn3tN71yRj/s0JQ3UwGso6ZIYYdT/skMcuMmS2L3iZQKBgQCm\nb6Pm2zzxEZ9lt5XUTsglbQnISA+1ILV2toS3g5kM7l9v8kgUqMY28DwVS08HpDtq\nDoJH5pBfHC+JZqWRdtHIuhPq+Qw74raSf0EqGX+xy0QX2LUGR9KphM2+iDwPXeo5\nbi4+yHmFqxeqCnwr+93wLPvh7G3APMFQWvadF2L3eQKBgQC7zGqilReyuPleRkc6\nud4YweWzLov2IOK4JM/2c6AxzhRzuIMMDEPFbEGSIKPNSabclBf/kdqySYy/ZbGO\nnQ2Z+hPmCf7hfXZBGdEvR/WRkohQE6WdD7jKc2mYgQCIv8AR7yr9Bau4fs1LQLCc\nmm1hHKbUb5R8edPKixtXWKowLA==\n-----END PRIVATE KEY-----\n"})
	if err != nil {
		return fmt.Errorf("Failed to set default cert - %s", err.Error())
	}

	// start https proxy
	err = router.StartTLS(config.RouteTls)
	if err != nil {
		return err
	}
	config.Log.Info("Proxy listening at https://%s...", config.RouteTls)

	// start proxy health checker (todo: make check interval(pulse) configurable)
	go router.StartHealth(20)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ROUTES
////////////////////////////////////////////////////////////////////////////////

func (self Nanobox) SetRoute(route core.Route) error {
	routes := router.Routes()
	// for idempotency
	for i := range routes {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			return nil
		}
	}

	routes = append(routes, self.rToRoute(route))
	return self.SetRoutes(self.routesToR(routes))
}

func (self Nanobox) DeleteRoute(route core.Route) error {
	routes := router.Routes()
	for i := range routes {
		if routes[i].SubDomain == route.SubDomain && routes[i].Domain == route.Domain && routes[i].Path == route.Path {
			routes = append(routes[:i], routes[i+1:]...)
			break
		}
	}
	return self.SetRoutes(self.routesToR(routes))
}

func (self Nanobox) SetRoutes(routes []core.Route) error {
	return router.UpdateRoutes(self.rToRoutes(routes))
}

func (self Nanobox) GetRoutes() ([]core.Route, error) {
	return self.routesToR(router.Routes()), nil
}

////////////////////////////////////////////////////////////////////////////////
// CERTS
////////////////////////////////////////////////////////////////////////////////

func (self Nanobox) SetCert(cert core.CertBundle) error {
	certs := router.Keys()
	// for idempotency
	for i := range certs {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			config.Log.Debug("Cert already added, skipping")
			return nil
		}
	}

	certs = append(certs, self.cToKey(cert))
	return self.SetCerts(self.keysToC(certs))
}

func (self Nanobox) SetDefaultCert(cert core.CertBundle) error {
	return router.SetDefaultCert(cert.Cert, cert.Key)
}

func (self Nanobox) DeleteCert(cert core.CertBundle) error {
	certs := router.Keys()
	for i := range certs {
		if certs[i].Cert == cert.Cert && certs[i].Key == cert.Key {
			certs = append(certs[:i], certs[i+1:]...)
			break
		}
	}

	return self.SetCerts(self.keysToC(certs))
}

func (self Nanobox) SetCerts(certs []core.CertBundle) error {
	return router.UpdateCerts(self.cToKeys(certs))
}

func (self Nanobox) GetCerts() ([]core.CertBundle, error) {
	return self.keysToC(router.Keys()), nil
}

func (self Nanobox) cToKeys(certs []core.CertBundle) []router.KeyPair {
	var keyPairs []router.KeyPair
	for i := range certs {
		keyPairs = append(keyPairs, router.KeyPair{Key: certs[i].Key, Cert: certs[i].Cert})
	}
	return keyPairs
}

func (self Nanobox) keysToC(certs []router.KeyPair) []core.CertBundle {
	var keyPairs []core.CertBundle
	for i := range certs {
		keyPairs = append(keyPairs, core.CertBundle{Key: certs[i].Key, Cert: certs[i].Cert})
	}
	return keyPairs
}

func (self Nanobox) rToRoutes(routes []core.Route) []router.Route {
	var rts []router.Route
	for i := range routes {
		rts = append(rts, router.Route{
			SubDomain:      routes[i].SubDomain,
			Domain:         routes[i].Domain,
			Path:           routes[i].Path,
			Targets:        routes[i].Targets,
			FwdPath:        routes[i].FwdPath,
			Page:           routes[i].Page,
			Endpoint:       routes[i].Endpoint,
			ExpectedCode:   routes[i].ExpectedCode,
			ExpectedBody:   routes[i].ExpectedBody,
			ExpectedHeader: routes[i].ExpectedHeader,
			Host:           routes[i].Host,
			Timeout:        routes[i].Timeout,
			Attempts:       routes[i].Attempts,
		})
	}
	return rts
}

func (self Nanobox) routesToR(routes []router.Route) []core.Route {
	var rts []core.Route
	for i := range routes {
		rts = append(rts, core.Route{
			SubDomain:      routes[i].SubDomain,
			Domain:         routes[i].Domain,
			Path:           routes[i].Path,
			Targets:        routes[i].Targets,
			FwdPath:        routes[i].FwdPath,
			Page:           routes[i].Page,
			Endpoint:       routes[i].Endpoint,
			ExpectedCode:   routes[i].ExpectedCode,
			ExpectedBody:   routes[i].ExpectedBody,
			ExpectedHeader: routes[i].ExpectedHeader,
			Host:           routes[i].Host,
			Timeout:        routes[i].Timeout,
			Attempts:       routes[i].Attempts,
		})
	}
	return rts
}

func (self Nanobox) cToKey(cert core.CertBundle) router.KeyPair {
	return router.KeyPair{Key: cert.Key, Cert: cert.Cert}
}

func (self Nanobox) rToRoute(route core.Route) router.Route {
	return router.Route{SubDomain: route.SubDomain, Domain: route.Domain, Path: route.Path, Targets: route.Targets, FwdPath: route.FwdPath, Page: route.Page}
}
