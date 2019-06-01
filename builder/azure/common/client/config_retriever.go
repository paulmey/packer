package client

// Method to resolve information about the user so that a client can be
// constructed to communicated with Azure.
//
// The following data are resolved.
//
// 1. TenantID

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func getSubscriptionFromIMDS() (string, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "http://169.254.169.254/metadata/instance/compute", nil)
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api-version", "2017-08-01")

	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	result := map[string]string{}
	err = json.Unmarshal(resp_body, &result)
	if err != nil {
		return "", err
	}

	return result["subscriptionId"], nil
}
