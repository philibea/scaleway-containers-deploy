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
	EnvRegion               = "INPUT_SCW_REGION"
	EnvPathRegistry         = "INPUT_SCW_REGISTRY"
	EnvProjectID            = "INPUT_SCW_PROJECT_ID"
	EnvSecretKey            = "INPUT_SCW_SECRET_KEY"
	EnvMemoryLimit          = "INPUT_SCW_MEMORY_LIMIT"
	EnvMinScale             = "INPUT_SCW_MIN_SCALE"
	EnvMaxScale             = "INPUT_SCW_MAX_SCALE"
	EnvMaxConcurrency       = "INPUT_SCW_MAX_CONCURRENCY"
	EnvCPULimit             = "INPUT_SCW_CPU_LIMIT"
	EnvSandbox              = "INPUT_SCW_SANDBOX"
	EnvRootZone             = "INPUT_ROOT_ZONE"
	EnvEnvironmentVariables = "INPUT_SCW_ENVIRONMENT_VARIABLES"
	EnvSecrets              = "INPUT_SCW_SECRETS"
)

var (
	Description                               = "this container was created automatically by a github-action"
	Port           uint32                     = 80
	MinScale       uint32                     = 1
	MaxScale       uint32                     = 5
	MaxConcurrency uint32                     = 5
	MemoryLimit    uint32                     = 256
	CPULimit       uint32                     = 70
	Sandbox        container.ContainerSandbox = container.ContainerSandboxV1
	Timeout        scw.Duration               = scw.Duration{
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

func GetContainerName(PathRegistry string) string {
	const maxLength = 34

	// Split the path by "/"
	pathSegments := strings.Split(PathRegistry, "/")
	if len(pathSegments) < 2 {
		// If the path doesn't have the expected format, return an empty string
		return ""
	}

	// The last part before the colon contains the container name
	lastSegment := pathSegments[len(pathSegments)-1]

	// Split by ":" to remove the tag (e.g., "latest")
	nameParts := strings.Split(lastSegment, ":")
	name := nameParts[0]

	// Remove unwanted characters
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "_", "")

	// Truncate if it exceeds the max length
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
) (*container.Container, error) {

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

		Hostname, err := SetDNSRecord(Client, Container, DNSName)

		if err != nil {
			fmt.Println("unable to set DNS record: ", err)
		}

		dns, err := WaitForDNSReady(Client, DNSName)

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
) (*container.Container, *container.Domain, error) {

	// Create or get a serverless container namespace
	namespaceContainer, err := GetContainersNamespace(Client, Region)

	if err != nil {
		fmt.Println("unable to get a namespace serverless container : ", err)
		os.Exit(1)
		return nil, nil, err
	}

	WaitForNamespaceReady(Client, namespaceContainer)


	ContainerName := GetContainerName(PathRegistry)

	Container, err := DeployContainer(Client, namespaceContainer, ContainerName, PathRegistry)

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

func main() {
	PathRegistry := os.Getenv(EnvPathRegistry)
	MaybeRegion := envOr(EnvRegion, "fr-par")
	Type := envOr(EnvType, "deploy")

	if PathRegistry == "" {
		fmt.Println("Env Registry is not set")
		os.Exit(1)
		return
	}

	Region, err := scw.ParseRegion(MaybeRegion)

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
		Container, Domain, err := Deploy(Client, Region, PathRegistry)

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
