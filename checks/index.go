package checks

func NewRegistry() map[string]func() Checker {
	return map[string]func() Checker{
		"tcp": func() Checker {
			return NewTcpChecker()
		},
		"grpc": func() Checker {
			return NewGrpcChecker()
		},
	}
}
