package node

import (
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
