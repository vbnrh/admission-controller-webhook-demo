#!/usr/bin/env bash

# Copyright (c) 2019 StackRox Inc.
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

# deploy.sh
#
# Sets up the environment for the admission controller webhook demo in the active cluster.

set -euo pipefail


# Set our known directories and parameters.
BASE_DIR=$(cd $(dirname $0)/..; pwd)
NAMESPACE="openshift-storage"
INSTALL_SELF_SIGNED_CERT=true



echo $BASE_DIR

if [ "${INSTALL_SELF_SIGNED_CERT}" == true ]; then
	${BASE_DIR}/hack/webhook-create-signed-cert.sh --namespace ${NAMESPACE}
fi
echo "Deploying service.yaml"
kubectl -n ${NAMESPACE} create -f ${BASE_DIR}/deployment/service.yaml
echo "Deployed service.yaml"

echo "Deploying webhook.yaml"
export NAMESPACE
cat ${BASE_DIR}/deployment/webhook.yaml | \
	${BASE_DIR}/hack/webhook-patch-ca-bundle.sh | \
	sed -e "s|\${NAMESPACE}|${NAMESPACE}|g" | \
	kubectl -n ${NAMESPACE} create -f -
echo "Deployed webhook.yaml"

echo "Deploying roles.yaml"
kubectl -n ${NAMESPACE} create -f ${BASE_DIR}/deployment/roles.yaml
echo "Deployed roles.yaml"
echo "Deploying deployment.yaml"
kubectl -n ${NAMESPACE} create -f ${BASE_DIR}/deployment/deployment.yaml
echo "Deployed deployment.yaml"


#basedir="$(dirname "$0")/deployment"
#keydir="$(mktemp -d)"


# Generate keys into a temporary directory.
#echo "Generating TLS keys ..."
#"${basedir}/generate-keys.sh" "$keydir"

# Create the `webhook-demo` namespace. This cannot be part of the YAML file as we first need to create the TLS secret,
# which would fail otherwise.
#echo "Creating Kubernetes objects ..."
#kubectl create namespace webhook-demo

# Create the TLS secret for the generated keys.
#oc -n openshift-storage create secret tls webhook-server-tls \
#    --cert "${keydir}/webhook-server-tls.crt" \
#    --key "${keydir}/webhook-server-tls.key"

# Read the PEM-encoded CA certificate, base64 encode it, and replace the `${CA_PEM_B64}` placeholder in the YAML
# template with it. Then, create the Kubernetes resources.
#ca_pem_b64="$(openssl base64 -A <"${keydir}/ca.crt")"
#sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"${basedir}/deployment.yaml.template" \
#    | oc create -f -

# Delete the key directory to prevent abuse (DO NOT USE THESE KEYS ANYWHERE ELSE).
#rm -rf "$keydir"

echo "The webhook server has been deployed and configured!"
