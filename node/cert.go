package node

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MoeclubM/V2bX/common/file"
	"github.com/MoeclubM/V2bX/conf"
	log "github.com/sirupsen/logrus"
)

func (c *Controller) renewCertTask() error {
	l, err := NewLego(c.CertConfig)
	if err != nil {
		log.WithField("tag", c.tag).Info("new lego error: ", err)
		return nil
	}
	err = l.RenewCert()
	if err != nil {
		log.WithField("tag", c.tag).Info("renew cert error: ", err)
		return nil
	}
	return nil
}

func (c *Controller) requestCert() error {
	if c.CertConfig == nil {
		c.CertConfig = &conf.CertConfig{}
	}
	if len(c.CertConfig.Certificate) > 0 || len(c.CertConfig.Key) > 0 {
		if len(c.CertConfig.Certificate) == 0 || len(c.CertConfig.Key) == 0 {
			return fmt.Errorf("inline certificate or key not exist")
		}
		return nil
	}
	if c.CertConfig.CertDomain == "" && c.info != nil {
		switch c.info.Type {
		case "vless":
			if c.info.VLESS != nil {
				c.CertConfig.CertDomain = c.info.VLESS.TlsSettings.ServerName
			}
		case "anytls":
			if c.info.AnyTls != nil {
				c.CertConfig.CertDomain = c.info.AnyTls.TlsSettings.ServerName
			}
		case "hysteria2":
			if c.info.Hysteria2 != nil {
				c.CertConfig.CertDomain = c.info.Hysteria2.TlsSettings.ServerName
			}
		}
		if c.CertConfig.CertDomain == "" && c.info.Common != nil {
			c.CertConfig.CertDomain = c.info.Common.ServerName
		}
	}
	if c.CertConfig.CertMode == "" {
		c.CertConfig.CertMode = "self"
	}
	c.ensureMachineCertPaths()
	c.normalizeCertPaths()
	if c.CertConfig.CertMode == "none" {
		if c.CertConfig.CertFile != "" && c.CertConfig.KeyFile != "" {
			return nil
		}
		return fmt.Errorf("tls node requires cert config")
	}
	switch c.CertConfig.CertMode {
	case "file":
		if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
			return fmt.Errorf("cert file path or key file path not exist")
		}
		// file 模式要判断现有证书是否是自签名或者过期。如是则提示用户手动调整文件
		if file.IsExist(c.CertConfig.CertFile) && file.IsExist(c.CertConfig.KeyFile) {
			if isSelfSignedOrInvalidCert(c.CertConfig.CertFile, c.CertConfig.CertDomain) {
				log.WithField("tag", c.tag).Warnf("Existing cert at %s is self-signed or invalid, please update your cert files manually.", c.CertConfig.CertFile)
			}
		}
	case "dns", "http":
		if c.CertConfig.CertDomain == "" {
			return fmt.Errorf("cert domain not exist")
		}
		if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
			return fmt.Errorf("cert file path or key file path not exist")
		}
		if c.CertConfig.CertMode == "dns" && c.CertConfig.Provider == "" {
			return fmt.Errorf("dns cert mode requires 'Provider' (e.g. cloudflare)")
		}
		if file.IsExist(c.CertConfig.CertFile) && file.IsExist(c.CertConfig.KeyFile) {
			if isSelfSignedOrInvalidCert(c.CertConfig.CertFile, c.CertConfig.CertDomain) {
				log.WithField("tag", c.tag).Infof("Existing cert at %s is self-signed or invalid for domain %s, removing old cert and requesting ACME...", c.CertConfig.CertFile, c.CertConfig.CertDomain)
				_ = os.Remove(c.CertConfig.CertFile)
				_ = os.Remove(c.CertConfig.KeyFile)
			} else {
				log.WithField("tag", c.tag).Infof("Domain %s certificate exists and is valid, using existing certificate.", c.CertConfig.CertDomain)
				return nil
			}
		}
		log.WithField("tag", c.tag).Infof("Requesting new ACME certificate for domain %s via %s mode...", c.CertConfig.CertDomain, c.CertConfig.CertMode)
		l, err := NewLego(c.CertConfig)
		if err != nil {
			log.WithField("tag", c.tag).Warnf("Create lego object error: %s, falling back to self-signed cert...", err)
			return generateSelfSslCertificate(c.CertConfig.CertDomain, c.CertConfig.CertFile, c.CertConfig.KeyFile)
		}

		// 如果是 http 模式，自动执行 80 端口智能借用与恢复机制
		if c.CertConfig.CertMode == "http" {
			yieldedSvc, yieldErr := yieldPort80()
			if yieldErr != nil {
				log.WithField("tag", c.tag).Warnf("80 端口自动借用尝试提醒: %s", yieldErr)
			}
			if yieldedSvc != nil {
				defer restorePort80(yieldedSvc)
			}
		}

		err = l.CreateCert()
		if err != nil {
			log.WithField("tag", c.tag).Warnf("ACME cert request failed (%s). Port 80 might be occupied or domain DNS not pointing to this IP.", err)
			log.WithField("tag", c.tag).Warnf("HINT: If port 80 is used by Nginx/Web, switch CertMode to 'dns' in panel or free port 80.")
			log.WithField("tag", c.tag).Warnf("Falling back to self-signed certificate to keep node running...")
			return generateSelfSslCertificate(c.CertConfig.CertDomain, c.CertConfig.CertFile, c.CertConfig.KeyFile)
		}
		log.WithField("tag", c.tag).Infof("ACME certificate for domain %s acquired successfully!", c.CertConfig.CertDomain)
	case "self":
		if c.CertConfig.CertDomain == "" {
			return fmt.Errorf("cert domain not exist")
		}
		if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
			return fmt.Errorf("cert file path or key file path not exist")
		}
		if file.IsExist(c.CertConfig.CertFile) && file.IsExist(c.CertConfig.KeyFile) {
			// 如果已有 ACME 颁发的合法证书，且域名匹配，则直接复用不被自签名覆盖
			if !isSelfSignedOrInvalidCert(c.CertConfig.CertFile, c.CertConfig.CertDomain) {
				return nil
			}
		}
		err := generateSelfSslCertificate(
			c.CertConfig.CertDomain,
			c.CertConfig.CertFile,
			c.CertConfig.KeyFile)
		if err != nil {
			return fmt.Errorf("generate self cert error: %s", err)
		}
	default:
		return fmt.Errorf("unsupported certmode: %s", c.CertConfig.CertMode)
	}
	return nil
}

