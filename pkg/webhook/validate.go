package webhook

import (
	"context"
	"fmt"
	rtypes "github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/woodliu/k8sopa/pkg/register"
	"github.com/woodliu/k8sopa/pkg/target"
	admissionv1 "k8s.io/api/admission/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"stash.weimob.com/devops/go_common/log"
	"stash.weimob.com/devops/go_common/stack"
	"strings"
)

// httpStatusWarning is the HTTP return code for displaying warning messages in admission webhook (supported in Kubernetes v1.19+)
// https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#response
const httpStatusWarning = 299

func (svr *Server) validate(ar *admissionv1.AdmissionReview) admission.Response {
	review := &target.AugmentedReview{AdmissionRequest: ar.Request}
	if ar.Request.Namespace != "" {
		ns, err := svr.k8sClientSet.CoreV1().Namespaces().Get(context.TODO(), ar.Request.Namespace, metav1.GetOptions{})
		if !k8serrors.IsNotFound(err) {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		review.Namespace = ns
	}

	resp, err := svr.frameworkClient.Review(context.Background(), review)
	if err != nil {
		log.Errorf("review request kind:%s namesapce:%s name:%s err:%s", review.AdmissionRequest.RequestKind.String(), review.AdmissionRequest.Namespace, review.AdmissionRequest.Name, err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	res := resp.Results()

	denyMsgs, warnMsgs := getValidationMessages(res, ar.Request)
	if len(denyMsgs) > 0 {
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Reason:  metav1.StatusReasonForbidden,
					Code:    http.StatusForbidden,
					Message: strings.Join(denyMsgs, "\n"),
				},
				Warnings: warnMsgs,
			},
		}
	}

	vResp := admission.Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Code: http.StatusOK,
			},
			Warnings: warnMsgs,
		},
	}
	if len(warnMsgs) > 0 {
		vResp.Result.Code = httpStatusWarning
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
