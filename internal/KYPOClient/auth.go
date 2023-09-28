package KYPOClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

func (c *Client) signIn() (string, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}

	httpClient := http.Client{Jar: jar}

	csrf, err := c.authorize(httpClient)
	if err != nil {
		return "", err
	}

	token, csrf, err := c.login(httpClient, csrf)
	if err != nil {
		return "", err
	}

	if token != "" {
		return token, err
	}

	return c.authorizeFirstTime(httpClient, csrf)
}

func (c *Client) authorize(httpClient http.Client) (string, error) {
	query := url.Values{}
	query.Add("response_type", "id_token token")
	query.Add("client_id", c.ClientID)
	query.Add("scope", "openid email profile")
	query.Add("redirect_uri", c.Endpoint)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/csirtmu-dummy-issuer-server/authorize?%s",
		c.Endpoint, query.Encode()), nil)
	if err != nil {
		return "", err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authorize failed, got HTTP code: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = res.Body.Close()
	if err != nil {
		return "", err
	}

	csrf, err := extractCsrf(string(body))
	if err != nil {
		return "", err
	}

	return csrf, nil
}

func extractCsrf(body string) (string, error) {
	csrfRegex := regexp.MustCompile("<input type=\"hidden\" name=\"_csrf\" value=\"([^\"]+)\" */>")
	matches := csrfRegex.FindStringSubmatch(body)
	if len(matches) != 2 {
		return "", errors.New("failed to match csrf token")
	}
	return matches[1], nil
}

func (c *Client) login(httpClient http.Client, csrf string) (string, string, error) {
	query := url.Values{}
	query.Add("username", c.Username)
	query.Add("password", c.Password)
	query.Add("_csrf", csrf)
	query.Add("submit", "Login")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/csirtmu-dummy-issuer-server/login",
		c.Endpoint), strings.NewReader(query.Encode()))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("login failed, got HTTP code: %d", res.StatusCode)
	}

	values, err := url.ParseQuery(res.Request.URL.Fragment)
	if err != nil {
		return "", "", err
	}

	token := values.Get("access_token")

	if token != "" {
		return token, "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	csrf, err = extractCsrf(string(body))
	if err != nil {
		return "", "", err
	}

	return "", csrf, nil
}

func (c *Client) authorizeFirstTime(httpClient http.Client, csrf string) (string, error) {
	query := url.Values{}
	query.Add("scope_openid", "openid")
	query.Add("scope_profile", "profile")
	query.Add("scope_email", "email")
	query.Add("remember", "until-revoked")
	query.Add("user_oauth_approval", "true")
	query.Add("authorize", "Authorize")
	query.Add("_csrf", csrf)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/csirtmu-dummy-issuer-server/authorize",
		c.Endpoint), strings.NewReader(query.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authorizeFirstTime failed, got HTTP code: %d", res.StatusCode)
	}

	values, err := url.ParseQuery(res.Request.URL.Fragment)
	if err != nil {
		return "", err
	}

	token := values.Get("access_token")
	if token == "" {
		return "", fmt.Errorf("authorizeFirstTime failed, token is empty")
	}
	return token, err
}

func (c *Client) authenticateKeycloak() (string, error) {
	query := url.Values{}
	query.Add("username", c.Username)
	query.Add("password", c.Password)
	query.Add("client_id", c.ClientID)
	query.Add("grant_type", "password")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/keycloak/realms/KYPO/protocol/openid-connect/token",
		c.Endpoint), strings.NewReader(query.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authenticateKeycloak failed, got HTTP code: %d", res.StatusCode)
	}

	accessToken := struct {
		AccessToken string `json:"access_token"`
	}{}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &accessToken)
	if err != nil {
		return "", err
	}

	return accessToken.AccessToken, nil
}
