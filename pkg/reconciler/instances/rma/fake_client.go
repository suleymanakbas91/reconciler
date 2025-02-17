package rma

import (
	"io/ioutil"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	helmfake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type FakeClient struct {
	clientset   *fake.Clientset
	helmStorage *storage.Storage
}

func NewFakeClient(clientset *fake.Clientset) *FakeClient {

	return &FakeClient{
		clientset:   clientset,
		helmStorage: storage.Init(driver.NewMemory()),
	}
}

func (c *FakeClient) KubernetesClientSet() (kubernetes.Interface, error) {
	return c.clientset, nil
}

func (c *FakeClient) HelmActionConfiguration(namespace string, log action.DebugLog) (*action.Configuration, error) {
	return &action.Configuration{
		Releases:     c.helmStorage,
		KubeClient:   &helmfake.FailingKubeClient{PrintingKubeClient: helmfake.PrintingKubeClient{Out: ioutil.Discard}},
		Capabilities: chartutil.DefaultCapabilities,
		Log:          log,
	}, nil
}
