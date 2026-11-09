package bce

import (
	"net/http"
	"time"

	"github.com/baidubce/bce-sdk-go/services/sms"
	"go.gh.ink/smsutils/v3/errors"
	"go.gh.ink/smsutils/v3/model"
	"go.gh.ink/toolbox/expr"
)

type Client struct {
	Client *sms.Client
	// JSON
	Marshal   func(any) ([]byte, error)
	Unmarshal func([]byte, any) error
}

type Driver struct{}

func (d Driver) NewClient(params model.DriverClientParam) (model.Client, error) {
	// Check credential
	ak, sk := params.Credential[AccessKeyID], params.Credential[SecretAccessKey]
	if ak == "" || sk == "" {
		return Client{}, errors.ErrDriverCredentialInvalid.WithDriverName(Name)
	}

	// Create bce client
	client, err := sms.NewClient(
		ak, sk, expr.Ternary(params.Credential[Endpoint] != "", params.Credential[Endpoint], DefaultEndpoint),
	)
	if err != nil {
		return Client{}, err
	}

	return Client{
		Client:    client,
		Marshal:   params.Marshal,
		Unmarshal: params.Unmarshal,
	}, nil
}

var _ model.ClientCtx = Client{}
var _ model.DriverCtx = Driver{}

// NewClientCtx adds context support without changing the legacy driver contract.
func (d Driver) NewClientCtx(params model.DriverClientParam, options model.ClientOptions) (model.ClientCtx, error) {
	if options.Timeout < 0 {
		return nil, errors.ErrDriverCredentialInvalid.WithDriverName(Name)
	}

	legacy, err := d.NewClient(params)
	if err != nil {
		return nil, err
	}
	c := legacy.(Client)
	if options.Transport != nil || options.Timeout > 0 {
		h := &http.Client{Timeout: 8 * time.Second}
		if c.Client.HTTPClient != nil {
			*h = *c.Client.HTTPClient
		}
		if options.Transport != nil {
			h.Transport = options.Transport
		}
		if options.Timeout > 0 {
			h.Timeout = options.Timeout
		}
		c.Client.HTTPClient = h
	}
	return c, nil
}
