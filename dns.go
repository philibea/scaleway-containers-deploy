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
	CNAME = "CNAME"
	ALIAS = "ALIAS"
)

var (
	TTL = uint32(360)
)

func WaitForDNSReady(client *scw.Client, DNSZone string) (*domain.DNSZone, error) {
	fmt.Println("waiting for dns zone to be ready")

	api := domain.NewAPI(client)

	dns, err := api.WaitForDNSZone(&domain.WaitForDNSZoneRequest{
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

	// ENV
	Prefix := os.Getenv(EnvDNSPrefix)
	RootZone := os.Getenv(EnvRootZone)

	api := domain.NewAPI(client)

	Data := Container.DomainName + "."

	var Name string = Container.Name
	var Type domain.RecordType = CNAME

	// Handle Prefix DNS

	if Prefix != "" {
		Name = Prefix

		fmt.Println("Update With Prefix Zone DNS - Delete", Prefix)
	}

	// Handle Root Zone Alias
	// Some DNS doesn't handle correctly CNAME on Root Zone.
	// We should use an Alias

	if RootZone == "true" {
		Name = ""
		Type = ALIAS
		fmt.Println("Update Root Zone DNS - Delete")
	}

	IDFields := &domain.RecordIdentifier{
		Name: Name,
		Data: &Data,
		Type: Type,
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

func SetDNSRecord(
	client *scw.Client,
	Container *container.Container,
	DNSZone string,
) (string, error) {

	// ENV
	Prefix := os.Getenv(EnvDNSPrefix)
	RootZone := os.Getenv(EnvRootZone)

	fmt.Println("Update Zone DNS - Add")

	api := domain.NewAPI(client)

	var Name string = Container.Name
	var Type domain.RecordType = CNAME

	// Handle Prefix DNS

	if Prefix != "" {
		Name = Prefix

		fmt.Println("Update With Prefix Zone DNS - Add", Prefix)
	}

	var Hostname string = Name + "." + DNSZone

	// Handle Root Zone Alias
	// Some DNS doesn't handle correctly CNAME on Root Zone.
	// We should use an Alias

	if RootZone == "true" {
		Name = ""
		Type = ALIAS
		Hostname = DNSZone
		fmt.Println("Update Root Zone DNS - Add")
	}

	Records := []*domain.Record{
		{
			Name: Name,
			Type: Type,
			TTL:  TTL,
			Data: Container.DomainName + ".",
		},
	}

	Data :=  Container.DomainName + "."

	IDFields := &domain.RecordIdentifier{
		Name: Name,
		Type: Type,
		TTL:  &TTL,
		Data: &Data,
	}

	Changes := []*domain.RecordChange{
		{
			Set: &domain.RecordChangeSet{
				IDFields: IDFields,
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

	fmt.Println("Hostname", Hostname)

	return Hostname,nil
}
