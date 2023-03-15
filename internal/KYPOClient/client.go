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

type Client struct {
	Endpoint   string
	ClientID   string
	HTTPClient *http.Client
	Token      string
	Username   string
	Password   string
}

func NewClientWithToken(endpoint, clientId, token string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		ClientID:   clientId,
		HTTPClient: http.DefaultClient,
		Token:      token,
	}

	return &client, nil
}

func NewClient(endpoint, clientId, username, password string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		ClientID:   clientId,
		HTTPClient: http.DefaultClient,
		Username:   username,
		Password:   password,
	}
	token, err := client.signIn()
	if err != nil {
		return nil, err
	}
	client.Token = token
	return &client, nil
}

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

var ErrNotFound = errors.New("not found")

func (c *Client) doRequest(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}

	return body, res.StatusCode, nil
}

type Definition struct {
	Id        int64     `json:"id" tfsdk:"id"`
	Url       string    `json:"url" tfsdk:"url"`
	Name      string    `json:"name" tfsdk:"name"`
	Rev       string    `json:"rev" tfsdk:"rev"`
	CreatedBy UserModel `json:"created_by" tfsdk:"created_by"`
}

type UserModel struct {
	Id         int64  `json:"id" tfsdk:"id"`
	Sub        string `json:"sub" tfsdk:"sub"`
	FullName   string `json:"full_name" tfsdk:"full_name"`
	GivenName  string `json:"given_name" tfsdk:"given_name"`
	FamilyName string `json:"family_name" tfsdk:"family_name"`
	Mail       string `json:"mail" tfsdk:"mail"`
}

type DefinitionRequest struct {
	Url string `json:"url"`
	Rev string `json:"rev"`
}

func (c *Client) GetDefinition(definitionID int64) (*Definition, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	definition := Definition{}

	if status == http.StatusNotFound {
		return nil, fmt.Errorf("definition %d %w", definitionID, ErrNotFound)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}

func (c *Client) CreateDefinition(url, rev string) (*Definition, error) {
	requestBody, err := json.Marshal(DefinitionRequest{url, rev})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/definitions", c.Endpoint), strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if status != http.StatusCreated {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	definition := Definition{}
	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}

func (c *Client) DeleteDefinition(definitionID int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if status != http.StatusNoContent && status != http.StatusNotFound {
		return fmt.Errorf("status: %d, body: %s", status, body)
	}

	return nil
}
