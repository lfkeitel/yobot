package librenms

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const authTokenHeader = "X-Auth-Token"

type Client struct {
	endpoint  *url.URL
	authToken string
	c         *http.Client
}

func NewClient(endpoint string) (*Client, error) {
	eurl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		endpoint: eurl,
		c:        &http.Client{},
	}, nil
}

func (c *Client) SkipTLSVerify() {
	c.c.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

type BaseAPI struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type System struct {
	LocalVer    string `json:"local_ver"`
	LocalSha    string `json:"local_sha"`
	LocalDate   string `json:"local_date"`
	LocalBranch string `json:"local_branch"`
	DBSchema    string `json:"db_schema"`
	PhpVer      string `json:"php_ver"`
	MySQLVer    string `json:"mysql_ver"`
	RrdtoolVer  string `json:"rrdtool_ver"`
	NetsnmpVer  string `json:"netsnmp_ver"`
}

type SystemAPI struct {
	BaseAPI
	System []System `json:"system"`
}

func (c *Client) Login(token string) error {
	c.authToken = token
	_, err := c.System()
	return err
}

func (c *Client) makeURL(endpoint string) *url.URL {
	eurl := copyURL(c.endpoint)
	eurl.Path = fmt.Sprintf("%s/api/v0%s", eurl.Path, endpoint)
	return eurl
}

func (c *Client) makeReq(method string, url *url.URL) *http.Request {
	r := &http.Request{
		Method: method,
		URL:    url,
		Header: make(http.Header),
	}
	r.Header.Add(authTokenHeader, c.authToken)
	return r
}

func (c *Client) doReq(r *http.Request) (*http.Response, error) {
	resp, err := c.c.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return nil, errors.New("unauthorized API token")
	}

	return resp, nil
}

func (c *Client) System() (*System, error) {
	r := c.makeReq(http.MethodGet, c.makeURL("/system"))
	resp, err := c.doReq(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data SystemAPI
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return &(data.System[0]), nil
}

type Device struct {
	ID          int    `json:"device_id"`
	Hostname    string `json:"hostname"`
	SysName     string `json:"sysName"`
	IP          string `json:"ip"`
	SnmpDisable int    `json:"snmp_disable"`
	SysContact  string `json:"sysContact"`
	Version     string `json:"version"`
	Hardware    string `json:"hardware"`
	OS          string `json:"os"`
	Status      bool   `json:"status"`
	Ignore      int    `json:"ignore"`
	Disabled    int    `json:"disabled"`
	Uptime      int64  `json:"uptime"`
	Systype     string `json:"type"`
	OSGroup     string `json:"os_group"`
}

type DeviceAPI struct {
	BaseAPI
	Devices []Device `json:"devices"`
}

func (c *Client) GetDevice(name string) (*Device, error) {
	r := c.makeReq(http.MethodGet, c.makeURL("/devices/"+url.PathEscape(name)))
	resp, err := c.doReq(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data DeviceAPI
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	// The API returns an almost empty device even if there's no match.
	// The safest is to make sure the ID is not zero.
	if len(data.Devices) > 0 && data.Devices[0].ID > 0 {
		return &data.Devices[0], nil
	}
	return nil, nil
}

func copyURL(u *url.URL) *url.URL {
	uu, _ := url.Parse(u.String())
	return uu
}
