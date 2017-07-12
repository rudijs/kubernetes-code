minikube start
#minikube start --insecure-registry "192.168.99.100/32"

minikube dashboard

kubectl create -f registry-rc.yaml 
kubectl create -f registry-svc.yaml 
kubectl create -f registry-pods.yaml

export MINIKUBE_IP=$(minikube ip)
export MINIKUBE_REGISTRY_PORT=$(kubectl get svc kube-registry --namespace=kube-system -o json | jq '.spec.ports[0].nodePort')

curl ${MINIKUBE_IP}:${MINIKUBE_REGISTRY_PORT}/v2/_catalog

docker-machine create --engine-insecure-registry ${MINIKUBE_IP}:${MINIKUBE_REGISTRY_PORT} -d virtualbox --virtualbox-hostonly-cidr 192.168.77.1/24 --virtualbox-memory '1024' --virtualbox-boot2docker-url https://releases.rancher.com/os/latest/rancheros.iso demo1

eval $(docker-machine env demo1)

docker pull mongo:latest  

docker tag mongo:latest ${MINIKUBE_IP}:${MINIKUBE_REGISTRY_PORT}/mongo:latest

docker push ${MINIKUBE_IP}:${MINIKUBE_REGISTRY_PORT}/mongo:latest   



kubectl create -f mongo.yaml
kubectl expose deployment mongo-deployment --type=NodePort
kubectl get services
eval $(docker-machine env myaxa-digital-account)
docker ps
docker exec -it mongo bash
kubectl get pods
kubectl exec -it mongo-deployment-3389199766-mz1xz bash

