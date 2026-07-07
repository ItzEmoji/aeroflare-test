// Package auth implements GitHub's OAuth device authorization flow
// (https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#device-flow)
// so that aeroflare can obtain a GitHub access token without a client secret
// or a browser redirect, and resolves credentials for GitHub, GitLab, and
// generic OCI registries from flags, environment variables, or the secrets
// manager.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// githubBaseURL is the root of GitHub's web/OAuth endpoints. It is a var
// (rather than a const) so tests can point it at a local test server.
var githubBaseURL = "https://github.com"

// DeviceCodeRequest is the body sent to POST /login/device/code to start the
// device authorization flow.
type DeviceCodeRequest struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
}

// DeviceCodeResponse is GitHub's reply to a device code request: a code for
// the device to poll with, a short code for the user to enter, the URL where
// they enter it, and the minimum polling interval in seconds.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
}

// RequestDeviceCode starts the GitHub OAuth device authorization flow for
// the given OAuth app client ID, returning the device/user codes the caller
// needs to complete authorization and then poll for a token.
func RequestDeviceCode(clientID string) (*DeviceCodeResponse, error) {
	return requestDeviceCode(clientID, githubBaseURL)
}

// requestDeviceCode is the baseURL-parameterized implementation behind
// RequestDeviceCode, allowing tests to substitute a fake GitHub server.
func requestDeviceCode(clientID, baseURL string) (*DeviceCodeResponse, error) {
	reqBodyBytes, err := json.Marshal(DeviceCodeRequest{
		ClientID: clientID,
		Scope:    "repo workflow write:packages read:packages",
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/login/device/code", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("unexpected status code: " + resp.Status + " body: " + string(body))
	}

	var result DeviceCodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TokenRequest is the body sent to POST /login/oauth/access_token to
// exchange a device code for an access token.
type TokenRequest struct {
	ClientID   string `json:"client_id"`
	DeviceCode string `json:"device_code"`
	GrantType  string `json:"grant_type"`
}

// TokenResponse is GitHub's reply to a token poll: either a non-empty
// AccessToken on success, or an Error code such as "authorization_pending"
// or "slow_down" while the user has not yet approved the request.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

// PollAccessToken repeatedly polls GitHub for an access token corresponding
// to deviceCode, honoring the server-specified interval (in seconds) between
// requests, until the user approves the request, an error occurs, or 15
// minutes elapse.
func PollAccessToken(clientID string, deviceCode string, interval int) (string, error) {
	return pollAccessToken(clientID, deviceCode, interval, githubBaseURL, time.Second)
}

// pollAccessToken is the baseURL/pollTimeUnit-parameterized implementation
// behind PollAccessToken. Tests substitute a fake server for baseURL and a
// smaller pollTimeUnit (e.g. time.Millisecond) so polling loops run quickly.
func pollAccessToken(clientID string, deviceCode string, interval int, baseURL string, pollTimeUnit time.Duration) (string, error) {
	if interval <= 0 {
		interval = 5
	}
	ticker := time.NewTicker(time.Duration(interval) * pollTimeUnit)
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	reqBodyBytes, err := json.Marshal(TokenRequest{
		ClientID:   clientID,
		DeviceCode: deviceCode,
		GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
	})
	if err != nil {
		return "", err
	}

	for {
		select {
		case <-ctx.Done():
			return "", errors.New("polling timed out")
		case <-ticker.C:
			reqCtx, reqCancel := context.WithTimeout(ctx, 10*time.Second)
			req, err := http.NewRequestWithContext(reqCtx, "POST", baseURL+"/login/oauth/access_token", bytes.NewBuffer(reqBodyBytes))
			if err != nil {
				reqCancel()
				return "", err
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				reqCancel()
				continue // retry on network error
			}

			if resp.StatusCode >= 500 {
				_ = resp.Body.Close()
				reqCancel()
				continue // transient error, retry
			}

			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if err != nil {
				reqCancel()
				continue
			}

			var result TokenResponse
			if err := json.Unmarshal(body, &result); err != nil {
				reqCancel()
				return "", err
			}

			if result.AccessToken != "" {
				reqCancel()
				return result.AccessToken, nil
			}

			if result.Error == "authorization_pending" {
				// User hasn't approved the request in their browser yet; keep polling.
				reqCancel()
				continue
			}
			if result.Error == "slow_down" {
				// GitHub is rate-limiting us; back off by widening the poll interval
				// as instructed by the OAuth device flow spec.
				interval += 5
				ticker.Reset(time.Duration(interval) * pollTimeUnit)
				reqCancel()
				continue
			}

			if result.Error != "" {
				reqCancel()
				return "", errors.New(result.Error)
			}

			reqCancel()
			return "", errors.New("invalid response from server")
		}
	}
}
