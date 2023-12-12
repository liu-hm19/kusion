package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

func TestGenerateLocalResources(t *testing.T) {
	g := genLocalMySQLGenerator()

	spec := &apiv1.Intent{}
	secret, err := g.generateLocalResources(g.mysql, spec)

	hostAddress := "testapp-db-local-service"
	username := g.mysql.Username
	password := g.generateLocalPassword(16)
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

func TestGenerateLocalSecret(t *testing.T) {
	g := genLocalMySQLGenerator()

	spec := &apiv1.Intent{}
	password, err := g.generateLocalSecret(spec)
	expectedPassword := g.generateLocalPassword(16)

	assert.NoError(t, err)
	assert.Equal(t, expectedPassword, password)
}

func TestGenerateLocalPVC(t *testing.T) {
	g := genLocalMySQLGenerator()

	spec := &apiv1.Intent{}
	err := g.generateLocalPVC(g.mysql, spec)

	assert.NoError(t, err)
}

func TestGenerateLocalDeployment(t *testing.T) {
	g := genLocalMySQLGenerator()

	spec := &apiv1.Intent{}
	err := g.generateLocalDeployment(g.mysql, spec)

	assert.NoError(t, err)
}

func TestGenerateLocalService(t *testing.T) {
	g := genLocalMySQLGenerator()

	spec := &apiv1.Intent{}
	svcName, err := g.generateLocalService(g.mysql, spec)
	expectedSvcName := "testapp-db-local-service"

	assert.NoError(t, err)
	assert.Equal(t, expectedSvcName, svcName)
}

func genLocalMySQLGenerator() *mysqlGenerator {
	project := &apiv1.Project{
		Name: "testproject",
	}
	stack := &apiv1.Stack{
		Name: "teststack",
	}
	appName := "testapp"
	workload := &workload.Workload{}
	mysql := &mysql.MySQL{
		Type:    "local",
		Version: "8.0",
	}
	ws := &apiv1.Workspace{
		Name: "testworkspace",
		Runtimes: &apiv1.RuntimeConfigs{
			Kubernetes: &apiv1.KubernetesConfig{
				KubeConfig: "/Users/username/testkubeconfig",
			},
		},
	}

	return &mysqlGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		mysql:    mysql,
		ws:       ws,
	}
}
