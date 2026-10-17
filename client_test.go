package calcom_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/psyb0t/go-calcom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAuthorization = "Bearer EXAMPLE-DO-NOT-USE"
	testAPIVersion    = "2024-06-11"
	testScheduleName  = "Office hours"
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *calcom.ClientWithResponses {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	auth := calcom.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", testAuthorization)

			return nil
		},
	)

	client, err := calcom.NewClientWithResponses(server.URL, auth)
	require.NoError(t, err)

	return client
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	_, err := io.WriteString(w, body)
	assert.NoError(t, err)
}

func TestGetMe(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/me", r.URL.Path)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
		writeTestJSON(t, w, `{"status":"success","data":{
			"id":7,"username":"example","email":"example@example.com",
			"name":null,"avatarUrl":null,"bio":null,"timeFormat":24,
			"defaultScheduleId":null,"weekStart":"Monday",
			"timeZone":"UTC","organizationId":null
		}}`)
	})

	res, err := client.MeControllerGetMeWithResponse(
		context.Background(),
		&calcom.MeControllerGetMeParams{},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	assert.Equal(t, float64(7), res.JSON200.Data.Id)
	assert.Equal(t, "example", res.JSON200.Data.Username)
	assert.Nil(t, res.JSON200.Data.Name)
	assert.Nil(t, res.JSON200.Data.DefaultScheduleId)
	assert.Nil(t, res.JSON200.Data.OrganizationId)
}

func TestGetSchedules(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/schedules", r.URL.Path)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
		assert.Equal(t, testAPIVersion, r.Header.Get("cal-api-version"))
		writeTestJSON(t, w, `{"status":"success","data":[{
			"id":42,"ownerId":7,"name":"Office hours",
			"timeZone":"UTC","availability":[],
			"isDefault":true,"overrides":[]
		}]}`)
	})

	params := &calcom.SchedulesController20240611GetSchedulesParams{
		CalApiVersion: testAPIVersion,
	}
	res, err := client.SchedulesController20240611GetSchedulesWithResponse(
		context.Background(),
		params,
	)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	require.Len(t, res.JSON200.Data, 1)
	assert.Equal(t, float64(42), res.JSON200.Data[0].Id)
	assert.Equal(t, testScheduleName, res.JSON200.Data[0].Name)
	assert.True(t, res.JSON200.Data[0].IsDefault)
}

func TestCreateSchedule(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/schedules", r.URL.Path)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
		assert.Equal(t, testAPIVersion, r.Header.Get("cal-api-version"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body calcom.CreateScheduleInput20240611

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, testScheduleName, body.Name)
		assert.Equal(t, "UTC", body.TimeZone)
		assert.True(t, body.IsDefault)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, `{"status":"success","data":{
			"id":42,"ownerId":7,"name":"Office hours","timeZone":"UTC",
			"availability":[],"isDefault":true,"overrides":[]
		}}`)
	})

	params := &calcom.SchedulesController20240611CreateScheduleParams{
		CalApiVersion: testAPIVersion,
	}
	body := calcom.CreateScheduleInput20240611{
		Name:      testScheduleName,
		TimeZone:  "UTC",
		IsDefault: true,
	}
	res, err := client.SchedulesController20240611CreateScheduleWithResponse(
		context.Background(),
		params,
		body,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode())
	require.NotNil(t, res.JSON201)
	assert.Equal(t, float64(42), res.JSON201.Data.Id)
	assert.Equal(t, testScheduleName, res.JSON201.Data.Name)
}

func TestUnauthorizedResponse(t *testing.T) {
	t.Parallel()

	const body = `{"status":"error","message":"Unauthorized"}`

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/me", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		writeTestJSON(t, w, body)
	})

	res, err := client.MeControllerGetMeWithResponse(
		context.Background(),
		&calcom.MeControllerGetMeParams{},
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode())
	assert.JSONEq(t, body, string(res.Body))
	assert.Nil(t, res.JSON200)
}

func TestMalformedResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeTestJSON(t, w, `{"status":`)
	})

	res, err := client.MeControllerGetMeWithResponse(
		context.Background(),
		&calcom.MeControllerGetMeParams{},
	)

	var syntaxError *json.SyntaxError

	require.ErrorAs(t, err, &syntaxError)
	assert.Nil(t, res)
}

func TestCancelledRequest(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("cancelled request reached the server")
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := client.MeControllerGetMeWithResponse(
		ctx,
		&calcom.MeControllerGetMeParams{},
	)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, res)
}

func TestReassignBooking(t *testing.T) {
	t.Parallel()

	const version = "2024-08-13"

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/bookings/booking-example/reassign", r.URL.Path)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
		assert.Equal(t, version, r.Header.Get("cal-api-version"))
		writeTestJSON(t, w, `{"status":"success","data":{
			"uid":"booking-example","reassignedTo":{"id":8},
			"extension":{"enabled":true}
		}}`)
	})
	params := &calcom.BookingsController20240813ReassignBookingParams{
		CalApiVersion: version,
	}
	res, err := client.BookingsController20240813ReassignBookingWithResponse(
		context.Background(),
		"booking-example",
		params,
	)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Equal(t, "booking-example", res.JSON200.Data["uid"])
	assert.Equal(t, map[string]any{"id": float64(8)},
		res.JSON200.Data["reassignedTo"])
	assert.Equal(t, map[string]any{"enabled": true},
		res.JSON200.Data["extension"])
}
