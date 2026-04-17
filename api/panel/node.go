package panel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/MoeclubM/V2bX/conf"
)

// Security type
const (
	None    = 0
	Tls     = 1
	Reality = 2
)

const (
	APIVersionV1 = "v1"
	APIVersionV2 = "v2"
)

type NodeInfo struct {
	APIVersion   string
	Id           int
	Type         string
	Security     int
	PushInterval time.Duration
	PullInterval time.Duration
	RawDNS       RawDNS
	Rules        Rules

	// origin
	VAllss      *VAllssNode
	Shadowsocks *ShadowsocksNode
	Trojan      *TrojanNode
	Tuic        *TuicNode
	AnyTls      *AnyTlsNode
	Naive       *NaiveNode
	Hysteria    *HysteriaNode
	Hysteria2   *Hysteria2Node
	Common      *CommonNode
}

type CommonNode struct {
	Protocol   string           `json:"protocol"`
	ListenIP   string           `json:"listen_ip"`
	Host       string           `json:"host"`
	ServerPort int              `json:"server_port"`
	ServerName string           `json:"server_name"`
	Routes     []Route          `json:"routes"`
	BaseConfig *BaseConfig      `json:"base_config"`
	CertConfig *conf.CertConfig `json:"cert_config"`
}

type Route struct {
	Id          int         `json:"id"`
	Match       interface{} `json:"match"`
	Action      string      `json:"action"`
	ActionValue string      `json:"action_value"`
}
type BaseConfig struct {
	PushInterval any `json:"push_interval"`
	PullInterval any `json:"pull_interval"`
}

// VAllssNode is vmess and vless node info
type VAllssNode struct {
	CommonNode
	Tls                 int                   `json:"tls"`
	TlsSettings         TlsSettings           `json:"tls_settings"`
	TlsSettingsBack     *TlsSettings          `json:"tlsSettings"`
	Network             string                `json:"network"`
	NetworkSettings     json.RawMessage       `json:"network_settings"`
	NetworkSettingsBack json.RawMessage       `json:"networkSettings"`
	Encryption          string                `json:"encryption"`
	EncryptionSettings  EncSettings           `json:"encryption_settings"`
	ServerName          string                `json:"server_name"`
	Multiplex           *conf.MultiplexConfig `json:"multiplex"`

	// vless only
	Flow          string        `json:"flow"`
	RealityConfig RealityConfig `json:"-"`
}

type TlsSettings struct {
	ServerName    string       `json:"server_name"`
	AllowInsecure bool         `json:"allow_insecure"`
	ECH           *ECHSettings `json:"ech"`
	Dest          string       `json:"dest"`
	ServerPort    string       `json:"server_port"`
	ShortId       string       `json:"short_id"`
	PrivateKey    string       `json:"private_key"`
	Mldsa65Seed   string       `json:"mldsa65Seed"`
	Xver          uint64       `json:"xver,string"`
}

type ECHSettings struct {
	Enabled         bool       `json:"enabled"`
	Config          StringList `json:"config,omitempty"`
	ConfigPath      string     `json:"config_path,omitempty"`
	QueryServerName string     `json:"query_server_name,omitempty"`
	Key             StringList `json:"key,omitempty"`
	KeyPath         string     `json:"key_path,omitempty"`
}

type EncSettings struct {
	Mode          string `json:"mode"`
	Ticket        string `json:"ticket"`
	ServerPadding string `json:"server_padding"`
	PrivateKey    string `json:"private_key"`
}

type RealityConfig struct {
	Xver         uint64 `json:"Xver"`
	MinClientVer string `json:"MinClientVer"`
	MaxClientVer string `json:"MaxClientVer"`
	MaxTimeDiff  string `json:"MaxTimeDiff"`
}

type ShadowsocksNode struct {
	CommonNode
	Cipher    string `json:"cipher"`
	ServerKey string `json:"server_key"`
}

type TrojanNode struct {
	CommonNode
	Tls                 int                   `json:"tls"`
	TlsSettings         TlsSettings           `json:"tls_settings"`
	TlsSettingsBack     *TlsSettings          `json:"tlsSettings"`
	Network             string                `json:"network"`
	NetworkSettings     json.RawMessage       `json:"network_settings"`
	NetworkSettingsBack json.RawMessage       `json:"networkSettings"`
	Multiplex           *conf.MultiplexConfig `json:"multiplex"`
}

