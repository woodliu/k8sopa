package client

import (
	"context"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/woodliu/k8sopa/pkg/register"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"stash.weimob.com/devops/go_common/log"
)

const (
	crdResource = "customresourcedefinitions"
)

var constraintTemplateGvr = schema.GroupVersionResource{Group: "templates.gatekeeper.sh", Version: "v1", Resource: "constrainttemplates"}

type constraintTemplate struct {
	client *Client
}

func (ct *constraintTemplate) OnAdd(obj interface{}) {
	var tmpl templates.ConstraintTemplate
	tmplUnstructured := obj.(*unstructured.Unstructured)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(tmplUnstructured.Object, &tmpl)
	if nil != err {
		log.Errorf("convert constraintTemplate from unstructured err:%s", err.Error())
		return
	}

	// 添加constraint template
	if _, err := ct.client.FrameworkClient.AddTemplate(&tmpl); nil != err {
		log.Errorf("add constraint template %s err:%s", tmpl.GetName(), err.Error())
		return
	}

	// 生成constraint crd对象
	crd, err := ct.client.FrameworkClient.CreateCRD(&tmpl)
	if nil != err {
		log.Errorf("create crd from constraint template %s err:", tmpl.GetName(), err.Error())
		ct.client.FrameworkClient.RemoveTemplate(context.TODO(), &tmpl)
		return
	}

	// 创建constraint crd
	var proposedCRD apiextensionsv1.CustomResourceDefinition
	err = register.Scheme.Convert(crd, proposedCRD, nil)
	if nil != err {
		log.Errorf("Convert constraint crd %s err:", crd.GetName(), err.Error())
		ct.client.FrameworkClient.RemoveTemplate(context.TODO(), &tmpl)
		return
	}

	err = ct.client.ApiExtensionsV1Client.Post().
		Resource(crdResource).
		VersionedParams(&metav1.CreateOptions{}, register.ParameterCodec).
		Body(&proposedCRD).
		Do(context.TODO()).
		Error()
	if nil != err {
		log.Errorf("create constraint crd %s err:%s", proposedCRD.GetName(), err.Error())
		ct.client.FrameworkClient.RemoveTemplate(context.TODO(), &tmpl)
		return
	}

	// 启动该constraint crd的informer
	constraintGvr := schema.GroupVersionResource{
		Group:    constraintGroup,
		Version:  constraintVersion,
		Resource: proposedCRD.GetName(),
	}
	ct.client.ConstraintStopCh[constraintGvr] = make(chan struct{})
	ct.client.startConstraintInformer(constraintGvr)

	return
}

// OnUpdate TODO:是否支持update，不支持的话禁用
func (ct *constraintTemplate) OnUpdate(oldObj, newObj interface{}) {
	log.Warnf("not implement！")
}

func (ct *constraintTemplate) OnDelete(obj interface{}) {
	var tmpl templates.ConstraintTemplate
	tmplUnstructured := obj.(*unstructured.Unstructured)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(tmplUnstructured.Object, &tmpl)
	if nil != err {
		log.Errorf("convert constraintTemplate from unstructured err:%s", err.Error())
		return
	}

	_, err = ct.client.FrameworkClient.RemoveTemplate(context.TODO(), &tmpl)
	if nil != err {
		log.Errorf("remove constraint template %s err:%s", tmpl.GetName(), err.Error())
		return
	}

	// 停止该constraint的informer
	ct.client.stopConstraintInformer(schema.GroupVersionResource{Group: constraintGroup, Version: constraintVersion, Resource: tmpl.GetName()})

	// 删除该constraint template的constraint crd
	ct.client.ApiExtensionsV1Client.Delete().
		Resource(crdResource).
		Name(tmpl.GetName()).
		Body(&metav1.DeleteOptions{}).
		Do(context.TODO()).
		Error()

	// TODO:需要在webhook中对constraint和constrainttemplate做权限管控
}
