package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/go-autorest/autorest/adal"
)

type AutoRefreshToken struct {
	adal.Token

	OauthConfig   adal.OAuthConfig
	AutoRefresh   bool
	RefreshWithin time.Duration
	Sender        adal.Sender
}

func NewAutoRefreshToken(token adal.Token, oauthConfig adal.OAuthConfig) AutoRefreshToken {
	return AutoRefreshToken{
		Token:         token,
		OauthConfig:   oauthConfig,
		AutoRefresh:   true,
		RefreshWithin: 5 * time.Minute,
		Sender:        http.DefaultClient,
	}
}

func (t *AutoRefreshToken) Refresh() error {
	return t.refresh(t.Token.Resource)
}

// RefreshExchange refreshes the token, but for a different resource.
func (t *AutoRefreshToken) RefreshExchange(resource string) error {
	return t.refresh(resource)
}

// EnsureFresh will refresh the token if it will expire within the refresh window (as set by
// RefreshWithin) and autoRefresh flag is on.
func (t *AutoRefreshToken) EnsureFresh() error {
	if t.AutoRefresh && t.WillExpireIn(t.RefreshWithin) {
		return t.Refresh()
	}
	return nil
}

func (t *AutoRefreshToken) refresh(resource string) error {
	if t.RefreshToken == "" {
		return fmt.Errorf("Token does not have a refresh token value, cannot refresh")
	}
	v := url.Values{}
	v.Set("resource", resource)
	v.Set("refresh_token", t.RefreshToken)
	v.Set("grant_type", "refresh_token")
	s := v.Encode()
	body := ioutil.NopCloser(strings.NewReader(s))
	req, err := http.NewRequest(http.MethodPost, t.OauthConfig.TokenEndpoint.String(), body)
	if err != nil {
		return fmt.Errorf("adal: Failed to build the refresh request. Error = '%v'", err)
	}

	req.ContentLength = int64(len(s))
	req.Header.Set(contentType, mimeTypeFormPost)
	log.Debug("Requesting new token")
	resp, err := t.Sender.Do(req)
	if err != nil {
		return fmt.Errorf("adal: Failed to execute the refresh request. Error = '%v'", err)
	}

	log.WithField("Status", resp.Status).Debug("Token requested")

	defer resp.Body.Close()
	rb, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return fmt.Errorf("adal: Refresh request failed. Status Code = '%d'. Failed reading response body: %s", resp.StatusCode, string(rb))
		}
		return fmt.Errorf("adal: Refresh request failed. Status Code = '%d'. Response body: %s", resp.StatusCode, string(rb))
	}

	if err != nil {
		return fmt.Errorf("adal: Failed to read a new service principal token during refresh. Error = '%v'", err)
	}
	if len(strings.Trim(string(rb), " ")) == 0 {
		return fmt.Errorf("adal: Empty service principal token received during refresh")
	}
	var token adal.Token
	err = json.Unmarshal(rb, &token)
	if err != nil {
		return fmt.Errorf("adal: Failed to unmarshal the service principal token during refresh. Error = '%v' JSON = '%s'", err, string(rb))
	}

	t.Token = token

	return nil
}

const (
	contentType      = "Content-Type"
	mimeTypeFormPost = "application/x-www-form-urlencoded"
)
