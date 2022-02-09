package validate

import (
	"context"
	"encoding/json"
	"fmt"
	rtypes "github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/pkg/errors"
	"github.com/woodliu/k8sopa/pkg/client"
	"github.com/woodliu/k8sopa/pkg/register"
	"github.com/woodliu/k8sopa/pkg/target"
	admissionv1 "k8s.io/api/admission/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"stash.weimob.com/devops/go_common/log"
	"stash.weimob.com/devops/go_common/stack"
	"strings"
	"unsafe"
)

type Chan struct {
	Request  chan []byte
	Response chan string
}

type validate struct {
	client *client.Client
	ch              *Chan
}

func New(ctx context.Context, host, token string)*validate {
	client, err := client.NewClient(host, token)
	if nil != err{
		panic("init client failed")
	}

	vldt := &validate{
		client: client,
		ch: &Chan{
			Request: make(chan []byte),
			Response: make(chan string),
		},
	}

	vldt.start(ctx)
	return vldt
}

func (vldt *validate) stop(){
	close(vldt.ch.Request)
}

func (vldt *validate) start(ctx context.Context){
	vldt.client.StartConstraintTmplInformer()

	go func() {
		for {
			select {
			case bytes := <- vldt.ch.Request:
				var resp Response
				object, err := runtime.Decode(unstructured.UnstructuredJSONScheme, bytes)
				if err != nil {
					resp = Errored(err)
					goto END
				}

				{
					unstructuredObj, ok := object.(*unstructured.Unstructured)
					if !ok {
						resp = Errored(errors.New("unstructured.Unstructured expected"))
						goto END
					}

					ar := &admissionv1.AdmissionRequest{
						Kind: metav1.GroupVersionKind{
							Group:   unstructuredObj.GroupVersionKind().Group,
							Version: unstructuredObj.GroupVersionKind().Version,
							Kind:    unstructuredObj.GroupVersionKind().Kind,
						},
						Object: runtime.RawExtension{
							Raw: bytes,
						},
					}
					resp = vldt.evaluate(ar)
				}

				END:
					jsonResp, _ := json.Marshal(resp)
					vldt.ch.Response <- *(*string)(unsafe.Pointer(&jsonResp))
			case <-ctx.Done():
				vldt.client.StopConstraintTmplInformer()
				vldt.stop()
			}
		}
	}()
}

func (vldt *validate) Validate(req []byte)string{ //TODO:request和respond应该是同一个结构体？
	vldt.ch.Request <- req
	return <-vldt.ch.Response
}

func (vldt *validate) evaluate(ar *admissionv1.AdmissionRequest) Response {
	review := &target.AugmentedReview{AdmissionRequest: ar}
	if ar.Namespace != "" {
		ns, err := vldt.client.K8sClientSet.CoreV1().Namespaces().Get(context.TODO(), ar.Namespace, metav1.GetOptions{})
		if !k8serrors.IsNotFound(err) {
			return Errored(err)
		}
		review.Namespace = ns
	}

	resp, err := vldt.client.FrameworkClient.Review(context.Background(), review)
	if err != nil {
		log.Errorf("review request kind:%s namesapce:%s name:%s err:%s", review.AdmissionRequest.RequestKind.String(), review.AdmissionRequest.Namespace, review.AdmissionRequest.Name, err)
		return Errored(err)
	}

	res := resp.Results()

	denyMsgs, warnMsgs := getValidationMessages(res, ar)
	if len(denyMsgs) > 0 {
		return Response{
			Allowed: false,
			Result: &metav1.Status{
				Reason:  metav1.StatusReasonForbidden,
				Message: strings.Join(denyMsgs, "\n"),
			},
			Warnings: warnMsgs,
		}
	}

	vResp := Response{
		Allowed: true,
		Warnings: warnMsgs,
	}

	return vResp
}

func getValidationMessages(res []*rtypes.Result, req *admissionv1.AdmissionRequest) ([]string, []string) {
	var denyMsgs, warnMsgs []string
	var resourceName string
	if len(res) > 0  {
		resourceName = req.Name
		if len(resourceName) == 0 && req.Object.Raw != nil {
			// On a CREATE operation, the client may omit name and
			// rely on the server to generate the name.
			obj := &unstructured.Unstructured{}
			if _, _, err := register.Deserializer.Decode(req.Object.Raw, nil, obj); err == nil {
				resourceName = obj.GetName()
			}
		}
	}
	for _, r := range res {
		if err := ValidateEnforcementAction(EnforcementAction(r.EnforcementAction)); err != nil {
			continue
		}

		logField := stack.FromKVs(map[string]interface{}{
			Process:"admission",
			EventType: "violation",
			ConstraintName: r.Constraint.GetName(),
			ConstraintGroup: r.Constraint.GroupVersionKind().Group,
			ConstraintAPIVersion: r.Constraint.GroupVersionKind().Version,
			ConstraintKind: r.Constraint.GetKind(),
			ConstraintAction: r.EnforcementAction,
			ResourceGroup: req.Kind.Group,
			ResourceAPIVersion: req.Kind.Version,
			ResourceKind: req.Kind.Kind,
			ResourceNamespace: req.Namespace,
			ResourceName: resourceName,
			RequestUsername: req.UserInfo.Username,
		})
		log.WithFields(logField).Info("denied admission")

		if r.EnforcementAction == string(Deny) {
			denyMsgs = append(denyMsgs, fmt.Sprintf("[%s] %s", r.Constraint.GetName(), r.Msg))
		}

		if r.EnforcementAction == string(Warn) {
			warnMsgs = append(warnMsgs, fmt.Sprintf("[%s] %s", r.Constraint.GetName(), r.Msg))
		}
	}
	return denyMsgs, warnMsgs
}
