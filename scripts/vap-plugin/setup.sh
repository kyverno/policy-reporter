#!/usr/bin/env bash
#
# Spins up a local kind cluster wired end-to-end for vap-plugin: a static
# --audit-webhook-config-file pointing at vap-plugin's Service, sample
# ValidatingAdmissionPolicy bindings (one Deny, one Audit), the
# openreports.io CRDs, and vap-plugin itself deployed via the Helm chart.
#
# Two things about how this is wired are the product of a real spike against
# a live cluster (k8s v1.36.1), not assumptions - see README.md "Report
# lifecycle" and "What's observable in the audit log":
#
#   - kube-apiserver runs with hostNetwork: true as a static pod, so it uses
#     the *node's* /etc/resolv.conf, not cluster DNS - a Service DNS name in
#     the audit-webhook-config-file will not resolve. This script works
#     around that by giving vap-plugin's Service a fixed ClusterIP (chosen
#     up front, before the cluster exists) and pointing the webhook config
#     at that IP directly. The apiserver *can* reach a Service ClusterIP
#     (verified: kube-proxy's rules apply at the node level), it just can't
#     resolve its DNS name.
#   - audit-webhook-mode stays at its default (batch/async), NOT blocking.
#     blocking was tried first and broke cluster bootstrap entirely: with
#     it, every audited API call - including kubeadm's own bootstrap
#     requests, made before vap-plugin's Service even exists - blocks
#     waiting to deliver its audit event to the (not yet reachable) webhook.
#     kubeadm init then times out. Async batching just drops undeliverable
#     events in the background instead. audit-webhook-batch-max-wait is
#     shortened so events still show up in `kubectl get reports` reasonably
#     promptly once vap-plugin is actually up, without going fully
#     synchronous.
#
# Usage:
#   ./scripts/vap-plugin/setup.sh              create/update the cluster and deploy
#   ./scripts/vap-plugin/setup.sh teardown     delete the kind cluster
#
# Env overrides: CLUSTER_NAME, NAMESPACE, SERVICE_IP, IMAGE_TAG

set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-policy-reporter-vap-plugin-dev}"
NAMESPACE="${NAMESPACE:-policy-reporter}"
SERVICE_IP="${SERVICE_IP:-10.96.0.200}"
SERVICE_PORT="8443"
IMAGE_TAG="${IMAGE_TAG:-dev}"
KIND_CONTEXT="kind-${CLUSTER_NAME}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$ROOT_DIR/../.." && pwd)"

log() { echo "==> $*" >&2; }
die() { echo "error: $*" >&2; exit 1; }

require_cmd() {
	command -v "$1" >/dev/null 2>&1 && return
	case "$1" in
	kind) die "missing required command: kind (https://kind.sigs.k8s.io/docs/user/quick-start/#installation)" ;;
	ko) die "missing required command: ko (go install github.com/google/ko@latest)" ;;
	*) die "missing required command: $1" ;;
	esac
}

teardown() {
	log "deleting kind cluster '${CLUSTER_NAME}' (if it exists)"
	kind delete cluster --name "$CLUSTER_NAME"
	log "removing generated TLS material and audit logs"
	rm -rf "${ROOT_DIR}/.local/${CLUSTER_NAME}"
}

if [[ "${1:-}" == "teardown" || "${1:-}" == "down" ]]; then
	require_cmd kind
	teardown
	exit 0
fi

for cmd in docker kind kubectl ko helm openssl; do
	require_cmd "$cmd"
done

# Only scratch space for files kind/kubeadm just need to read once at
# cluster-creation time (the audit-webhook kubeconfig, the CA cert copy,
# etc). Anything that needs to keep existing for the cluster's whole
# lifetime - the audit log (bind-mounted into the running apiserver
# container) and the TLS material (re-read on every script run) - lives
# under the persistent CERT_DIR instead; putting those here too was tried
# first and broke both: this directory is deleted when the script exits.
WORK_DIR="$(mktemp -d -t vap-plugin-setup)"
cleanup() { rm -rf "$WORK_DIR"; }
trap cleanup EXIT

