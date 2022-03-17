# operator-application

See below for instructions how to set up and run the application operator as well as the used commands for the development of it.

The following instructions assume that you use the managed Kubernetes service on the IBM Cloud. You can also use any other Kubernetes service or OpenShift.

There are three ways to run the operator:

1) [Local Go Operator](#setup-and-local-usage) 
2) [Kubernetes Operator manually deployed](#setup-and-manual-deployment)
3) [Kubernetes Operator deployed via OLM](#setup-and-deployment-via-operator-lifecycle-manager)

### Prerequisites

* [operator-sdk](https://sdk.operatorframework.io/docs/installation/) (comes with Golang)
* git
* kubectl
* docker
* [ibmcloud](https://cloud.ibm.com/docs/cli?topic=cli-install-ibmcloud-cli) (if IBM Cloud is used)

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
$ kubectl create ns database
$ kubectl apply -f ../operator-database/config/crd/bases/database.sample.third.party_databases.yaml
```

Build and push the Operator Image:

```
$ export REGISTRY='docker.io'
$ export ORG='nheidloff'
$ export IMAGE='application-controller:v1'
$ make generate
$ make manifests
$ make docker-build IMG="$REGISTRY/$ORG/$IMAGE"
$ docker login $REGISTRY
$ docker push "$REGISTRY/$ORG/$IMAGE"
```

Deploy Operator:

Namespace will be the project name defined in 'Project' plus '-system'.

```
$ make deploy IMG="$REGISTRY/$ORG/$IMAGE"
$ export OPERATOR_NAMESPACE='operator-application-system'
$ kubectl get all -n $OPERATOR_NAMESPACE
```

Test Operator: 

```
$ kubectl apply -f config/samples/application.sample_v1alpha1_application.yaml
$ kubectl delete -f config/samples/application.sample_v1alpha1_application.yaml
```

The sample endpoint can be triggered via '<your-ip>:30548/hello':

```
$ ibmcloud ks worker ls --cluster niklas-heidloff-fra02-b3c.4x16
$ open http://159.122.86.194:30548/hello
```

Delete Operator:

```
$ make undeploy
```

### Setup and Deployment via Operator Lifecycle Manager

Follow the same steps as above in the section [Setup and manual Deployment](#setup-and-manual-deployment) up to the step 'Deploy Operator'.

Build and push the Bundle Image:

```
$ export REGISTRY='docker.io'
$ export ORG='nheidloff'
$ export BUNDLEIMAGE="application-controller-bundle:v1"
$ make bundle-build BUNDLE_IMG="$REGISTRY/$ORG/$BUNDLEIMAGE"
$ docker push "$REGISTRY/$ORG/$BUNDLEIMAGE"
```

```
$ operator-sdk run bundle "$REGISTRY/$ORG/$BUNDLEIMAGE" -n operators
```

To test the operator, follow the instructions at the bottom of the section [Setup and manual Deployment](#setup-and-manual-deployment).

### Development Commands

Commands used for the project creation:

```
$ operator-sdk init --domain ibm.com --repo github.com/nheidloff/operator-sample-go/operator-application
$ operator-sdk create api --group application.sample --version v1alpha1 --kind Application --resource --controller
$ make generate
$ make manifests
```

Commands used for the bundle creation:

```
$ export REGISTRY='docker.io'
$ export ORG='nheidloff'
$ export IMAGE='application-controller:v10'
$ make bundle IMG="$REGISTRY/$ORG/$IMAGE"
```