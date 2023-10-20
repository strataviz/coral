#!/bin/bash

openssl genrsa -out ca.key 2048

openssl req -new -x509 -days 365 -key ca.key \
  -subj "/C=US/CN=coral-stvz-io-webhook"\
  -out ca.crt

openssl req -newkey rsa:2048 -nodes -keyout server.key \
  -subj "/C=US/CN=coral-stvz-io-webhook" \
  -out server.csr

openssl x509 -req \
  -extfile <(printf "subjectAltName=DNS:coral-stvz-io-webhook.coral.svc") \
  -days 365 \
  -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt

cat ca.crt | base64 | fold > cabundle.crt

cat > config/overlays/dev/mutating-webhooks.yaml << EOF
---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: na
webhooks:
- name: mbuilder.stvz.io
  clientConfig:
    caBundle: "$(awk '{printf "%s\\n", $0}' cabundle.crt)"
    service:
      name: coral-stvz-io-webhook
      namespace: coral
      port: 9443
EOF

cat > config/overlays/dev/validating-webhooks.yaml << EOF
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: na
webhooks:
- name: vbuilder.stvz.io
  clientConfig:
    caBundle: "$(awk '{printf "%s\\n", $0}' cabundle.crt)"
    service:
      name: coral-stvz-io-webhook
      namespace: coral
      port: 9443
EOF

mv server.crt ./config/overlays/dev/tls.crt
mv server.key ./config/overlays/dev/tls.key

rm ca.crt ca.key ca.srl server.csr cabundle.crt
