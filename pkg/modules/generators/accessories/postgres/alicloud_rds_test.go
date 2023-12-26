package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"

	"kusionstack.io/kusion/pkg/modules/inputs"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestGenerateAlicloudResources(t *testing.T) {
	g := genAlicloudPostgreSQLGenerator()

	spec := &apiv1.Intent{}
	secret, err := g.generateAlicloudResources(g.postgres, spec)

	hostAddress := "$kusion_path.aliyun:alicloud:alicloud_db_connection:testapp.connection_string"
	username := g.postgres.Username
	password := "$kusion_path.hashicorp:random:random_password:testapp-db.result"
	data := make(map[string]string)
	data["hostAddress"] = hostAddress
	data["username"] = username
	data["password"] = password
	expectedSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      g.appName + dbResSuffix,
			Namespace: g.project.Name,
		},
		StringData: data,
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secret)
}

func TestGenerateAlicloudDBInstance(t *testing.T) {
	g := genAlicloudPostgreSQLGenerator()
	alicloudProvider := &inputs.Provider{}
	alicloudProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])
	_ = alicloudProvider.SetString(alicloudProviderURL)
	alicloudProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])

	alicloudDBInstanceID, r := g.generateAlicloudDBInstance(alicloudProviderRegion, alicloudProvider, g.postgres)
	expectedAlicloudDBInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	expectedRes := apiv1.Resource{
		ID:   "aliyun:alicloud:alicloud_db_instance:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"category":                 g.postgres.Category,
			"db_instance_storage_type": "cloud_essd",
			"engine":                   "PostgreSQL",
			"engine_version":           g.postgres.Version,
			"instance_charge_type":     "Serverless",
			"instance_storage":         g.postgres.Size,
			"instance_type":            g.postgres.InstanceType,
			"security_ips":             g.postgres.SecurityIPs,
			"serverless_config": []alicloudServerlessConfig{
				{
					AutoPause:   false,
					SwitchForce: false,
					MaxCapacity: 8,
					MinCapacity: 1,
				},
			},
			"vswitch_id": g.postgres.SubnetID,
		},
		Extensions: map[string]interface{}{
			"provider": alicloudProviderURL,
			"providerMeta": map[string]interface{}{
				"region": alicloudProviderRegion,
			},
			"resourceType": "alicloud_db_instance",
		},
	}

	assert.Equal(t, expectedAlicloudDBInstanceID, alicloudDBInstanceID)
	assert.Equal(t, expectedRes, r)
}

func TestGenerateAlicloudDBConnection(t *testing.T) {
	g := genAlicloudPostgreSQLGenerator()
	alicloudProvider := &inputs.Provider{}
	alicloudProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])
	_ = alicloudProvider.SetString(alicloudProviderURL)
	alicloudProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])

	dbInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	alicloudDBConnectionID, r := g.generateAlicloudDBConnection(dbInstanceID, alicloudProviderRegion, alicloudProvider)
	expectedAlicloudDBConnectionID := "aliyun:alicloud:alicloud_db_connection:testapp"
	expectedRes := apiv1.Resource{
		ID:   "aliyun:alicloud:alicloud_db_connection:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"instance_id": "$kusion_path.aliyun:alicloud:alicloud_db_instance:testapp.id",
		},
		Extensions: map[string]interface{}{
			"provider": alicloudProviderURL,
			"providerMeta": map[string]interface{}{
				"region": alicloudProviderRegion,
			},
			"resourceType": "alicloud_db_connection",
		},
	}

	assert.Equal(t, expectedAlicloudDBConnectionID, alicloudDBConnectionID)
	assert.Equal(t, expectedRes, r)
}

func TestGenerateAlicloudRDSAccount(t *testing.T) {
	g := genAlicloudPostgreSQLGenerator()
	alicloudProvider := &inputs.Provider{}
	alicloudProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])
	_ = alicloudProvider.SetString(alicloudProviderURL)
	alicloudProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AlicloudProvider])

	accountName := g.postgres.Username
	randomPasswordID := "hashicorp:random:random_password:testapp-db"
	alicloudDBInstanceID := "aliyun:alicloud:alicloud_db_instance:testapp"
	r := g.generateAlicloudRDSAccount(accountName, randomPasswordID, alicloudDBInstanceID, alicloudProviderRegion, alicloudProvider, g.postgres)

	expectedRes := apiv1.Resource{
		ID:   "aliyun:alicloud:alicloud_rds_account:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"account_name":     accountName,
			"account_password": "$kusion_path.hashicorp:random:random_password:testapp-db.result",
			"account_type":     "Super",
			"db_instance_id":   "$kusion_path.aliyun:alicloud:alicloud_db_instance:testapp.id",
		},
		Extensions: map[string]interface{}{
			"provider": alicloudProviderURL,
			"providerMeta": map[string]interface{}{
				"region": alicloudProviderRegion,
			},
			"resourceType": "alicloud_rds_account",
		},
	}

	assert.Equal(t, expectedRes, r)
}

func genAlicloudPostgreSQLGenerator() *postgresGenerator {
	project := &apiv1.Project{
		Name: "testproject",
	}
	stack := &apiv1.Stack{
		Name: "teststack",
	}
	appName := "testapp"
	workload := &workload.Workload{}
	postgres := &postgres.PostgreSQL{
		Type:           "cloud",
		Version:        "5.7",
		Size:           20,
		InstanceType:   "postgres.n2.serverless.1c",
		Category:       "serverless_basic",
		PrivateRouting: false,
		SubnetID:       "test_subnet_id",
		Username:       defaultUsername,
		SecurityIPs:    defaultSecurityIPs,
	}
	ws := &apiv1.Workspace{
		Name: "testworkspace",
		Runtimes: &apiv1.RuntimeConfigs{
			Kubernetes: &apiv1.KubernetesConfig{
				KubeConfig: "/Users/username/testkubeconfig",
			},
			Terraform: apiv1.TerraformConfig{
				"random": &apiv1.ProviderConfig{
					Source:  "hashicorp/random",
					Version: "3.5.1",
				},
				"alicloud": &apiv1.ProviderConfig{
					Source:  "aliyun/alicloud",
					Version: "1.209.1",
					GenericConfig: apiv1.GenericConfig{
						"region": "cn-beijing",
					},
				},
			},
		},
		Modules: apiv1.ModuleConfigs{
			"postgres": &apiv1.ModuleConfig{
				Default: apiv1.GenericConfig{
					"cloud":          "alicloud",
					"size":           20,
					"instanceType":   "postgres.n2.serverless.1c",
					"category":       "serverless_basic",
					"privateRouting": false,
					"subnetID":       "test_subnet_id",
				},
			},
		},
	}

	return &postgresGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		postgres: postgres,
		ws:       ws,
	}
}
