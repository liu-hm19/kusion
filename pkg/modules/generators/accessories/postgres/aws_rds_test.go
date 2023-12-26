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

func TestGenerateAWSResources(t *testing.T) {
	g := genAWSPostgreSQLGenerator()

	spec := &apiv1.Intent{}
	secret, err := g.generateAWSResources(g.postgres, spec)

	hostAddress := "$kusion_path.hashicorp:aws:aws_db_instance:testapp.address"
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

func TestGenerateAWSSecurityGroup(t *testing.T) {
	g := genAWSPostgreSQLGenerator()
	awsProvider := &inputs.Provider{}
	awsProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AWSProvider])
	_ = awsProvider.SetString(awsProviderURL)
	awsProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AWSProvider])

	awsSecurityGroupID, r, err := g.generateAWSSecurityGroup(awsProvider, awsProviderRegion, g.postgres)
	expectedAWSSecurityGroupID := "hashicorp:aws:aws_security_group:testapp-db"
	expectedRes := apiv1.Resource{
		ID:   "hashicorp:aws:aws_security_group:testapp-db",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"egress": []awsSecurityGroupTraffic{
				{
					CidrBlocks: []string{"0.0.0.0/0"},
					Protocol:   "-1",
					FromPort:   0,
					ToPort:     0,
				},
			},
			"ingress": []awsSecurityGroupTraffic{
				{
					CidrBlocks: []string{"0.0.0.0/0"},
					Protocol:   "tcp",
					FromPort:   3306,
					ToPort:     3306,
				},
			},
		},
		Extensions: map[string]interface{}{
			"provider": awsProviderURL,
			"providerMeta": map[string]interface{}{
				"region": awsProviderRegion,
			},
			"resourceType": awsSecurityGroup,
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedAWSSecurityGroupID, awsSecurityGroupID)
	assert.Equal(t, expectedRes, r)
}

func TestGenerateAWSDBInstance(t *testing.T) {
	g := genAWSPostgreSQLGenerator()
	awsProvider := &inputs.Provider{}
	awsProviderURL, _ := inputs.GetProviderURL(g.ws.Runtimes.Terraform[inputs.AWSProvider])
	_ = awsProvider.SetString(awsProviderURL)
	awsProviderRegion, _ := inputs.GetProviderRegion(g.ws.Runtimes.Terraform[inputs.AWSProvider])

	awsSecurityGroupID := "hashicorp:aws:aws_security_group:testapp-db"
	randomPasswordID := "hashicorp:random:random_password:testapp-db"

	awsDBInstanceID, r := g.generateAWSDBInstance(awsProviderRegion, awsSecurityGroupID, randomPasswordID, awsProvider, g.postgres)
	expectedAWSDBInstanceID := "hashicorp:aws:aws_db_instance:testapp"
	expectedRes := apiv1.Resource{
		ID:   "hashicorp:aws:aws_db_instance:testapp",
		Type: "Terraform",
		Attributes: map[string]interface{}{
			"allocated_storage":   g.postgres.Size,
			"engine":              dbEngine,
			"engine_version":      g.postgres.Version,
			"identifier":          g.appName,
			"instance_class":      g.postgres.InstanceType,
			"password":            "$kusion_path.hashicorp:random:random_password:testapp-db.result",
			"publicly_accessible": true,
			"skip_final_snapshot": true,
			"username":            g.postgres.Username,
			"vpc_security_group_ids": []string{
				"$kusion_path.hashicorp:aws:aws_security_group:testapp-db.id",
			},
		},
		Extensions: map[string]interface{}{
			"provider": awsProviderURL,
			"providerMeta": map[string]interface{}{
				"region": awsProviderRegion,
			},
			"resourceType": awsDBInstance,
		},
	}

	assert.Equal(t, expectedAWSDBInstanceID, awsDBInstanceID)
	assert.Equal(t, expectedRes, r)
}

func genAWSPostgreSQLGenerator() *postgresGenerator {
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
		InstanceType:   "db.t3.micro",
		PrivateRouting: false,
		Username:       defaultUsername,
		Category:       defaultCategory,
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
				"aws": &apiv1.ProviderConfig{
					Source:  "hashicorp/aws",
					Version: "5.0.1",
					GenericConfig: apiv1.GenericConfig{
						"region": "us-east-1",
					},
				},
			},
		},
		Modules: apiv1.ModuleConfigs{
			"postgres": &apiv1.ModuleConfig{
				Default: apiv1.GenericConfig{
					"cloud":          "aws",
					"size":           20,
					"instanceType":   "db.t3.micro",
					"privateRouting": false,
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