kctl() { kubectl --context "$KIND_CONTEXT" "$@"; }

CERT_DIR="$ROOT_DIR/.local/${CLUSTER_NAME}"
mkdir -p "$CERT_DIR/audit-logs"

CLUSTER_EXISTS=0
if kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
	CLUSTER_EXISTS=1
fi

if [[ "$CLUSTER_EXISTS" -eq 0 ]]; then
	log "generating a self-signed CA and server certificate for ${SERVICE_IP}"
	openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
		-keyout "$WORK_DIR/ca.key" -out "$CERT_DIR/ca.crt" \
		-subj "/CN=vap-plugin-dev-ca" >/dev/null 2>&1

	cat >"$WORK_DIR/server.ext" <<EOF
subjectAltName = IP:${SERVICE_IP},DNS:policy-reporter-vap-plugin.${NAMESPACE}.svc,DNS:policy-reporter-vap-plugin.${NAMESPACE}.svc.cluster.local
EOF

	openssl req -newkey rsa:2048 -nodes \
		-keyout "$CERT_DIR/tls.key" -out "$WORK_DIR/server.csr" \
		-subj "/CN=${SERVICE_IP}" >/dev/null 2>&1

	openssl x509 -req -in "$WORK_DIR/server.csr" \
		-CA "$CERT_DIR/ca.crt" -CAkey "$WORK_DIR/ca.key" -CAcreateserial \
		-out "$CERT_DIR/tls.crt" -days 3650 -extfile "$WORK_DIR/server.ext" >/dev/null 2>&1

	# Copied (not referenced directly), so we're not relying on ROOT_DIR
	# still being at this path once mounted into the kind node container.
	cp "$ROOT_DIR/deploy/audit-policy.yaml" "$WORK_DIR/audit-policy.yaml"

	cat >"$WORK_DIR/audit-webhook-kubeconfig.yaml" <<EOF
apiVersion: v1
kind: Config
clusters:
  - name: policy-reporter-vap-plugin
    cluster:
      server: https://${SERVICE_IP}:${SERVICE_PORT}/webhook
      certificate-authority: /etc/kubernetes/vap-plugin/ca.crt
contexts:
  - name: policy-reporter-vap-plugin
    context:
      cluster: policy-reporter-vap-plugin
current-context: policy-reporter-vap-plugin
EOF

	cat >"$WORK_DIR/kind-config.yaml" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraMounts:
      - hostPath: ${WORK_DIR}/audit-policy.yaml
        containerPath: /etc/kubernetes/vap-plugin/audit-policy.yaml
        readOnly: true
      - hostPath: ${WORK_DIR}/audit-webhook-kubeconfig.yaml
        containerPath: /etc/kubernetes/vap-plugin/audit-webhook-kubeconfig.yaml
        readOnly: true
      - hostPath: ${CERT_DIR}/ca.crt
        containerPath: /etc/kubernetes/vap-plugin/ca.crt
        readOnly: true
      - hostPath: ${CERT_DIR}/audit-logs
        containerPath: /var/log/vap-plugin
    kubeadmConfigPatches:
      - |
        kind: ClusterConfiguration
        apiServer:
          extraArgs:
            audit-policy-file: /etc/kubernetes/vap-plugin/audit-policy.yaml
            audit-webhook-config-file: /etc/kubernetes/vap-plugin/audit-webhook-kubeconfig.yaml
            audit-webhook-batch-max-wait: 1s
            audit-log-path: /var/log/vap-plugin/audit.log
            audit-log-format: json
          extraVolumes:
            - name: vap-plugin-audit
              hostPath: /etc/kubernetes/vap-plugin
              mountPath: /etc/kubernetes/vap-plugin
              readOnly: true
              pathType: Directory
            - name: vap-plugin-audit-logs
              hostPath: /var/log/vap-plugin
              mountPath: /var/log/vap-plugin
              pathType: DirectoryOrCreate
