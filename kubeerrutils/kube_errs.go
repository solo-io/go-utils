package kubeerrutils

import (
	"strings"

	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/validation"
)

const ObjectIsBeingDeletedErrorMsg = "object is being deleted"

func IsImmutableErr(err error) bool {
	if err != nil {
		return kubeerrs.IsInvalid(err) && strings.Contains(err.Error(), validation.FieldImmutableErrorMsg)
	}
	return false
}

// is the error AlreadyExists && the resource is not terminating?
func IsAlreadyExists(err error) bool {
	if err != nil {
		return kubeerrs.IsAlreadyExists(err) && !strings.Contains(err.Error(), ObjectIsBeingDeletedErrorMsg)
	}
	return false

}
