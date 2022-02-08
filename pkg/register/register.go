package register

import (
	frameworksapis "github.com/open-policy-agent/frameworks/constraint/pkg/apis"
	v1 "k8s.io/api/admission/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var AddToSchemes runtime.SchemeBuilder
var Scheme         = runtime.NewScheme()
var Codecs         = serializer.NewCodecFactory(Scheme)
var Deserializer   = Codecs.UniversalDeserializer()
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	AddToSchemes = append(AddToSchemes, v1.AddToScheme)
	AddToSchemes = append(AddToSchemes, apiextensionsv1.AddToScheme)
	AddToSchemes = append(AddToSchemes, frameworksapis.AddToScheme)
	utilruntime.Must(AddToScheme(Scheme))
}

func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
