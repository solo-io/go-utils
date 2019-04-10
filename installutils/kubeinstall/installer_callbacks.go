package kubeinstall

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CallbackOptions interface {
	PreInstall() error
	PostInstall() error
	PreCreate(res *unstructured.Unstructured) error
	PostCreate(res *unstructured.Unstructured) error
	PreUpdate(res *unstructured.Unstructured) error
	PostUpdate(res *unstructured.Unstructured) error
	PreDelete(res *unstructured.Unstructured) error
	PostDelete(res *unstructured.Unstructured) error
}

type CallbackOption struct {
	OnPreInstall  func() error
	OnPostInstall func() error
	OnPreCreate   func(res *unstructured.Unstructured) error
	OnPostCreate  func(res *unstructured.Unstructured) error
	OnPreUpdate   func(res *unstructured.Unstructured) error
	OnPostUpdate  func(res *unstructured.Unstructured) error
	OnPreDelete   func(res *unstructured.Unstructured) error
	OnPostDelete  func(res *unstructured.Unstructured) error
}

func (cb *CallbackOption) PreInstall() error {
	if cb.OnPreInstall != nil {
		return cb.OnPreInstall()
	}
	return nil
}

func (cb *CallbackOption) PostInstall() error {
	if cb.OnPostInstall != nil {
		return cb.OnPostInstall()
	}
	return nil
}

func (cb *CallbackOption) PreCreate(res *unstructured.Unstructured) error {
	if cb.OnPreCreate != nil {
		return cb.OnPreCreate(res)
	}
	return nil
}

func (cb *CallbackOption) PostCreate(res *unstructured.Unstructured) error {
	if cb.OnPostCreate != nil {
		return cb.OnPostCreate(res)
	}
	return nil
}

func (cb *CallbackOption) PreUpdate(res *unstructured.Unstructured) error {
	if cb.OnPreUpdate != nil {
		return cb.OnPreUpdate(res)
	}
	return nil
}

func (cb *CallbackOption) PostUpdate(res *unstructured.Unstructured) error {
	if cb.OnPostUpdate != nil {
		return cb.OnPostUpdate(res)
	}
	return nil
}

func (cb *CallbackOption) PreDelete(res *unstructured.Unstructured) error {
	if cb.OnPreDelete != nil {
		return cb.OnPreDelete(res)
	}
	return nil
}

func (cb *CallbackOption) PostDelete(res *unstructured.Unstructured) error {
	if cb.OnPostDelete != nil {
		return cb.OnPostDelete(res)
	}
	return nil
}

func initCallbacks() []CallbackOptions {
	return []CallbackOptions{
		&CallbackOption{OnPreCreate: setInstallationAnnotation},
	}
}
