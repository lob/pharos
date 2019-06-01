package kubeconfig

import (
	"errors"
	"os"
	"reflect"

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
func ConfigFromFile(filename string) (*clientcmdapi.Config, error) {
	kubeConfig, err := clientcmd.LoadFromFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if kubeConfig == nil {
		return kubeConfig, errors.New("unable to load kubeconfig file")
	}

	// If the kubeconfig struct is empty, this indicates an error retrieving the kubeconfig
	if reflect.DeepEqual(kubeConfig, clientcmdapi.NewConfig()) {
		return nil, errors.New("kubeconfig file is malformed or empty")
	}

	return kubeConfig, nil
}
