package metadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Metadata struct {
	ComputeInfo `json:"compute"`
}

type ComputeInfo struct {
	SubscriptionID string `json:"subscriptionId"`
}

func Get() (md Metadata, err error) {
	fetched, err := fetcher()
	if err != nil {
		return md, err
	}

	err = json.Unmarshal(fetched, &md)
	if err != nil {
		return Metadata{}, err
	}

	return md, nil
}

//curl -H Metadata:true "http://169.254.169.254/metadata/instance?api-version=2017-08-01"
var fetcher = func() ([]byte, error) {
	c := http.Client{}
	c.Transport = &http.Transport{} // force new transport to avoid proxies

	req, err := http.NewRequest(http.MethodGet, "http://169.254.169.254/metadata/instance?api-version=2017-08-01", nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating HTTP request for metadata service: %v", err)
	}
	req.Header.Set("Metadata", "true")

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error executing HTTP request for metadata service: %v", err)
	}
	defer res.Body.Close()

	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading data from metadata service: %v", err)
	}

	return d, nil
}
