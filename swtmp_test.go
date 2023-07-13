package swtpm

import (
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
)

func TestSwtpm(t *testing.T) {
	tpm := NewSwtpm(t.TempDir())
	socket, err := tpm.Socket()
	if err != nil {
		t.Fatal(err)
	}
	rwc, err := transport.OpenTPM(socket)
	if err != nil {
		t.Fatal(err)
	}

	getCmd := tpm2.GetCapability{
		Capability: tpm2.TPMCapAlgs,
	}
	_, err = getCmd.Execute(rwc)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
