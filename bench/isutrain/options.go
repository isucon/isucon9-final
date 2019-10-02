package isutrain

type ClientOption func(o *ClientOptions)

type ClientOptions struct {
	wantStatusCode int
	wantIsOK       bool
	autoAssert     bool
}

func newClientOptions(statusCode int, opts ...ClientOption) *ClientOptions {
	o := &ClientOptions{
		wantStatusCode: statusCode,
		wantIsOK:       true,
		autoAssert:     true,
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

func IsOKOpt(isOK bool) ClientOption {
	return func(o *ClientOptions) {
		o.wantIsOK = isOK
	}
}

func DisableAssertOpt() ClientOption {
	return func(o *ClientOptions) {
		o.autoAssert = false
	}
}
