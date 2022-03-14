package controllers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"time"

	"golang.org/x/crypto/ripemd160"

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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

const image = "docker.io/nheidloff/simple-microservice:latest"
const port int32 = 8081
const nodePort int32 = 30548
const labelKey = "app"
const labelValue = "myapplication"
const greetingMessage = "World"
const secretGreetingMessageLabel = "GREETING_MESSAGE"

// for simplication purposes database properties are hardcoded
const databaseUser string = "name"
const databasePassword string = "password"
const databaseUrl string = "url"
const databaseCertificate string = "certificate"

const finalizer = "database.sample.third.party/finalizer"
const hashLabelName = "application.sample.ibm.com/hash"

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
	err = reconciler.setConditionResourceFound(ctx, application)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconciler.checkPrerequisites() == false {
		log.Info("Prerequisites not fulfilled")
		err = reconciler.setConditionFailed(ctx, application, CONDITION_REASON_FAILED_INSTALL_READY)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second * 60}, fmt.Errorf("Prerequisites not fulfilled")
	}
	err = reconciler.setConditionInstallReady(ctx, application)
	if err != nil {
		return ctrl.Result{}, err
	}

	fmt.Println("Custom Resource Values:")
	fmt.Printf("- Name: %s\n", application.Name)
	fmt.Printf("- Namespace: %s\n", application.Namespace)
	fmt.Printf("- Version: %s\n", application.Spec.Version)
	fmt.Printf("- AmountPods: %d\n", application.Spec.AmountPods)
	fmt.Printf("- DatabaseName: %s\n", application.Spec.DatabaseName)
	fmt.Printf("- DatabaseNamespace: %s\n", application.Spec.DatabaseNamespace)

	reconciler.setGlobalVariables(application)

	isApplicationMarkedToBeDeleted := application.GetDeletionTimestamp() != nil
	if isApplicationMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(application, finalizer) {
			if err := reconciler.finalizeApplication(ctx, application); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(application, finalizer)
			err := reconciler.Update(ctx, application)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	database := &databasesamplev1alpha1.Database{}
	databaseDefinition := reconciler.defineDatabase(application)
	err = reconciler.Get(ctx, types.NamespacedName{Name: application.Spec.DatabaseName, Namespace: application.Spec.DatabaseNamespace}, database)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Database resource " + application.Spec.DatabaseName + " not found. Creating or re-creating database")
			err = reconciler.setConditionDatabaseExists(ctx, application, CONDITION_STATUS_FALSE)
			if err != nil {
				return ctrl.Result{}, err
			}
			err = reconciler.Create(ctx, databaseDefinition)
			if err != nil {
				log.Info("Failed to create database resource. Re-running reconcile.")
				return ctrl.Result{}, err
			} else {
				return ctrl.Result{RequeueAfter: time.Second * 1}, nil // delay the next loop run since database creation can take time
			}
		} else {
			log.Info("Failed to get database resource " + application.Spec.DatabaseName + ". Re-running reconcile.")
			return ctrl.Result{}, err
		}
	}

	// TODO: Create schema
	// see https://github.com/IBM/multi-tenancy/blob/a181c562b788f7b5fad99e09b441f93e4489b72f/operator/ecommerceapplication/postgresHelper/postgresHelper.go

	// TODO: Check if database and schema exist
	err = reconciler.setConditionDatabaseExists(ctx, application, CONDITION_STATUS_TRUE)
	if err != nil {
		return ctrl.Result{}, err
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
		// for simplication purposes secrets are not updated - see deployment section
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
		specHashTarget := reconciler.getHashForSpec(&deploymentDefinition.Spec)
		specHashActual := reconciler.getHashFromLabels(deployment.Labels)
		if specHashActual != specHashTarget {
			var current int32 = *deployment.Spec.Replicas
			var expected int32 = *deploymentDefinition.Spec.Replicas
			if current != expected {
				deployment.Spec.Replicas = &expected
				deployment.Labels = reconciler.setHashToLabels(deployment.Labels, specHashTarget)
				err = reconciler.Update(ctx, deployment)
				if err != nil {
					log.Info("Failed to update deployment resource. Re-running reconcile.")
					return ctrl.Result{}, err
				}
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
		// for simplication purposes secrets are not updated - see deployment section
	}

	// commented out for dev productivity
	/*
		if !controllerutil.ContainsFinalizer(application, finalizer) {
			controllerutil.AddFinalizer(application, finalizer)
			err = reconciler.Update(ctx, application)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	*/
	err = reconciler.setConditionSucceeded(ctx, application)
	if err != nil {
		return ctrl.Result{}, err
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

	specHashActual := reconciler.getHashForSpec(&deployment.Spec)
	deployment.Labels = reconciler.setHashToLabels(nil, specHashActual)

	ctrl.SetControllerReference(application, deployment, reconciler.Scheme)
	return deployment
}

func (reconciler *ApplicationReconciler) checkPrerequisites() bool {
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

func (reconciler *ApplicationReconciler) setGlobalVariables(application *applicationsamplev1alpha1.Application) {
	secretName = application.Name + "-secret-greeting"
	deploymentName = application.Name + "-deployment-microservice"
	serviceName = application.Name + "-service-microservice"
	containerName = application.Name + "-microservice"
	// TODO: handle application.Spec.Version
}

func (reconciler *ApplicationReconciler) finalizeApplication(ctx context.Context, application *applicationsamplev1alpha1.Application) error {
	database := &databasesamplev1alpha1.Database{}
	err := reconciler.Get(ctx, types.NamespacedName{Name: application.Spec.DatabaseName, Namespace: application.Spec.DatabaseNamespace}, database)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
	}
	return fmt.Errorf("Database not deleted yet")
}

const CONDITION_STATUS_TRUE = "True"
const CONDITION_STATUS_FALSE = "False"
const CONDITION_STATUS_UNKNOWN = "Unknown"

// status of RESOURCE_FOUND can only be True, otherwise there is no condition
const CONDITION_TYPE_RESOURCE_FOUND = "ResourceFound"
const CONDITION_REASON_RESOURCE_FOUND = "ResourceFound"
const CONDITION_MESSAGE_RESOURCE_FOUND = "Resource found in k18n"

func (reconciler *ApplicationReconciler) setConditionResourceFound(ctx context.Context,
	application *applicationsamplev1alpha1.Application) error {

	if !reconciler.containsCondition(ctx, application, CONDITION_REASON_RESOURCE_FOUND) {
		return reconciler.appendCondition(ctx, application, CONDITION_TYPE_RESOURCE_FOUND, CONDITION_STATUS_TRUE,
			CONDITION_REASON_RESOURCE_FOUND, CONDITION_MESSAGE_RESOURCE_FOUND)
	}
	return nil
}

// status of INSTALL_READY can only be True, otherwise there is a failure condition
const CONDITION_TYPE_INSTALL_READY = "InstallReady"
const CONDITION_REASON_INSTALL_READY = "AllRequirementsMet"
const CONDITION_MESSAGE_INSTALL_READY = "All requirements met, attempting install"

func (reconciler *ApplicationReconciler) setConditionInstallReady(ctx context.Context,
	application *applicationsamplev1alpha1.Application) error {

	reconciler.deleteCondition(ctx, application, CONDITION_TYPE_FAILED, CONDITION_REASON_FAILED_INSTALL_READY)
	if !reconciler.containsCondition(ctx, application, CONDITION_REASON_INSTALL_READY) {
		return reconciler.appendCondition(ctx, application, CONDITION_TYPE_INSTALL_READY, CONDITION_STATUS_TRUE,
			CONDITION_REASON_INSTALL_READY, CONDITION_MESSAGE_INSTALL_READY)
	}
	return nil
}

// status of FAILED can only be True
const CONDITION_TYPE_FAILED = "Failed"
const CONDITION_REASON_FAILED_INSTALL_READY = "RequirementsNotMet"
const CONDITION_MESSAGE_FAILED_INSTALL_READY = "Not all requirements met"

func (reconciler *ApplicationReconciler) setConditionFailed(ctx context.Context,
	application *applicationsamplev1alpha1.Application, reason string) error {

	var message string
	switch reason {
	case CONDITION_REASON_FAILED_INSTALL_READY:
		message = CONDITION_MESSAGE_FAILED_INSTALL_READY
	}

	if !reconciler.containsCondition(ctx, application, reason) {
		return reconciler.appendCondition(ctx, application, CONDITION_TYPE_FAILED, CONDITION_STATUS_TRUE,
			reason, message)
	}
	return nil
}

// status of DATABASE_EXISTS can be True or False
const CONDITION_TYPE_DATABASE_EXISTS = "DatabaseExists"
const CONDITION_REASON_DATABASE_EXISTS = "DatabaseExists"
const CONDITION_MESSAGE_DATABASE_EXISTS = "The database exists"

func (reconciler *ApplicationReconciler) setConditionDatabaseExists(ctx context.Context,
	application *applicationsamplev1alpha1.Application, status metav1.ConditionStatus) error {

	if !reconciler.containsCondition(ctx, application, CONDITION_REASON_DATABASE_EXISTS) {
		return reconciler.appendCondition(ctx, application, CONDITION_TYPE_DATABASE_EXISTS, status,
			CONDITION_REASON_DATABASE_EXISTS, CONDITION_MESSAGE_DATABASE_EXISTS)
	} else {
		currentStatus := reconciler.getConditionStatus(ctx, application, CONDITION_TYPE_DATABASE_EXISTS)
		if currentStatus != status {
			reconciler.deleteCondition(ctx, application, CONDITION_TYPE_DATABASE_EXISTS, CONDITION_REASON_DATABASE_EXISTS)
			return reconciler.appendCondition(ctx, application, CONDITION_TYPE_DATABASE_EXISTS, status,
				CONDITION_REASON_DATABASE_EXISTS, CONDITION_MESSAGE_DATABASE_EXISTS)
		}
	}
	return nil
}

// status of SUCCEEDED can only be True
const CONDITION_TYPE_SUCCEEDED = "Succeeded"
const CONDITION_REASON_SUCCEEDED = "InstallSucceeded"
const CONDITION_MESSAGE_SUCCEEDED = "Application has been installed"

func (reconciler *ApplicationReconciler) setConditionSucceeded(ctx context.Context,
	application *applicationsamplev1alpha1.Application) error {

	if !reconciler.containsCondition(ctx, application, CONDITION_REASON_SUCCEEDED) {
		return reconciler.appendCondition(ctx, application, CONDITION_TYPE_SUCCEEDED, CONDITION_STATUS_TRUE,
			CONDITION_REASON_SUCCEEDED, CONDITION_MESSAGE_SUCCEEDED)
	}
	return nil
}

// status of DELETION_REQUEST_RECEIVED can only be True
const CONDITION_TYPE_DELETION_REQUEST_RECEIVED = "DeletionRequestReceived"
const CONDITION_REASON_DELETION_REQUEST_RECEIVED = "DeletionRequestReceived"
const CONDITION_MESSAGE_DELETION_REQUEST_RECEIVED = "Application is supposed to be deleted"

func (reconciler *ApplicationReconciler) setConditionDeletionRequestReceived(ctx context.Context,
	application *applicationsamplev1alpha1.Application) error {

	if !reconciler.containsCondition(ctx, application, CONDITION_REASON_DELETION_REQUEST_RECEIVED) {
		return reconciler.appendCondition(ctx, application, CONDITION_TYPE_DELETION_REQUEST_RECEIVED, CONDITION_STATUS_TRUE,
			CONDITION_REASON_DELETION_REQUEST_RECEIVED, CONDITION_MESSAGE_DELETION_REQUEST_RECEIVED)
	}
	return nil
}

func (reconciler *ApplicationReconciler) getConditionStatus(ctx context.Context, application *applicationsamplev1alpha1.Application,
	typeName string) metav1.ConditionStatus {

	var output metav1.ConditionStatus = CONDITION_STATUS_UNKNOWN
	for _, condition := range application.Status.Conditions {
		if condition.Type == typeName {
			output = condition.Status
		}
	}
	return output
}

func (reconciler *ApplicationReconciler) deleteCondition(ctx context.Context, application *applicationsamplev1alpha1.Application,
	typeName string, reason string) error {

	log := log.FromContext(ctx)
	var newConditions = make([]metav1.Condition, 0)
	for _, condition := range application.Status.Conditions {
		if condition.Type != typeName && condition.Reason != reason {
			newConditions = append(newConditions, condition)
		}
	}
	application.Status.Conditions = newConditions

	err := reconciler.Client.Status().Update(ctx, application)
	if err != nil {
		log.Info("Application resource status update failed.")
	}
	return nil
}

func (reconciler *ApplicationReconciler) appendCondition(ctx context.Context, application *applicationsamplev1alpha1.Application,
	typeName string, status metav1.ConditionStatus, reason string, message string) error {

	log := log.FromContext(ctx)
	time := metav1.Time{Time: time.Now()}
	condition := metav1.Condition{Type: typeName, Status: status, Reason: reason, Message: message, LastTransitionTime: time}
	application.Status.Conditions = append(application.Status.Conditions, condition)

	err := reconciler.Client.Status().Update(ctx, application)
	if err != nil {
		log.Info("Application resource status update failed.")
	}
	return nil
}

func (reconciler *ApplicationReconciler) containsCondition(ctx context.Context,
	application *applicationsamplev1alpha1.Application, reason string) bool {

	output := false
	for _, condition := range application.Status.Conditions {
		if condition.Reason == reason {
			output = true
		}
	}
	return output
}

func (reconciler *ApplicationReconciler) getHashForSpec(specStruct interface{}) string {
	byteArray, _ := json.Marshal(specStruct)
	var hasher hash.Hash
	hasher = ripemd160.New()
	hasher.Reset()
	hasher.Write(byteArray)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (reconciler *ApplicationReconciler) setHashToLabels(labels map[string]string, specHashActual string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[hashLabelName] = specHashActual
	return labels
}

func (reconciler *ApplicationReconciler) getHashFromLabels(labels map[string]string) string {
	return labels[hashLabelName]
}
