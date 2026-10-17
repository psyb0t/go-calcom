package calcom_test

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-calcom"
)

// This is the README usage example. It requires real credentials.
//
//nolint:testableexamples // requires real credentials, compile check only
func Example() {
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
