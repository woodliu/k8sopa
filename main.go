package main

import (
    "encoding/json"
    "github.com/golang/glog"
    informer2 "github.com/woodliu/k8sopa/pkg/informer"
    "gopkg.in/yaml.v3"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "os"
    "os/signal"
    "stash.weimob.com/devops/go_common/log"
    "syscall"
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
    host := "https://192.168.118.148:6443"
    token := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImZSdnF0dFR1M2xMSzhJSVIxQ09tZ25HcWJTY01md1owUkcyYmlENmtiSDAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4tNzU1bjQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjQzMjM3MzgzLTI5MjktNDJiNi1iMDNhLWQ3MDA4ODE2MDg1ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.oe98ZLj4gJDC0RQ6JwVHR2POBgzDmBTEVS2eYI4kL6Je9Q-NAcZhtwODPg6io40uxHubukYv57-V7CeNu0gjaXmKTs6Ou4VhX_C_zDMNScxX4V3IrHampuSHdS20SBVLSgzrOIHkaFieYSfWLR4jszUeG3OVfP-FW1zK6qxt7lX1R-rsIXYilE0s-TdsIxSLzblGi8GOUIDMfiUpwBOtBldtQ591YDwUcdOahAyzyicNrsOMXqENwObENBdasw-ZTmYS7qZryzIPSFIQDRWbotbjKiAW2OWZbfDV09wivnpU08iL41ZNmhmLpV9MCgTk8rLkziLEgSDih_yTswbEBA"

    informer,err := informer2.NewInformer(host,token)
    if nil != err {
        log.Error(err)
        return
    }

    informer.StartConstraintTmplInformer()
    // listening OS shutdown singal
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
    <-signalChan

    glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")

    // TODO:exclude

}
//D:\code\gosrc\src\github.com\open-policy-agent\gatekeeper\pkg\webhook\policy.go  Handle函数处理了exclude, 应该不需要的，rego中有做exclude处理。此处的exclude应该是exclude特定处理
//D:\code\gosrc\src\github.com\open-policy-agent\gatekeeper\pkg\controller\config\process\excluder.go excluder.go


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
