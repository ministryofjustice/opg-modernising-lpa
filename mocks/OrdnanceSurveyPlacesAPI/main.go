package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/env"
)

var resultsJsonString = `
{
	"header": {
		"uri": "https://api.os.uk/search/places/v1/postcode?postcode=%[1]s",
		"query": "postcode=%[1]s",
		"offset": 0,
		"totalresults": 2,
		"format": "JSON",
		"dataset": "DPA",
		"lr": "EN,CY",
		"maxresults": 100,
		"epoch": "96",
		"output_srs": "EPSG:27700"
	},
	"results": [ {
		"DPA": {
			"UPRN": "100071390703",
			"UDPRN": "432175",
			"ADDRESS": "123, FAKE STREET, FAKETON, SOMEVILLE, %[1]s",
			"BUILDING_NUMBER": "123",
			"THOROUGHFARE_NAME": "FAKE STREET",
			"DEPENDENT_LOCALITY": "FAKETON",
			"POST_TOWN": "SOMEVILLE",
			"POSTCODE": "%[1]s",
			"RPC": "1",
			"X_COORDINATE": 407783.0,
			"Y_COORDINATE": 281505.0,
			"STATUS": "APPROVED",
			"LOGICAL_STATUS_CODE": "1",
			"CLASSIFICATION_CODE": "RD04",
			"CLASSIFICATION_CODE_DESCRIPTION": "Terraced",
			"LOCAL_CUSTODIAN_CODE": 4605,
			"LOCAL_CUSTODIAN_CODE_DESCRIPTION": "SOMEVILLE",
			"COUNTRY_CODE": "E",
			"COUNTRY_CODE_DESCRIPTION": "This record is within England",
			"POSTAL_ADDRESS_CODE": "D",
			"POSTAL_ADDRESS_CODE_DESCRIPTION": "A record which is linked to PAF",
			"BLPU_STATE_CODE": "2",
			"BLPU_STATE_CODE_DESCRIPTION": "In use",
			"TOPOGRAPHY_LAYER_TOID": "osgb1000020369531",
			"LAST_UPDATE_DATE": "10/02/2016",
			"ENTRY_DATE": "16/04/2001",
			"BLPU_STATE_DATE": "29/04/2013",
			"LANGUAGE": "EN",
			"MATCH": 1.0,
			"MATCH_DESCRIPTION": "EXACT",
			"DELIVERY_POINT_SUFFIX": "1Q"
		}
	}, {
		"DPA": {
			"UPRN": "100071390703",
			"UDPRN": "432175",
			"ADDRESS": "456, FAKE STREET, FAKETON, SOMEVILLE, %[1]s",
			"BUILDING_NUMBER": "456",
			"THOROUGHFARE_NAME": "FAKE STREET",
			"DEPENDENT_LOCALITY": "FAKETON",
			"POST_TOWN": "SOMEVILLE",
			"POSTCODE": "%[1]s",
			"RPC": "1",
			"X_COORDINATE": 407783.0,
			"Y_COORDINATE": 281505.0,
			"STATUS": "APPROVED",
			"LOGICAL_STATUS_CODE": "1",
			"CLASSIFICATION_CODE": "RD04",
			"CLASSIFICATION_CODE_DESCRIPTION": "Terraced",
			"LOCAL_CUSTODIAN_CODE": 4605,
			"LOCAL_CUSTODIAN_CODE_DESCRIPTION": "SOMEVILLE",
			"COUNTRY_CODE": "E",
			"COUNTRY_CODE_DESCRIPTION": "This record is within England",
			"POSTAL_ADDRESS_CODE": "D",
			"POSTAL_ADDRESS_CODE_DESCRIPTION": "A record which is linked to PAF",
			"BLPU_STATE_CODE": "2",
			"BLPU_STATE_CODE_DESCRIPTION": "In use",
			"TOPOGRAPHY_LAYER_TOID": "osgb1000020369531",
			"LAST_UPDATE_DATE": "10/02/2016",
			"ENTRY_DATE": "16/04/2001",
			"BLPU_STATE_DATE": "29/04/2013",
			"LANGUAGE": "EN",
			"MATCH": 1.0,
			"MATCH_DESCRIPTION": "EXACT",
			"DELIVERY_POINT_SUFFIX": "1Q"
		}
	}]
}
`

func main() {
	port := env.Get("PORT", "8080")

	http.HandleFunc("/search/places/v1/postcode", func(w http.ResponseWriter, r *http.Request) {
		postcode := r.URL.Query().Get("postcode")
		log.Println("postcode searched:", postcode)
		w.Write([]byte(fmt.Sprintf(resultsJsonString, postcode)))
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
