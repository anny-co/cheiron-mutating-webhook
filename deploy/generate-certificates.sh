#!/usr/bin/env bash

# Copyright (c) 2021 anny UG (haftungsbeschränkt)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Generate the private key for the webhook server
openssl genrsa -out webhook-server-tls.key 2048
# Generate a Certificate Signing Request (CSR) for the private key
openssl req -new -key webhook-server-tls.key -subj "/CN=Cheiron Mutating Webhook Server/O=edit" -out webhook-server-tls.csr

CSR=$(cat webhook-server-tls.csr | base64 | tr -d "\n")

cat <<EOF > csr.yaml
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
    name: cheiron-mutating-webhook-server-tls
spec:
    request: $CSR
    signerName: kubernetes.io/kube-apiserver-client
    # expirationSeconds: 2592000
    usages:
    - client auth
    - server auth
EOF

kubectl apply -f csr.yaml

kubectl get csr

kubectl certificate approve cheiron-mutating-webhook-server-tls

sleep 2

kubectl get csr cheiron-mutating-webhook-server-tls -o jsonpath='{.status.certificate}'| base64 -d > webhook-server-tls.crt

kubectl -n cheiron-system create secret tls cheiron-mutating-webhook-server-tls \
    --cert "webhook-server-tls.crt" \
    --key "webhook-server-tls.key"

kubectl delete csr cheiron-mutating-webhook-server-tls

rm webhook-server-tls.crt
rm webhook-server-tls.key
rm webhook-server-tls.csr
rm csr.yaml