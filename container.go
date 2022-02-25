package main

import (
	"fmt"
	"os"
	"strconv"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

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

func GetOrCreateContainersNamespace(
	client *scw.Client,
	Region scw.Region,
) (*container.Namespace, error) {

	// OPTIONAL ENV VARIABLES
	ContainersNamespaceId := os.Getenv(EnvContainerNamespaceID)

	api := container.NewAPI(client)

	if ContainersNamespaceId != "" {

		namespace, err := api.GetNamespace(&container.GetNamespaceRequest{
			Region:      Region,
			NamespaceID: ContainersNamespaceId,
		})

		if err != nil {
			fmt.Println("unable to get namespace: ", err)
		} else {
			return namespace, nil
		}
	}

	Description := "Namespace created by a github action ( philibea/scaleway-action-container )"

	createdNamespace, err := api.CreateNamespace(&container.CreateNamespaceRequest{
		Description: &Description,
		Region:      Region,
	})

	if err != nil {
		fmt.Println("unable to create namespace: ", err)
		return nil, err
	}

	return createdNamespace, nil
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

	port, _ := strconv.ParseInt(envOr(EnvContainerPort, "80"), 10, 32)

	Port := uint32(port)

	updatedContainer, err := api.UpdateContainer(&container.UpdateContainerRequest{
		Region:        Container.Region,
		ContainerID:   Container.ID,
		RegistryImage: &PathRegistry,
		Redeploy:      &Redeploy,
		Port:          &Port,
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

	port, _ := strconv.ParseInt(envOr(EnvContainerPort, "80"), 10, 32)

	Port := uint32(port)

	createdContainer, err := api.CreateContainer(&container.CreateContainerRequest{
		Description:    &Description,
		MaxConcurrency: &MaxConcurrency,
		MaxScale:       &MaxScale,
		MemoryLimit:    &MemoryLimit,
		MinScale:       &MinScale,
		Name:           ContainerName,
		NamespaceID:    NamespaceContainer.ID,
		Port:           &Port,
		Region:         NamespaceContainer.Region,
		RegistryImage:  &PathRegistry,
		Timeout:        &Timeout,
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
