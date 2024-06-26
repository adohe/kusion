package generators

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
)

type namespaceGenerator struct {
	namespace string
}

func NewNamespaceGenerator(namespace string) (modules.Generator, error) {
	return &namespaceGenerator{
		namespace: namespace,
	}, nil
}

func NewNamespaceGeneratorFunc(namespace string) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewNamespaceGenerator(namespace)
	}
}

func (g *namespaceGenerator) Generate(i *v1.Spec) error {
	if i.Resources == nil {
		i.Resources = make(v1.Resources, 0)
	}

	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{Name: g.namespace},
	}

	// Avoid generating duplicate namespaces with the same ID.
	id := modules.KubernetesResourceID(ns.TypeMeta, ns.ObjectMeta)
	for _, res := range i.Resources {
		if res.ID == id {
			return nil
		}
	}

	return modules.AppendToSpec(v1.Kubernetes, id, i, ns)
}
