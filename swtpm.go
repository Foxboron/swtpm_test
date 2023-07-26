package swtpm

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/google/go-tpm/tpm2/transport"
)

func CreateUserConfigFiles(dir string) error {
	cmd := exec.Command("swtpm_setup", "--create-config-files")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("XDG_CONFIG_HOME=%s", dir),
	)
	return cmd.Run()
}

type Swtpm struct {
	c             *exec.Cmd
	Tpmstate      string
	Flags         []string
	SetupFlags    []string
	LockNVRam     bool
	ServerSocket  string
	ControlSocket string
	tpm           transport.TPMCloser
}

var _ transport.TPMCloser = &Swtpm{}

func (s *Swtpm) Send(input []byte) ([]byte, error) {
	return s.tpm.Send(input)
}

func (s *Swtpm) Setup() error {
	if err := CreateUserConfigFiles(s.Tpmstate); err != nil {
		return fmt.Errorf("failed to create user config files: %s", err)
	}
	args := []string{
		"--tpm2",
		"--tpmstate", s.Tpmstate,
		"--create-ek-cert",
		"--create-platform-cert",
		"--lock-nvram",
	}
	cmd := exec.Command("swtpm_setup", args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("XDG_CONFIG_HOME=%s", s.Tpmstate),
	)
	return cmd.Run()
}

func (s *Swtpm) Socket() (string, error) {
	permall := path.Join(s.Tpmstate, "tpm2-00.permall")
	if _, err := os.Stat(permall); os.IsNotExist(err) {
		if err := s.Setup(); err != nil {
			return "", fmt.Errorf("failed swtmp setup: %w", err)
		}
	}
	args := []string{
		"socket",
		"--tpm2",
		"--tpmstate", fmt.Sprintf("dir=%s", s.Tpmstate),
		"--server", fmt.Sprintf("type=unixio,path=%s", s.ServerSocket),
		"--ctrl", fmt.Sprintf("type=unixio,path=%s", s.ControlSocket),
		"--flags", "not-need-init,startup-clear",
		// "--log", "file=logs,level=20",
	}
	s.c = exec.Command("swtpm", args...)
	s.c.Env = append(os.Environ(),
		fmt.Sprintf("XDG_CONFIG_HOME=%s", s.Tpmstate),
	)
	err := s.c.Start()
	time.Sleep(time.Millisecond * 100)
	return s.ServerSocket, err
}

func (s *Swtpm) Close() error {
	time.Sleep(time.Millisecond * 100)
	s.c.Process.Signal(syscall.SIGTERM)
	s.c.Process.Wait()
	if s.tpm != nil {
		return s.tpm.Close()
	}
	return nil
}

func NewSwtpm(dir string) *Swtpm {
	// Old API
	return &Swtpm{
		Tpmstate:      dir,
		Flags:         []string{"not-need-init", "startup-clear"},
		LockNVRam:     true,
		SetupFlags:    []string{"--create-ek-cert", "--create-platform-cert", "--lock-nvram"},
		ServerSocket:  path.Join(dir, "swtp-sock"),
		ControlSocket: path.Join(dir, "swtp-sock.ctrl"),
	}
}

func OpenSwtpm(dir string) (transport.TPMCloser, error) {
	// TPM Transport option
	swtpm := NewSwtpm(dir)

	s, err := swtpm.Socket()
	if err != nil {
		return nil, err
	}

	tpm, err := transport.OpenTPM(s)
	if err != nil {
		return nil, err
	}
	swtpm.tpm = tpm
	return swtpm, nil
}
