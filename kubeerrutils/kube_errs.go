package kubeerrutils

import (
	"strings"

	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/validation"
)

func IsImmutableErr(err error) bool {
	if err != nil {
		return kubeerrs.IsInvalid(err) && strings.Contains(err.Error(), validation.FieldImmutableErrorMsg)
	}
	return false
}
