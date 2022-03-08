# operator-sample-go

Work in progress ...

## Usage

operator-sdk init --domain nheidloff --repo github.com/nheidloff/operator-sample-go

operator-sdk create api --group cache --version v1alpha1 --kind MyApplication --resource --controller

make generate

make manifests

ibmcloud login -a cloud.ibm.com -r eu-de -g resource-group-niklas-heidloff7 --sso

ibmcloud ks cluster config --cluster xxxxxxx

make install run

kubectl apply -f operator/config/samples/cache_v1alpha1_myapplication.yaml

kubectl create ns test1

kubectl config set-context --current --namespace=test1

kubectl get secret -l app=myapplication

ibmcloud ks worker ls --cluster niklas-heidloff-fra02-b3c.4x16

http://159.122.86.194:30548/hello

kubectl apply -f kubernetes/secret.yaml
kubectl apply -f kubernetes/microservice-deployment.yaml 
kubectl apply -f kubernetes/microservice-service.yaml

kubectl delete -f kubernetes/microservice-service.yaml
kubectl delete -f kubernetes/microservice-deployment.yaml 
kubectl delete -f kubernetes/secret.yaml

make build run