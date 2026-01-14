package main

import (
	"context"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// InternalLoadBalancerSpec defines the desired state of InternalLoadBalancer
type InternalLoadBalancerSpec struct {
	DeploymentName string `json:"deploymentName"`
	TraefikAPIURL  string `json:"traefikApiUrl"`
	Image          string `json:"image"`
}

// InternalLoadBalancer is the Schema for the internal load balancer CRD
type InternalLoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              InternalLoadBalancerSpec `json:"spec,omitempty"`
}

type InternalLoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InternalLoadBalancer `json:"items"`
}

func (in *InternalLoadBalancer) DeepCopyObject() runtime.Object {
	out := new(InternalLoadBalancer)
	*out = *in
	return out
}
func (in *InternalLoadBalancerList) DeepCopyObject() runtime.Object {
	out := new(InternalLoadBalancerList)
	*out = *in
	return out
}

// Reconciler
type ILBReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ILBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var ilb InternalLoadBalancer
	if err := r.Get(ctx, req.NamespacedName, &ilb); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deployName := ilb.Spec.DeploymentName
	image := ilb.Spec.Image
	if image == "" {
		image = "yourrepo/k8s-internal-loadbalancer:latest"
	}
	traefikAPIURL := ilb.Spec.TraefikAPIURL

	// Get the target deployment and its selector
	var dep appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: ilb.Namespace, Name: deployName}, &dep); err != nil {
		return ctrl.Result{}, err
	}
	selector := dep.Spec.Selector.MatchLabels

	// Create the load balancer deployment
	ilbDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ilb.Name + "-ilb",
			Namespace: ilb.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": ilb.Name + "-ilb"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": ilb.Name + "-ilb"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "ilb",
						Image: image,
						Env: []corev1.EnvVar{
							{Name: "POD_LABELS", Value: mapToSelector(selector)},
							{Name: "TRAEFIK_API_URL", Value: traefikAPIURL},
							{Name: "POD_NAMESPACE", ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
							}},
						},
					}},
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(&ilb, ilbDeploy, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	err := r.Client.Patch(ctx, ilbDeploy, client.Apply, client.ForceOwnership, client.FieldOwner("ilb-operator"))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Service (ClusterIP)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ilb.Name + "-ilb",
			Namespace: ilb.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": ilb.Name + "-ilb"},
			Ports: []corev1.ServicePort{{
				Name:       "tcp",
				Port:       3333,
				TargetPort: intstrFromInt(3333),
				Protocol:   corev1.ProtocolTCP,
			}},
		},
	}
	if err := ctrl.SetControllerReference(&ilb, svc, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	err = r.Client.Patch(ctx, svc, client.Apply, client.ForceOwnership, client.FieldOwner("ilb-operator"))
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func int32Ptr(i int32) *int32 { return &i }
func intstrFromInt(i int) intstr.IntOrString {
	return intstr.IntOrString{Type: intstr.Int, IntVal: int32(i)}
}
func mapToSelector(m map[string]string) string {
	// Converts map to "key1=value1,key2=value2"
	var s []string
	for k, v := range m {
		s = append(s, k+"="+v)
	}
	return strings.Join(s, ",")
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	scheme := runtime.NewScheme()
	_ = apiextv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	scheme.AddKnownTypes(schema.GroupVersion{Group: "ilb.example.com", Version: "v1"}, &InternalLoadBalancer{}, &InternalLoadBalancerList{})

	mgr, err := manager.New(ctrl.GetConfigOrDie(), manager.Options{
		Scheme: scheme,
	})
	if err != nil {
		os.Exit(1)
	}

	if err = controller.New("ilb-controller", mgr, controller.Options{
		Reconciler: &ILBReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
	}); err != nil {
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
