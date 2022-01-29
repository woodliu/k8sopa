package informer

import (
    "context"
    "github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
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

var constraintTemplateGvr = schema.GroupVersionResource{Group: "templates.gatekeeper.sh", Version: "v1",Resource: "constrainttemplates"}

type constraintTemplate struct {
    informer *Informer
}

func (ct *constraintTemplate)OnAdd(obj interface{}){
    var tmpl templates.ConstraintTemplate
    tmplUnstructured := obj.(*unstructured.Unstructured)
    err := runtime.DefaultUnstructuredConverter.FromUnstructured(tmplUnstructured.Object, &tmpl)
    if nil != err {
        log.Errorf("convert constraintTemplate from unstructured err:%s",err.Error())
        return
    }

    // 添加constraint template
    if _,err := ct.informer.FrameworkClient.AddTemplate(&tmpl);nil != err{
        log.Errorf("add constraint template %s err:%s", tmpl.GetName(), err.Error())
        return
    }

    // 生成constraint crd对象
    crd,err := ct.informer.FrameworkClient.CreateCRD(&tmpl)
    if nil != err {
        log.Errorf("create crd from constraint template %s err:", tmpl.GetName(), err.Error())
        ct.informer.FrameworkClient.RemoveTemplate(context.TODO(),&tmpl)
        return
    }

    // 创建constraint crd
    var  proposedCRD apiextensionsv1.CustomResourceDefinition
    Scheme.Convert(crd, proposedCRD, nil)
    err = ct.informer.ApiExtensionsV1Client.Post().
        Resource(crdResource).
        VersionedParams(&metav1.CreateOptions{}, ParameterCodec).
        Body(&proposedCRD).
        Do(context.TODO()).
        Error()
    if nil != err {
        log.Errorf("create constraint crd %s err:%s",proposedCRD.GetName(),err.Error())
        ct.informer.FrameworkClient.RemoveTemplate(context.TODO(),&tmpl)
        return
    }

    // 启动该constraint crd的informer
    constraintGvr := schema.GroupVersionResource{
        Group: constraintGroup,
        Version: constraintVersion,
        Resource: proposedCRD.GetName(),
    }
    ct.informer.ConstraintStopCh[constraintGvr] = make(chan struct{})
    ct.informer.startConstraintInformer(constraintGvr)

    return
}

// TODO:是否支持update，不支持的话禁用
func (ct *constraintTemplate)OnUpdate(oldObj, newObj interface{}){
    log.Warnf("not implement！")
}

func (ct *constraintTemplate)OnDelete(obj interface{}){
    var tmpl templates.ConstraintTemplate
    tmplUnstructured := obj.(*unstructured.Unstructured)
    err := runtime.DefaultUnstructuredConverter.FromUnstructured(tmplUnstructured.Object, &tmpl)
    if nil != err {
        log.Errorf("convert constraintTemplate from unstructured err:%s",err.Error())
        return
    }

    _,err = ct.informer.FrameworkClient.RemoveTemplate(context.TODO(),&tmpl)
    if nil != err {
        log.Errorf("remove constraint template %s err:%s",tmpl.GetName(),err.Error())
        return
    }

    // 停止该constraint的informer
    ct.informer.stopConstraintInformer(schema.GroupVersionResource{Group: constraintGroup, Version: constraintVersion, Resource: tmpl.GetName()})

    // 删除该constraint template的constraint crd
    ct.informer.ApiExtensionsV1Client.Delete().
        Resource(crdResource).
        Name(tmpl.GetName()).
        Body(&metav1.DeleteOptions{}).
        Do(context.TODO()).
        Error()

    // TODO:需要在webhook中对constraint和constrainttemplate做权限管控
}


