package graph

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/third_party/terraform/dag"
)

func TestResourceNode_Execute(t *testing.T) {
	type fields struct {
		BaseNode baseNode
		Action   models.ActionType
		state    *apiv1.Resource
	}
	type args struct {
		operation models.Operation
	}

	const Jack = "jack"
	const Pony = "pony"
	const Eric = "eric"
	mf := &apiv1.Spec{Resources: []apiv1.Resource{
		{
			ID:   Pony,
			Type: runtime.Kubernetes,
			Attributes: map[string]interface{}{
				"c": "d",
			},
			DependsOn: []string{Jack},
		},
		{
			ID:   Eric,
			Type: runtime.Kubernetes,
			Attributes: map[string]interface{}{
				"a": ImplicitRefPrefix + "jack.a.b",
			},
			DependsOn: []string{Pony},
		},
		{
			ID:   Jack,
			Type: runtime.Kubernetes,
			Attributes: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c",
				},
			},
			DependsOn: nil,
		},
	}}

	priorStateResourceIndex := map[string]*apiv1.Resource{}
	for i, resource := range mf.Resources {
		priorStateResourceIndex[resource.ResourceKey()] = &mf.Resources[i]
	}

	newResourceState := &apiv1.Resource{
		ID:   Eric,
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"a": ImplicitRefPrefix + "jack.a.b",
		},
		DependsOn: []string{Pony},
	}

	illegalResourceState := &apiv1.Resource{
		ID:   Eric,
		Type: runtime.Kubernetes,
		Attributes: map[string]interface{}{
			"a": ImplicitRefPrefix + "jack.notExist",
		},
		DependsOn: []string{Pony},
	}

	graph := &dag.AcyclicGraph{}
	graph.Add(&RootNode{})

	tests := []struct {
		name   string
		fields fields
		args   args
		want   v1.Status
	}{
		{
			name: "update",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   models.Update,
				state:    newResourceState,
			},
			args: args{operation: models.Operation{
				OperationType:           models.Apply,
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				IgnoreFields:            []string{"not_exist_field"},
				MsgCh:                   make(chan models.Message),
				Lock:                    &sync.Mutex{},
				RuntimeMap:              map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
				Release:                 &apiv1.Release{},
			}},
			want: nil,
		},
		{
			name: "delete",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   models.Delete,
				state:    newResourceState,
			},
			args: args{operation: models.Operation{
				OperationType:           models.Apply,
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan models.Message),
				Lock:                    &sync.Mutex{},
				RuntimeMap:              map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
			}},
			want: nil,
		},
		{
			name: "illegalRef",
			fields: fields{
				BaseNode: baseNode{ID: Jack},
				Action:   models.Update,
				state:    illegalResourceState,
			},
			args: args{operation: models.Operation{
				OperationType:           models.Apply,
				CtxResourceIndex:        priorStateResourceIndex,
				PriorStateResourceIndex: priorStateResourceIndex,
				StateResourceIndex:      priorStateResourceIndex,
				MsgCh:                   make(chan models.Message),
				Lock:                    &sync.Mutex{},
				RuntimeMap:              map[apiv1.Type]runtime.Runtime{runtime.Kubernetes: &kubernetes.KubernetesRuntime{}},
			}},
			want: v1.NewErrorStatusWithMsg(v1.IllegalManifest, "can't find specified value in resource:jack by ref:jack.notExist"),
		},
	}
	for _, tt := range tests {
		mockey.PatchConvey(tt.name, t, func() {
			rn := &ResourceNode{
				baseNode: &tt.fields.BaseNode,
				Action:   tt.fields.Action,
				resource: tt.fields.state,
			}
			mockey.Mock(mockey.GetMethod(tt.args.operation.RuntimeMap[runtime.Kubernetes], "Apply")).To(
				func(k *kubernetes.KubernetesRuntime, ctx context.Context, request *runtime.ApplyRequest) *runtime.ApplyResponse {
					mockState := *newResourceState
					mockState.Attributes["a"] = "c"
					return &runtime.ApplyResponse{
						Resource: &mockState,
					}
				}).Build()
			mockey.Mock(mockey.GetMethod(tt.args.operation.RuntimeMap[runtime.Kubernetes], "Delete")).To(
				func(k *kubernetes.KubernetesRuntime, ctx context.Context, request *runtime.DeleteRequest) *runtime.DeleteResponse {
					return &runtime.DeleteResponse{Status: nil}
				}).Build()
			mockey.Mock(mockey.GetMethod(tt.args.operation.RuntimeMap[runtime.Kubernetes], "Read")).To(
				func(k *kubernetes.KubernetesRuntime, ctx context.Context, request *runtime.ReadRequest) *runtime.ReadResponse {
					return &runtime.ReadResponse{Resource: request.PriorResource}
				}).Build()
			mockey.Mock((*models.Operation).UpdateReleaseState).Return(nil).Build()

			assert.Equalf(t, tt.want, rn.Execute(&tt.args.operation), "Execute(%v)", tt.args.operation)
		})
	}
}

