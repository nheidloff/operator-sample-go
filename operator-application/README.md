# operator-application

See below for instructions how to set up and run the application operator as well as the used commands for the development of it.

### Setup and Usage

The instructions below assume that you use the managed Kubernetes service on the IBM Cloud. You can also use any other Kubernetes service or OpenShift.

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

### Development Commands

Manual Setup of the Application Resources only:

```
$ kubectl apply -f kubernetes/secret.yaml
$ kubectl apply -f kubernetes/microservice-deployment.yaml 
$ kubectl apply -f kubernetes/microservice-service.yaml
$ kubectl delete -f kubernetes/microservice-service.yaml
$ kubectl delete -f kubernetes/microservice-deployment.yaml 
$ kubectl delete -f kubernetes/secret.yaml
```

Commands used for the Project Creation:

```
$ operator-sdk init --domain ibm.com --repo github.com/nheidloff/operator-sample-go/operator-application
$ operator-sdk create api --group application.sample --version v1alpha1 --kind Application --resource --controller
$ make generate
$ make manifests
```