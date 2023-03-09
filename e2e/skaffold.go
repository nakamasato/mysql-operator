package e2e

import (
	"fmt"
	"os"
	"os/exec"
)

type Skaffold struct {
	KubeconfigPath string
}

func (s *Skaffold) run(tail bool) error {
	args := []string{"run", "--kubeconfig", s.KubeconfigPath}
	if tail {
		args = append(args, "--tail")
	}
	return s.execute(args...)
}

func (s *Skaffold) delete() error {
	return s.execute(
		"delete",
		"--kubeconfig",
		s.KubeconfigPath,
	)
}

func (s *Skaffold) execute(args ...string) error {
	cmd := exec.Command(
		"skaffold",
		args...,
	)
	// cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run skaffold command. %v", err)
	}
	fmt.Println("skaffold completed")
	return nil
}
