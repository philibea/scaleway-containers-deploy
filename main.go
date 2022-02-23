package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	EnvType                 = "INPUT_TYPE" // deploy || teardown
	EnvAccessKey            = "INPUT_SCW_ACCESS_KEY"
	EnvProjectID            = "INPUT_SCW_PROJECT_ID"
	EnvContainerNamespaceID = "INPUT_SCW_CONTAINERS_NAMESPACE_ID"
	EnvContainerPort        = "INPUT_SCW_CONTAINER_PORT"
	EnvPathRegistry         = "INPUT_SCW_REGISTRY"
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
	// ScalewayProjectID := os.Getenv(EnvProjectID)

	// Create a Scaleway client
	client, err := scw.NewClient(
		scw.WithAuth(ScalewayAccessKey, ScalewaySecretKey),
		// scw.WithDefaultProjectID(ScalewayProjectID),
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

	var name string
	// rg.fr-par.scw.cloud/testing/images:latest

	splitPath := strings.Split(PathRegistry, "/")
	name = splitPath[2]
	name = strings.ReplaceAll(name, ":", "")
	name = strings.ReplaceAll(name, "-", "")

	return name
}

func Deploy() {
	fmt.Println("Deploy")

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

	//Create or get a serverless container namespace
	namespaceContainer, err := GetOrCreateContainersNamespace(Client, Region)

	WaitForNamespaceReady(Client, namespaceContainer)

	if err != nil {
		fmt.Println("unable to create or get a namespace serverless container : ", err)
		os.Exit(1)
		return
	}

	ContainerName := GetContainerName(PathRegistry)

	// deploy a container
	container, err := CreateContainerAndDeploy(Client, namespaceContainer, Region, PathRegistry, ContainerName)

	if err != nil {
		fmt.Println("unable to create or deploy a serverless container : ", err)
		os.Exit(1)
		return
	}

	readyContainer, _ := WaitForContainerReady(Client, container)

	fmt.Printf("::set-output name=container_url::%v\n", readyContainer.DomainName)
	fmt.Printf("::set-output name=url::https://%v\n", readyContainer.DomainName)
	fmt.Printf("::set-output name=scw_container_id::%v\n", readyContainer.ID)
	fmt.Printf("::set-output name=scw_namespace_id::%v\n", namespaceContainer.ID)

}

func Teardown() {

}

func main() {

	Type := os.Getenv(EnvType)

	if Type == "deploy" {
		Deploy()

	}
	if Type == "teardown" {
		Teardown()
	}

	// if DNS is set, need to set the DNS with the container endpoint in CNAME
	// Then we need to create endpoint custom Domain on containers

	// if ScalewayCustomeDNS == "" {
	// 	println("ScalewayCustomDNS")
	// }

}
