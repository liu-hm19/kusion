package generators

import (
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/generator/appconfiguration/provider"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component/container"
)

// kubernetesResourceID returns the unique ID of a Kubernetes resource
// based on its type and metadata.
func kubernetesResourceID(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:code-city:code-citydev
	id := typeMeta.APIVersion + ":" + typeMeta.Kind + ":"
	if objectMeta.Namespace != "" {
		id += objectMeta.Namespace + ":"
	}
	id += objectMeta.Name
	return id
}

// terraformResourceID returns the unique ID of a Terraform resource
// based on its provider, type and name information.
func terraformResourceID(provider *provider.Provider, resourceType, resourceName string) (string, error) {
	// resource id example: hashicorp:aws:aws_db_instance:wordpressdev
	if provider.Namespace == "" || provider.Name == "" {
		return "", fmt.Errorf("insufficient provider information: %s", provider.URL)
	}

	id := provider.Namespace + ":" + provider.Name + ":" + resourceType + ":" + resourceName

	return id, nil
}

// providerExtensions returns the extended information of provider
// based on the provider and type of the resource.
func providerExtensions(provider *provider.Provider, providerMeta provider.ProviderMeta, resourceType string) map[string]interface{} {
	return map[string]interface{}{
		"provider":     provider.URL,
		"providerMeta": providerMeta,
		"resourceType": resourceType,
	}
}

// callGeneratorFuncs calls each NewGeneratorFunc in the given slice
// and returns a slice of Generator instances.
func callGeneratorFuncs(newGenerators ...NewGeneratorFunc) ([]Generator, error) {
	gs := make([]Generator, 0, len(newGenerators))
	for _, newGenerator := range newGenerators {
		if g, err := newGenerator(); err != nil {
			return nil, err
		} else {
			gs = append(gs, g)
		}
	}
	return gs, nil
}

// callGenerators calls the Generate method of each Generator instance
// returned by the given NewGeneratorFuncs.
func callGenerators(spec *models.Spec, newGenerators ...NewGeneratorFunc) error {
	gs, err := callGeneratorFuncs(newGenerators...)
	if err != nil {
		return err
	}
	for _, g := range gs {
		if err := g.Generate(spec); err != nil {
			return err
		}
	}
	return nil
}

// appendToSpec adds a Kubernetes resource to a spec's resources
// slice.
func appendToSpec(resourceID string, resource any, spec *models.Spec) error {
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}
	r := models.Resource{
		ID:         resourceID,
		Type:       generator.Kubernetes,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: nil,
	}
	spec.Resources = append(spec.Resources, r)
	return nil
}

// uniqueComponentName returns a unique name for a component based on
// its project and name.
func uniqueComponentName(projectName, compName string) string {
	return projectName + "-" + compName
}

// uniqueComponentLabels returns a map of labels that identify a
// component based on its project and name.
func uniqueComponentLabels(projectName, compName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      projectName,
		"app.kubernetes.io/component": compName,
	}
}

// int32Ptr returns a pointer to an int32 value.
func int32Ptr(i int32) *int32 {
	return &i
}

// foreachOrderedContainers executes the given function on each
// container in the map in order of their keys.
func foreachOrderedContainers(
	m map[string]container.Container,
	f func(containerName string, c container.Container) error,
) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		if err := f(k, v); err != nil {
			return err
		}
	}

	return nil
}

func toOrderedContainers(appContainers map[string]container.Container) ([]corev1.Container, error) {
	// Create a slice of containers based on the component's
	// containers.
	containers := []corev1.Container{}
	if err := foreachOrderedContainers(appContainers, func(containerName string, c container.Container) error {
		// Create a slice of env vars based on the container's
		// envvars.
		envs := []corev1.EnvVar{}
		for k, v := range c.Env {
			envs = append(envs, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}

		// Create a container object and append it to the containers
		// slice.
		containers = append(containers, corev1.Container{
			Name:       containerName,
			Image:      c.Image,
			Command:    c.Command,
			Args:       c.Args,
			WorkingDir: c.WorkingDir,
			Env:        envs,
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return containers, nil
}

// foreachOrderedComponents executes the given function on each
// component in the map in order of their keys.
func foreachOrderedComponents(
	m map[string]component.Component,
	f func(compName string, comp component.Component) error,
) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		if err := f(k, v); err != nil {
			return err
		}
	}

	return nil
}

// dependencyWithKusionPath returns the implicit resource dependency path
// based on the resource id and name with the "$kusion_path" prefix.
func dependencyWithKusionPath(id, name string) string {
	return "$kusion_path." + id + "." + name
}

// terraformResource returns the Terraform resource in the form of models.Resource
func terraformResource(id string, dependsOn []string, attrs, exts map[string]interface{}) models.Resource {
	return models.Resource{
		ID:         id,
		Type:       generator.Terraform,
		Attributes: attrs,
		DependsOn:  dependsOn,
		Extensions: exts,
	}
}