type TuicNode struct {
	CommonNode
	TlsSettings       TlsSettings `json:"tls_settings"`
	CongestionControl string      `json:"congestion_control"`
	ZeroRTTHandshake  bool        `json:"zero_rtt_handshake"`
}

type AnyTlsNode struct {
	CommonNode
	TlsSettings   TlsSettings `json:"tls_settings"`
	PaddingScheme StringList  `json:"padding_scheme,omitempty"`
}

type NaiveNode struct {
	CommonNode
	Tls             int          `json:"tls"`
	TlsSettings     TlsSettings  `json:"tls_settings"`
	TlsSettingsBack *TlsSettings `json:"tlsSettings"`
}

type HysteriaNode struct {
	CommonNode
	TlsSettings TlsSettings `json:"tls_settings"`
	UpMbps      int         `json:"up_mbps"`
	DownMbps    int         `json:"down_mbps"`
	Obfs        string      `json:"obfs"`
}

type Hysteria2Node struct {
	CommonNode
	TlsSettings             TlsSettings `json:"tls_settings"`
	Ignore_Client_Bandwidth bool        `json:"ignore_client_bandwidth"`
	UpMbps                  int         `json:"up_mbps"`
	DownMbps                int         `json:"down_mbps"`
	ObfsType                string      `json:"obfs"`
	ObfsPassword            string      `json:"obfs-password"`
}

type RawDNS struct {
	DNSMap  map[string]map[string]interface{}
	DNSJson []byte
}

type Rules struct {
	Regexp   []string
	Protocol []string
}

type nodeResponseMeta struct {
	Protocol string `json:"protocol"`
	Version  int    `json:"version"`
}

type StringList []string

func (s *StringList) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single == "" {
			*s = nil
			return nil
		}
		*s = []string{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*s = many
		return nil
	}
	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	values := make([]string, 0, len(raw))
	for i := range raw {
		item, ok := raw[i].(string)
		if !ok {
			return fmt.Errorf("item %d is not string", i)
		}
		values = append(values, item)
	}
	*s = values
	return nil
}

func normalizeResponseNodeType(requestedType string, meta nodeResponseMeta) (string, error) {
	nodeType := strings.ToLower(strings.TrimSpace(meta.Protocol))
	if nodeType == "" {
		nodeType = strings.ToLower(strings.TrimSpace(requestedType))
	}
	switch nodeType {
	case "v2ray":
		nodeType = "vmess"
	case "hysteria":
		if meta.Version == 2 {
			nodeType = "hysteria2"
		}
	case "v2node":
		nodeType = ""
	}
	if nodeType == "" {
		return "", fmt.Errorf("missing node protocol in response")
	}
	return nodeType, nil
}

func normalizeConfiguredNodeType(nodeType string) (string, error) {
	normalized, err := normalizeResponseNodeType(nodeType, nodeResponseMeta{})
	if err != nil {
		return "", fmt.Errorf("missing node protocol in config")
	}
	return normalized, nil
}

func (c *Client) GetNodeInfo() (node *NodeInfo, err error) {
	path, err := c.serverPath("config")
	if err != nil {
		return nil, err
	}
	r, err := c.client.
		R().
		SetHeader("If-None-Match", c.nodeEtag).
		ForceContentType("application/json").
		Get(path)

	if err = c.checkResponse(r, path, err); err != nil {
		return nil, err
	}
	if r.StatusCode() == 304 {
		return nil, nil
	}
	hash := sha256.Sum256(r.Body())
	newBodyHash := hex.EncodeToString(hash[:])
	if c.responseBodyHash == newBodyHash {
		return nil, nil
	}
	c.responseBodyHash = newBodyHash
	c.nodeEtag = r.Header().Get("ETag")

	defer func() {
		if r.RawBody() != nil {
			r.RawBody().Close()
		}
	}()

	if r == nil {
		return nil, fmt.Errorf("received nil response")
	}
	if c.useV2API {
		node, err = decodeV2NodeInfo(r.Body(), c.NodeId, c.NodeType)
	} else {
		node, err = decodeV1NodeInfo(r.Body(), c.NodeId, c.NodeType)
	}
	if err != nil {
		return nil, err
	}
	if c.NodeType != node.Type {
		c.NodeType = node.Type
		c.setQueryParams(node.Type)
	}

	return node, nil
}

