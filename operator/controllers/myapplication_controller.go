package controllers

import (
	"context"
	"fmt"

	cachev1alpha1 "github.com/nheidloff/operator-sample-go/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MyApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.nheidloff,resources=myapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.nheidloff,resources=myapplications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.nheidloff,resources=myapplications/finalizers,verbs=update
func (reconciler *MyApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	fmt.Println("Reconcile started")

	myApplication := &cachev1alpha1.MyApplication{}
	err := reconciler.Get(ctx, req.NamespacedName, myApplication)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("MyApplication resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Info("Failed to get MyApplication resource. Re-running reconcile.")
		return ctrl.Result{}, err
	}

	fmt.Printf("Name: %s\n", myApplication.Name)
	fmt.Printf("Namespace: %s\n", myApplication.Namespace)
	fmt.Printf("Size: %d\n", myApplication.Spec.Size)

	secret := &corev1.Secret{}
	secretName := myApplication.Name + "-secret-greeting"
	greetingMessage := "World"
	err = reconciler.Get(ctx, types.NamespacedName{Name: secretName, Namespace: myApplication.Namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Secret resource " + secretName + " not found. Creating or re-creating secret")
			secretDefinition := reconciler.defineSecret(myApplication, secretName, "GREETING_MESSAGE", greetingMessage)
			err = reconciler.Create(ctx, secretDefinition)
			if err != nil {
				log.Info("Failed to create secret resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Failed to get secret resource " + secretName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	}

	deployment := &appsv1.Deployment{}
	deploymentName := myApplication.Name + "-deployment-microservice"
	err = reconciler.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: myApplication.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Deployment resource " + deploymentName + " not found. Creating or re-creating deployment")
			deploymentDefinition := reconciler.defineDeployment(myApplication)
			err = reconciler.Create(ctx, deploymentDefinition)
			if err != nil {
				log.Info("Failed to create deployment resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Failed to get deployment resource " + deploymentName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	}

	service := &corev1.Service{}
	serviceName := myApplication.Name + "-service-microservice"
	err = reconciler.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: myApplication.Namespace}, service)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Service resource " + serviceName + " not found. Creating or re-creating service")
			serviceDefinition := reconciler.defineService(myApplication)
			err = reconciler.Create(ctx, serviceDefinition)
			if err != nil {
				log.Info("Failed to create service resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Failed to get service resource " + serviceName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (reconciler *MyApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.MyApplication{}).
		Owns(&appsv1.Deployment{}).
		Complete(reconciler)
}

func (reconciler *MyApplicationReconciler) defineService(myApplication *cachev1alpha1.MyApplication) *corev1.Service {
	labels := map[string]string{"app": "myapplication"}
	serviceName := myApplication.Name + "-service-microservice"
	var port int32 = 8081
	var nodePort int32 = 30548

	service := &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: myApplication.Namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{{
				Port:     port,
				NodePort: nodePort,
				Protocol: "TCP",
				TargetPort: intstr.IntOrString{
					IntVal: port,
				},
			}},
			Selector: labels,
		},
	}

	ctrl.SetControllerReference(myApplication, service, reconciler.Scheme)
	return service
}

func (reconciler *MyApplicationReconciler) defineSecret(myApplication *cachev1alpha1.MyApplication, name string, key string, value string) *corev1.Secret {
	stringData := make(map[string]string)
	stringData[key] = value

	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: myApplication.Namespace},
		Immutable:  new(bool),
		Data:       map[string][]byte{},
		StringData: stringData,
		Type:       "Opaque",
	}

	ctrl.SetControllerReference(myApplication, secret, reconciler.Scheme)
	return secret
}

func (reconciler *MyApplicationReconciler) defineDeployment(myApplication *cachev1alpha1.MyApplication) *appsv1.Deployment {
	replicas := myApplication.Spec.Size
	deploymentName := myApplication.Name + "-deployment-microservice"
	secretName := myApplication.Name + "-secret-greeting"
	labels := map[string]string{"app": "myapplication"}
	image := "docker.io/nheidloff/simple-microservice:latest"
	containerName := "microservice"
	port := 8081

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: myApplication.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: image,
						Name:  containerName,
						Ports: []corev1.ContainerPort{{
							ContainerPort: int32(port),
						}},
						Env: []corev1.EnvVar{{
							Name: "GREETING_MESSAGE",
							ValueFrom: &v1.EnvVarSource{
								SecretKeyRef: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: secretName,
									},
									Key: "GREETING_MESSAGE",
								},
							}},
						},
						ReadinessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{Path: "/q/health/live", Port: intstr.FromInt(port)},
							},
							InitialDelaySeconds: 20,
						},
						LivenessProbe: &v1.Probe{
							Handler: v1.Handler{
								HTTPGet: &v1.HTTPGetAction{Path: "/q/health/ready", Port: intstr.FromInt(port)},
							},
							InitialDelaySeconds: 40,
						},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(myApplication, deployment, reconciler.Scheme)
	return deployment
}
