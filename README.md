# vault-PKI-k8s-demo

#Install and start minikube:  
https://github.com/kubernetes/minikube#installation  
#Install helm:  
https://github.com/kubernetes/helm#install  
#Install nginx ingress:  
```
helm install --name nginx-ingress stable/nginx-ingress --set controller.service.type="NodePort" --set controller.service.nodePorts.https="32443" --set controller.service.nodePorts.http="32080"
```
#Install consul:  
```
helm install --name consul stable/consul --set antiAffinity=soft
```
#Check consul:
```
minikube service consul-consul-ui
```
#Install Vault chart:
```
helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
helm install --name=vault incubator/vault --set vault.config.storage.consul.address="consul-consul:8500",vault.config.storage.consul.path="vault" --set vault.dev="false"
```
#Login to vault:
```
export POD_NAME=$(kubectl get pods --namespace default -l "app=vault" -o jsonpath="{.items[0].metadata.name}")
kubectl exec -it $POD_NAME sh
```
#Init vault:
```
export VAULT_ADDR='http://127.0.0.1:8200'
vault init -key-shares=1 -key-threshold=1
vault unseal <Unseal key>
vault auth <Root token>
```
#Configure PKI backend:
```
vault mount pki
vault write pki/root/generate/internal common_name=example.com ttl=87600h
vault write pki/roles/example-dot-com     allowed_domains=example.com     allow_subdomains=true max_ttl=1m
```
#Write root token to the pki-demo:
```
sed 's/__ROOT_TOKEN__/<Root token>/g' pki-demo.yaml.sed > pki-demo.yaml
```
#Create ingress, deploy ehocserver and create certficate job:
```
kubectl create -f pki-demo.yaml
```
#Add site to hosts:
```
echo $(minikube ip) pkidemo.default.example.com >> /etc/hosts
```
#Go to site url
```
xdg-open https://pkidemo.default.example.com:32443
```
#Or check it with openssl:
```
openssl s_client -servername pkidemo.default.example.com -connect $(minikube ip):32443 2>/dev/null | openssl x509 -inform pem -noout -issuer -serial -subject -dates
```
#Recreate certificate and check again:
```
kubectl delete job vault-secret-pki
kubectl apply -f pki-demo.yaml
```

#Cleanup:
```
helm delete --purge consul vault nginx-ingress
kubectl delete -f pki-demo.yaml
kubectl delete secret pkidemo.default.example.com
kubectl delete pvc $(kubectl get pvc -l "component=consul-consul" -o jsonpath="{.items[*].metadata.name}")
```