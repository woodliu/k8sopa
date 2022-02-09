package client

import (
	"context"
	"github.com/open-policy-agent/frameworks/constraint/pkg/apis/templates/v1beta1"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/woodliu/k8sopa/pkg/register"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

func getUnversionedTmpl(obj interface{})(*templates.ConstraintTemplate,error){
	var v1beta1Tmpl v1beta1.ConstraintTemplate
	var unversionedTmpl templates.ConstraintTemplate
	unstructuredTmpl := obj.(*unstructured.Unstructured)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredTmpl.Object, &v1beta1Tmpl)
	if nil != err {
		return nil,err
	}

	// 添加constraint template
	if err := register.Scheme.Convert(&v1beta1Tmpl, &unversionedTmpl, nil); err != nil {
		return nil,err
	}

	return &unversionedTmpl,nil
}

func (ct *constraintTemplate) OnAdd(obj interface{}) {
	unversionedTmpl,err := getUnversionedTmpl(obj)
	if nil != err{
		log.Errorf("get unversioned constraint template err:%s",err.Error())
		return
	}

	if _, err := ct.client.FrameworkClient.AddTemplate(unversionedTmpl); nil != err {
		log.Errorf("add constraint template %s err:%s", unversionedTmpl.GetName(), err.Error())
		return
	}

	// 生成constraint crd对象
	crd, err := ct.client.FrameworkClient.CreateCRD(unversionedTmpl)
	if nil != err {
		log.Errorf("create crd from constraint template %s err:", unversionedTmpl.GetName(), err.Error())
		ct.client.FrameworkClient.RemoveTemplate(context.TODO(), unversionedTmpl)
		return
	}

	// 创建constraint crd
	var proposedCRD apiextensionsv1.CustomResourceDefinition
	err = register.Scheme.Convert(crd, &proposedCRD, nil)
	if nil != err {
		log.Errorf("Convert constraint crd %s err:", crd.GetName(), err.Error())
		ct.client.FrameworkClient.RemoveTemplate(context.TODO(), unversionedTmpl)
		return
	}

	err = ct.client.ApiExtensionsV1Client.Post().
		Resource(crdResource).
		VersionedParams(&metav1.CreateOptions{}, register.ParameterCodec).
		Body(&proposedCRD).
		Do(context.TODO()).
		Error()
	if nil != err {
		if ins,ok := err.(*errors.StatusError);ok{
			if ins.ErrStatus.Reason == metav1.StatusReasonAlreadyExists{
				goto CONTINUE
			}
		}
		log.Errorf("create constraint crd %s err:%s", proposedCRD.Spec.Names.Plural, err.Error())
		ct.client.FrameworkClient.RemoveTemplate(context.TODO(), unversionedTmpl)
		return
	}

	CONTINUE:
	// 启动该constraint crd的informer
	constraintGvr := schema.GroupVersionResource{
		Group:    constraintGroup,
		Version:  constraintVersion,
		Resource: proposedCRD.Spec.Names.Plural,
	}

	ct.client.ConstraintStopCh[constraintGvr] = make(chan struct{})
	ct.client.startConstraintInformer(constraintGvr)

	return
}

func (ct *constraintTemplate) OnUpdate(oldObj, newObj interface{}) {
	oldTmpl,err := getUnversionedTmpl(oldObj)
	if nil != err{
		log.Errorf("get old unversioned constraint template err:%s",err.Error())
		return
	}

	newTmpl,err := getUnversionedTmpl(newObj)
	if nil != err{
		log.Errorf("get new unversioned constraint template err:%s",err.Error())
		return
	}

	_,err = ct.client.FrameworkClient.RemoveTemplate(context.TODO(), oldTmpl)
	if nil != err{
		log.Errorf("remove old constraint:%s when update err:%s",oldTmpl.GetName(),err.Error())
		return
	}

	ct.client.FrameworkClient.AddTemplate(newTmpl)
	if nil != err{
		log.Errorf("add updated constraint:%s when update err:%s",newTmpl.GetName(),err.Error())
		return
	}
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
