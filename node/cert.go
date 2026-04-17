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
		return fmt.Errorf("tls node requires cert config")
	}
	if len(c.CertConfig.Certificate) > 0 || len(c.CertConfig.Key) > 0 {
		if len(c.CertConfig.Certificate) == 0 || len(c.CertConfig.Key) == 0 {
			return fmt.Errorf("inline certificate or key not exist")
		}
		return nil
	}
	c.ensureMachineCertPaths()
	c.normalizeCertPaths()
	if c.CertConfig.CertMode == "" || c.CertConfig.CertMode == "none" {
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
	case "dns", "http":
		if c.CertConfig.CertDomain == "" {
			return fmt.Errorf("cert domain not exist")
		}
		if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
			return fmt.Errorf("cert file path or key file path not exist")
		}
		if file.IsExist(c.CertConfig.CertFile) && file.IsExist(c.CertConfig.KeyFile) {
			return nil
		}
		l, err := NewLego(c.CertConfig)
		if err != nil {
			return fmt.Errorf("create lego object error: %s", err)
		}
		err = l.CreateCert()
		if err != nil {
			return fmt.Errorf("create lego cert error: %s", err)
		}
	case "self":
		if c.CertConfig.CertDomain == "" {
			return fmt.Errorf("cert domain not exist")
		}
		if c.CertConfig.CertFile == "" || c.CertConfig.KeyFile == "" {
			return fmt.Errorf("cert file path or key file path not exist")
		}
		if file.IsExist(c.CertConfig.CertFile) && file.IsExist(c.CertConfig.KeyFile) {
			return nil
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
	f, err := os.OpenFile(certPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	err = pem.Encode(f, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	if err != nil {
		return err
	}
	f, err = os.OpenFile(keyPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	err = pem.Encode(f, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return err
	}
	return nil
}
