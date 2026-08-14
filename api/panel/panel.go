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
		// legacy alias kept for compatibility — treated as vless
		nodeType = "vless"
	case
		"v2node",
		"hysteria2",
		"anytls",
		"vless":
	default:
		return nil, fmt.Errorf("unsupported Node type: %s (only vless/anytls/hysteria2 are supported in this build)", c.NodeType)
	}
	if c.MachineID <= 0 {
		return nil, fmt.Errorf("machine id is required")
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
		"machine_id": strconv.Itoa(c.MachineID),
		"token":      c.Token,
	}
	if c.NodeId > 0 {
		params["node_id"] = strconv.Itoa(c.NodeId)
	}
	if c.NodeId > 0 && nodeType != "" {
		params["node_type"] = nodeType
	}
	c.client.SetQueryParams(params)
}

func (c *Client) serverPath(path string) (string, error) {
	return "/api/v2/server/" + path, nil
}
