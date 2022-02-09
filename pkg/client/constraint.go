package client

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"stash.weimob.com/devops/go_common/log"
)

type constraint struct {
	client *Client
}

const (
	constraintGroup   = "constraints.gatekeeper.sh"
	constraintVersion = "v1beta1"
)

func (c *constraint) OnAdd(obj interface{}) {
	constraintUnstructured := obj.(*unstructured.Unstructured)
	_, err := c.client.FrameworkClient.AddConstraint(context.TODO(), constraintUnstructured)
	if nil != err {
		log.Errorf("add constraint %s err:%s", constraintUnstructured.GetName(), err.Error())
		return
	}
}

func (c *constraint) OnUpdate(oldObj, newObj interface{}) {
	oldConstraint := oldObj.(*unstructured.Unstructured)
	newConstraint := newObj.(*unstructured.Unstructured)
	_,err := c.client.FrameworkClient.RemoveConstraint(context.TODO(), oldConstraint)
	if nil != err{
		log.Errorf("remove old constraint:%s when update err:%s",oldConstraint.GetName(),err.Error())
		return
	}

	c.client.FrameworkClient.AddConstraint(context.TODO(), newConstraint)
	if nil != err{
		log.Errorf("add updated constraint:%s when update err:%s",newConstraint.GetName(),err.Error())
		return
	}
}

func (c *constraint) OnDelete(obj interface{}) {
	constraintUnstructured := obj.(*unstructured.Unstructured)
	_, err := c.client.FrameworkClient.RemoveConstraint(context.TODO(), constraintUnstructured)
	if nil != err {
		log.Errorf("remove constraint %s err:%s", constraintUnstructured.GetName(), err.Error())
		return
	}
}
