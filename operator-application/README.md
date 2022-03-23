# operator-application

See below for instructions how to set up and run the application operator as well as the used commands for the development of it.

The following instructions assume that you use the managed Kubernetes service on the IBM Cloud. You can also use any other Kubernetes service or OpenShift.

There are three ways to run the operator:

1) [Local Go Operator](#setup-and-local-usage) 
2) [Kubernetes Operator manually deployed](#setup-and-manual-deployment)
3) [Kubernetes Operator deployed via OLM](#setup-and-deployment-via-operator-lifecycle-manager)
    * via operator-sdk
    * via kubectl

### Prerequisites

* [operator-sdk](https://sdk.operatorframework.io/docs/installation/) (comes with Golang)
* git
* kubectl
* docker
* [ibmcloud](https://cloud.ibm.com/docs/cli?topic=cli-install-ibmcloud-cli) (if IBM Cloud is used)

*Image Registry*

```
$ export REGISTRY='docker.io'
$ export ORG='nheidloff'
$ export IMAGE='application-controller:v22'
$ export BUNDLE_IMAGE="application-controller-bundle:v16"
$ export CATALOG_IMAGE="application-controller-catalog:v1"
```

### Setup and local Usage

Get the code:

```
$ https://github.com/nheidloff/operator-sample-go.git
$ cd operator-application
$ code .
```

Login to Kubernetes:

```
$ ibmcloud login -a cloud.ibm.com -r eu-de -g resource-group-niklas-heidloff7 --sso
$ ibmcloud ks cluster config --cluster xxxxxxx
```

Configure Kubernetes:

```
$ kubectl create ns test1
$ kubectl config set-context --current --namespace=test1
$ kubectl create ns database
$ kubectl apply -f ../operator-database/config/crd/bases/database.sample.third.party_databases.yaml
```

From a terminal in VSCode run these commands:

```
$ make install run
$ kubectl apply -f config/samples/application.sample_v1alpha1_application.yaml
```

To debug, press F5 (Run - Start Debugging) instead of 'make install run'. The directory 'operator-application' needs to be root in VSCode.

The sample endpoint can be triggered via '<your-ip>:30548/hello':

```
$ ibmcloud ks worker ls --cluster niklas-heidloff-fra02-b3c.4x16
$ open http://159.122.86.194:30548/hello
```

All resources can be deleted:

```
$ kubectl delete -f config/samples/application.sample_v1alpha1_application.yaml
```

### Setup and manual Deployment

Get the code:

```
$ https://github.com/nheidloff/operator-sample-go.git
$ cd operator-application
$ code .
```

Login to Kubernetes:

```
$ ibmcloud login -a cloud.ibm.com -r eu-de -g resource-group-niklas-heidloff7 --sso
$ ibmcloud ks cluster config --cluster xxxxxxx
```

Configure Kubernetes:

```
$ kubectl create ns test1
$ kubectl config set-context --current --namespace=test1
$ kubectl apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
```

Deploy Database Operator:

Before running the application-controller-bundle (the application operator), the database operator needs to be deployed since it is defined as 'required' in the application CSV.

```
$ kubectl create ns database
$ cd ../operator-database
$ export BUNDLE_IMAGE_DATABASE="database-controller-bundle:v1"
$ operator-sdk run bundle "$REGISTRY/$ORG/$BUNDLE_IMAGE_DATABASE" -n operators
$ cd ../operator-application
```

Build and push the Operator Image:

```
$ make generate manifests
$ docker login $REGISTRY
$ make docker-build docker-push IMG="$REGISTRY/$ORG/$IMAGE"
```

Deploy Operator:

```
$ make deploy IMG="$REGISTRY/$ORG/$IMAGE"
$ export OPERATOR_NAMESPACE='operator-application-system'
$ kubectl get all -n $OPERATOR_NAMESPACE
```

Test Operator: 

```
$ kubectl apply -f config/samples/application.sample_v1alpha1_application.yaml
```

The sample endpoint can be triggered via '<your-ip>:30548/hello':

```
$ ibmcloud ks worker ls --cluster niklas-heidloff-fra02-b3c.4x16
$ open http://159.122.86.194:30548/hello
```

Delete Resources:

```
$ kubectl delete -f config/samples/application.sample_v1alpha1_application.yaml
$ make undeploy IMG="$REGISTRY/$ORG/$IMAGE"
```

### Setup and Deployment via Operator Lifecycle Manager

Follow the same steps as above in the section [Setup and manual Deployment](#setup-and-manual-deployment) up to the step 'Deploy Operator'.

Install the Operator Lifecycle Manager (OLM):

```
$ operator-sdk olm install latest 
$ kubectl get all -n olm
```

Build and push the Bundle Image:

```
$ make bundle-build docker-push BUNDLE_IMG="$REGISTRY/$ORG/$BUNDLE_IMAGE" IMG="$REGISTRY/$ORG/$BUNDLE_IMAGE"
```

**Deploy the Operator**

There are two ways to deploy the operator:

1) operator-sdk (all necessary resources are created)
2) kubectl (resources defined in yaml)

*operator-sdk:*

```
$ operator-sdk run bundle "$REGISTRY/$ORG/$BUNDLE_IMAGE" -n operators
```

*kubectl:*

```
$ kubectl apply -f olm/catalogsource.yaml
$ kubectl apply -f olm/subscription.yaml 
$ kubectl get installplans -n operators
$ kubectl -n operators patch installplan install-xxxxx -p '{"spec":{"approved":true}}' --type merge
```

To test the operator, follow the instructions at the bottom of the section [Setup and manual Deployment](#setup-and-manual-deployment).

Verify Installation:

```
$ kubectl get all -n operators
$ kubectl get catalogsource operator-application-catalog -n operators -oyaml
$ kubectl get subscriptions operator-application-v0-0-1-sub -n operators -oyaml
$ kubectl get csv operator-application.v0.0.1 -n operators -oyaml
$ kubectl get installplans -n operators
$ kubectl get installplans install-xxxxx -n operators -oyaml
$ kubectl get operators operator-application.operators -n operators -oyaml
```

Delete Resources (operator-sdk):

```
$ kubectl delete -f config/samples/application.sample_v1alpha1_application.yaml
$ operator-sdk cleanup operator-application -n operators --delete-all
$ kubectl apply -f ../operator-database/config/crd/bases/database.sample.third.party_databases.yaml
$ operator-sdk olm uninstall
```

Delete Resources (kubectl):

```
$ kubectl delete -f config/samples/application.sample_v1alpha1_application.yaml
$ kubectl delete -f olm/catalogsource.yaml
$ kubectl delete -f olm/subscription.yaml
$ operator-sdk olm uninstall
```

### Development Commands

Commands for the project creation:

```
$ operator-sdk init --domain ibm.com --repo github.com/nheidloff/operator-sample-go/operator-application
$ operator-sdk create api --group application.sample --version v1alpha1 --kind Application --resource --controller
$ make generate
$ make manifests
```

Commands for the bundle creation:

```
$ make bundle IMG="$REGISTRY/$ORG/$IMAGE"
```

Commands for the webhook creations:

```
$ operator-sdk create webhook --group application.sample --version v1alpha1 --kind Application --defaulting --programmatic-validation --conversion
$ make manifests
$ make install
$ make run ENABLE_WEBHOOKS=false
```

Command to create catalog:

```
$ make catalog-build docker-push CATALOG_IMG="$REGISTRY/$ORG/$CATALOG_IMAGE" BUNDLE_IMGS="$REGISTRY/$ORG/$BUNDLE_IMAGE" IMG="$REGISTRY/$ORG/$CATALOG_IMAGE"
```