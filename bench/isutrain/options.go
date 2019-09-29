package isutrain

type ClientOption func(o *ClientOptions)

type ClientOptions struct {
	wantStatusCode int
	autoAssert     bool
}

func newClientOptions(statusCode int, opts ...ClientOption) *ClientOptions {
	o := &ClientOptions{
		wantStatusCode: statusCode,
		autoAssert:     true,
	}
	for _, opt := range opts {
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
