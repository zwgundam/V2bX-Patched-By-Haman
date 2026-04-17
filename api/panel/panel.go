package panel

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MoeclubM/V2bX/conf"
	"github.com/go-resty/resty/v2"
)

// Panel is the interface for different panel's api.

type Client struct {
	client           *resty.Client
	APIHost          string
	APISendIP        string
	Token            string
	MachineID        int
	NodeType         string
	NodeId           int
	useV2API         bool
	serverPathPrefix string
	nodeEtag         string
	userEtag         string
	responseBodyHash string
	UserList         *UserListBody
	AliveMap         *AliveMap
}

func New(c *conf.ApiConfig) (*Client, error) {
	var client *resty.Client
	if c.APISendIP != "" {
		client = resty.NewWithLocalAddr(&net.TCPAddr{
			IP: net.ParseIP(c.APISendIP),
		})
	} else {
		client = resty.New()
	}
	client.SetRetryCount(3)
	if c.Timeout > 0 {
		client.SetTimeout(time.Duration(c.Timeout) * time.Second)
	} else {
		client.SetTimeout(5 * time.Second)
	}
	client.OnError(func(req *resty.Request, err error) {
		var v *resty.ResponseError
		if errors.As(err, &v) {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			logrus.Error(v.Err)
		}
	})
	client.SetBaseURL(c.APIHost)
	// Check node type
	nodeType := strings.ToLower(strings.TrimSpace(c.NodeType))
	switch nodeType {
	case "v2ray":
		nodeType = "vmess"
	case
		"v2node",
		"vmess",
		"trojan",
		"shadowsocks",
		"naive",
		"hysteria",
		"hysteria2",
		"tuic",
		"anytls",
		"vless":
	default:
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}
	cli := &Client{
		client:    client,
		Token:     c.Key,
		APIHost:   c.APIHost,
		APISendIP: c.APISendIP,
		MachineID: c.MachineID,
		NodeType:  nodeType,
		NodeId:    c.NodeID,
		UserList:  &UserListBody{},
		AliveMap:  &AliveMap{},
	}
	cli.setQueryParams(nodeType)
	return cli, nil
}

func (c *Client) setQueryParams(nodeType string) {
	params := map[string]string{
		"node_id": strconv.Itoa(c.NodeId),
		"token":   c.Token,
	}
	if nodeType != "" {
		params["node_type"] = nodeType
	}
	if c.MachineID > 0 {
		params["machine_id"] = strconv.Itoa(c.MachineID)
	}
	c.client.SetQueryParams(params)
}

func (c *Client) ensureServerAPI() error {
	if c.serverPathPrefix != "" {
		return nil
	}
	r, err := c.client.R().
		ForceContentType("application/json").
		Post("/api/v2/server/handshake")
	if err != nil {
		return fmt.Errorf("request %s failed: %s", c.assembleURL("/api/v2/server/handshake"), err)
	}
	if r != nil && r.StatusCode() < 400 {
		c.useV2API = true
		c.serverPathPrefix = "/api/v2/server"
		return nil
	}
	if c.MachineID > 0 {
		if err = c.checkResponse(r, "/api/v2/server/handshake", nil); err != nil {
			return err
		}
	}
	if r != nil && r.StatusCode() == 404 {
		c.serverPathPrefix = "/api/v1/server/UniProxy"
		return nil
	}
	if err = c.checkResponse(r, "/api/v2/server/handshake", nil); err != nil {
		return err
	}
	c.serverPathPrefix = "/api/v1/server/UniProxy"
	return nil
}

func (c *Client) serverPath(path string) (string, error) {
	if err := c.ensureServerAPI(); err != nil {
		return "", err
	}
	return c.serverPathPrefix + "/" + path, nil
}
