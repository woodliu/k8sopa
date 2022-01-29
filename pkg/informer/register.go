package informer

import (
    frameworksapis "github.com/open-policy-agent/frameworks/constraint/pkg/apis"
    apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/serializer"
)

var AddToSchemes runtime.SchemeBuilder
var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init(){
    AddToSchemes = append(AddToSchemes, apiextensionsv1.AddToScheme)
    AddToSchemes = append(AddToSchemes, frameworksapis.AddToScheme)
    AddToScheme(Scheme)
}

func AddToScheme(s *runtime.Scheme) error {
    return AddToSchemes.AddToScheme(s)
}