package bce

import (
	"context"

	bcesdk "github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/sms/api"
	"go.gh.ink/smsutils/v3/errors"
	"go.gh.ink/smsutils/v3/httpclient"
	"go.gh.ink/smsutils/v3/model"
	"go.gh.ink/smsutils/v3/utils"
)

// SendMessage retains the original API and delegates to the context-aware path.
func (c Client) SendMessage(dest string, sender string, template string, vars model.Vars) error {
	return c.SendMessageCtx(context.Background(), dest, sender, template, vars)
}

// SendMessageCtx binds cancellation to the actual provider HTTP request.
func (c Client) SendMessageCtx(ctx context.Context, dest, sender, template string, vars model.Vars) (err error) {
	if ctx == nil {
		return errors.ErrMessageInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	defer func() {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	if err := utils.ValidateVars(vars, true); err != nil {
		return err
	}

	// Try to parse number
	dest, _, _, _, err = utils.ProcessNumberForChinese(dest)
	if err != nil {
		return err
	}

	// Preprocess vars
	params := make(map[string]any)
	for _, v := range vars {
		params[v.Key] = v.Value
	}

	// Construct args
	sendSmsArgs := &api.SendSmsArgs{
		Mobile:      dest,
		Template:    template,
		SignatureId: sender,
		ContentVar:  params,
	}

	// Send request
	if c.Client == nil || c.Client.BceClient == nil || c.Client.Config == nil {
		return errors.ErrDriverSendFailed.WithDriverName(Name)
	}
	sdk := *c.Client
	base := *sdk.BceClient
	config := *base.Config
	config.Retry = bcesdk.NewBackOffRetryPolicy(0, 0, 0)
	base.Config = &config
	base.HTTPClient = httpclient.BindClient(ctx, base.HTTPClient)
	sdk.BceClient = &base
	result, err := sdk.SendSms(sendSmsArgs)
	if err != nil {
		return err
	}
	if result != nil {
		// Success
		if result.Code == "1000" {
			return nil
		}

		return errors.ErrDriverSendFailed.
			WithDriverName(Name).
			WithDriverCode(result.Code).
			WithDriverMessage(result.Message).
			WithDriverRequestID(result.RequestId).
			WithDriverResponse(result)
	}
	return errors.ErrDriverSendFailed.WithDriverName(Name)
}
