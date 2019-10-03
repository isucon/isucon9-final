package isutrain

type ClientOption func(o *ClientOptions)

type ClientOptions struct {
	wantStatusCode int
	wantIsOK       bool

	// Client側で自動アサーションを行うか否か
	autoAssert bool

	// 検索結果の座席数をアサーションするか否か
	seatCount       int
	assertSeatCount bool
}

func newClientOptions(statusCode int, opts ...ClientOption) *ClientOptions {
	o := &ClientOptions{
		wantStatusCode:  statusCode,
		wantIsOK:        true,
		autoAssert:      true,
		seatCount:       0,
		assertSeatCount: false,
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

func EnableAssertSeatCountOpt(count int) ClientOption {
	return func(o *ClientOptions) {
		o.seatCount = count
		o.assertSeatCount = true
	}
}
