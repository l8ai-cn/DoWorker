package coordinator

import (
	"context"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

type k8sPodSpec struct {
	Name              string
	Namespace         string
	Image             string
	ImagePullPolicy   string
	BackendURL        string
	GRPCEndpoint      string
	RelayBaseURL      string
	RunnerNodeID      string
	RunnerOrgSlug     string
	MaxConcurrentPods int
	SSLHostPath       string
	SSLSecretName     string
}

const runnerPodTemplate = `apiVersion: v1
kind: Pod
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: agentsmesh-runner
    app.kubernetes.io/component: coordinator-provisioned
spec:
  restartPolicy: Never
  containers:
    - name: runner
      image: {{ .Image }}
      imagePullPolicy: {{ .ImagePullPolicy }}
      command: ["/usr/local/bin/runner-entrypoint.sh"]
      env:
        - name: BACKEND_URL
          value: {{ quote .BackendURL }}
        - name: GRPC_ENDPOINT
          value: {{ quote .GRPCEndpoint }}
        - name: RELAY_BASE_URL
          value: {{ quote .RelayBaseURL }}
        - name: RUNNER_NODE_ID
          value: {{ quote .RunnerNodeID }}
        - name: RUNNER_ORG_SLUG
          value: {{ quote .RunnerOrgSlug }}
        - name: MAX_CONCURRENT_PODS
          value: {{ quote (printf "%d" .MaxConcurrentPods) }}
        - name: AGENTSMESH_MCP_BIND
          value: "0.0.0.0"
{{- if .SSLHostPath }}
      volumeMounts:
        - name: ssl
          mountPath: /app/ssl
          readOnly: true
{{- else if .SSLSecretName }}
      volumeMounts:
        - name: ssl
          mountPath: /app/ssl
          readOnly: true
{{- end }}
{{- if .SSLHostPath }}
  volumes:
    - name: ssl
      hostPath:
        path: {{ .SSLHostPath }}
        type: Directory
{{- else if .SSLSecretName }}
  volumes:
    - name: ssl
      secret:
        secretName: {{ .SSLSecretName }}
{{- end }}
`

func renderRunnerPod(spec k8sPodSpec) ([]byte, error) {
	funcs := template.FuncMap{
		"quote": strconv.Quote,
	}
	tpl, err := template.New("runner-pod").Funcs(funcs).Parse(runnerPodTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, spec); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func kubectlBaseArgs(cfg k8sLauncherConfig) []string {
	args := make([]string, 0, 4)
	if kubeconfig := strings.TrimSpace(cfg.Kubeconfig); kubeconfig != "" {
		args = append(args, "--kubeconfig", kubeconfig)
	}
	if ns := strings.TrimSpace(cfg.Namespace); ns != "" {
		args = append(args, "-n", ns)
	}
	return args
}

func podPhase(ctx context.Context, l *K8sLauncher, name string) (string, error) {
	args := append(kubectlBaseArgs(l.cfg), "get", "pod", name, "-o", "jsonpath={.status.phase}")
	stdout, stderr, err := l.runner.Run(ctx, l.cfg.Kubectl, args...)
	if err != nil {
		if strings.Contains(stderr, "NotFound") {
			return "", nil
		}
		return "", fmt.Errorf("kubectl get pod: %w: %s", err, strings.TrimSpace(stderr))
	}
	return strings.TrimSpace(stdout), nil
}
