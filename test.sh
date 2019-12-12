#!/usr/bin/env bash


echo "[0] TEST CREATE"

kubectl delete hr test-nginx
kubectl delete cm test-nginx
kubectl create  cm test-nginx --from-file=tests/values.yaml

kubectl captain create test-nginx --chart=stable/nginx-ingress --version=1.26.2 --set=a=b -w --timeout=30 --configmap=test-nginx

kubectl get hr test-nginx -o yaml

if [ $(kubectl get hr test-nginx -o json | jq -r .status.phase) != "Synced" ]
then
   echo "HelmRequest test-nginx not synced"
   exit 1
fi


echo "HelmRequest test-nginx synced"


echo "[1] TEST UPGRADE FAILED"

kubectl captain upgrade test-nginx --repo no


until [ $(kubectl get hr test-nginx -o json | jq -r .status.phase)  == "Failed" ]
do
    echo "Wait for hr to be failed..."
    sleep 1
done


echo "[2] TEST UPGGRADE"
kubectl delete cm test-nginx-2
kubectl create  cm test-nginx-2 --from-file=tests/values.yaml
kubectl captain upgrade test-nginx --repo=stable --version=1.26.2 -w --timeout=30 --configmap=test-nginx-2

kubectl captain upgrade test-nginx  --version=1.26.1 -w --timeout=30
if [ $(kubectl get hr test-nginx -o json | jq -r .status.phase) != "Synced" ]
then
    echo "HelmRequest test-nginx not synced"
    exit 1
fi

if [ $(kubectl get hr test-nginx -o json | jq -r .spec.version) != "1.26.1" ]
then
    echo "HelmRequest test-nginx not synced"
    exit 1
fi

kubectl get hr test-nginx -o yaml


echo "[3] TEST ROLLBACK"

kubectl captain rollback test-nginx

if [ $(kubectl get hr test-nginx -o json | jq -r .spec.version) != "1.26.2" ]
then
    echo "HelmRequest rollback version error"
    exit 1
fi


echo "[4] TEST CREATE CHARTREPO"
kubectl delete ctr test-repo -n captain
kubectl captain create-repo test-repo --url=https://alauda.github.io/captain-test-charts/ -n captain -w --timeout=30
if [ $(kubectl get ctr test-repo -n captain -o json | jq -r .status.phase) != "Synced" ]
then
    echo "ChartRepo test-repo not synced"
    exit 1
fi


kubectl delete cm test-nginx
kubectl delete ctr test-repo -n captain
kubectl delete hr test-nginx 
