package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	// async "github.com/scaleway/scaleway-sdk-go/scw/internal/async"
)

const timeout = 10 * time.Minute
const retryInterval = 5 * time.Second

const (
	NamespaceStatusUnknown  = container.NamespaceStatus("unknown")
	NamespaceStatusReady    = container.NamespaceStatus("ready")
	NamespaceStatusDeleting = container.NamespaceStatus("deleting")
	NamespaceStatusError    = container.NamespaceStatus("error")
	NamespaceStatusLocked   = container.NamespaceStatus("locked")
	NamespaceStatusCreating = container.NamespaceStatus("creating")
	NamespaceStatusPending  = container.NamespaceStatus("pending")
)

func waitForNamespaceReady(client *scw.Client, NamespaceContainer *container.Namespace) (*container.Namespace, error) {
	fmt.Println("waiting for namespace to be ready")

	api := container.NewAPI(client)

	terminalStatus := map[container.NamespaceStatus]struct{}{
		container.NamespaceStatusReady:  {},
		container.NamespaceStatusLocked: {},
		container.NamespaceStatusError:  {},
	}

	namespace, err := WaitSync(&WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			namespace, err := api.GetNamespace(&container.GetNamespaceRequest{
				NamespaceID: NamespaceContainer.ID,
				Region:      NamespaceContainer.Region,
			})
			if err != nil {
				return nil, false, err
			}

			_, isTerminal := terminalStatus[namespace.Status]

			return namespace, isTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: LinearIntervalStrategy(retryInterval),
	})

	if err != nil {
		return nil, fmt.Errorf("unable to wait for namespace to be ready: %s", err)
	}

	return namespace.(*container.Namespace), nil

}

func WaitForContainerReady(
	client *scw.Client,
	Container *container.Container,
) (*container.Container, error) {

	fmt.Println("waiting for container to be ready")

	api := container.NewAPI(client)

	terminalStatus := map[container.ContainerStatus]struct{}{
		container.ContainerStatusReady:  {},
		container.ContainerStatusError:  {},
		container.ContainerStatusLocked: {},
	}

	_container, err := WaitSync(&WaitSyncConfig{
		Get: func() (interface{}, bool, error) {
			container, err := api.GetContainer(&container.GetContainerRequest{
				ContainerID: Container.ID,
				Region:      Container.Region,
			})
			if err != nil {
				return nil, false, err
			}

			_, isTerminal := terminalStatus[container.Status]

			return container, isTerminal, nil
		},
		Timeout:          timeout,
		IntervalStrategy: LinearIntervalStrategy(retryInterval),
	})

	if err != nil {
		return nil, fmt.Errorf("unable to wait for container to be ready: %s", err)
	}

	return _container.(*container.Container), nil
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