func Test_removeNestedField(t *testing.T) {
	t.Run("remove nested field", func(t *testing.T) {
		e1 := []interface{}{
			map[string]interface{}{"f": "f1", "g": "g1"},
		}
		e2 := []interface{}{
			map[string]interface{}{"f": "f2", "g": "g2"},
		}

		c := []interface{}{
			map[string]interface{}{"d": "d1", "e": e1},
			map[string]interface{}{"d": "d2", "e": e2},
		}

		a := map[string]interface{}{
			"b": 1,
			"c": c,
		}

		obj := map[string]interface{}{
			"a": a,
		}

		removeNestedField(obj, "a", "c", "e", "f")
		assert.Len(t, e1[0], 1)
		assert.Len(t, e2[0], 1)

		removeNestedField(obj, "a", "c", "e", "g")
		assert.Empty(t, e1[0])
		assert.Empty(t, e2[0])

		removeNestedField(obj, "a", "c", "e")
		assert.Len(t, c[0], 1)
		assert.Len(t, c[1], 1)

		removeNestedField(obj, "a", "c", "d")
		assert.Len(t, c[0], 0)
		assert.Len(t, c[1], 0)

		removeNestedField(obj, "a", "c")
		assert.Len(t, a, 1)

		removeNestedField(obj, "a", "b")
		assert.Len(t, a, 0)

		removeNestedField(obj, "a")
		assert.Empty(t, obj)
	})

	t.Run("remove spec.ports.targetPort", func(t *testing.T) {
		ports := []interface{}{
			map[string]interface{}{
				"port":       80,
				"protocol":   "TCP",
				"targetPort": 80,
			},
		}

		spec := map[string]interface{}{
			"clusterIP": "172.16.128.40",
			"ports":     ports,
		}

		obj := map[string]interface{}{
			"spec": spec,
		}

		removeNestedField(obj, "spec", "ports", "targetPort")
		assert.Len(t, ports[0], 2)
	})
}

func TestParseExternalSecretDataRef(t *testing.T) {
	tests := []struct {
		name       string
		dataRefStr string
		want       *apiv1.ExternalSecretRef
		wantErr    bool
	}{
		{
			name:       "invalid data ref string",
			dataRefStr: "$%#//invalid",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "only secret name",
			dataRefStr: "ref://secret-name",
			want: &apiv1.ExternalSecretRef{
				Name: "secret-name",
			},
			wantErr: false,
		},
		{
			name:       "secret name with version",
			dataRefStr: "ref://secret-name?version=1",
			want: &apiv1.ExternalSecretRef{
				Name:    "secret-name",
				Version: "1",
			},
			wantErr: false,
		},
		{
			name:       "secret name with property and version",
			dataRefStr: "ref://secret-name/property?version=1",
			want: &apiv1.ExternalSecretRef{
				Name:     "secret-name",
				Property: "property",
				Version:  "1",
			},
			wantErr: false,
		},
		{
			name:       "nested secret name with property and version",
			dataRefStr: "ref://customer/acme/customer_name?version=1",
			want: &apiv1.ExternalSecretRef{
				Name:     "customer/acme",
				Property: "customer_name",
				Version:  "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExternalSecretDataRef(tt.dataRefStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseExternalSecretDataRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseExternalSecretDataRef() got = %v, want %v", got, tt.want)
			}
		})
	}
}
