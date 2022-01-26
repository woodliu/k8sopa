package main

import (
    "encoding/json"
    "fmt"
    "github.com/open-policy-agent/frameworks/constraint/pkg/apis/templates/v1beta1"
    "github.com/open-policy-agent/frameworks/constraint/pkg/client"
    "github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/local"
    "github.com/woodliu/k8sopa/pkg/target"
    "gopkg.in/yaml.v3"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/discovery"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/informers"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "sigs.k8s.io/controller-runtime/pkg/controller"
    "sigs.k8s.io/controller-runtime/pkg/handler"
    "sigs.k8s.io/controller-runtime/pkg/manager"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    "sigs.k8s.io/controller-runtime/pkg/source"
    "strings"
    "time"
)


const (
    crdName                          = "constrainttemplates.templates.gatekeeper.sh"
    constraintsGV                    = "constraints.gatekeeper.sh/v1beta1"
    msgSize                          = 256
    defaultAuditInterval             = 60
    defaultConstraintViolationsLimit = 20
    defaultListLimit                 = 500
    defaultAPICacheDir               = "/tmp/audit"
)

func main(){

    driver := local.New()

    backend, err := client.NewBackend(client.Driver(driver))
    if err != nil {
        return
    }

    cl, err := backend.NewClient(client.Targets(&target.K8sValidationTarget{}))


    restConfig := &rest.Config{
        Host:            "https://192.168.118.148:6443",
        BearerToken:     "eyJhbGciOiJSUzI1NiIsImtpZCI6ImZSdnF0dFR1M2xMSzhJSVIxQ09tZ25HcWJTY01md1owUkcyYmlENmtiSDAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4tNzU1bjQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjQzMjM3MzgzLTI5MjktNDJiNi1iMDNhLWQ3MDA4ODE2MDg1ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.oe98ZLj4gJDC0RQ6JwVHR2POBgzDmBTEVS2eYI4kL6Je9Q-NAcZhtwODPg6io40uxHubukYv57-V7CeNu0gjaXmKTs6Ou4VhX_C_zDMNScxX4V3IrHampuSHdS20SBVLSgzrOIHkaFieYSfWLR4jszUeG3OVfP-FW1zK6qxt7lX1R-rsIXYilE0s-TdsIxSLzblGi8GOUIDMfiUpwBOtBldtQ591YDwUcdOahAyzyicNrsOMXqENwObENBdasw-ZTmYS7qZryzIPSFIQDRWbotbjKiAW2OWZbfDV09wivnpU08iL41ZNmhmLpV9MCgTk8rLkziLEgSDih_yTswbEBA",
        TLSClientConfig: rest.TLSClientConfig{Insecure: true},
    }

    //dynClient, err := dynamic.NewForConfig(restConfig)
    //if err != nil {
    //    return
    //}
    discoveryCli,err := discovery.NewDiscoveryClientForConfig(restConfig)
    if nil != err{
        return
    }
    clientSet, err := kubernetes.NewForConfig(restConfig)

    factory := informers.NewSharedInformerFactory(clientSet, 5*time.Second)
    factory.InformerFor()
    //
    //SchemeGroupVersion = schema.GroupVersion{Group: "templates.gatekeeper.sh", Version: "v1beta1"}
    //var r = schema.GroupVersionResource{Group: "operator.victoriametrics.com", Version: "v1beta1", Resource: "vmrules"}
    //list, err := dynClient.Resource(r).Namespace("vm").List(context.TODO(), metav1.ListOptions{})
    //if err != nil {
    //    return
    //}
    //dynClient.
    //rlist := v1beta1.VMRuleList{}
    ///* 也可以直接使用json方式解析到结构体中
    //   data, _ := list.MarshalJSON()
    //   if err := json.Unmarshal(data, &rlist); err != nil {
    //       return
    //   }
    //*/
    //runtime.DefaultUnstructuredConverter.FromUnstructured(list.UnstructuredContent(), &rlist)
    //fmt.Println(rlist)

}
//D:\code\gosrc\src\github.com\open-policy-agent\gatekeeper\pkg\webhook\policy.go  Handle函数处理了exclude
//D:\code\gosrc\src\github.com\open-policy-agent\gatekeeper\pkg\controller\config\process\excluder.go excluder.go

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
//watch参考
func add(mgr manager.Manager, r reconcile.Reconciler) error {
    // Create a new controller
    c, err := controller.New(ctrlName, mgr, controller.Options{Reconciler: r})
    if err != nil {
        return err
    }


    // Watch for changes to ConstraintTemplate
    err = c.Watch(&source.Kind{Type: &v1beta1.ConstraintTemplate{}}, &handler.EnqueueRequestForObject{})
    if err != nil {
        return err
    }

    // Watch for changes to ConstraintTemplateStatus
    err = c.Watch(
        &source.Kind{Type: &statusv1beta1.ConstraintTemplatePodStatus{}},
        handler.EnqueueRequestsFromMapFunc(constrainttemplatestatus.PodStatusToConstraintTemplateMapper(true)),
    )
    if err != nil {
        return err
    }

    // Watch for changes to Constraint CRDs
    err = c.Watch(
        &source.Kind{Type: &apiextensionsv1.CustomResourceDefinition{}},
        &handler.EnqueueRequestForOwner{
            OwnerType:    &v1beta1.ConstraintTemplate{},
            IsController: true,
        },
    )
    if err != nil {
        return err
    }

    return nil
}


// list参考
func (am *Manager) getAllConstraintKinds() ([]schema.GroupVersionKind, error) {
    discoveryClient, err := discovery.NewDiscoveryClientForConfig(am.mgr.GetConfig())
    if err != nil {
        return nil, err
    }
    l, err := discoveryClient.ServerResourcesForGroupVersion(constraintsGV)
    if err != nil {
        return nil, err
    }
    resourceGV := strings.Split(constraintsGV, "/")
    group := resourceGV[0]
    version := resourceGV[1]
    // We have seen duplicate GVK entries on shifting to status client, remove them
    unique := make(map[schema.GroupVersionKind]bool)
    for i := range l.APIResources {
        unique[schema.GroupVersionKind{Group: group, Version: version, Kind: l.APIResources[i].Kind}] = true
    }
    var ret []schema.GroupVersionKind
    for gvk := range unique {
        ret = append(ret, gvk)
    }
    return ret, nil
}

func readUnstructured(bytes []byte) (*unstructured.Unstructured, error) {
    u := &unstructured.Unstructured{
        Object: make(map[string]interface{}),
    }
    err := parseYAML(bytes, u)
    if err != nil {
        return nil, err
    }
    return u, nil
}


func parseYAML(yamlBytes []byte, v interface{}) error {
    // Pass through JSON since k8s parsing logic doesn't fully handle objects
    // parsed directly from YAML. Without passing through JSON, the OPA client
    // panics when handed scalar types it doesn't recognize.
    obj := make(map[string]interface{})

    err := yaml.Unmarshal(yamlBytes, obj)
    if err != nil {
        return err
    }

    jsonBytes, err := json.Marshal(obj)
    if err != nil {
        return err
    }

    return parseJSON(jsonBytes, v)
}

func parseJSON(jsonBytes []byte, v interface{}) error {
    return json.Unmarshal(jsonBytes, v)
}
