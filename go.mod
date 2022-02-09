module github.com/woodliu/k8sopa

go 1.16

require (
	github.com/golang/glog v1.0.0
	github.com/open-policy-agent/frameworks/constraint v0.0.0-20220125153457-52025b371d65
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.23.1
	k8s.io/apiextensions-apiserver v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v0.23.1
	stash.weimob.com/devops/go_common v0.0.0-20220113055242-eda8ed9cda60
)

replace (
	k8s.io/api v0.23.1 => k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.23.1 => k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.23.1 => k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.23.1 => k8s.io/client-go v0.22.2
)
