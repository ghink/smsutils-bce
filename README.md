# smsutils-bce

Baidu Cloud Engine (BCE) SMS driver for [smsutils](https://go.gh.ink/smsutils).

This module implements the `model.Driver` / `model.Client` interfaces of the core
`smsutils` library on top of the Baidu Cloud `bce-sdk-go` SMS service.

Driver name: `bce`

## Installation

```bash
go get go.gh.ink/smsutils/bce/v3
```

## Usage

Blank-import this package so it registers itself, then configure it via the core client.

```go
import (
	"go.gh.ink/smsutils/v3/client"
	"go.gh.ink/smsutils/v3/model"

	_ "go.gh.ink/smsutils/bce/v3"
)

clients, err := client.NewClient(model.Config{
	Credentials: model.C{
		"bce": {
			"accessKeyID":     "<your-access-key-id>",
			"secretAccessKey": "<your-secret-access-key>",
			// Optional override:
			// "endpoint": "sms.bj.baidubce.com",
		},
	},
})
if err != nil {
	panic(err)
}

// "sender" is the BCE SignatureId; "smsTpl..." is the template code.
err = clients["bce"].SendMessage("+8617601205205", "<signature-id>", "smsTpl:abc123", model.Vars{
	{Key: "code", Value: "1234"},
})
```

## Context-aware sending

The existing `client.NewClient` and `SendMessage` APIs remain available. With the
updated local core and this driver, use the same configuration with the new factory:

```go
// Additional imports: "context", "time".
clients, err := client.NewClientCtx(config, model.ClientOptions{Timeout: 5 * time.Second})
if err != nil {
    return err
}
ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
defer cancel()
err = clients["bce"].SendMessageCtx(ctx, dest, sender, template, vars)
```

`config` has the credential layout shown above. `ClientOptions.Transport` optionally
injects a concurrency-safe HTTP transport. Context cancellation reaches the HTTP request;
a timeout does not prove the provider did not accept the SMS, so avoid blind retries.
Nil variable entries and malformed provider responses are rejected.

These additions require the updated core source. Existing dependency versions are
intentionally unchanged until release. For local testing, create a Go workspace containing
this module and the sibling `smsutils/v3` core module; see the core repository's
`CONTEXT.md` (local path: `../smsutils/CONTEXT.md`) for the complete six-module
commands and API contracts.

## Credentials

| Key | Constant | Required | Default | Description |
|-----|----------|----------|---------|-------------|
| `accessKeyID` | `AccessKeyID` | Yes | — | BCE Access Key ID |
| `secretAccessKey` | `SecretAccessKey` | Yes | — | BCE Secret Access Key |
| `endpoint` | `Endpoint` | No | `sms.bj.baidubce.com` | API endpoint |

Missing `accessKeyID` or `secretAccessKey` yields `errors.ErrDriverCredentialInvalid`.

## Behavior

- The destination is normalized via `utils.ProcessNumberForChinese` before sending.
- `template` maps to the BCE `Template`; `sender` maps to `SignatureId`.
- `vars` are passed **by key** as `ContentVar` (`map[string]any`), so each `Var.Key` must
  match a placeholder name in your template.
- A BCE result `Code` of `1000` indicates success. Any other code produces
  `errors.ErrDriverSendFailed`, decorated with the provider code, message, request ID and raw
  result.

## Requirements

- Go 1.25.0+
- `go.gh.ink/smsutils/v3`
- Baidu Cloud SDK: `github.com/baidubce/bce-sdk-go`

## License

[Apache License 2.0](LICENSE)
