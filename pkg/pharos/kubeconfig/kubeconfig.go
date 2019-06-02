package kubeconfig

import (
	"errors"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

// CurrentCluster returns current cluster name from current context
func CurrentCluster(kubeConfigFile string) (string, error) {
	kubeConfig := clientcmd.GetConfigFromFileOrDie(FilePath(kubeConfigFile))

	context := (kubeConfig.Contexts[kubeConfig.CurrentContext])
	if context == nil {
		return "", errors.New("no context found")
	}

	return context.Cluster, nil
}

// FilePath returns final kubeconfig file path (defaults to "$HOME/.kube/config" if empty string is passed in)
func FilePath(kubeConfigFile string) string {
	if kubeConfigFile == "" {
		kubeConfigFile = os.Getenv("HOME") + "/.kube/config"
	}

	return kubeConfigFile
}
