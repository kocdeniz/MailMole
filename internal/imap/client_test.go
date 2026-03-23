package imap

import (
	"testing"
)

func TestConnectIPRequiresInsecureOptIn(t *testing.T) {
	t.Setenv("MAILMOLE_ALLOW_INSECURE_IP_TLS", "")
	cfg := Config{
		Host:     "127.0.0.1",
		Port:     993,
		Username: "user",
		Password: "pass",
		TLS:      true,
	}

	_, err := Connect(cfg)
	if err == nil {
		t.Fatalf("expected error for TLS over IP without opt-in")
	}

}
