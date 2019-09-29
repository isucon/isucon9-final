package isutrain

type ClientOption func(o *ClientOptions)

type ClientOptions struct {
	wantStatusCode  int
	autoAssert      bool
	trainSeatsCount int
}

func newClientOptions(statusCode int, opts ...ClientOption) *ClientOptions {
	o := &ClientOptions{
		wantStatusCode:  statusCode,
		autoAssert:      true,
		trainSeatsCount: 2,
	}
	if len(opts) == 0 {
		return o
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(o)
	}
	return o
}

func StatusCodeOpt(statusCode int) ClientOption {
	return func(o *ClientOptions) {
		o.wantStatusCode = statusCode
	}
}

func DisableAssertOpt() ClientOption {
	return func(o *ClientOptions) {
		o.autoAssert = false
	}
}

func TrainSeatsCountOpt(count int) ClientOption {
	return func(o *ClientOptions) {
		o.trainSeatsCount = count
	}
}
