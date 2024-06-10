package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"os"
	"time"
)

type GrpcParams struct {
	Host     string
	Port     int
	CertFile string
	CAFile   string
	KeyFile  string
}

func NewGrpcChecker() Checker {

	type result struct {
		s *Status
		e error
	}

	return NewFromTemplate[GrpcParams](func(ctx context.Context, config *GrpcParams) (*Status, error) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()

		ch := make(chan result, 1)

		go func() {

			tlsConfig, err := loadTLSConfig(config.CertFile, config.KeyFile, config.CAFile)

			if err != nil {
				fmt.Println(err.Error())
				ch <- result{
					s: StatusUnknown,
				}
				return
			}

			// TODO - add credentials
			c, err := grpc.NewClient(fmt.Sprintf("%s:%d", config.Host, config.Port),
				grpc.WithTransportCredentials(tlsConfig))
			if err != nil {
				fmt.Println(err.Error())
				ch <- result{
					s: StatusUnknown,
				}
				return
			}
			defer c.Close()
			resp, err := grpc_health_v1.NewHealthClient(c).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
			if err != nil {
				fmt.Println(err.Error())
				ch <- result{
					s: StatusUnhealthy,
				}
				return
			}
			if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
				ch <- result{
					s: StatusHealthy,
				}
				return
			} else {
				fmt.Println("not healthy")
				ch <- result{
					s: StatusUnhealthy,
				}
				return
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

func loadTLSConfig(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), nil
}
