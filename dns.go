package main

import (
	"fmt"
	"os"
	"time"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	waitForDNSDefaultTimeout = 15 * time.Minute
	defaultRetryInterval     = 5 * time.Second
)

const (
	TYPE = "CNAME"
)

var (
	TTL = uint32(360)
)

func WaitForDNSZone(
	client *scw.Client,
	DNSZone string,
) (*domain.DNSZone, error) {
	fmt.Println("waiting for container to be ready")

	api := container.NewAPI(client)

	dns, err := api.WaitForDNSZone(&domain.WaitForDNSRequest{
		DNSZone: DNSZone,
	})

	if err != nil {
		return dns, err
	}

	return dns, nil
}

func DeleteDNSRecord(
	client *scw.Client,
	Container *container.Container,
	DNSZone string,
) (*domain.UpdateDNSZoneRecordsResponse, error) {
	fmt.Println("Update Zone DNS - Delete")

	api := domain.NewAPI(client)

	Data := Container.DomainName + "."

	Prefix := os.Getenv(EnvDNSPrefix)

	var Name string = Container.Name

	if Prefix != "" {
		Name = Prefix
	}

	IDFields := &domain.RecordIdentifier{
		Name: Name,
		Data: &Data,
		Type: TYPE,
		TTL:  &TTL,
	}

	Changes := []*domain.RecordChange{
		{
			Delete: &domain.RecordChangeDelete{
				IDFields: IDFields,
			},
		},
	}

	records, err := api.UpdateDNSZoneRecords(
		&domain.UpdateDNSZoneRecordsRequest{
			DNSZone: DNSZone,
			Changes: Changes,
		})

	if err != nil {

		return nil, err
	}

	return records, nil
}

func AddDNSRecord(
	client *scw.Client,
	Container *container.Container,
	DNSZone string,
) (string, error) {

	fmt.Println("Update Zone DNS - Add")

	api := domain.NewAPI(client)

	Prefix := os.Getenv(EnvDNSPrefix)

	var Name string = Container.Name

	if Prefix != "" {
		Name = Prefix
	}

	Records := []*domain.Record{
		{
			Name: Name,
			Type: TYPE,
			TTL:  TTL,
			Data: Container.DomainName + ".",
		},
	}

	Changes := []*domain.RecordChange{
		{
			Add: &domain.RecordChangeAdd{
				Records: Records,
			},
		},
	}

	_, err := api.UpdateDNSZoneRecords(
		&domain.UpdateDNSZoneRecordsRequest{
			DNSZone: DNSZone,
			Changes: Changes,
		})

	if err != nil {
		return "", err
	}

	Hostname := Name + "." + DNSZone

	return Hostname, nil
}
