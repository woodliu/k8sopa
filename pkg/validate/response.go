package validate

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Response is the output of an admission handler.
// It contains a response indicating if a given
// operation is allowed, as well as a set of patches
// to mutate the object in the case of a mutating admission handler.
type Response struct {
	Allowed bool
	Result *metav1.Status
	Warnings []string
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

// Errored creates a new Response for error-handling a request.
func Errored(err error) Response {
	return Response{
		Allowed: false,
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// ValidationResponse returns a response for admitting a request.
func ValidationResponse(allowed bool, reason string) Response {
	resp := Response{
		Allowed: allowed,
	}
	if len(reason) > 0 {
		resp.Result.Reason = metav1.StatusReason(reason)
	}
	return resp
}
