package main

import (
	"fmt"
	"os"
	"strconv"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func getSecrets() []*container.Secret {
	SecretsMap := getKeyValue(EnvSecrets)
	Secrets := make([]*container.Secret, 0)

	for key, value := range SecretsMap {
		Secrets = append(Secrets, &container.Secret{
			Key:   key,
			Value: &value,
		})
	}

	return Secrets
}

func GetSandboxVersion() container.ContainerSandbox {

	sandbox := envOr(EnvSandbox, Sandbox.String())

	if sandbox == "v1" {
		return container.ContainerSandboxV1
	}

	if sandbox == "v2" {
		return container.ContainerSandboxV2
	}

	return container.ContainerSandboxUnknownSandbox
}


func getContainerEnvVariables() container.Container {

	port, _ := strconv.ParseInt(envOr(EnvContainerPort, fmt.Sprint(Port)), 10, 32)
	memoryLimit, _ := strconv.ParseInt(envOr(EnvMemoryLimit, fmt.Sprint(MemoryLimit)), 10, 32)
	minScale, _ := strconv.ParseInt(envOr(EnvMinScale, fmt.Sprint(MinScale)), 10, 32)
	maxScale, _ := strconv.ParseInt(envOr(EnvMaxScale, fmt.Sprint(MaxScale)), 10, 32)
	maxConcurrency, _ := strconv.ParseInt(envOr(EnvMaxConcurrency, fmt.Sprint(MaxConcurrency)), 10, 32)
	cpuLimit, _ := strconv.ParseInt(envOr(EnvCPULimit, fmt.Sprint(CPULimit)), 10, 32)

	Env := container.Container{
		Port:           uint32(port),
		MemoryLimit:    uint32(memoryLimit),
		MinScale:       uint32(minScale),
		MaxScale:       uint32(maxScale),
		MaxConcurrency: uint32(maxConcurrency),
		CPULimit:       uint32(cpuLimit),
		Sandbox:        GetSandboxVersion(),
	}

	return Env

}

func WaitForNamespaceReady(client *scw.Client, NamespaceContainer *container.Namespace) (*container.Namespace, error) {
	fmt.Println("waiting for namespace to be ready")

	api := container.NewAPI(client)

	namespace, err := api.WaitForNamespace(&container.WaitForNamespaceRequest{
		Region:      NamespaceContainer.Region,
		NamespaceID: NamespaceContainer.ID,
	})

	if err != nil {
		return namespace, err
	}

	return namespace, nil

}

func WaitForContainerReady(
	client *scw.Client,
	Container *container.Container,
) (*container.Container, error) {

	fmt.Println("waiting for container to be ready")

	api := container.NewAPI(client)

	container, err := api.WaitForContainer(&container.WaitForContainerRequest{
		Region:      Container.Region,
		ContainerID: Container.ID,
	})

	if err != nil {
		return container, err
	}

	return container, nil
}

func GetContainer(
	client *scw.Client,
	Region scw.Region,
	ContainerName string,
) (*container.Container, error) {

	ContainersNamespaceId := os.Getenv(EnvContainerNamespaceID)

	api := container.NewAPI(client)

	if ContainersNamespaceId != "" {

		containersPointer, _ := api.ListContainers(&container.ListContainersRequest{
			Region:      Region,
			NamespaceID: ContainersNamespaceId,
			Name:        &ContainerName,
		})

		if len(containersPointer.Containers) == 0 {
			return nil, fmt.Errorf("container %s not found", ContainerName)
		}

		container := containersPointer.Containers[0]

		return container, nil

	}

	return nil, fmt.Errorf("namespace id not found")
}

func DeleteContainer(
	client *scw.Client,
	Region scw.Region,
	Container *container.Container,
) (*container.Container, error) {

	api := container.NewAPI(client)

	container, err := api.DeleteContainer(&container.DeleteContainerRequest{
		Region:      Region,
		ContainerID: Container.ID,
	})

	return container, err

}

func GetContainersNamespace(
	client *scw.Client,
	Region scw.Region,
) (*container.Namespace, error) {

	// OPTIONAL ENV VARIABLES
	ContainersNamespaceId := os.Getenv(EnvContainerNamespaceID)

	if ContainersNamespaceId == "" {

		return nil, fmt.Errorf("containers namespace id not found")
	}

	api := container.NewAPI(client)

	namespace, err := api.GetNamespace(&container.GetNamespaceRequest{
		Region:      Region,
		NamespaceID: ContainersNamespaceId,
	})

	if err != nil {
		fmt.Println("unable to get namespace: ", err)
	} else {
		return namespace, nil
	}

	return namespace, nil
}

func isContainerAlreadyCreated(
	client *scw.Client,
	Namespace *container.Namespace,
	ContainerName string,
) (*container.Container, error) {

	api := container.NewAPI(client)

	containers, err := api.ListContainers(&container.ListContainersRequest{
		Region:      Namespace.Region,
		NamespaceID: Namespace.ID,
		Name:        &ContainerName,
	})

	if err != nil {
		return nil, err
	}

	if len(containers.Containers) == 0 {
		return nil, nil
	}

	return containers.Containers[0], nil
}

func UpdateDeployedContainer(
	client *scw.Client,
	Container *container.Container,
	PathRegistry string,
) (*container.Container, error) {

	api := container.NewAPI(client)

	Redeploy := true

	containerEnv := getContainerEnvVariables()
	Secrets := getSecrets()
    EnvironmentVariables := getKeyValue(EnvEnvironmentVariables)

	updatedContainer, err := api.UpdateContainer(&container.UpdateContainerRequest{
		Region:                     Container.Region,
		ContainerID:                Container.ID,
		RegistryImage:              &PathRegistry,
		Redeploy:                   &Redeploy,
		EnvironmentVariables:       &EnvironmentVariables,
		SecretEnvironmentVariables: Secrets,
		MemoryLimit:                &containerEnv.MemoryLimit,
		MinScale:                   &containerEnv.MinScale,
		MaxScale:                   &containerEnv.MaxScale,
		CPULimit:                   &containerEnv.CPULimit,
		Port:                       &containerEnv.Port,
		MaxConcurrency:             &containerEnv.MaxConcurrency,
		Sandbox:                    containerEnv.Sandbox,
	})

	if err != nil {
		return nil, err
	}

	return updatedContainer, nil
}

func CreateContainerAndDeploy(
	client *scw.Client,
	NamespaceContainer *container.Namespace,
	PathRegistry string,
	ContainerName string,
) (*container.Container, error) {

	api := container.NewAPI(client)

	containerEnv := getContainerEnvVariables()
	Secrets := getSecrets()
	EnvironmentVariables := getKeyValue(EnvEnvironmentVariables)


	createdContainer, err := api.CreateContainer(&container.CreateContainerRequest{
		Description:                &Description,
		Name:                       ContainerName,
		NamespaceID:                NamespaceContainer.ID,
		Region:                     NamespaceContainer.Region,
		RegistryImage:              &PathRegistry,
		Timeout:                    &Timeout,
		EnvironmentVariables:       &EnvironmentVariables,
		SecretEnvironmentVariables: Secrets,
		MemoryLimit:                &containerEnv.MemoryLimit,
		MinScale:                   &containerEnv.MinScale,
		MaxScale:                   &containerEnv.MaxScale,
		CPULimit:                   &containerEnv.CPULimit,
		Port:                       &containerEnv.Port,
		MaxConcurrency:             &containerEnv.MaxConcurrency,
		Sandbox:                    containerEnv.Sandbox,
	})

	if err != nil {
		fmt.Println("unable to create container: ", err)
		return nil, err
	}

	deployedContainer, err := api.DeployContainer(&container.DeployContainerRequest{
		Region:      NamespaceContainer.Region,
		ContainerID: createdContainer.ID,
	})

	if err != nil {
		fmt.Println("unable to deploy container: ", err)
		return nil, err
	}

	return deployedContainer, nil
}

func SetCustomDomainContainer(
	client *scw.Client,
	Container *container.Container,
	Hostname string,
) (*container.Domain, error) {

	if Hostname == "" {
		return nil, fmt.Errorf("Hostname is required")
	}

	if len(Hostname) > 63 {
		return nil, fmt.Errorf("Hostname cannot be longer than 63 characters")
	}

	api := container.NewAPI(client)

	ResListDomains, _ := api.ListDomains(&container.ListDomainsRequest{
		Region:      Container.Region,
		ContainerID: Container.ID,
	})

	for _, domain := range ResListDomains.Domains {
		if domain.Hostname == Hostname {
			return domain, nil
		}
	}

	container, err := api.CreateDomain(
		&container.CreateDomainRequest{
			Region:      Container.Region,
			ContainerID: Container.ID,
			Hostname:    Hostname,
		},
	)

	if err != nil {
		return nil, err
	}

	return container, nil

}
