package tests_test

import (
  "sigs.k8s.io/kustomize/k8sdeps/kunstruct"
  "sigs.k8s.io/kustomize/k8sdeps/transformer"
  "sigs.k8s.io/kustomize/pkg/fs"
  "sigs.k8s.io/kustomize/pkg/loader"
  "sigs.k8s.io/kustomize/pkg/resmap"
  "sigs.k8s.io/kustomize/pkg/resource"
  "sigs.k8s.io/kustomize/pkg/target"
  "testing"
)

func writeApiServiceBase(th *KustTestHarness) {
  th.writeF("/manifests/pipeline/api-service/base/deployment.yaml", `
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  labels:
    app: ml-pipeline
  name: ml-pipeline
spec:
  selector:
    matchLabels:
      app: ml-pipeline
  template:
    metadata:
      labels:
        app: ml-pipeline
    spec:
      containers:
      - env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: gcr.io/ml-pipeline/api-server:0.1.14
        imagePullPolicy: IfNotPresent
        name: ml-pipeline-api-server
        ports:
        - containerPort: 8888
        - containerPort: 8887
      serviceAccountName: ml-pipeline
`)
  th.writeF("/manifests/pipeline/api-service/base/role-binding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  labels:
    app: ml-pipeline
  name: ml-pipeline
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ml-pipeline
subjects:
- kind: ServiceAccount
  name: ml-pipeline
  namespace: kubeflow
`)
  th.writeF("/manifests/pipeline/api-service/base/role.yaml", `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  labels:
    app: ml-pipeline
  name: ml-pipeline
rules:
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - create
  - get
  - list
  - watch
  - update
  - patch
  - delete
- apiGroups:
  - kubeflow.org
  resources:
  - scheduledworkflows
  verbs:
  - create
  - get
  - list
  - update
  - patch
  - delete
`)
  th.writeF("/manifests/pipeline/api-service/base/sa.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ml-pipeline
`)
  th.writeF("/manifests/pipeline/api-service/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: ml-pipeline
  name: ml-pipeline
spec:
  ports:
  - name: http
    port: 8888
    protocol: TCP
    targetPort: 8888
  - name: grpc
    port: 8887
    protocol: TCP
    targetPort: 8887
  selector:
    app: ml-pipeline
`)
  th.writeK("/manifests/pipeline/api-service/base", `
resources:
- deployment.yaml
- role-binding.yaml
- role.yaml
- sa.yaml
- service.yaml

images:
- name: gcr.io/ml-pipeline/api-server
  newTag: '0.1.14'
`)
}

func TestApiServiceBase(t *testing.T) {
  th := NewKustTestHarness(t, "/manifests/pipeline/api-service/base")
  writeApiServiceBase(th)
  m, err := th.makeKustTarget().MakeCustomizedResMap()
  if err != nil {
    t.Fatalf("Err: %v", err)
  }
  targetPath := "../pipeline/api-service/base"
  fsys := fs.MakeRealFS()
    _loader, loaderErr := loader.NewLoader(targetPath, fsys)
    if loaderErr != nil {
      t.Fatalf("could not load kustomize loader: %v", loaderErr)
    }
    rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
    kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl())
    if err != nil {
      th.t.Fatalf("Unexpected construction error %v", err)
    }
  n, err := kt.MakeCustomizedResMap()
  if err != nil {
    t.Fatalf("Err: %v", err)
  }
  expected, err := n.EncodeAsYaml()
  th.assertActualEqualsExpected(m, string(expected))
}