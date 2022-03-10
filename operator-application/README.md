# operator-sample-go: operator

Work in progress ...

## Usage

operator-sdk init --domain ibm.com --repo github.com/nheidloff/operator-sample-go/operator-application

operator-sdk create api --group application.sample --version v1alpha1 --kind Application --resource --controller

make generate

make manifests

ibmcloud login -a cloud.ibm.com -r eu-de -g resource-group-niklas-heidloff7 --sso

ibmcloud ks cluster config --cluster xxxxxxx

kubectl create ns test1

kubectl config set-context --current --namespace=test1

make install run

kubectl apply -f config/samples/application.sample_v1alpha1_application.yaml

kubectl apply -f kubernetes/secret.yaml
kubectl apply -f kubernetes/microservice-deployment.yaml 
kubectl apply -f kubernetes/microservice-service.yaml

ibmcloud ks worker ls --cluster niklas-heidloff-fra02-b3c.4x16

http://159.122.86.194:30548/hello

kubectl delete -f kubernetes/microservice-service.yaml
kubectl delete -f kubernetes/microservice-deployment.yaml 
kubectl delete -f kubernetes/secret.yaml

