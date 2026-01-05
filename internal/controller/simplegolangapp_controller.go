/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/zooeymoon1989/kubernetesStudyOperatorSDK/api/v1alpha1"
)

const Image = "zooeymoon1989/simple-golang:latest"
const ContainerName = "simple-golang"
const ContainerPortName = "http"
const ContainerPort = 80

// SimpleGolangAppReconciler reconciles a SimpleGolangApp object
type SimpleGolangAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// RBAC（一定要加，不然会 403）
// +kubebuilder:rbac:groups=apps.osuk8s.site,resources=simplegolangapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.osuk8s.site,resources=simplegolangapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.osuk8s.site,resources=simplegolangapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimpleGolangApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *SimpleGolangAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	var cr appsv1alpha1.SimpleGolangApp

	if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// defaults
	replicas := int32(1)
	// 如果spec.replicas不为空，那么赋值给replicas
	if cr.Spec.Replicas != nil {
		replicas = *cr.Spec.Replicas
	}

	port := int32(ContainerPort)
	if cr.Spec.Port != nil {
		port = *cr.Spec.Port
	}

	image := cr.Spec.Image
	if image == "" {
		image = Image
	}

	//label
	labels := map[string]string{
		"app.kubernetes.io/name":    cr.Name,
		"app.kubernetes.io/part-of": ContainerName,
	}

	depName := cr.Name + "-deployment"
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: cr.Namespace,
		},
	}
	if err := ctrl.SetControllerReference(&cr, dep, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, dep, func() error {
		dep.Labels = labels
		dep.Spec.Replicas = &replicas
		dep.Spec.Template.Spec.Containers = []corev1.Container{}
		dep.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

		dep.Spec.Template.Spec.Containers = []corev1.Container{{
			Name:            ContainerName,
			Image:           Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports: []corev1.ContainerPort{{
				ContainerPort: ContainerPort,
				Name:          ContainerPortName,
			}},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/ping",
						Port: intstr.FromInt32(ContainerPort),
					},
				},
				InitialDelaySeconds: 5,
				PeriodSeconds:       10,
				TimeoutSeconds:      2,
				FailureThreshold:    3,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/ping",
						Port: intstr.FromInt32(ContainerPort),
					},
				},
				InitialDelaySeconds: 2,
				PeriodSeconds:       5,
				TimeoutSeconds:      2,
				FailureThreshold:    3,
			},
		}}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	// 3) Desired Service
	svcName := cr.Name + "-svc"
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: cr.Namespace,
		},
	}

	if err := ctrl.SetControllerReference(&cr, svc, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Labels = labels
		// 不要碰 ClusterIP
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = labels
		svc.Spec.Ports = []corev1.ServicePort{{
			Name:       "http",
			Port:       port,
			TargetPort: intstr.FromString(strconv.Itoa(ContainerPort)),
			Protocol:   corev1.ProtocolTCP,
		}}
		return nil
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	// 4) Update Status
	var currentDep appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, &currentDep); err == nil {
		cr.Status.ReadyReplicas = currentDep.Status.ReadyReplicas
		cr.Status.ServiceName = svcName
		if err := r.Status().Update(ctx, &cr); err != nil {
			logger.Error(err, "failed to update deployment")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleGolangAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SimpleGolangApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("simplegolangapp").
		Complete(r)
}
