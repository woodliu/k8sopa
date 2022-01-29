package informer

import (
    "fmt"
    frameworkClient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
    "github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/local"
    "github.com/woodliu/k8sopa/pkg/target"
    apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/dynamic/dynamicinformer"
    "k8s.io/client-go/rest"
)

//type CnsrtInformer interface {
//    StartConstraintTmplInformer()
//    StopConstraintTmplInformer()
//    StartConstraintInformer()
//    StopConstraintInformer()
//}

type Informer struct {
    FrameworkClient              *frameworkClient.Client
    ApiExtensionsV1Client        *rest.RESTClient
    DynamicSharedInformerFactory dynamicinformer.DynamicSharedInformerFactory
    ConstraintTmplStopCh         chan struct{}
    ConstraintStopCh             map[schema.GroupVersionResource] chan struct{}
}

func NewRestCfg(host,token string)*rest.Config{
    return &rest.Config{
        Host:            host,
        BearerToken:     token,
        TLSClientConfig: rest.TLSClientConfig{Insecure: true},
    }
}

func NewCrdClient(restCfg *rest.Config)(*rest.RESTClient,error){
    restCfg.APIPath = "/apis"
    restCfg.GroupVersion = &apiextensionsv1.SchemeGroupVersion
    restCfg.NegotiatedSerializer = Codecs.WithoutConversion()
    if restCfg.UserAgent == "" {
        restCfg.UserAgent = rest.DefaultKubernetesUserAgent()
    }
    return rest.RESTClientFor(restCfg)
}

func NewInformer(host,token string)(*Informer,error){
    driver := local.New()
    backend, err := frameworkClient.NewBackend(frameworkClient.Driver(driver))
    if err != nil {
        return nil,err
    }

    frameworkClient, err := backend.NewClient(frameworkClient.Targets(&target.K8sValidationTarget{}))
    if err != nil {
        return nil,err
    }

    restCfg := NewRestCfg(host,token)
    dynamicClient,err := dynamic.NewForConfig(restCfg)
    if nil != err {
        return nil,err
    }

    apiExtensionsV1Client,err := NewCrdClient(restCfg)
    if nil != err {
        return nil,err
    }

    return &Informer{
        FrameworkClient:              frameworkClient,
        DynamicSharedInformerFactory: dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0),
        ApiExtensionsV1Client:        apiExtensionsV1Client,
        ConstraintTmplStopCh:         make(chan struct{}),
        ConstraintStopCh:             make(map[schema.GroupVersionResource]chan struct{}),
    },nil
}

func (i *Informer) StartConstraintTmplInformer(){
    dynamicInformer := i.DynamicSharedInformerFactory.ForResource(constraintTemplateGvr)
    dynamicInformer.Informer().AddEventHandler(&constraintTemplate{})

    fmt.Println("Start constraint template informer")
    go i.DynamicSharedInformerFactory.Start(i.ConstraintTmplStopCh)
}

func (i *Informer) startConstraintInformer(gvr schema.GroupVersionResource){
    dynamicInformer := i.DynamicSharedInformerFactory.ForResource(gvr)
    dynamicInformer.Informer().AddEventHandler(&constraint{})

    go i.DynamicSharedInformerFactory.Start(i.ConstraintStopCh[gvr])
}

func (i *Informer) stopConstraintInformer(gvr schema.GroupVersionResource) {
    close(i.ConstraintStopCh[gvr])
    delete(i.ConstraintStopCh, gvr)
}
