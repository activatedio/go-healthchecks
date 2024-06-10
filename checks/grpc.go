package checks

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
			c, err := grpc.NewClient(fmt.Sprintf("%s:%d", config.Host, config.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				ch <- result{
					s: StatusUnhealthy,
				}
			}
			defer c.Close()
			resp, err := grpc_health_v1.NewHealthClient(c).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
			if err != nil {
				ch <- result{
					s: StatusUnhealthy,
				}
			}
			if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
				ch <- result{
					s: StatusHealthy,
				}
			} else {
				ch <- result{
					s: StatusHealthy,
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