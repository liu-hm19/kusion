package generators

import (
	"fmt"
	"net"
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
	randomPassword       = "random_password"
	awsSecurityGroup     = "aws_security_group"
	awsDBInstance        = "aws_db_instance"
	alicloudDBInstance   = "alicloud_db_instance"
	alicloudDBConnection = "alicloud_db_connection"
	alicloudRDSAccount   = "alicloud_rds_account"
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

type alicloudServerlessConfig struct {
	AutoPause   bool `yaml:"auto_pause" json:"auto_pause"`
	MaxCapacity int  `yaml:"max_capacity" json:"max_capacity"`
	MinCapacity int  `yaml:"min_capacity" json:"min_capacity"`
	SwitchForce bool `yaml:"switch_force" json:"switch_force"`
}

var (
	tfProviderAWS          = os.Getenv("TF_PROVIDER_AWS")
	tfProviderAlicloud     = os.Getenv("TF_PROVIDER_ALICLOUD")
	tfProviderRandom       = os.Getenv("TF_PROVIDER_RANDOM")
	awsProviderRegion      = os.Getenv("AWS_PROVIDER_REGION")
	alicloudProviderRegion = os.Getenv("ALICLOUD_PROVIDER_REGION")
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

	// skip rendering for empty rds instance.
	db := g.comp.Database
	if db.Type == "" {
		return nil
	}

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
	// set the alicloud and random provider.
	randomProvider := &provider.Provider{}
	if err := randomProvider.SetString(tfProviderRandom); err != nil {
		return err
	}

	alicloudProvider := &provider.Provider{}
	if err := alicloudProvider.SetString(tfProviderAlicloud); err != nil {
		return err
	}

	// build alicloud_db_instance.
	alicloudDBInstanceID, r, err := generateAlicloudDBInstance(g.projectName, g.stackName, alicloudProviderRegion, alicloudProvider, db)
	if err != nil {
		return err
	}
	spec.Resources = append(spec.Resources, r)

	// build alicloud_db_connection for alicloud_db_instance.
	var alicloudDBConnectionID string
	if isPublicAccessible(db.SecurityIPs) {
		alicloudDBConnectionID, r, err = generateAlicloudDBConnection(g.projectName, g.stackName, alicloudDBInstanceID, alicloudProviderRegion, alicloudProvider, db)
		if err != nil {
			return nil
		}
		spec.Resources = append(spec.Resources, r)
	}

	// build random_password for alicloud_rds_account.
	randomPasswordID, r, err := generateTFRandomPassword(g.projectName, g.stackName, randomProvider)
	if err != nil {
		return err
	}
	spec.Resources = append(spec.Resources, r)

	// build alicloud_rds_account.
	r, err = generateAlicloudRDSAccount(g.projectName, g.stackName, db.Username, randomPasswordID, alicloudDBInstanceID, alicloudProviderRegion, alicloudProvider, db)
	if err != nil {
		return err
	}
	spec.Resources = append(spec.Resources, r)

	// inject the database username, password and host address into k8s secret.
	password := dependencyWithKusionPath(randomPasswordID, "result")
	hostAddress := dependencyWithKusionPath(alicloudDBInstanceID, "connection_string")
	if !db.PrivateRouting {
		hostAddress = dependencyWithKusionPath(alicloudDBConnectionID, "connection_string")
	}

	return generateDBSecret(g.projectName, g.stackName, db.Username, password, hostAddress, spec)
}

func generateAlicloudDBInstance(projectName, stackName, providerRegion string,
	pvd *provider.Provider, db *database.Database,
) (string, models.Resource, error) {
	dbAttrs := map[string]interface{}{
		"category":         db.Category,
		"engine":           db.Engine,
		"engine_version":   db.Version,
		"instance_storage": db.Size,
		"instance_type":    db.InstanceType,
		"security_ips":     db.SecurityIPs,
		"vswitch_id":       db.AlicloudVSwitchID,
	}

	// set serverless specific attributes.
	if strings.Contains(db.Category, "serverless") {
		dbAttrs["db_instance_storage_type"] = "cloud_essd"
		dbAttrs["instance_charge_type"] = "Serverless"

		serverlessConfig := alicloudServerlessConfig{
			MaxCapacity: 8,
			MinCapacity: 1,
		}
		if db.Engine == "SQLServer" {
			serverlessConfig.MinCapacity = 2
		} else if db.Engine == "MySQL" {
			serverlessConfig.AutoPause = false
			serverlessConfig.SwitchForce = false
		}
		dbAttrs["serverless_config"] = []alicloudServerlessConfig{
			serverlessConfig,
		}
	}

	id, err := terraformResourceID(pvd, alicloudDBInstance, projectName+stackName)
	if err != nil {
		return "", models.Resource{}, err
	}

	return id, terraformResource(id, nil, dbAttrs, providerExtensions(pvd, provider.ProviderMeta{
		Region: providerRegion,
	}, alicloudDBInstance)), nil
}

func generateAlicloudDBConnection(projectName, stackName, dbInstanceID, providerRegion string,
	pvd *provider.Provider, db *database.Database,
) (string, models.Resource, error) {
	dbConnectionAttrs := map[string]interface{}{
		"instance_id": dependencyWithKusionPath(dbInstanceID, "id"),
	}

	id, err := terraformResourceID(pvd, alicloudDBConnection, projectName+stackName)
	if err != nil {
		return "", models.Resource{}, err
	}

	return id, terraformResource(id, nil, dbConnectionAttrs, providerExtensions(pvd, provider.ProviderMeta{
		Region: providerRegion,
	}, alicloudDBConnection)), nil
}

func generateAlicloudRDSAccount(projectName, stackName, accountName, randomPwdID, dbInstanceID, providerRegion string,
	pvd *provider.Provider, db *database.Database,
) (models.Resource, error) {
	rdsAccountAttrs := map[string]interface{}{
		"account_name":     accountName,
		"account_password": dependencyWithKusionPath(randomPwdID, "result"),
		"account_type":     "Super",
		"db_instance_id":   dependencyWithKusionPath(dbInstanceID, "id"),
	}

	id, err := terraformResourceID(pvd, alicloudRDSAccount, projectName+stackName)
	if err != nil {
		return models.Resource{}, err
	}

	return terraformResource(id, nil, rdsAccountAttrs, providerExtensions(pvd, provider.ProviderMeta{
		Region: providerRegion,
	}, alicloudRDSAccount)), nil
}

func generateAWSSecurityGroup(projectName, stackName, providerRegion string,
	pvd *provider.Provider, db *database.Database,
) (string, models.Resource, error) {
	// securityIPs should be in the format of IP address or Classes Inter-Domain
	// Routing (CIDR) mode.
	for _, ip := range db.SecurityIPs {
		if !isIPAddress(ip) && !isCIDR(ip) {
			return "", models.Resource{}, fmt.Errorf("illegal security ip format")
		}
	}

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
		"publicly_accessible": isPublicAccessible(db.SecurityIPs),
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
	data["username"] = username
	data["password"] = password
	data["hostAddress"] = hostAddress

	// create the k8s secret and append to the spec.
	secret := &v1.Secret{
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
		kubernetesResourceID(secret.TypeMeta, secret.ObjectMeta),
		secret,
		spec,
	)
}

func isPublicAccessible(securityIPs []string) bool {
	var parsedIP net.IP
	for _, ip := range securityIPs {
		if isIPAddress(ip) {
			parsedIP = net.ParseIP(ip)
		} else if isCIDR(ip) {
			parsedIP, _, _ = net.ParseCIDR(ip)
		}

		if parsedIP != nil && !parsedIP.IsPrivate() {
			return true
		}
	}

	return false
}

func isIPAddress(ipStr string) bool {
	ip := net.ParseIP(ipStr)

	return ip != nil
}

func isCIDR(cidrStr string) bool {
	_, _, err := net.ParseCIDR(cidrStr)

	return err == nil
}
