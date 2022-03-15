package applicationcontroller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	applicationsamplev1alpha1 "github.com/nheidloff/operator-sample-go/operator-application/api/v1alpha1"
	databasesamplev1alpha1 "github.com/nheidloff/operator-sample-go/operator-database/api/v1alpha1"
)

const finalizer = "database.sample.third.party/finalizer"

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

	_, err = reconciler.reconcileDatabase(ctx, application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Create schema and sample data and check if data from the database can be accessed
	// see https://github.com/IBM/multi-tenancy/blob/a181c562b788f7b5fad99e09b441f93e4489b72f/operator/ecommerceapplication/postgresHelper/postgresHelper.go
	// see http://heidloff.net/article/creating-database-schemas-kubernetes-operators/

	err = reconciler.setConditionDatabaseExists(ctx, application, CONDITION_STATUS_TRUE)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = reconciler.reconcileSecret(ctx, application)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = reconciler.reconcileDeployment(ctx, application)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = reconciler.reconcileService(ctx, application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Note: Commented out for dev productivity only
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
		// Note: Possible, but not used in this scenario
		//Owns(&databasesamplev1alpha1.Database{}).
		Complete(reconciler)
}

func (reconciler *ApplicationReconciler) setGlobalVariables(application *applicationsamplev1alpha1.Application) {
	secretName = application.Name + "-secret-greeting"
	deploymentName = application.Name + "-deployment-microservice"
	serviceName = application.Name + "-service-microservice"
	containerName = application.Name + "-microservice"
	// TODO: Handle application.Spec.Version
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

// Note: Status of RESOURCE_FOUND can only be True, otherwise there is no condition
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

// Note: Status of INSTALL_READY can only be True, otherwise there is a failure condition
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

// Note: Status of FAILED can only be True
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

// Note: Status of DATABASE_EXISTS can be True or False
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

// Note: Status of SUCCEEDED can only be True
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

// Note: Status of DELETION_REQUEST_RECEIVED can only be True
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
