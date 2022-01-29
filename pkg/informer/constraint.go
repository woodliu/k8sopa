package informer

import (
    "context"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "stash.weimob.com/devops/go_common/log"
)

type constraint struct {
    informer *Informer
}

const (
    constraintGroup = "constraints.gatekeeper.sh"
    constraintVersion = "v1beta1"
)

func (c *constraint)OnAdd(obj interface{}){
    constraintUnstructured := obj.(*unstructured.Unstructured)
    _,err := c.informer.FrameworkClient.AddConstraint(context.TODO(), constraintUnstructured)
    if nil != err {
        log.Errorf("add constraint %s err:%s",constraintUnstructured.GetName(), err.Error())
        return
    }
}

func (c *constraint)OnUpdate(oldObj, newObj interface{}){
    log.Warnf("not implement！")
}

func (c *constraint)OnDelete(obj interface{}){
    constraintUnstructured := obj.(*unstructured.Unstructured)
    _,err := c.informer.FrameworkClient.RemoveConstraint(context.TODO(), constraintUnstructured)
    if nil != err {
        log.Errorf("remove constraint %s err:%s",constraintUnstructured.GetName(), err.Error())
        return
    }
}