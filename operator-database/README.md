# operator-sample-go: operator

Work in progress ...

## Usage

operator-sdk init --domain third.party --repo github.com/nheidloff/operator-sample-go/operator-database

operator-sdk create api --group database.sample --version v1alpha1 --kind Database --resource --controller

make generate

make manifests

ibmcloud login -a cloud.ibm.com -r eu-de -g resource-group-niklas-heidloff7 --sso

ibmcloud ks cluster config --cluster xxxxxxx

kubectl create ns database

make install run

kubectl apply -f config/samples/database.sample_v1alpha1_database.yaml -n database 

kubectl delete -f config/samples/database.sample_v1alpha1_database.yaml -n database 
