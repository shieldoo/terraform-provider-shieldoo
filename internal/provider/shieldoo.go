package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Group struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	ObjectId string `json:"objectId"`
}

type FirewallRule struct {
	Protocol string  `json:"protocol"`
	Port     string  `json:"port"`
	Host     string  `json:"host"`
	Groups   []Group `json:"groups"`
}

type Firewall struct {
	Id       string         `json:"id"`
	Name     string         `json:"name"`
	RulesIn  []FirewallRule `json:"rulesIn"`
	RulesOut []FirewallRule `json:"rulesOut"`
}

type Listener struct {
	ListenPort  int    `json:"listenPort"`
	Protocol    string `json:"protocol"`
	ForwardPort int    `json:"forwardPort"`
	ForwardHost string `json:"forwardHost"`
	Description string `json:"description"`
}

type Server struct {
	Id             string                   `json:"id"`
	Name           string                   `json:"name"`
	Groups         []Group                  `json:"groups"`
	Firewall       Firewall                 `json:"firewall"`
	Listeners      []Listener               `json:"listeners"`
	Autoupdate     bool                     `json:"autoupdate"`
	IpAddress      string                   `json:"ipAddress"`
	Description    string                   `json:"description"`
	Configuration  string                   `json:"configuration"`
	OSUpdatePolicy ServerOSAutoupdatePolicy `json:"osUpdatePolicy"`
}

type ServerOSAutoupdatePolicy struct {
	Enabled                   bool `json:"enabled"`
	SecurityAutoupdateEnabled bool `json:"securityAutoupdateEnabled"`
	AllAutoupdateEnabled      bool `json:"allAutoupdateEnabled"`
	RestartAfterUpdate        bool `json:"restartAfterUpdate"`
	UpdateHour                int  `json:"updateHour"`
}

type ShieldooJWTData struct {
	jwt.RegisteredClaims
	ShieldooClaims map[string]string `json:"shieldoo"`
}

type ShieldooClient struct {
	uri    string
	apiKey string
}

func (c *ShieldooClient) ListGroups() ([]Group, error) {
	data, err := c.callApi("GET", "groups", "", "", nil)
	if err != nil {
		return nil, err
	}
	var groups []Group
	err = json.Unmarshal([]byte(data), &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (c *ShieldooClient) GetServer(name string) (*Server, error) {
	if c.uri == "https://mockup" && c.apiKey == "mockup" {
		return &Server{
			Id:            "mockup",
			Name:          name,
			Configuration: "mockup",
			IpAddress:     "mockup",
		}, nil
	}
	data, err := c.callApi("GET", "servers", name, "", nil)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(data, "[") {
		data = strings.TrimPrefix(data, "[")
		data = strings.TrimSuffix(data, "]")
	}
	var server Server
	err = json.Unmarshal([]byte(data), &server)
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (c *ShieldooClient) DeleteServer(id string) error {
	_, err := c.callApi("DELETE", "servers", "", id, nil)
	return err
}

func (c *ShieldooClient) CreateServer(server *Server) (*Server, error) {
	data, err := c.callApi("POST", "servers", "", "", server)
	if err != nil {
		return nil, err
	}
	var newServer Server
	err = json.Unmarshal([]byte(data), &newServer)
	if err != nil {
		return nil, err
	}
	return &newServer, nil
}

func (c *ShieldooClient) UpdateServer(server *Server) (*Server, error) {
	data, err := c.callApi("PUT", "servers", "", server.Id, server)
	if err != nil {
		return nil, err
	}
	var newServer Server
	err = json.Unmarshal([]byte(data), &newServer)
	if err != nil {
		return nil, err
	}
	return &newServer, nil
}

func (c *ShieldooClient) GetFirewall(name string) (*Firewall, error) {
	if c.uri == "https://mockup" && c.apiKey == "mockup" {
		return &Firewall{
			Id:   "mockup",
			Name: name,
		}, nil
	}
	data, err := c.callApi("GET", "firewalls", name, "", nil)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(data, "[") {
		data = strings.TrimPrefix(data, "[")
		data = strings.TrimSuffix(data, "]")
	}
	var firewall Firewall
	err = json.Unmarshal([]byte(data), &firewall)
	if err != nil {
		return nil, err
	}
	return &firewall, nil
}

func (c *ShieldooClient) DeleteFirewall(id string) error {
	if c.uri == "https://mockup" && c.apiKey == "mockup" {
		return nil
	}
	_, err := c.callApi("DELETE", "firewalls", "", id, nil)
	return err
}

func (c *ShieldooClient) CreateFirewall(firewall *Firewall) (*Firewall, error) {
	if c.uri == "https://mockup" && c.apiKey == "mockup" {
		return &Firewall{
			Id:   "mockup",
			Name: firewall.Name,
		}, nil
	}
	data, err := c.callApi("POST", "firewalls", "", "", firewall)
	if err != nil {
		return nil, err
	}
	var ret Firewall
	err = json.Unmarshal([]byte(data), &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *ShieldooClient) UpdateFirewall(firewall *Firewall) (*Firewall, error) {
	if c.uri == "https://mockup" && c.apiKey == "mockup" {
		return &Firewall{
			Id:   "mockup",
			Name: firewall.Name,
		}, nil
	}
	data, err := c.callApi("PUT", "firewalls", "", firewall.Id, firewall)
	if err != nil {
		return nil, err
	}
	var ret Firewall
	err = json.Unmarshal([]byte(data), &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (c *ShieldooClient) generateJWTAccessToken() (string, error) {
	instance := c.shieldooExtractDomainFromUri()
	// prepare claims for token
	claims := ShieldooJWTData{
		RegisteredClaims: jwt.RegisteredClaims{
			// set token lifetime in timestamp
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
		},
		ShieldooClaims: map[string]string{
			"instance": instance,
		},
	}

	// generate a string using claims and HS256 algorithm
	tokenString := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// sign the generated key using secretKey
	token, err := tokenString.SignedString([]byte(c.apiKey))

	return token, err
}

func (c *ShieldooClient) shieldooExtractDomainFromUri() string {
	// extract domain from uri
	parsedURL, err := url.Parse(c.uri)
	ret := ""
	if err == nil {
		ret = parsedURL.Hostname()
	}
	return ret
}

func (c *ShieldooClient) callApi(method string, entity string, name string, id string, data interface{}) (string, error) {
	// create Jwt token
	token, err := c.generateJWTAccessToken()
	if err != nil {
		return "", err
	}
	// call REST API
	httpClient := &http.Client{}
	myurl := c.uri + "/cliapi/" + entity
	if id != "" {
		myurl += "/" + url.QueryEscape(id)
	}
	if name != "" {
		// url encode name
		myurl += "?name=" + url.QueryEscape(name)
	}
	var buff *bytes.Buffer
	// convert data to json if it is not nil
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		buff = bytes.NewBuffer(jsonData)
	} else {
		buff = &bytes.Buffer{}
	}
	req, err := http.NewRequest(method, myurl, buff)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthToken", token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return string(body), errors.New(resp.Status)
	}
	return strings.TrimSpace(string(body)), nil
}
