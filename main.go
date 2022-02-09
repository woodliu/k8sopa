package main

import (
	"context"
	"github.com/golang/glog"
	"github.com/woodliu/k8sopa/pkg/validate"
	"os"
	"os/signal"
	"stash.weimob.com/devops/go_common/log"
	"syscall"
	"time"
)

// TODO:CRD的gvk名称修改 framework的createCrd接口生成的group就是constraints.gatekeeper.sh，重命名？
// TODO:目前只能对资源的格式做限制，无法对资源的操作做限制--是否可以支持?
var host = "https://192.168.118.148:6443"
var token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImZSdnF0dFR1M2xMSzhJSVIxQ09tZ25HcWJTY01md1owUkcyYmlENmtiSDAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4tNzU1bjQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjQzMjM3MzgzLTI5MjktNDJiNi1iMDNhLWQ3MDA4ODE2MDg1ZCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.oe98ZLj4gJDC0RQ6JwVHR2POBgzDmBTEVS2eYI4kL6Je9Q-NAcZhtwODPg6io40uxHubukYv57-V7CeNu0gjaXmKTs6Ou4VhX_C_zDMNScxX4V3IrHampuSHdS20SBVLSgzrOIHkaFieYSfWLR4jszUeG3OVfP-FW1zK6qxt7lX1R-rsIXYilE0s-TdsIxSLzblGi8GOUIDMfiUpwBOtBldtQ591YDwUcdOahAyzyicNrsOMXqENwObENBdasw-ZTmYS7qZryzIPSFIQDRWbotbjKiAW2OWZbfDV09wivnpU08iL41ZNmhmLpV9MCgTk8rLkziLEgSDih_yTswbEBA"

func main() {
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()
	vldt := validate.New(ctx, host, token)
	time.Sleep(time.Second)
	resp := vldt.Validate([]byte(`{
  "kind": "Namespace",
  "apiVersion": "v1",
  "metadata": {
    "name": "cos-opa",
    "labels": {
      "kubernetes.io/metadata.name": "cos-opa"
    },
  "spec": {
    "finalizers": [
      "kubernetes"
    ]
  }
}}`))

	log.Info(resp)
	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
}


//D:\code\gosrc\src\github.com\open-policy-agent\gatekeeper\pkg\webhook\policy.go  Handle函数处理了exclude, 应该不需要的，rego中有做exclude处理。此处的exclude应该是exclude特定处理
//D:\code\gosrc\src\github.com\open-policy-agent\gatekeeper\pkg\controller\config\process\excluder.go excluder.go
