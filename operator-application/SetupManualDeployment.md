# Setup and manual Deployment

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
$ operator-sdk run bundle "docker.io/nheidloff/database-controller-bundle:v1" -n operators
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