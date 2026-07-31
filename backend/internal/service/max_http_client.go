package service

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	_ "embed"
)

// MAX API uses TLS certificates issued by Russian Trusted CA (MinDigital).
// Alpine Docker images do not include these roots by default.
//
//go:embed certs/russian_trusted_root_ca.pem
var russianTrustedRootCAPEM []byte

//go:embed certs/russian_trusted_sub_ca.pem
var russianTrustedSubCAPEM []byte

func newMaxHTTPClient() *http.Client {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	_, _ = pool.AppendCertsFromPEM(russianTrustedRootCAPEM), pool.AppendCertsFromPEM(russianTrustedSubCAPEM)

	return &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}
}
