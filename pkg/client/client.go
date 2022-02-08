package client

import (
	frameworkClient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/local"
	"github.com/woodliu/k8sopa/pkg/register"
	"github.com/woodliu/k8sopa/pkg/target"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"stash.weimob.com/devops/go_common/log"
)

type Client struct {
	FrameworkClient              *frameworkClient.Client
	ApiExtensionsV1Client        *rest.RESTClient
	DynamicSharedInformerFactory dynamicinformer.DynamicSharedInformerFactory
	ConstraintTmplStopCh         chan struct{}
	ConstraintStopCh             map[schema.GroupVersionResource]chan struct{}
}

func NewRestCfg(host, token string) *rest.Config {
	return &rest.Config{
		Host:            host,
		BearerToken:     token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
}

func NewApiExtensionsV1Client(restCfg *rest.Config) (*rest.RESTClient, error) {
	restCfg.APIPath = "/apis"
	restCfg.GroupVersion = &apiextensionsv1.SchemeGroupVersion
	restCfg.NegotiatedSerializer = register.Codecs.WithoutConversion()
	if restCfg.UserAgent == "" {
		restCfg.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	return rest.RESTClientFor(restCfg)
}

func NewClient(host, token string) (*Client, error) {
	driver := local.New()
	backend, err := frameworkClient.NewBackend(frameworkClient.Driver(driver))
	if err != nil {
		return nil, err
	}

	frameworkClient, err := backend.NewClient(frameworkClient.Targets(&target.K8sValidationTarget{}))
	if err != nil {
		return nil, err
	}

	restCfg := NewRestCfg(host, token)
	dynamicClient, err := dynamic.NewForConfig(restCfg)
	if nil != err {
		return nil, err
	}

	apiExtensionsV1Client, err := NewApiExtensionsV1Client(restCfg)
	if nil != err {
		return nil, err
	}

	return &Client{
		FrameworkClient:              frameworkClient,
		DynamicSharedInformerFactory: dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0),
		ApiExtensionsV1Client:        apiExtensionsV1Client,
		ConstraintTmplStopCh:         make(chan struct{}),
		ConstraintStopCh:             make(map[schema.GroupVersionResource]chan struct{}),
	}, nil
}

func StartConstraintTmplInformer(host,token string)*Client{
	client, err := NewClient(host, token)
	if nil != err {
		panic(err)
	}

	dynamicInformer := client.DynamicSharedInformerFactory.ForResource(constraintTemplateGvr)
	dynamicInformer.Informer().AddEventHandler(&constraintTemplate{})

	log.Info("start constraint template informer")
	go client.DynamicSharedInformerFactory.Start(client.ConstraintTmplStopCh)
	return client
}

func (c *Client)StopConstraintTmplInformer()  {
	close(c.ConstraintTmplStopCh)
}

func (c *Client)startConstraintInformer(gvr schema.GroupVersionResource) {
	dynamicInformer := c.DynamicSharedInformerFactory.ForResource(gvr)
	dynamicInformer.Informer().AddEventHandler(&constraint{})

	go c.DynamicSharedInformerFactory.Start(c.ConstraintStopCh[gvr])
}

func (c *Client)stopConstraintInformer(gvr schema.GroupVersionResource) {
	close(c.ConstraintStopCh[gvr])
	delete(c.ConstraintStopCh, gvr)
}
