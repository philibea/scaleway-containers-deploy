package main

// CGO_ENABLED=0 go build -ldflags="-w -s" -v -o scw-container-deploy .
// SCW_REGION="fr-par" SCW_CONTAINERS_NAMESPACE_ID="ae28eaf1-3b94-4660-bce0-9b0e0a5d1062" SCW_SECRET_KEY="d49d3492-f500-4ab3-b7b1-48da542c310f" SCW_ACCESS_KEY="SCWR65FCAEYQTJAKVJK0"  ./scw-container-deploy

import (
	"fmt"
	"os"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	registry "github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/logger"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	EnvAccessKey            = "SCW_ACCESS_KEY"
	EnvSecretKey            = "SCW_SECRET_KEY"
	EnvRegion               = "SCW_REGION"
	EnvContainerNamespaceID = "SCW_CONTAINERS_NAMESPACE_ID"
	EnvRegistryNamespaceID  = "SCW_REGISTRY_NAMESPACE_ID"
	EnvRegistryImage        = "SCW_REGISTRY_IMAGE"
	EnvRegistryTag          = "SCW_REGISTRY_TAG"
	EnvContainerConfig      = "SCW_CONTAINER_CONFIG"
)

var (
	Description                 = "this container was created per a githuh-action"
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

func GetOrCreateContainersNamespace(
	client *scw.Client,
	NamespaceID string,
	Region scw.Region,
) (*container.Namespace, error) {

	api := container.NewAPI(client)

	namespace, err := api.GetNamespace(&container.GetNamespaceRequest{
		Region:      Region,
		NamespaceID: NamespaceID,
	})

	if err != nil {
		return nil, err
	}

	if namespace.ID != "" {
		return namespace, nil
	}

	createdNamespace, err := api.CreateNamespace(&container.CreateNamespaceRequest{
		Region: Region,
		Name:   "githuh-action",
	})

	if err != nil {
		logger.Errorf("Unable to create namespace: %s", err)

		return nil, err
	}

	return createdNamespace, nil
}

func FindRegistryNamespace(
	client *scw.Client,
	Region scw.Region,
	NamespaceId string,
) (*registry.Namespace, error) {
	api := registry.NewAPI(client)

	res, err := api.GetNamespace(&registry.GetNamespaceRequest{
		Region:      Region,
		NamespaceID: NamespaceId,
	})

	return res, err
}

func FindImageByName(
	client *scw.Client,
	Region scw.Region,
	NamespaceId string,
	ImageName string,
) (*registry.Image, error) {

	api := registry.NewAPI(client)

	res, err := api.ListImages(&registry.ListImagesRequest{
		Region:      Region,
		NamespaceID: &NamespaceId,
		Name:        &ImageName,
	})

	if res.Images != nil && len(res.Images) > 0 || err != nil {

		image := res.Images[0]

		return image, nil

	}

	return nil, fmt.Errorf("image not found")
}

func FindTagByName(
	client *scw.Client,
	Region scw.Region,
	ImageID string,
	TagName string,
) (*registry.Tag, error) {

	api := registry.NewAPI(client)

	res, err := api.ListTags(&registry.ListTagsRequest{
		Region:  Region,
		ImageID: ImageID,
		Name:    &TagName,
	})

	if res.Tags != nil && len(res.Tags) > 0 || err != nil {

		tag := res.Tags[0]

		return tag, nil

	}

	return nil, fmt.Errorf("tag not found")
}

func CreateContainer(
	client *scw.Client,
	NamespaceContainer *container.Namespace,
	NamespaceRegistry *registry.Namespace,
	Region scw.Region,
	RegistryImage *registry.Image,
	RegistryTag *registry.Tag,
	Port uint32,
) (*container.Container, error) {

	api := container.NewAPI(client)

	registryImage := "rg." + scw.Region.String(Region) + ".scw.cloud/" + NamespaceRegistry.Name + "/" + RegistryImage.Name + ":" + RegistryTag.Name

	fmt.Print(registryImage)

	container, err := api.CreateContainer(&container.CreateContainerRequest{
		Description:   &Description,
		MaxScale:      &MaxScale,
		MinScale:      &MinScale,
		Name:          "githuh-action",
		NamespaceID:   NamespaceContainer.ID,
		Region:        Region,
		RegistryImage: &registryImage,
		Timeout:       &Timeout,
		Port:          &Port,
	})

	if err != nil {
		logger.Errorf("Unable to create container: %s", err)
		return nil, err
	}

	return container, nil
}

func CreateClient() (*scw.Client, scw.Region, error) {

	// required to initialize the client
	ScalewayAccessKey := os.Getenv(EnvAccessKey)
	ScalewaySecretKey := os.Getenv(EnvSecretKey)

	// optional
	ScalewayRegion := os.Getenv(EnvRegion)

	// check if the region is valid
	Region, _ := scw.ParseRegion(ScalewayRegion)

	// Create a Scaleway client
	client, err := scw.NewClient(
		scw.WithAuth(ScalewayAccessKey, ScalewaySecretKey),
	)

	if err != nil {

		return client, Region, err
	}

	return client, Region, nil

}

func main() {

	// OPTIONAL ENV VARIABLES
	ContainersNamespaceId := os.Getenv(EnvContainerNamespaceID)
	RegistryNamespaceId := os.Getenv(EnvRegistryNamespaceID)
	ImageName := os.Getenv(EnvRegistryImage) // node
	ImageTag := os.Getenv(EnvRegistryTag)    // latest
	// ScalewayCustomeDNS := os.Getenv("SCW_CUSTOM_DNS")

	Client, Region, err := CreateClient()

	if err != nil {
		return
	}

	//Create or get a container namespace
	namespaceContainer, _ := GetOrCreateContainersNamespace(Client, ContainersNamespaceId, Region)

	// registry
	namespaceRegistry, _ := FindRegistryNamespace(Client, Region, RegistryNamespaceId)
	image, _ := FindImageByName(Client, Region, RegistryNamespaceId, ImageName)
	tag, _ := FindTagByName(Client, Region, image.ID, ImageTag)

	// deploy a container
	container, _ := CreateContainer(Client, namespaceContainer, namespaceRegistry, Region, image, tag, Port)

	fmt.Println("container", container)

	// if DNS is set, need to set the DNS with the container endpoint in CNAME
	// Then we need to create endpoint custom Domain on containers

	// if ScalewayCustomeDNS == "" {
	// 	println("ScalewayCustomeDNS")
	// }

}
