package puller

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"path"
)

type (
	// Client is an HTTP client to connect to ChartMuseum
	client struct {
		*http.Client
		opts options
	}
)

// Option configures the client with the provided options.
func (its *client) Option(opts ...option) *client {
	for _, opt := range opts {
		opt(&its.opts)
	}
	return its
}

func (its *client) downloadFile(filePath string) (*http.Response, error) {
	u, err := url.Parse(its.opts.url)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, filePath)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	if its.opts.username != "" && its.opts.password != "" {
		req.SetBasicAuth(its.opts.username, its.opts.password)
	}

	return its.Do(req)
}

func newClient(opts ...option) (*client, error) {
	var c client
	c.Client = &http.Client{}
	c.Option(timeout(30))
	c.Option(opts...)
	c.Timeout = c.opts.timeout

	//Enable tls config if configured
	tr, err := newTransport(
		c.opts.insecureSkipVerify,
	)
	if err != nil {
		return nil, err
	}

	c.Transport = tr

	return &c, nil
}

func newTransport(insecureSkipVerify bool) (*http.Transport, error) {
	transport := &http.Transport{}
	config := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	transport.TLSClientConfig = config
	transport.Proxy = http.ProxyFromEnvironment
	return transport, nil
}
