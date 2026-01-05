@@Meta.Output="/deploy/test.yaml"

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: kubesphere
    component: {{.Config.ProjectName}}-dev
    tier: backend
  name: {{.Config.ProjectName}}-dev
  namespace: gencode
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  selector:
    matchLabels:
      app: kubesphere
      component: {{.Config.ProjectName}}-dev
      tier: backend
  template:
    metadata:
      labels:
        app: kubesphere
        component: {{.Config.ProjectName}}-dev
        tier: backend
    spec:
      containers:
        - env:
            - name: CACHE_IGNORE
              value: js|html
            - name: CACHE_PUBLIC_EXPIRATION
              value: 3d
          image: $REGISTRY/$DOCKERHUB_NAMESPACE/{{.Config.ProjectName}}:SNAPSHOT-$SAFE_BRANCH_NAME-$BUILD_NUMBER
          readinessProbe:
            tcpSocket:
              port: 8080
            timeoutSeconds: 10
            failureThreshold: 30
            periodSeconds: 5
          imagePullPolicy: Always
          name: {{.Config.ProjectName}}
          ports:
            - containerPort: 8080
              protocol: TCP
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      terminationGracePeriodSeconds: 30

---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: kubesphere
    component: {{.Config.ProjectName}}-dev
  name: {{.Config.ProjectName}}-dev
  namespace: gencode
spec:
  ports:
    - name: http
      port: 8080
      protocol: TCP
      targetPort: 8080
  selector:
    app: kubesphere
    component: {{.Config.ProjectName}}-dev
    tier: backend
  sessionAffinity: None
  type: NodePort