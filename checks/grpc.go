package checks

import (
	"context"
	"crypto/tls"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
	"time"
)

type GrpcParams struct {
	Host string
	Port int
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

			// TODO - add credentials
			c, err := grpc.NewClient(fmt.Sprintf("%s:%d", config.Host, config.Port),
				grpc.WithTransportCredentials(
					credentials.NewTLS(&tls.Config{
						InsecureSkipVerify: true,
					})))
			if err != nil {
				fmt.Println(err.Error())
				ch <- result{
					s: StatusUnhealthy,
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
