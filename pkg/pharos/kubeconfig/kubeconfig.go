package kubeconfig

import (
	"os"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// CurrentCluster returns current cluster name from current context
func CurrentCluster(kubeConfigFile *string) string {
	kubeConfig := Config(FilePath(kubeConfigFile))
	context := (kubeConfig.Contexts[kubeConfig.CurrentContext]).Cluster
	return context
}

// FilePath returns correct kubeconfig file path
func FilePath(kubeConfigFile *string) *string {
	if *kubeConfigFile == "/.kube/config" {
		home := os.Getenv("HOME")
		*kubeConfigFile = home + *kubeConfigFile
	}

	return kubeConfigFile
}

// Config returns a struct containing kubeconfig information from a kubeconfig file
func Config(kubeConfigFile *string) *clientcmdapi.Config {
	return clientcmd.GetConfigFromFileOrDie(*kubeConfigFile)
}
