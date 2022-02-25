package main

import (
	"fmt"

	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// func ListDNS(
// 	client *scw.Client,
// 	DNSName string,
// ) {
// 	fmt.Println("waiting for dns to be ready")

// 	api := domain.NewAPI(client)

// 	ZonesDNS, err := api.ListDNSZones(&domain.ListDNSZonesRequest{
// 		DNSZone: DNSName,
// 	})

// }

func SetDNSRecord(
	client *scw.Client,
	DNSZone string,
	Container *container.Container,

) (*domain.UpdateDNSZoneRecordsResponse, error) {

	fmt.Println("Update Zone DNS")

	api := domain.NewAPI(client)

	Data := Container.DomainName + "."

	Records := []*domain.Record{
		{
			Name: Container.Name,
			Type: "CNAME",
			TTL:  60,
			Data: Data,
		},
	}

	Changes := []*domain.RecordChange{
		{
			Add: &domain.RecordChangeAdd{
				Records: Records,
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
