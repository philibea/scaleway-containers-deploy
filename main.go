package main

import (
	"fmt"
	"os"
	"strings"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	EnvAccessKey            = "INPUT_SCW_ACCESS_KEY"
	EnvContainerNamespaceID = "INPUT_SCW_CONTAINERS_NAMESPACE_ID"
	EnvContainerPort        = "INPUT_SCW_CONTAINER_PORT"
	EnvDNS                  = "INPUT_SCW_DNS"
	EnvPathRegistry         = "INPUT_SCW_REGISTRY"
	EnvProjectID            = "INPUT_SCW_PROJECT_ID"
	EnvSecretKey            = "INPUT_SCW_SECRET_KEY"
)

var (
	Description                 = "this container was created automatically by a github-action"
	Port           uint32       = 80
	MinScale       uint32       = 1
	MaxScale       uint32       = 5
	MaxConcurrency uint32       = 5
	MemoryLimit    uint32       = 1024
	Timeout        scw.Duration = scw.Duration{
		Seconds: 60,
		Nanos:   0,
	}
)

func PrintOutputGithubActionVariables(Container *container.Container, Domain *container.Domain) {

	if Domain != nil {
		fmt.Printf("::set-output name=url::https://%v\n", Domain.Hostname)
		fmt.Printf("::set-output name=container_url::%v\n", Container.DomainName)
		fmt.Printf("::set-output name=scw_container_id::%v\n", Container.ID)
		fmt.Printf("::set-output name=scw_namespace_id::%v\n", Container.ID)
	} else {
		fmt.Printf("::set-output name=container_url::%v\n", Container.DomainName)
		fmt.Printf("::set-output name=url::https://%v\n", Container.DomainName)
		fmt.Printf("::set-output name=scw_container_id::%v\n", Container.ID)
		fmt.Printf("::set-output name=scw_namespace_id::%v\n", Container.ID)

	}

}

func envOr(name, def string) string {
	if d, ok := os.LookupEnv(name); ok {
		return d
	}
	return def
}

func CreateClient(Region scw.Region) (*scw.Client, error) {

	// required to initialize the client
	ScalewayAccessKey := os.Getenv(EnvAccessKey)
	ScalewaySecretKey := os.Getenv(EnvSecretKey)
	ScalewayProjectID := os.Getenv(EnvProjectID)

	// Create a Scaleway client
	client, err := scw.NewClient(
		scw.WithAuth(ScalewayAccessKey, ScalewaySecretKey),
		scw.WithDefaultProjectID(ScalewayProjectID),
	)

	if err != nil {

		return client, err
	}

	return client, nil

}

func GetRegionFromRegistryPath(PathRegistry string) (scw.Region, error) {
	region := strings.Split(PathRegistry, ".")[1]

	Region, err := scw.ParseRegion(region)

	if err != nil {
		return Region, err
	}

	return Region, nil
}

func GetContainerName(PathRegistry string) string {

	const maxLength = 20

	var name string
	// rg.fr-par.scw.cloud/testing/images:latest

	splitPath := strings.Split(PathRegistry, "/")
	name = splitPath[2]
	name = strings.ReplaceAll(name, ":", "")
	name = strings.ReplaceAll(name, "-", "")

	if len(name) > maxLength {
		name = name[:maxLength]
	}

	return name
}

func DeployContainer(Client *scw.Client, Namespace *container.Namespace, ContainerName string, PathRegistry string) (*container.Container, error) {

	fmt.Println("Container Name: ", ContainerName)

	ExistingContainer, _ := isContainerAlreadyCreated(Client, Namespace, ContainerName)

	if ExistingContainer != nil {

		// container already exists and need to be updated

		fmt.Println("Container already exists and will be updated", ExistingContainer)

		Container, err := UpdateDeployedContainer(Client, ExistingContainer, PathRegistry)

		if err != nil {
			fmt.Println("unable to redeploy this serverless container : ", err)
			os.Exit(1)
			return Container, err
		}

		container, err := WaitForContainerReady(Client, Container)

		return container, err

	} else {
		Container, err := CreateContainerAndDeploy(Client, Namespace, PathRegistry, ContainerName)

		if err != nil {
			fmt.Println("unable to create or deploy a serverless container : ", err)
			os.Exit(1)
			return Container, err
		}
		container, err := WaitForContainerReady(Client, Container)

		return container, err
	}
}

func SetupDomain(Client *scw.Client, Container *container.Container) (*container.Domain, error) {
	DNSName := os.Getenv(EnvDNS)

	if DNSName != "" {

		_, err := SetDNSRecord(Client, DNSName, Container)

		if err != nil {
			fmt.Println("unable to set DNS record: ", err)
		}

		Hostname := Container.Name + "." + DNSName

		ContainerDomain, err := SetCustomDomainContainer(Client, Container, Hostname)

		if err != nil {
			fmt.Println("unable to set DNS record: ", err)
			return nil, err
		}

		println("ContainerDomain", ContainerDomain.Hostname, ContainerDomain.Status)

		return ContainerDomain, nil
	}

	return nil, nil
}

func main() {
	PathRegistry := os.Getenv(EnvPathRegistry)

	if PathRegistry == "" {
		fmt.Println("Env Registry is not set")
		os.Exit(1)
		return
	}

	Region, err := GetRegionFromRegistryPath(PathRegistry)

	if err != nil {
		fmt.Println("Registry should respact format", err)
		os.Exit(1)
		return

	}

	Client, err := CreateClient(Region)

	if err != nil {
		fmt.Println("unable to create client: ", err)
		os.Exit(1)
		return
	}

	// Create or get a serverless container namespace
	namespaceContainer, err := GetOrCreateContainersNamespace(Client, Region)

	WaitForNamespaceReady(Client, namespaceContainer)

	if err != nil {
		fmt.Println("unable to create or get a namespace serverless container : ", err)
		os.Exit(1)
		return
	}

	ContainerName := GetContainerName(PathRegistry)

	Container, err := DeployContainer(Client, namespaceContainer, ContainerName, PathRegistry)

	if err != nil {
		fmt.Println("unable to deploy a serverless container : ", err)
		os.Exit(1)
		return
	}

	Domain, _ := SetupDomain(Client, Container)

	PrintOutputGithubActionVariables(Container, Domain)

}
