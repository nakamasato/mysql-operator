package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Kind struct {
	Ctx context.Context
	Name string
	KubeconfigPath string
}

func (k *Kind) createCluster() error {
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

func (k *Kind) deleteCluster() error {
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

func newKind(ctx context.Context, name, kubeconfigPath string) *Kind {
	return &Kind{Ctx: ctx, Name: name, KubeconfigPath: kubeconfigPath}
}
