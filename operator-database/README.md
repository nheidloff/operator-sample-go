# operator-sample-go: operator

Work in progress ...

## Usage

operator-sdk init --domain nheidloff --repo github.com/nheidloff/operator-database

operator-sdk create api --group cache --version v1alpha1 --kind ExternalDatabase --resource --controller

make generate

make manifests

ibmcloud login -a cloud.ibm.com -r eu-de -g resource-group-niklas-heidloff7 --sso

ibmcloud ks cluster config --cluster xxxxxxx

make install run

kubectl apply -f config/samples/cache_v1alpha1_externaldatabase.yaml

kubectl create ns test1

kubectl config set-context --current --namespace=test1


