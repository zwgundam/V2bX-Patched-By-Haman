package node

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoeclubM/V2bX/api/panel"
	"github.com/MoeclubM/V2bX/conf"
)

func TestRequestCertRequiresLocalConfigForTLSNode(t *testing.T) {
	controller := &Controller{
		Options: &conf.Options{
			CertConfig: conf.NewCertConfig(),
		},
	}
	err := controller.requestCert()
	if err == nil {
		t.Fatal("expected request cert error")
	}
	if !strings.Contains(err.Error(), "cert config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestCertAcceptsInlineCertificate(t *testing.T) {
	controller := &Controller{
		Options: &conf.Options{
			CertConfig: &conf.CertConfig{
				Certificate: []string{"cert"},
				Key:         []string{"key"},
			},
		},
	}
	if err := controller.requestCert(); err != nil {
		t.Fatalf("request cert error: %v", err)
	}
}

func TestEnsureMachineCertPathsFillsMissingPathsForV2ACME(t *testing.T) {
	controller := &Controller{
		apiClient: &panel.Client{
			MachineID: 1,
			NodeId:    179,
		},
		Options: &conf.Options{
			CertConfig: &conf.CertConfig{
				CertMode:   "http",
				CertDomain: "node.example.com",
				Email:      "admin@example.com",
			},
		},
	}
	controller.ensureMachineCertPaths()
	if controller.CertConfig.CertFile != filepath.Join("/etc/V2bX", "cert", "machine-1", "node-179.pem") {
		t.Fatalf("unexpected cert path: %s", controller.CertConfig.CertFile)
	}
	if controller.CertConfig.KeyFile != filepath.Join("/etc/V2bX", "cert", "machine-1", "node-179.key") {
		t.Fatalf("unexpected key path: %s", controller.CertConfig.KeyFile)
	}
}

func TestNormalizeCertPathsResolvesPlaceholders(t *testing.T) {
	controller := &Controller{
		apiClient: &panel.Client{
			MachineID: 2,
			NodeId:    7,
		},
		Options: &conf.Options{
			CertConfig: &conf.CertConfig{
				CertDomain: "node.example.com",
				Email:      "admin@example.com",
				CertFile:   "/etc/V2bX/{machine_id}/{node_id}/{domain}.pem",
				KeyFile:    "/etc/V2bX/{machine_id}/{node_id}/{email}.key",
			},
		},
	}
	controller.normalizeCertPaths()
	if controller.CertConfig.CertFile != "/etc/V2bX/2/7/node.example.com.pem" {
		t.Fatalf("unexpected cert path: %s", controller.CertConfig.CertFile)
	}
	if controller.CertConfig.KeyFile != "/etc/V2bX/2/7/admin@example.com.key" {
		t.Fatalf("unexpected key path: %s", controller.CertConfig.KeyFile)
	}
}

func TestRequestCertRejectsMissingCertDomainForACME(t *testing.T) {
	controller := &Controller{
		Options: &conf.Options{
			CertConfig: &conf.CertConfig{
				CertMode: "http",
				CertFile: "/tmp/test.pem",
				KeyFile:  "/tmp/test.key",
			},
		},
	}
	err := controller.requestCert()
	if err == nil {
		t.Fatal("expected request cert error")
	}
	if !strings.Contains(err.Error(), "cert domain") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestCertDefaultsToSelfSignedFromTLSServerName(t *testing.T) {
	dir := t.TempDir()
	controller := &Controller{
		info: &panel.NodeInfo{
			Type: "anytls",
			AnyTls: &panel.AnyTlsNode{
				TlsSettings: panel.TlsSettings{
					ServerName: "node.example.com",
				},
			},
		},
		Options: &conf.Options{
			CertConfig: &conf.CertConfig{
				CertFile: filepath.Join(dir, "node.pem"),
				KeyFile:  filepath.Join(dir, "node.key"),
			},
		},
	}
	if err := controller.requestCert(); err != nil {
		t.Fatalf("request cert error: %v", err)
	}
	if controller.CertConfig.CertMode != "self" {
		t.Fatalf("unexpected cert mode: %s", controller.CertConfig.CertMode)
	}
	if controller.CertConfig.CertDomain != "node.example.com" {
		t.Fatalf("unexpected cert domain: %s", controller.CertConfig.CertDomain)
	}
	if _, err := os.Stat(controller.CertConfig.CertFile); err != nil {
		t.Fatalf("stat cert file error: %v", err)
	}
	if _, err := os.Stat(controller.CertConfig.KeyFile); err != nil {
		t.Fatalf("stat key file error: %v", err)
	}
}

func TestRequestCertFallsBackToCommonServerName(t *testing.T) {
	dir := t.TempDir()
	controller := &Controller{
		info: &panel.NodeInfo{
			Type: "hysteria",
			Common: &panel.CommonNode{
				ServerName: "sni.example.com",
			},
		},
		Options: &conf.Options{
			CertConfig: &conf.CertConfig{
				CertFile: filepath.Join(dir, "node.pem"),
				KeyFile:  filepath.Join(dir, "node.key"),
			},
		},
	}
	if err := controller.requestCert(); err != nil {
		t.Fatalf("request cert error: %v", err)
	}
	if controller.CertConfig.CertMode != "self" {
		t.Fatalf("unexpected cert mode: %s", controller.CertConfig.CertMode)
	}
	if controller.CertConfig.CertDomain != "sni.example.com" {
		t.Fatalf("unexpected cert domain: %s", controller.CertConfig.CertDomain)
	}
}
