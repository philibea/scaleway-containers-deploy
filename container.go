package main

import (
	"fmt"
	"os"
	"strconv"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func WaitForNamespaceReady(client *scw.Client, NamespaceContainer *container.Namespace) (*container.Namespace, error) {
	api := container.NewAPI(client)

	namespace, err := api.WaitForNamespace(
		&container.WaitForNamespaceRequest{
			NamespaceID: NamespaceContainer.ID,
			Region:      NamespaceContainer.Region,
		},
	)

	return namespace, err

}
func WaitForContainerReady(client *scw.Client, Container *container.Container) (*container.Container, error) {

	api := container.NewAPI(client)

	container, err := api.WaitForContainer(
		&container.WaitForContainerRequest{
			ContainerID: Container.ID,
			Region:      Container.Region,
		},
	)

	return container, err

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

	Description := "Namespace created by a github action"

	// Will allow to set project ID inside Client Request
	api.ListNamespaces(&container.ListNamespacesRequest{
		Region: Region,
	})

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

// func UpdateContainer(
// 	client *scw.Client,
// 	NamespaceContainer *container.Namespace,
// 	Region scw.Region,
// 	PathRegistry string,
// 	ContainerName string,
// ) (*container.Container, error) {

// 	api := container.NewAPI(client)
// 	// api.WaitForNamespace()
// }

func CreateContainerAndDeploy(
	client *scw.Client,
	NamespaceContainer *container.Namespace,
	Region scw.Region,
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
		Region:         Region,
		RegistryImage:  &PathRegistry,
		Timeout:        &Timeout,
	})

	if err != nil {
		fmt.Println("unable to create container: ", err)
		return nil, err
	}

	deployedContainer, err := api.DeployContainer(&container.DeployContainerRequest{
		Region:      Region,
		ContainerID: createdContainer.ID,
	})

	if err != nil {
		fmt.Println("unable to deploy container: ", err)
		return nil, err
	}

	return deployedContainer, nil
}