func (c *Controller) ensureMachineCertPaths() {
	if c.CertConfig == nil || c.apiClient == nil || c.apiClient.MachineID <= 0 || c.apiClient.NodeId <= 0 {
		return
	}
	switch c.CertConfig.CertMode {
	case "http", "dns", "self":
	default:
		return
	}
	baseDir := filepath.Join("/etc/V2bX", "cert", fmt.Sprintf("machine-%d", c.apiClient.MachineID))
	if c.CertConfig.CertFile == "" {
		c.CertConfig.CertFile = filepath.Join(baseDir, fmt.Sprintf("node-%d.pem", c.apiClient.NodeId))
	}
	if c.CertConfig.KeyFile == "" {
		c.CertConfig.KeyFile = filepath.Join(baseDir, fmt.Sprintf("node-%d.key", c.apiClient.NodeId))
	}
}

func (c *Controller) normalizeCertPaths() {
	if c.CertConfig == nil {
		return
	}
	machineID := 0
	nodeID := 0
	if c.apiClient != nil {
		machineID = c.apiClient.MachineID
		nodeID = c.apiClient.NodeId
	}
	if c.CertConfig.CertFile != "" {
		c.CertConfig.CertFile = resolveCertPath(c.CertConfig.CertFile, c.CertConfig, machineID, nodeID)
	}
	if c.CertConfig.KeyFile != "" {
		c.CertConfig.KeyFile = resolveCertPath(c.CertConfig.KeyFile, c.CertConfig, machineID, nodeID)
	}
}

func resolveCertPath(rawPath string, certConfig *conf.CertConfig, machineID, nodeID int) string {
	if rawPath == "" || certConfig == nil {
		return rawPath
	}
	replacer := strings.NewReplacer(
		"{domain}", certConfig.CertDomain,
		"{email}", certConfig.Email,
		"{machine_id}", strconv.Itoa(machineID),
		"{node_id}", strconv.Itoa(nodeID),
	)
	return replacer.Replace(rawPath)
}

func generateSelfSslCertificate(domain, certPath, keyPath string) error {
	_ = os.MkdirAll(filepath.Dir(certPath), 0755)
	_ = os.MkdirAll(filepath.Dir(keyPath), 0755)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		Version:      3,
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames:              []string{domain},
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(30, 0, 0),
	}
	cert, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	if err != nil {
		return err
	}
	certFile, err := os.OpenFile(certPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	if err != nil {
		return err
	}
	keyFile, err := os.OpenFile(keyPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	err = pem.Encode(keyFile, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return err
	}
	return nil
}

// 检查现有证书是否为自签名证书、即将在24小时内过期或域名不匹配
func isSelfSignedOrInvalidCert(certPath string, targetDomain string) bool {
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return true
	}
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return true
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return true
	}
	// 1. 如果是自签名证书 (Issuer == Subject)，判定为无效需重新申请正式证书
	if cert.Issuer.String() == cert.Subject.String() {
		return true
	}
	// 2. 如果证书即将在 24 小时内过期，判定为需更新
	if time.Now().Add(24 * time.Hour).After(cert.NotAfter) {
		return true
	}
	// 3. 检查域名匹配度
	if targetDomain != "" {
		matched := false
		for _, dnsName := range cert.DNSNames {
			if dnsName == targetDomain || (strings.HasPrefix(dnsName, "*.") && strings.HasSuffix(targetDomain, dnsName[1:])) {
				matched = true
				break
			}
		}
		if !matched && cert.Subject.CommonName != targetDomain {
			return true
		}
	}
	return false
}
