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
	apiVersion := strings.ToLower(strings.TrimSpace(c.APIVersion))
	if apiVersion == "" {
		if c.MachineID > 0 || nodeType == "v2node" {
			apiVersion = "v2"
		} else {
			apiVersion = "v1"
		}
	}
	switch apiVersion {
	case "v1":
		cli.serverPathPrefix = "/api/v1/server/UniProxy"
	case "v2":
		cli.useV2API = true
		cli.serverPathPrefix = "/api/v2/server"
	default:
		return nil, fmt.Errorf("unsupported api version: %s", c.APIVersion)
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

func (c *Client) serverPath(path string) (string, error) {
	return c.serverPathPrefix + "/" + path, nil
}
