package e2e

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Skaffold struct {
	KubeconfigPath string
}

func (s *Skaffold) run() {
	s.execute(
		"run",
		"--kubeconfig",
		s.KubeconfigPath,
	)
}

func (s *Skaffold) delete() {
	s.execute(
		"delete",
		"--kubeconfig",
		s.KubeconfigPath,
	)
}

func (s *Skaffold) execute(args ...string) {
	cmd := exec.Command(
		"skaffold",
		args...,
	)
	// cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal("failed to run skaffold command.", err)
	}
	fmt.Println("skaffold completed")
}
