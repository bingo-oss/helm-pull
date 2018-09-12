package puller

import (
	"time"
)

type (
	// Option allows specifying various settings
	option func(*options)

	// options specify optional settings
	options struct {
		url                string
		username           string
		password           string
		timeout            time.Duration
		insecureSkipVerify bool
	}
)

// URL specifies the chart repo URL
func repourl(url string) option {
	return func(opts *options) {
		opts.url = url
	}
}

// Username is HTTP basic auth username
func username(username string) option {
	return func(opts *options) {
		opts.username = username
	}
}

// Password is HTTP basic auth password
func password(password string) option {
	return func(opts *options) {
		opts.password = password
	}
}

// Timeout specifies the duration (in seconds) before timing out request
func timeout(timeout int64) option {
	return func(opts *options) {
		opts.timeout = time.Duration(timeout) * time.Second
	}
}

//InsecureSkipVerify to indicate if verify the certificate when connecting
func insecureSkipVerify(insecureSkipVerify bool) option {
	return func(opts *options) {
		opts.insecureSkipVerify = insecureSkipVerify
	}
}
