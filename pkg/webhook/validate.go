package webhook

import (
    "context"
    frameworkClient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
    "github.com/woodliu/k8sopa/pkg/target"
    admissionv1 "k8s.io/api/admission/v1"
    k8serrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "net/http"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
    "stash.weimob.com/devops/go_common/log"
    "strings"
)

type WebhookServer struct {
    k8sClientSet *kubernetes.Clientset
    frameworkClient              *frameworkClient.Client
}

func (whsvr *WebhookServer) validate(ar *admissionv1.AdmissionReview) (admission.Response,error) {
    review := &target.AugmentedReview{AdmissionRequest: ar.Request}
    if ar.Request.Namespace != "" {
        ns,err := whsvr.k8sClientSet.CoreV1().Namespaces().Get(context.TODO(),ar.Request.Namespace,metav1.GetOptions{})
        if !k8serrors.IsNotFound(err) {
            return nil, err
        }

        review.Namespace = ns
    }

    resp, err := whsvr.frameworkClient.Review(context.Background(), review)
    if err != nil {
        log.Errorf("review request kind:%s namesapce:%s name:%s err:%s",review.AdmissionRequest.RequestKind.String(),review.AdmissionRequest.Namespace,review.AdmissionRequest.Name,err)
        return admission.Errored(http.StatusInternalServerError, err), err
    }


    res := resp.Results()
    denyMsgs, warnMsgs := h.getValidationMessages(res, &req)

    if len(denyMsgs) > 0 {
        requestResponse = denyResponse
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

    requestResponse = allowResponse
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