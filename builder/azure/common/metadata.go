package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func (azureClient) GetComputeMetadata() (ComputeMetadata, error) {
	rv := ComputeMetadata{}

	req, err := http.NewRequest(http.MethodGet,
		"http://169.254.169.254/metadata/instance/compute?api-version=2017-08-01", nil)
	if err != nil {
		return rv, err
	}
	req.Header.Set("Metadata", "true ")

	res, err := Sender.Do(req)
	if err != nil {
		return rv, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return rv, fmt.Errorf("Unexpected status code (%d) from IMDS: %s",
			res.StatusCode, res.Status)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return rv, err
	}
	return rv, json.Unmarshal(data, &rv)
}

type ComputeMetadata struct {
	Location string `json:"location"`
	OsType   string `json:"osType"`

	SubscriptionID    string `json:"subscriptionId"`
	ResourceGroupName string `json:"resourceGroupName"`
	Name              string `json:"name"`
}
