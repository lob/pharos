package kubeconfig

import (
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// CurrentCluster returns current context name.
func CurrentCluster(kubeConfigFile string) (string, error) {
	kubeConfig, err := configFromFile(filePath(kubeConfigFile))
	if err != nil {
		return "", errors.Wrap(err, "unable to load config file")
	}

	// Make sure that current context exists.
	context := kubeConfig.Contexts[kubeConfig.CurrentContext]
	if context == nil {
		return "", errors.New("context not found")
	}

	return kubeConfig.CurrentContext, nil
}

// filePath returns final kubeconfig file path.
// It defaults to "$HOME/.kube/config" if empty string is passed in.
func filePath(kubeConfigFile string) string {
	if kubeConfigFile == "" {
		kubeConfigFile = os.Getenv("HOME") + "/.kube/config"
	}

	return kubeConfigFile
}

// configFromFile returns a struct containing kubeconfig information from a file.
// Missing files result in empty kubeconfig struct and an error.
// Function source: https://github.com/kubernetes/client-go/blob/88ff0afc48bbf242f66f2f0c8d5c26b253e6561c/tools/clientcmd/config.go#L471
func configFromFile(fileName string) (*clientcmdapi.Config, error) {
	kubeConfig, err := clientcmd.LoadFromFile(fileName)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if kubeConfig == nil {
		return clientcmdapi.NewConfig(), errors.New("file not found")
	}
	return kubeConfig, nil
}
