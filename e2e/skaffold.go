package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Skaffold struct {
	KubeconfigPath string
	cmd            *exec.Cmd
}

func (s *Skaffold) run(ctx context.Context) error {
	fmt.Println("run")
	args := []string{"run", "--kubeconfig", s.KubeconfigPath, "--tail"}
	s.cmd = exec.CommandContext(
		ctx,
		"skaffold",
		args...,
	)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	return s.cmd.Start() // Run in background
}

func (s *Skaffold) cleanup() error {
	fmt.Println("skaffold cleanup")
	fmt.Println("skaffold kill process")
	errKill := s.cmd.Process.Kill()
	s.cmd = exec.Command(
		"skaffold",
		"delete",
		"--kubeconfig",
		s.KubeconfigPath,
	)
	fmt.Println("skaffold delete")
	errRun := s.cmd.Run()
	if errKill != nil {
		return errKill
	} else if errRun != nil {
		return errRun
	}
	return nil
}
