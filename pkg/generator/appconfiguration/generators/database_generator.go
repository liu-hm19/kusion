package generators

import (
	"fmt"
	"os"
	"strings"

	"kusionstack.io/kusion/pkg/generator/appconfiguration/provider"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/accessories/database"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	randomPassword   = "random_password"
	awsSecurityGroup = "aws_security_group"
	awsDBInstance    = "aws_db_instance"
	dbUsernameKey    = "DATABASE_USERNAME"
	dbPasswordKey    = "DATABASE_PASSWORD"
	dbHostKey        = "DATABASE_HOST_ADDRESS"
)

type databaseGenerator struct {
	projectName string
	stackName   string
	comp        *component.Component
}

type awsSecurityGroupTraffic struct {
	CidrBlocks []string `yaml:"cidr_blocks" json:"cidr_blocks"`
	Protocol   string   `yaml:"protocol" json:"protocol"`
	FromPort   int      `yaml:"from_port" json:"from_port"`
	ToPort     int      `yaml:"to_port" json:"to_port"`
}

var (
	tfProviderAWS     = os.Getenv("TF_PROVIDER_AWS")
	tfProviderRandom  = os.Getenv("TF_PROVIDER_RANDOM")
	awsProviderRegion = os.Getenv("AWS_PROVIDER_REGION")
)

func NewDatabaseGenerator(projectName, stackName string, comp *component.Component) (Generator, error) {
	if len(projectName) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &databaseGenerator{
		projectName: projectName,
		stackName:   stackName,
		comp:        comp,
	}, nil
}

func NewDatabaseGeneratorFunc(projectName, stackName string, comp *component.Component) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewDatabaseGenerator(projectName, stackName, comp)
	}
}

func (g *databaseGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	db := g.comp.Database

	switch strings.ToLower(db.Type) {
	case "aws":
		return g.generateAWSResources(&db, spec)
	case "alicloud":
		return g.generateAlicloudResources(&db, spec)
	default:
		return fmt.Errorf("unsupported database type: %s", db.Type)
	}
}

func (g *databaseGenerator) generateAWSResources(db *database.Database, spec *models.Spec) error {
	// set the aws and random provider.
	randomProvider := &provider.Provider{}
	if err := randomProvider.SetString(tfProviderRandom); err != nil {
		return err
	}

	awsProvider := &provider.Provider{}
	if err := awsProvider.SetString(tfProviderAWS); err != nil {
		return err
	}

	// build random_password for aws_db_instance.
	randomPasswordID, r, err := generateTFRandomPassword(g.projectName, g.stackName, randomProvider)
	if err != nil {
		return err
	}
	spec.Resources = append(spec.Resources, r)

	// build aws_security group for aws_db_instance.
	awsSecurityGroupID, r, err := generateAWSSecurityGroup(g.projectName, g.stackName, awsProviderRegion, awsProvider, db)
	if err != nil {
		return err
	}
	spec.Resources = append(spec.Resources, r)

	// build aws_db_instance.
	awsDBInstanceID, r, err := generateAWSDBInstance(g.projectName, g.stackName, awsProviderRegion, awsSecurityGroupID, randomPasswordID, awsProvider, db)
	if err != nil {
		return err
	}
	spec.Resources = append(spec.Resources, r)

	// inject the database username, password and host address into k8s secret.
	password := dependencyWithKusionPath(randomPasswordID, "result")
	hostAddress := dependencyWithKusionPath(awsDBInstanceID, "address")

	return generateDBSecret(g.projectName, g.stackName, db.Username, password, hostAddress, spec)
}

func (g *databaseGenerator) generateAlicloudResources(db *database.Database, spec *models.Spec) error {
	// TODO: implement generator logic for alicloud rds instance.
	panic("not implemented yet")
}

func generateAWSSecurityGroup(projectName, stackName, providerRegion string,
	pvd *provider.Provider, db *database.Database,
) (string, models.Resource, error) {
	sgAttrs := map[string]interface{}{
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
				CidrBlocks: db.SecurityIPs,
				Protocol:   "tcp",
				FromPort:   3306,
				ToPort:     3306,
			},
		},
	}

	id, err := terraformResourceID(pvd, awsSecurityGroup, projectName+stackName)
	if err != nil {
		return "", models.Resource{}, err
	}

	return id, terraformResource(id, nil, sgAttrs, providerExtensions(pvd, provider.ProviderMeta{
		Region: providerRegion,
	}, awsSecurityGroup)), nil
}

func generateTFRandomPassword(projectName, stackName string,
	pvd *provider.Provider,
) (string, models.Resource, error) {
	pswAttrs := map[string]interface{}{
		"length":           16,
		"special":          true,
		"override_special": "!#$%&*()-_=+[]{}<>:?",
	}

	id, err := terraformResourceID(pvd, randomPassword, projectName+stackName+"-db")
	if err != nil {
		return "", models.Resource{}, err
	}

	return id, terraformResource(id, nil, pswAttrs, providerExtensions(pvd, provider.ProviderMeta{}, randomPassword)), nil
}

func generateAWSDBInstance(projectName, stackName, providerRegion, awsSecurityGroupID, randomPasswordID string,
	pvd *provider.Provider, db *database.Database,
) (string, models.Resource, error) {
	dbAttrs := map[string]interface{}{
		"allocated_storage":   db.Size,
		"engine":              db.Engine,
		"engine_version":      db.Version,
		"identifier":          projectName + stackName,
		"instance_class":      db.InstanceType,
		"password":            dependencyWithKusionPath(randomPasswordID, "result"),
		"publicly_accessible": db.AccessInternet,
		"skip_final_snapshot": true,
		"username":            db.Username,
		"vpc_security_groups_ids": []string{
			dependencyWithKusionPath(awsSecurityGroupID, "id"),
		},
	}

	id, err := terraformResourceID(pvd, awsDBInstance, projectName+stackName)
	if err != nil {
		return "", models.Resource{}, err
	}

	return id, terraformResource(id, nil, dbAttrs, providerExtensions(pvd, provider.ProviderMeta{
		Region: providerRegion,
	}, awsDBInstance)), nil
}

func generateDBSecret(projectName, stackName, username, password, hostAddress string, spec *models.Spec) error {
	// create the data map of k8s secret which stores the database username, password
	// and host address.
	data := make(map[string]string)
	data[dbUsernameKey] = username
	data[dbPasswordKey] = password
	data[dbHostKey] = hostAddress

	// create the k8s secret and append to the spec.
	cm := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectName + stackName + "-db",
			Namespace: projectName,
		},
		StringData: data,
	}

	return appendToSpec(
		kubernetesResourceID(cm.TypeMeta, cm.ObjectMeta),
		cm,
		spec,
	)
}
