package webhook

import (
    "net/http"

    jsonpatch "gomodules.xyz/jsonpatch/v2"
    admissionv1 "k8s.io/api/admission/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Response is the output of an admission handler.
// It contains a response indicating if a given
// operation is allowed, as well as a set of patches
// to mutate the object in the case of a mutating admission handler.
type Response struct {
    // Patches are the JSON patches for mutating webhooks.
    // Using this instead of setting Response.Patch to minimize
    // overhead of serialization and deserialization.
    // Patches set here will override any patches in the response,
    // so leave this empty if you want to set the patch response directly.
    Patches []jsonpatch.JsonPatchOperation
    // AdmissionResponse is the raw admission response.
    // The Patch field in it will be overwritten by the listed patches.
    admissionv1.AdmissionResponse
}

// Allowed constructs a response indicating that the given operation
// is allowed (without any patches).
func Allowed(reason string) Response {
    return ValidationResponse(true, reason)
}

// Denied constructs a response indicating that the given operation
// is not allowed.
func Denied(reason string) Response {
    return ValidationResponse(false, reason)
}

// Patched constructs a response indicating that the given operation is
// allowed, and that the target object should be modified by the given
// JSONPatch operations.
func Patched(reason string, patches ...jsonpatch.JsonPatchOperation) Response {
    resp := Allowed(reason)
    resp.Patches = patches

    return resp
}

// Errored creates a new Response for error-handling a request.
func Errored(code int32, err error) Response {
    return Response{
        AdmissionResponse: admissionv1.AdmissionResponse{
            Allowed: false,
            Result: &metav1.Status{
                Code:    code,
                Message: err.Error(),
            },
        },
    }
}

// ValidationResponse returns a response for admitting a request.
func ValidationResponse(allowed bool, reason string) Response {
    code := http.StatusForbidden
    if allowed {
        code = http.StatusOK
    }
    resp := Response{
        AdmissionResponse: admissionv1.AdmissionResponse{
            Allowed: allowed,
            Result: &metav1.Status{
                Code: int32(code),
            },
        },
    }
    if len(reason) > 0 {
        resp.Result.Reason = metav1.StatusReason(reason)
    }
    return resp
}

// PatchResponseFromRaw takes 2 byte arrays and returns a new response with json patch.
// The original object should be passed in as raw bytes to avoid the roundtripping problem
// described in https://github.com/kubernetes-sigs/kubebuilder/issues/510.
func PatchResponseFromRaw(original, current []byte) Response {
    patches, err := jsonpatch.CreatePatch(original, current)
    if err != nil {
        return Errored(http.StatusInternalServerError, err)
    }
    return Response{
        Patches: patches,
        AdmissionResponse: admissionv1.AdmissionResponse{
            Allowed: true,
            PatchType: func() *admissionv1.PatchType {
                if len(patches) == 0 {
                    return nil
                }
                pt := admissionv1.PatchTypeJSONPatch
                return &pt
            }(),
        },
    }
}

// validationResponseFromStatus returns a response for admitting a request with provided Status object.
func validationResponseFromStatus(allowed bool, status metav1.Status) Response {
    resp := Response{
        AdmissionResponse: admissionv1.AdmissionResponse{
            Allowed: allowed,
            Result:  &status,
        },
    }
    return resp
}

// WithWarnings adds the given warnings to the Response.
// If any warnings were already given, they will not be overwritten.
func (r Response) WithWarnings(warnings ...string) Response {
    r.AdmissionResponse.Warnings = append(r.AdmissionResponse.Warnings, warnings...)
    return r
}
