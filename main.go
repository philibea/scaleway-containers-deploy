package main

import (
	"fmt"
	"os"
	"strings"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	EnvType                 = "INPUT_TYPE"
	EnvAccessKey            = "INPUT_SCW_ACCESS_KEY"
	EnvContainerNamespaceID = "INPUT_SCW_CONTAINERS_NAMESPACE_ID"
	EnvContainerPort        = "INPUT_SCW_CONTAINER_PORT"
	EnvDNS                  = "INPUT_SCW_DNS"
	EnvDNSPrefix            = "INPUT_SCW_DNS_PREFIX"
	EnvPathRegistry         = "INPUT_SCW_REGISTRY"
	EnvProjectID            = "INPUT_SCW_PROJECT_ID"
	EnvSecretKey            = "INPUT_SCW_SECRET_KEY"
	EnvMemoryLimit          = "INPUT_SCW_MEMORY_LIMIT"
	EnvRootZone             = "INPUT_ROOT_ZONE"
	EnvEnvironmentVariables = "INPUT_SCW_ENVIRONMENT_VARIABLES"
	EnvSecrets              = "INPUT_SCW_SECRETS"
)

var (
	Description                 = "this container was created automatically by a github-action"
	Port           uint32       = 80
	MinScale       uint32       = 1
	MaxScale       uint32       = 5
	MaxConcurrency uint32       = 5
	MemoryLimit    uint32       = 256
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

	// Create a Scaleway client
	client, err := scw.NewClient(
		scw.WithAuth(ScalewayAccessKey, ScalewaySecretKey),
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

	const maxLength = 34

	var name string
	// rg.fr-par.scw.cloud/testing/images:latest

	// splitPath := strings.Split(PathRegistry, "/")
	// name = splitPath[2]
	// name = strings.ReplaceAll(name, ":", "")
	// name = strings.ReplaceAll(name, "-", "")

	// limitation of naming container with 20 characters
	splitPath := strings.Split(PathRegistry, ":")
	name = splitPath[1]

	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "_", "")

	if len(name) > maxLength {
		name = name[:maxLength]
	}

	return name
}

func DeployContainer(
	Client *scw.Client,
	Namespace *container.Namespace,
	ContainerName string,
	PathRegistry string,
	EnvironmentVariables map[string]string,
	Secrets []*container.Secret,
) (*container.Container, error) {

	fmt.Println("Container Name: ", ContainerName)

	ExistingContainer, _ := isContainerAlreadyCreated(Client, Namespace, ContainerName)

	if ExistingContainer != nil {

		// container already exists and need to be updated

		fmt.Println("Container already exists and will be updated", ExistingContainer)

		Container, err := UpdateDeployedContainer(Client, ExistingContainer, PathRegistry, EnvironmentVariables, Secrets)

		if err != nil {
			fmt.Println("unable to redeploy this serverless container : ", err)
			os.Exit(1)
			return Container, err
		}

		container, err := WaitForContainerReady(Client, Container)

		return container, err

	} else {
		Container, err := CreateContainerAndDeploy(Client, Namespace, PathRegistry, EnvironmentVariables, Secrets, ContainerName)

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

		Hostname, err := AddDNSRecord(Client, Container, DNSName)

		if err != nil {
			fmt.Println("unable to set DNS record: ", err)
		}

		dns, err := WaitForDNS(Client, DNSName)

		if err != nil {
			fmt.Println(dns)
			fmt.Println("unable to wait for DNS record: ", err)
		}

		ContainerDomain, err := SetCustomDomainContainer(Client, Container, Hostname)

		if err != nil {
			fmt.Println("unable to set x on Container: ", err)
			return nil, err
		}

		println("ContainerDomain", ContainerDomain.Hostname, ContainerDomain.Status)

		return ContainerDomain, nil
	}

	return nil, nil
}

func Deploy(
	Client *scw.Client,
	Region scw.Region,
	PathRegistry string,
	EnvironmentVariables map[string]string,
	Secrets []*container.Secret,
) (*container.Container, *container.Domain, error) {

	// Create or get a serverless container namespace
	namespaceContainer, err := GetContainersNamespace(Client, Region)

	WaitForNamespaceReady(Client, namespaceContainer)

	if err != nil {
		fmt.Println("unable to create or get a namespace serverless container : ", err)
		os.Exit(1)
		return nil, nil, err
	}

	ContainerName := GetContainerName(PathRegistry)

	Container, err := DeployContainer(Client, namespaceContainer, ContainerName, PathRegistry, EnvironmentVariables, Secrets)

	if err != nil {
		fmt.Println("unable to deploy a serverless container : ", err)
		os.Exit(1)
		return nil, nil, err
	}

	Domain, err := SetupDomain(Client, Container)

	if err != nil {
		fmt.Println("unable to setup dns : ", err)
	}

	return Container, Domain, nil

}

func Teardown(Client *scw.Client, Region scw.Region, PathRegistry string) (*container.Container, error) {

	ContainerName := GetContainerName(PathRegistry)
	Container, err := GetContainer(Client, Region, ContainerName)

	if err != nil {
		return nil, err
	}

	DNSName := os.Getenv(EnvDNS)

	if DNSName != "" {

		_, err := DeleteDNSRecord(Client, Container, DNSName)

		if err != nil {
			fmt.Println("unable to remove DNS record: ", err)
		}
	}

	ContainerDeleted, err := DeleteContainer(Client, Region, Container)

	if err != nil {
		return nil, err
	}

	fmt.Printf("Container %v deleted\n", ContainerDeleted.Name)

	return ContainerDeleted, nil

}

func getKeyValue(key string) map[string]string {
	KeyValue := make(map[string]string)
	EnvironmentKeyValues := strings.Split(os.Getenv(key), ",")

	for _, env := range EnvironmentKeyValues {
		splitEnv := strings.Split(env, "=")

		if len(splitEnv) == 2 {
			KeyValue[splitEnv[0]] = splitEnv[1]
		}
	}

	return KeyValue
}

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

func main() {
	PathRegistry := os.Getenv(EnvPathRegistry)
	Type := envOr(EnvType, "deploy")
	EnvironmentVariables := getKeyValue(EnvEnvironmentVariables)
	Secrets := getSecrets()

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

	if Type == "deploy" {
		Container, Domain, err := Deploy(Client, Region, PathRegistry, EnvironmentVariables, Secrets)

		if err != nil {
			fmt.Println("unable to deploy: ", err)
			os.Exit(1)
			return

		} else {
			PrintOutputGithubActionVariables(Container, Domain)
		}
	}

	if Type == "teardown" {
		deletedContainer, err := Teardown(Client, Region, PathRegistry)

		if err != nil {
			fmt.Println("unable to teardown container: ", err)
			os.Exit(1)
			return
		} else {
			PrintOutputGithubActionVariables(deletedContainer, nil)
		}
	}
}
