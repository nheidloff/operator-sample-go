package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	applicationsamplev1alpha1 "github.com/nheidloff/operator-sample-go/operator-application/api/v1alpha1"
	databasesamplev1alpha1 "github.com/nheidloff/operator-sample-go/operator-database/api/v1alpha1"
)

var kubernetesServerVersion string
var runsOnOpenShift bool = false

var secretName string
var deploymentName string
var serviceName string
var containerName string

var image = "docker.io/nheidloff/simple-microservice:latest"
var port int32 = 8081
var nodePort int32 = 30548
var labelKey = "app"
var labelValue = "myapplication"
var greetingMessage = "World"
var secretGreetingMessageLabel = "GREETING_MESSAGE"

// for simplication purposes database properties are hardcoded
var databaseUser string = "name"
var databasePassword string = "password"
var databaseUrl string = "url"
var databaseCertificate string = "certificate"

var managerConfig *rest.Config

type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=application.sample.ibm.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=application.sample.ibm.com,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=application.sample.ibm.com,resources=applications/finalizers,verbs=update
func (reconciler *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	fmt.Println("Reconcile started")

	application := &applicationsamplev1alpha1.Application{}
	err := reconciler.Get(ctx, req.NamespacedName, application)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Application resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Info("Failed to getyApplication resource. Re-running reconcile.")
		return ctrl.Result{}, err
	}

	if checkPrerequisites() == false {
		log.Info("Prerequisites not fulfilled")
		return ctrl.Result{}, fmt.Errorf("Prerequisites not fulfilled")
	}

	fmt.Println("Custom Resource Values:")
	fmt.Printf("- Name: %s\n", application.Name)
	fmt.Printf("- Namespace: %s\n", application.Namespace)
	fmt.Printf("- Version: %s\n", application.Spec.Version)
	fmt.Printf("- AmountPods: %d\n", application.Spec.AmountPods)
	fmt.Printf("- DatabaseName: %s\n", application.Spec.DatabaseName)
	fmt.Printf("- DatabaseNamespace: %s\n", application.Spec.DatabaseNamespace)

	setGlobalVariables(application)

	database := &databasesamplev1alpha1.Database{}
	databaseDefinition := reconciler.defineDatabase(application)
	err = reconciler.Get(ctx, types.NamespacedName{Name: application.Spec.DatabaseName, Namespace: application.Spec.DatabaseNamespace}, database)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Database resource " + application.Spec.DatabaseName + " not found. Creating or re-creating database")
			err = reconciler.Create(ctx, databaseDefinition)
			if err != nil {
				log.Info("Failed to create database resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Failed to get database resource " + application.Spec.DatabaseName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	}

	secret := &corev1.Secret{}
	secretDefinition := reconciler.defineSecret(application)
	err = reconciler.Get(ctx, types.NamespacedName{Name: secretName, Namespace: application.Namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Secret resource " + secretName + " not found. Creating or re-creating secret")
			err = reconciler.Create(ctx, secretDefinition)
			if err != nil {
				log.Info("Failed to create secret resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Failed to get secret resource " + secretName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	} else {
		// for simplication purposes secrets are not updated - see deployment
	}

	deployment := &appsv1.Deployment{}
	deploymentDefinition := reconciler.defineDeployment(application)
	err = reconciler.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: application.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Deployment resource " + deploymentName + " not found. Creating or re-creating deployment")
			err = reconciler.Create(ctx, deploymentDefinition)
			if err != nil {
				log.Info("Failed to create deployment resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Failed to get deployment resource " + deploymentName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	} else {
		// TODO: use hash of the spec sections to check whether deployed resource needs to be updated
		// e.g. https://github.com/kubernetes/kubernetes/blob/master/pkg/util/hash/hash.go

		var currentPointer *int32 = deployment.Spec.Replicas
		var current int32 = *currentPointer
		var expected int32 = application.Spec.AmountPods
		if current != expected {
			err = reconciler.Update(ctx, deploymentDefinition)
			if err != nil {
				log.Info("Failed to update deployment resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		}
	}

	serviceDefinition := reconciler.defineService(application)
	service := &corev1.Service{}
	err = reconciler.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: application.Namespace}, service)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Service resource " + serviceName + " not found. Creating or re-creating service")
			err = reconciler.Create(ctx, serviceDefinition)
			if err != nil {
				log.Info("Failed to create service resource. Re-running reconcile.")
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Failed to get service resource " + serviceName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	} else {
		// for simplication purposes secrets are not updated - see deployment
	}

	return ctrl.Result{}, nil
}

func (reconciler *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	managerConfig = mgr.GetConfig()

	return ctrl.NewControllerManagedBy(mgr).
		For(&applicationsamplev1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		//Owns(&databasesamplev1alpha1.Database{}). // possible, but not used in this scenario
		Complete(reconciler)
}

func (reconciler *ApplicationReconciler) defineService(application *applicationsamplev1alpha1.Application) *corev1.Service {
	labels := map[string]string{labelKey: labelValue}

	service := &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: application.Namespace, Labels: labels},
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

	ctrl.SetControllerReference(application, service, reconciler.Scheme)
	return service
}

func (reconciler *ApplicationReconciler) defineSecret(application *applicationsamplev1alpha1.Application) *corev1.Secret {
	stringData := make(map[string]string)
	stringData[secretGreetingMessageLabel] = greetingMessage

	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: application.Namespace},
		Immutable:  new(bool),
		Data:       map[string][]byte{},
		StringData: stringData,
		Type:       "Opaque",
	}

	ctrl.SetControllerReference(application, secret, reconciler.Scheme)
	return secret
}

func (reconciler *ApplicationReconciler) defineDatabase(application *applicationsamplev1alpha1.Application) *databasesamplev1alpha1.Database {
	database := &databasesamplev1alpha1.Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      application.Spec.DatabaseName,
			Namespace: application.Spec.DatabaseNamespace,
		},
		Spec: databasesamplev1alpha1.DatabaseSpec{
			User:        databaseUser,
			Password:    databasePassword,
			Url:         databaseUrl,
			Certificate: databaseCertificate,
		},
	}

	//ctrl.SetControllerReference(application, database, reconciler.Scheme) // possible, but not used in this scenario
	return database
}

func (reconciler *ApplicationReconciler) defineDeployment(application *applicationsamplev1alpha1.Application) *appsv1.Deployment {
	replicas := application.Spec.AmountPods
	labels := map[string]string{labelKey: labelValue}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: application.Namespace,
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
							ContainerPort: port,
						}},
						Env: []corev1.EnvVar{{
							Name: secretGreetingMessageLabel,
							ValueFrom: &v1.EnvVarSource{
								SecretKeyRef: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: secretName,
									},
									Key: secretGreetingMessageLabel,
								},
							}},
						},
						ReadinessProbe: &v1.Probe{
							ProbeHandler: v1.ProbeHandler{
								HTTPGet: &v1.HTTPGetAction{Path: "/q/health/live", Port: intstr.IntOrString{
									IntVal: port,
								}},
							},
							InitialDelaySeconds: 20,
						},
						LivenessProbe: &v1.Probe{
							ProbeHandler: v1.ProbeHandler{
								HTTPGet: &v1.HTTPGetAction{Path: "/q/health/ready", Port: intstr.IntOrString{
									IntVal: port,
								}},
							},
							InitialDelaySeconds: 40,
						},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(application, deployment, reconciler.Scheme)
	return deployment
}

func checkPrerequisites() bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(managerConfig)
	if err == nil {
		serverVersion, err := discoveryClient.ServerVersion()
		if err == nil {
			kubernetesServerVersion = serverVersion.String()
			fmt.Println("Kubernetes Server Version: " + kubernetesServerVersion)

			apiGroup, _, err := discoveryClient.ServerGroupsAndResources()
			if err == nil {
				for i := 0; i < len(apiGroup); i++ {
					if apiGroup[i].Name == "route.openshift.io" {
						runsOnOpenShift = true
					}
				}
			}
		}
	}
	// TODO: check correct Kubernetes version and distro
	return true
}

func setGlobalVariables(application *applicationsamplev1alpha1.Application) {
	secretName = application.Name + "-secret-greeting"
	deploymentName = application.Name + "-deployment-microservice"
	serviceName = application.Name + "-service-microservice"
	containerName = application.Name + "-microservice"
	// TODO: handle application.Spec.Version
}
