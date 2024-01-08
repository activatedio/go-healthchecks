package checks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

type TcpParams struct {
	Host string
	Port int
}

func NewTcpChecker() Checker {

	type result struct {
		s *Status
		e error
	}

	return NewFromTemplate[TcpParams](func(ctx context.Context, config *TcpParams) (*Status, error) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()

		ch := make(chan result, 1)

		go func() {

			c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
			if err != nil {
				var _e *net.OpError
				if errors.As(err, &_e) {
					s := _e.Err.Error()
					if s == "connect: connection refused" {
						ch <- result{
							s: StatusUnhealthy,
						}
						return
					}
				}
				ch <- result{
					s: StatusUnknown,
					e: err,
				}
				return
			}
			defer c.Close()
			ch <- result{
				s: StatusHealthy,
			}
			return
		}()

		select {
		case <-ctx.Done():
			return StatusUnhealthy, nil
		case r := <-ch:
			return r.s, r.e
		}

	})
}
