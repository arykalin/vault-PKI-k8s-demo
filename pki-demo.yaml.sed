apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: ingress-example
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  tls:
  - hosts:
    - pkidemo.default.example.com
    secretName: pkidemo.default.example.com
  rules:
  - host: pkidemo.default.example.com
    http:
      paths:
      - path: /
        backend:
          serviceName: echoserver
          servicePort: 8080
---
apiVersion: batch/v1
kind: Job
metadata:
  name: vault-secret-pki
spec:
  template:
    spec:
      containers:
      - name: getsecretfromvault
        image: arykalin/getsecretfromvault:latest
        imagePullPolicy: IfNotPresent
        command: ["bash", "-c", "/go/src/app/getSecretFromVault"]
        env:
        - name: VAULT_ADDR
          value: "http://vault-vault:8200"
        - name: CERT_NAME
          value: "pkidemo.default.example.com"
        - name: ROLE_NAME
          value: "example-dot-com"
        - name: VAULT_TOKEN
          value: "__ROOT_TOKEN__"
      restartPolicy: Never
  backoffLimit: 4
#---
#apiVersion: batch/v1beta1
#kind: CronJob
#metadata:
#  name: vault-secret-pki
#spec:
#  schedule: "*/1 * * * *"
#  jobTemplate:
#    spec:
#      template:
#        spec:
#          containers:
#          - name: getsecretfromvault
#            image: arykalin/getsecretfromvault:latest
#            command: ["/bin/sleep", "200000"]
#            env:
#            - name: VAULT_ADDR
#              value: "vault-vault:8200"
#            - name: CERT_NAME
#              value: "pkidemo.default.example.com"
#            - name: ROLE_NAME
#              value: "example-dot-com"
#            - name: VAULT_TOKEN
#              value: "__ROOT_TOKEN__"
#          restartPolicy: OnFailure
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: echoserver
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: echoserver
    spec:
      containers:
      - name: echoserver
        image: gcr.io/google_containers/echoserver:1.4
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: echoserver
spec:
  ports:
  - port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    run: echoserver
  sessionAffinity: None
  type: NodePort
status:
  loadBalancer: {}
