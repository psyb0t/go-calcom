# go-calcom

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-calcom.svg)](https://pkg.go.dev/github.com/psyb0t/go-calcom)
[![CI](https://github.com/psyb0t/go-calcom/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-calcom/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-calcom/badges/version.svg)](https://github.com/psyb0t/go-calcom/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-calcom/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-calcom/badges/importers.svg)](https://github.com/psyb0t/go-calcom/blob/badges/importers.md)

Unofficial Go client for [Cal.com API v2](https://cal.com/docs/api-reference/v2/introduction), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the public OpenAPI document in the Cal.com repository. I am not affiliated with Cal.com.

The package is `calcom`, in the repository root. Everything in `client.gen.go` is generated. Don't edit it by hand, regenerate it. The supplied spec covers 82 paths and 121 operations for bookings, event types, available slots, schedules, calendars, teams, webhooks and profile. Its title is `Cal.diy API v2`, version `1.0.0`.

## Install

```sh
go get github.com/psyb0t/go-calcom
```

## Usage

Every endpoint in the spec has a typed method on `ClientWithResponses`. Responses decode into `JSON200` and the other status fields.

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-calcom"
)

func main() {
	token := "Bearer " + os.Getenv("CALCOM_API_KEY")
	auth := calcom.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", token)

			return nil
		},
	)

	client, err := calcom.NewClientWithResponses("https://api.cal.com", auth)
	if err != nil {
		panic(err)
	}

	res, err := client.MeControllerGetMeWithResponse(
		context.Background(),
		&calcom.MeControllerGetMeParams{Authorization: token},
	)
	if err != nil {
		panic(err)
	}

	if res.StatusCode() != http.StatusOK || res.JSON200 == nil {
		panic(res.Status())
	}

	_ = res.JSON200.Data.Username
}
```

The same code is in `example_test.go`, so `make test` checks that it compiles. It requires real credentials and is not run by the test suite.

The spec declares no server URL, so pass the host yourself: `https://api.cal.com`, as shown in the [profile endpoint docs](https://cal.com/docs/api-reference/v2/me/get-my-profile), or another deployment. Generated paths already contain `/v2`, so do not include it in the base URL.

Method names come from the spec's operation IDs. `MeControllerGetMeWithResponse` reads the profile and `BookingsController20240813GetBookingsWithResponse` lists bookings. The error return covers transport and decoding failures; inspect `StatusCode()` and the parsed status field for API failures.

## Auth

Cal.com uses `Authorization: Bearer <token>` with an API key or supported access token. Read the credential from `CALCOM_API_KEY`; the request editor above attaches the header. The spec expresses authentication as operation header parameters rather than `securitySchemes`, so required `Authorization` fields must also be populated in the generated params.

Many endpoints require a `cal-api-version` header. Populate `CalApiVersion` in that operation's params: bookings use `2024-08-13`, event types `2024-06-14`, schedules `2024-06-11`, and available slots `2024-09-04`. These dates select endpoint versions. `/v2/me` has no version header in the spec. Check the generated parameter comments for each endpoint; defaults in the spec are not filled in automatically.

Some endpoints accept the deprecated Platform headers `x-cal-client-id` and `x-cal-secret-key`. API access and supported authentication depend on the endpoint and account plan. See the [provider introduction](https://cal.com/docs/api-reference/v2/introduction).

## Regenerating

The spec is not committed. `make generate` downloads the current OpenAPI document into the gitignored `api/` directory and regenerates the client:

```sh
make generate
```

`gen.go` declares the steps, so `make generate` is just `go generate ./...`:

- `scripts/fetch-spec.sh` downloads `https://raw.githubusercontent.com/calcom/cal.com/main/docs/api-reference/v2/openapi.json` to `api/openapi.json`.
- `scripts/patch-spec.sh` writes `api/openapi.patched.json`. The source reuses operation IDs for the singular and plural calendar event GET and PATCH routes. The patch gives the singular aliases distinct IDs and retains every route.
- The patch also replaces the reassignment response's self-referential `data` schema with an open object. Its description names two schemas absent from the document. Returned fields are preserved in a `map[string]any`.
- oapi-codegen runs with `config.yaml`. The config maps unformatted `number` fields to `float64` instead of `float32`.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-calcom) teaches a coding agent how to use go-calcom when it writes Go. It works in agents that read `.agents/skills/`, and installs through the clients below once listed in the marketplace.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-calcom@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-calcom@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

The release workflow is configured to publish the skill to ClawHub. After publication:

```bash
openclaw skills install @psyb0t/go-calcom
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs only Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the spec and regenerate
```

## License

[MIT](LICENSE). The generated code comes from the public specification in the Cal.com repository. The specification itself is not redistributed here and declares no license or terms URL in its metadata. The client license does not grant rights to the upstream specification or hosted service.
