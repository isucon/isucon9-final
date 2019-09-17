package util

import (
	"errors"
	"net/url"
)

var (
	ErrURLHostEmpty = errors.New("URLのホスト部が空です")
)

func ParseURL(urlStr string) (*url.URL, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if u.Host == "" {
		return nil, ErrURLHostEmpty
	}

	return &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}, nil
}
