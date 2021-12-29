# Testing

## Controller test

Tools:

- [envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)

## e2e test

Currently, there are two ways for e2e tests.

1. [kuttl](https://kuttl.dev/) (will be deleted)
1. [kind](https://kind.sigs.k8s.io/) + [skaffold](https://skaffold.dev/) + [Ginkgo](https://onsi.github.io/ginkgo/) + [Gomage](https://onsi.github.io/gomega/)
