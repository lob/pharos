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
	_, ok := kubeConfig.Contexts[kubeConfig.CurrentContext]
	if !ok {
		return "", errors.New("context not found")
	}

	return kubeConfig.CurrentContext, nil
}

// SwitchCluster switches current context to given cluster or context name.
func SwitchCluster(kubeConfigFile string, context string) error {
	kubeConfigFile = filePath(kubeConfigFile)
	kubeConfig, err := configFromFile(kubeConfigFile)
	if err != nil {
		return err
	}

	// Check if there is a context corresponding to the given context name or cluster.
	_, ok := (kubeConfig.Contexts[context])
	if !ok {
		return errors.New("cluster does not exist in context")
	}

	// Switch to new cluster.
	kubeConfig.CurrentContext = context

	return clientcmd.WriteToFile(*kubeConfig, kubeConfigFile)
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
// Does not differentiate between errors resulting from a missing file and errors
// from reading from a malformed config.
// Function source: https://github.com/kubernetes/client-go/blob/88ff0afc48bbf242f66f2f0c8d5c26b253e6561c/tools/clientcmd/config.go#L471
func configFromFile(fileName string) (*clientcmdapi.Config, error) {
	kubeConfig, err := clientcmd.LoadFromFile(fileName)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}