EOF

	log "creating kind cluster '${CLUSTER_NAME}'"
	kind create cluster --name "$CLUSTER_NAME" --config "$WORK_DIR/kind-config.yaml"
else
	log "kind cluster '${CLUSTER_NAME}' already exists, reusing it"
	if [[ ! -f "$CERT_DIR/tls.crt" ]]; then
		die "cluster exists but $CERT_DIR is missing its TLS material; run '$0 teardown' and re-run this script"
	fi
fi

log "installing openreports.io CRDs"
kctl apply -f https://raw.githubusercontent.com/openreports/reports-api/refs/heads/main/config/install.yaml

log "creating namespace '${NAMESPACE}'"
kctl create namespace "$NAMESPACE" --dry-run=client -o yaml | kctl apply -f -

log "creating vap-plugin-tls secret"
kctl -n "$NAMESPACE" create secret tls vap-plugin-tls \
	--cert="$CERT_DIR/tls.crt" --key="$CERT_DIR/tls.key" \
	--dry-run=client -o yaml | kctl apply -f -

log "deploying policy-reporter via Helm"
# config.report.reportDenied is turned on here (the app default is off,
# see README "Reporting Deny results") so this demo still shows both
# sample policies' Reports - the whole point of running it.
helm upgrade --install policy-reporter "$PROJECT_DIR/charts/policy-reporter" \
	--kube-context "$KIND_CONTEXT" \
	--namespace "$NAMESPACE" \
	--set ui.enabled=true \
	--set plugin.vap.enabled=true \
	--set plugin.vap.service.clusterIP="$SERVICE_IP" \
	--set plugin.vap.tls.existingSecret=vap-plugin-tls \
	--set plugin.vap.config.report.reportDenied=false \
	--wait --timeout=2m

# Applied only now, after vap-plugin's own Deployment exists, as extra
# safety margin: the sample Deny binding's namespaceSelector excludes the
# vap-plugin namespace (see deploy/sample-vap-policies.yaml) specifically
# so it can never block vap-plugin's own pods - past initial creation too,
# since a rollout restart's new replica is a Create request like any other,
# and ValidatingAdmissionPolicy evaluates at admission time on every one of
# those, not just the first. A first version of this without that exclusion
# blocked every future rollout of vap-plugin once the policy was live, not
# just the initial bootstrap - worth the reminder for any cluster-wide Deny
# policy that doesn't carve out its own reporting/infra namespaces.
log "applying sample ValidatingAdmissionPolicy + bindings"
kctl apply -f "$ROOT_DIR/deploy/sample-vap-policies.yaml"

cat >&2 <<EOF

==> Ready.

Cluster:    ${CLUSTER_NAME} (context: ${KIND_CONTEXT})
Namespace:  ${NAMESPACE}
Webhook:    https://${SERVICE_IP}:${SERVICE_PORT}/webhook
Audit log:  docker exec ${CLUSTER_NAME}-control-plane cat /var/log/vap-plugin/audit.log

Trigger the sample policies:
  kubectl --context ${KIND_CONTEXT} create ns demo
  kubectl --context ${KIND_CONTEXT} run denied  -n demo --image=nginx --labels="owner=alice"          # blocked (missing 'team')
  kubectl --context ${KIND_CONTEXT} run flagged -n demo --image=nginx --labels="team=payments"         # allowed, flagged (missing 'owner')

Watch reports appear:
  kubectl --context ${KIND_CONTEXT} get reports -A -w

Tail vap-plugin's logs:
  kubectl --context ${KIND_CONTEXT} -n ${NAMESPACE} logs -l app.kubernetes.io/name=vap-plugin -f

Tear down:
  ./scripts/setup.sh teardown
EOF
