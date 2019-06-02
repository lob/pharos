package kubeconfig

import (
	"errors"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// CurrentCluster returns current cluster name from current context
func CurrentCluster(kubeConfigFile string) (string, error) {
	kubeConfig, err := ConfigFromFile(FilePath(kubeConfigFile))
	if err != nil {
		return "", err
	}

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

// ConfigFromFile returns a struct containing kubeconfig information from a file
// Empty or missing files result in empty kubeconfig structs, not an error
// Function source: https://github.com/kubernetes/client-go/blob/88ff0afc48bbf242f66f2f0c8d5c26b253e6561c/tools/clientcmd/config.go#L471
func ConfigFromFile(filename string) (*clientcmdapi.Config, error) {
	kubeConfig, err := clientcmd.LoadFromFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if kubeConfig == nil {
		kubeConfig = clientcmdapi.NewConfig()
	}

	return kubeConfig, nil
}
