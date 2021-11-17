package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Kind struct {
	Ctx            context.Context
	Name           string
	KubeconfigPath string
	LazyMode       bool
}

func (k *Kind) createCluster() error {
	isRunning := k.checkCluster()
	if k.LazyMode && isRunning {
		fmt.Println("kind cluster is already running")
		return nil
	}
	cmd := exec.CommandContext(
		k.Ctx,
		"kind",
		"create",
		"cluster",
		"--name",
		k.Name,
		"--kubeconfig",
		k.KubeconfigPath,
		"--wait", // block until the control plane reaches a ready status
		"30s",
	)
	// cmd.Path = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("start creating kind cluster")
	return cmd.Run()
}

func (k *Kind) checkCluster() bool {
	out, err := exec.CommandContext(
		k.Ctx,
		"kind",
		"get",
		"clusters",
	).Output()

	if err != nil {
		return false
	}
	clusters := strings.Split(string(out), "\n")
	return stringInSlice(k.Name, clusters)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (k *Kind) deleteCluster() error {
	if k.LazyMode {
		fmt.Println("keep kind cluster running for next time.")
		return nil
	}
	cmd := exec.CommandContext(
		k.Ctx,
		"kind",
		"delete",
		"cluster",
		"--name",
		k.Name,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (k *Kind) checkVersion() error {
	cmd := exec.Command("kind", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Printf("kind version: %s\n", out.String())
	return nil
}

func newKind(ctx context.Context, name, kubeconfigPath string, lazyMode bool) *Kind {
	return &Kind{
		Ctx:            ctx,
		Name:           name,
		KubeconfigPath: kubeconfigPath,
		LazyMode:       lazyMode,
	}
}
