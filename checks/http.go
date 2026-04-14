package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

type HttpParams struct {
	Host     string
	Port     int
	Path     string
	TLS      bool
	CertFile string
	CAFile   string
	KeyFile  string
}

func NewHttpChecker() Checker {

	type result struct {
		s *Status
		e error
	}

	return NewFromTemplate[HttpParams](func(ctx context.Context, config *HttpParams) (*Status, error) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()

		ch := make(chan result, 1)

		go func() {

			scheme := "http"
			if config.TLS {
				scheme = "https"
			}

			url := fmt.Sprintf("%s://%s:%d%s", scheme, config.Host, config.Port, config.Path)

			client := &http.Client{
				Timeout: time.Second,
			}

			if config.TLS {
				tlsConfig := &tls.Config{}

				if config.CertFile != "" && config.KeyFile != "" {
					certificate, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
					if err != nil {
						fmt.Println(err.Error())
						ch <- result{
							s: StatusUnknown,
						}
						return
					}
					tlsConfig.Certificates = []tls.Certificate{certificate}
				}

				if config.CAFile != "" {
					ca, err := os.ReadFile(config.CAFile)
					if err != nil {
						fmt.Println(err.Error())
						ch <- result{
							s: StatusUnknown,
						}
						return
					}
					capool := x509.NewCertPool()
					if !capool.AppendCertsFromPEM(ca) {
						fmt.Println("failed to append CA certificate")
						ch <- result{
							s: StatusUnknown,
						}
						return
					}
					tlsConfig.RootCAs = capool
				}

				client.Transport = &http.Transport{
					TLSClientConfig: tlsConfig,
				}
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				fmt.Println(err.Error())
				ch <- result{
					s: StatusUnknown,
					e: err,
				}
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				ch <- result{
					s: StatusUnhealthy,
				}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				ch <- result{
					s: StatusHealthy,
				}
			} else {
				ch <- result{
					s: StatusUnhealthy,
				}
			}
		}()

		select {
		case <-ctx.Done():
			return StatusUnhealthy, nil
		case r := <-ch:
			return r.s, r.e
		}
	})
}
