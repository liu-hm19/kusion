package monitoring

import (
	"fmt"
	"strings"
	"testing"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/modules/inputs/monitoring"
)

type Fields struct {
	project *project.Project
	monitor *monitoring.Monitor
	appName string
}

type Args struct {
	spec *intent.Intent
}

type TestCase struct {
	name    string
	fields  Fields
	args    Args
	want    *intent.Intent
	wantErr bool
}

func BuildMonitoringTestCase(
	projectName, appName string,
	interval, timeout prometheusv1.Duration,
	path, port, scheme string,
	monitorType project.MonitorType,
	operatorMode bool,
) *TestCase {
	var endpointType string
	var monitorKind project.MonitorType
	if monitorType == "Service" {
		monitorKind = "ServiceMonitor"
		endpointType = "endpoints"
	} else if monitorType == "Pod" {
		monitorKind = "PodMonitor"
		endpointType = "podMetricsEndpoints"
	}
	expectedResources := make([]intent.Resource, 0)
	if operatorMode {
		expectedResources = []intent.Resource{
			{
				ID:   fmt.Sprintf("monitoring.coreos.com/v1:%s:%s:%s-%s-monitor", monitorKind, projectName, appName, strings.ToLower(string(monitorType))),
				Type: "Kubernetes",
				Attributes: map[string]interface{}{
					"apiVersion": "monitoring.coreos.com/v1",
					"kind":       string(monitorKind),
					"metadata": map[string]interface{}{
						"creationTimestamp": nil,
						"name":              fmt.Sprintf("%s-%s-monitor", appName, strings.ToLower(string(monitorType))),
						"namespace":         projectName,
					},
					"spec": map[string]interface{}{
						endpointType: []interface{}{
							map[string]interface{}{
								"bearerTokenSecret": map[string]interface{}{
									"key": "",
								},
								"interval":      string(interval),
								"scrapeTimeout": string(timeout),
								"path":          path,
								"port":          port,
								"scheme":        scheme,
							},
						},
						"namespaceSelector": make(map[string]interface{}),
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"kusion_monitoring_appname": appName,
							},
						},
					},
				},
				DependsOn: nil,
				Extensions: map[string]interface{}{
					"GVK": fmt.Sprintf("monitoring.coreos.com/v1, Kind=%s", string(monitorKind)),
				},
			},
		}
	}
	testCase := &TestCase{
		name: fmt.Sprintf("%s-%s", projectName, appName),
		fields: Fields{
			project: &project.Project{
				ProjectConfiguration: project.ProjectConfiguration{
					Name: projectName,
					Prometheus: &project.PrometheusConfig{
						OperatorMode: operatorMode,
						MonitorType:  monitorType,
					},
				},
				Path: "/test-project",
			},
			monitor: &monitoring.Monitor{
				Interval: interval,
				Timeout:  timeout,
				Path:     path,
				Port:     port,
				Scheme:   scheme,
			},
			appName: appName,
		},
		args: Args{
			spec: &intent.Intent{},
		},
		want: &intent.Intent{
			Resources: expectedResources,
		},
		wantErr: false,
	}
	return testCase
}

func TestMonitoringGenerator_Generate(t *testing.T) {
	tests := []TestCase{
		*BuildMonitoringTestCase("test-project", "test-app", "15s", "5s", "/metrics", "web", "http", "Service", true),
		*BuildMonitoringTestCase("test-project", "test-app", "15s", "5s", "/metrics", "web", "http", "Pod", true),
		*BuildMonitoringTestCase("test-project", "test-app", "30s", "15s", "/metrics", "8080", "http", "Service", false),
		*BuildMonitoringTestCase("test-project", "test-app", "30s", "15s", "/metrics", "8080", "http", "Pod", false),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &monitoringGenerator{
				project: tt.fields.project,
				monitor: tt.fields.monitor,
				appName: tt.fields.appName,
			}
			if err := g.Generate(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, tt.args.spec)
		})
	}
}