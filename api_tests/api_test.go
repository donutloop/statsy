package api_tests

import (
	"bytes"
	"github.com/donutloop/statsy/internal/api"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"
)

func TestValidFlow(t *testing.T) {

	os.Setenv("SERVICE_ENV_FILE", "../services.env")

	a := api.NewAPI(true)
	a.Bootstrap()
	a.Start()
	defer a.Stop()

	jsonData := `{"customerID":1,"tagID":2,"userID":"aaaaaaaa-bbbb-cccc-1111-222222222222","remoteIP":"123.234.56.78","timestamp":1500000000}`

	resp, err := http.Post(a.Server.TestURL+"/customer/stats", "application/json", bytes.NewReader([]byte(jsonData)))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := httputil.DumpResponse(resp, true)
		if err != nil {
			t.Fatal(err)
		}

		t.Fatal(string(respBody))
	}

	resp, err = http.Get(a.Server.TestURL + "/customer/stats/1/day/1500000000")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal("response code is bad", resp.StatusCode)
	}

	respBody, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(respBody))
}

func TestBadUserAgentFlow(t *testing.T) {

	os.Setenv("SERVICE_ENV_FILE", "../services.env")

	a := api.NewAPI(true)
	a.Bootstrap()
	a.Start()
	defer a.Stop()

	jsonData := `{"customerID":1,"tagID":2,"userID":"aaaaaaaa-bbbb-cccc-1111-222222222222","remoteIP":"123.234.56.78","timestamp":1500000000}`

	req, err := http.NewRequest(http.MethodPost, a.Server.TestURL+"/customer/stats", bytes.NewReader([]byte(jsonData)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("User-Agent", "Googlebot-News")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("response code is bad", resp.StatusCode)
	}
}

func TestBadRemoteIPFlow(t *testing.T) {

	os.Setenv("SERVICE_ENV_FILE", "../services.env")

	a := api.NewAPI(true)
	a.Bootstrap()
	a.Start()
	defer a.Stop()

	jsonData := `{"customerID":1,"tagID":2,"userID":"aaaaaaaa-bbbb-cccc-1111-222222222222","remoteIP":"213.070.64.33","timestamp":1500000000}`

	req, err := http.NewRequest(http.MethodPost, a.Server.TestURL+"/customer/stats", bytes.NewReader([]byte(jsonData)))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("response code is bad", resp.StatusCode)
	}
}

func TestCustomerInactiveFlow(t *testing.T) {

	os.Setenv("SERVICE_ENV_FILE", "../services.env")

	a := api.NewAPI(true)
	a.Bootstrap()
	a.Start()
	defer a.Stop()

	jsonData := `{"customerID":3,"tagID":2,"userID":"aaaaaaaa-bbbb-cccc-1111-222222222222","remoteIP":"213.080.64.33","timestamp":1500000000}`

	req, err := http.NewRequest(http.MethodPost, a.Server.TestURL+"/customer/stats", bytes.NewReader([]byte(jsonData)))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("response code is bad", resp.StatusCode)
	}
}
