package node

import (
	"strings"
	"testing"

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