func decodeV1NodeInfo(body []byte, nodeID int, requestedType string) (*NodeInfo, error) {
	nodeType, err := normalizeConfiguredNodeType(requestedType)
	if err != nil {
		return nil, err
	}
	node := newNodeInfo(nodeID, APIVersionV1)
	node.Type = nodeType
	return decodeNodeInfoBody(body, node)
}

func decodeV2NodeInfo(body []byte, nodeID int, requestedType string) (*NodeInfo, error) {
	meta := nodeResponseMeta{}
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, fmt.Errorf("decode node metadata error: %s", err)
	}
	nodeType, err := normalizeResponseNodeType(requestedType, meta)
	if err != nil {
		return nil, err
	}
	node := newNodeInfo(nodeID, APIVersionV2)
	node.Type = nodeType
	return decodeNodeInfoBody(body, node)
}

func newNodeInfo(nodeID int, apiVersion string) *NodeInfo {
	return &NodeInfo{
		APIVersion: apiVersion,
		Id:         nodeID,
		RawDNS: RawDNS{
			DNSMap:  make(map[string]map[string]interface{}),
			DNSJson: []byte(""),
		},
	}
}

func decodeNodeInfoBody(body []byte, node *NodeInfo) (*NodeInfo, error) {
	var cm *CommonNode
	switch node.Type {
	case "vmess", "vless":
		rsp := &VAllssNode{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		if len(rsp.NetworkSettingsBack) > 0 {
			rsp.NetworkSettings = rsp.NetworkSettingsBack
			rsp.NetworkSettingsBack = nil
		}
		if rsp.TlsSettingsBack != nil {
			rsp.TlsSettings = *rsp.TlsSettingsBack
			rsp.TlsSettingsBack = nil
		}
		cm = &rsp.CommonNode
		node.VAllss = rsp
		node.Security = node.VAllss.Tls
	case "shadowsocks":
		rsp := &ShadowsocksNode{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode shadowsocks params error: %s", err)
		}
		cm = &rsp.CommonNode
		node.Shadowsocks = rsp
		node.Security = None
	case "trojan":
		rsp := &TrojanNode{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode trojan params error: %s", err)
		}
		if len(rsp.NetworkSettingsBack) > 0 {
			rsp.NetworkSettings = rsp.NetworkSettingsBack
			rsp.NetworkSettingsBack = nil
		}
		if rsp.TlsSettingsBack != nil {
			rsp.TlsSettings = *rsp.TlsSettingsBack
			rsp.TlsSettingsBack = nil
		}
		cm = &rsp.CommonNode
		node.Trojan = rsp
		if rsp.Tls == Reality {
			node.Security = Reality
		} else {
			node.Security = Tls
		}
	case "tuic":
		rsp := &TuicNode{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode tuic params error: %s", err)
		}
		cm = &rsp.CommonNode
		node.Tuic = rsp
		node.Security = Tls
	case "anytls":
		rsp := &AnyTlsNode{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode anytls params error: %s", err)
		}
		cm = &rsp.CommonNode
		node.AnyTls = rsp
		node.Security = Tls
	case "hysteria":
		rsp := &HysteriaNode{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode hysteria params error: %s", err)
		}
		cm = &rsp.CommonNode
		node.Hysteria = rsp
		node.Security = Tls
	case "hysteria2":
		rsp := &Hysteria2Node{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode hysteria2 params error: %s", err)
		}
		cm = &rsp.CommonNode
		node.Hysteria2 = rsp
		node.Security = Tls
	case "naive":
		rsp := &NaiveNode{}
		if err := json.Unmarshal(body, rsp); err != nil {
			return nil, fmt.Errorf("decode naive params error: %s", err)
		}
		if rsp.TlsSettingsBack != nil {
			rsp.TlsSettings = *rsp.TlsSettingsBack
			rsp.TlsSettingsBack = nil
		}
		cm = &rsp.CommonNode
		node.Naive = rsp
		node.Security = rsp.Tls
	default:
		return nil, fmt.Errorf("unsupported node type returned by panel: %s", node.Type)
	}
	if cm == nil {
		return nil, fmt.Errorf("decode node params error: missing common config")
	}
	if err := finalizeNodeInfo(node, cm); err != nil {
		return nil, err
	}
	return node, nil
}

func finalizeNodeInfo(node *NodeInfo, cm *CommonNode) error {
	for i := range cm.Routes {
		matchs, parseErr := parseRouteMatch(cm.Routes[i].Match)
		if parseErr != nil {
			return fmt.Errorf("decode route[%d] match error: %w", i, parseErr)
		}
		if len(matchs) == 0 {
			continue
		}
		switch cm.Routes[i].Action {
		case "block":
			for _, v := range matchs {
				if strings.HasPrefix(v, "protocol:") {
					node.Rules.Protocol = append(node.Rules.Protocol, strings.TrimPrefix(v, "protocol:"))
				} else {
					node.Rules.Regexp = append(node.Rules.Regexp, strings.TrimPrefix(v, "regexp:"))
				}
			}
		case "dns":
			domains := append([]string(nil), matchs...)
			if matchs[0] != "main" {
				node.RawDNS.DNSMap[strconv.Itoa(i)] = map[string]interface{}{
					"address": cm.Routes[i].ActionValue,
					"domains": domains,
				}
			} else {
				node.RawDNS.DNSJson = []byte(strings.Join(matchs[1:], ""))
			}
		}
	}
	if cm.BaseConfig != nil {
		node.PushInterval = intervalToTime(cm.BaseConfig.PushInterval)
		node.PullInterval = intervalToTime(cm.BaseConfig.PullInterval)
	}
	node.Common = cm
	cm.Routes = nil
	cm.BaseConfig = nil
	return nil
}

func (n *NodeInfo) ECH() *ECHSettings {
	switch n.Type {
	case "vmess", "vless":
		if n.VAllss != nil {
			return n.VAllss.TlsSettings.ECH
		}
	case "trojan":
		if n.Trojan != nil {
			return n.Trojan.TlsSettings.ECH
		}
	case "naive":
		if n.Naive != nil {
			return n.Naive.TlsSettings.ECH
		}
	case "tuic":
		if n.Tuic != nil {
			return n.Tuic.TlsSettings.ECH
		}
	case "anytls":
		if n.AnyTls != nil {
			return n.AnyTls.TlsSettings.ECH
		}
	case "hysteria":
		if n.Hysteria != nil {
			return n.Hysteria.TlsSettings.ECH
		}
	case "hysteria2":
		if n.Hysteria2 != nil {
			return n.Hysteria2.TlsSettings.ECH
		}
	}
	return nil
}

func parseRouteMatch(match interface{}) ([]string, error) {
	var raw []string
	switch v := match.(type) {
	case nil:
		return nil, nil
	case string:
		raw = strings.Split(v, ",")
	case []string:
		raw = v
	case []interface{}:
		raw = make([]string, 0, len(v))
		for i := range v {
			value, ok := v[i].(string)
			if !ok {
				return nil, fmt.Errorf("item %d is not string", i)
			}
			raw = append(raw, value)
		}
	default:
		return nil, fmt.Errorf("unsupported type %T", match)
	}
	matches := make([]string, 0, len(raw))
	for i := range raw {
		item := strings.TrimSpace(raw[i])
		if item != "" {
			matches = append(matches, item)
		}
	}
	return matches, nil
}

func intervalToTime(i interface{}) time.Duration {
	switch v := i.(type) {
	case nil:
		return 0
	case int:
		return time.Duration(v) * time.Second
	case int8:
		return time.Duration(v) * time.Second
	case int16:
		return time.Duration(v) * time.Second
	case int32:
		return time.Duration(v) * time.Second
	case int64:
		return time.Duration(v) * time.Second
	case uint:
		return time.Duration(v) * time.Second
	case uint8:
		return time.Duration(v) * time.Second
	case uint16:
		return time.Duration(v) * time.Second
	case uint32:
		return time.Duration(v) * time.Second
	case uint64:
		return time.Duration(v) * time.Second
	case float32:
		return time.Duration(float64(v) * float64(time.Second))
	case float64:
		return time.Duration(v * float64(time.Second))
	case string:
		if v == "" {
			return 0
		}
		seconds, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return time.Duration(seconds * float64(time.Second))
	default:
		return 0
	}
}
