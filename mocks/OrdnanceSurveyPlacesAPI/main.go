package main

import (
	"log"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/env"
)

var resultsJsonString = `
{
	"header": {
		"uri": "https://api.os.uk/search/places/v1/postcode?postcode=B14 7ED",
		"query": "postcode=B14 7ED",
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
		"DPA" : {
		  "UPRN" : "100071428503",
		  "UDPRN" : "431891",
		  "ADDRESS" : "1, RICHMOND PLACE, BIRMINGHAM, B14 7ED",
		  "BUILDING_NUMBER" : "1",
		  "THOROUGHFARE_NAME" : "RICHMOND PLACE",
		  "POST_TOWN" : "BIRMINGHAM",
		  "POSTCODE" : "B14 7ED",
		  "RPC" : "1",
		  "X_COORDINATE" : 407823.0,
		  "Y_COORDINATE" : 281658.0,
		  "STATUS" : "APPROVED",
		  "LOGICAL_STATUS_CODE" : "1",
		  "CLASSIFICATION_CODE" : "RD04",
		  "CLASSIFICATION_CODE_DESCRIPTION" : "Terraced",
		  "LOCAL_CUSTODIAN_CODE" : 4605,
		  "LOCAL_CUSTODIAN_CODE_DESCRIPTION" : "BIRMINGHAM",
		  "COUNTRY_CODE" : "E",
		  "COUNTRY_CODE_DESCRIPTION" : "This record is within England",
		  "POSTAL_ADDRESS_CODE" : "D",
		  "POSTAL_ADDRESS_CODE_DESCRIPTION" : "A record which is linked to PAF",
		  "BLPU_STATE_CODE" : "2",
		  "BLPU_STATE_CODE_DESCRIPTION" : "In use",
		  "TOPOGRAPHY_LAYER_TOID" : "osgb1000020370592",
		  "LAST_UPDATE_DATE" : "10/02/2016",
		  "ENTRY_DATE" : "16/04/2001",
		  "BLPU_STATE_DATE" : "29/04/2013",
		  "LANGUAGE" : "EN",
		  "MATCH" : 1.0,
		  "MATCH_DESCRIPTION" : "EXACT",
		  "DELIVERY_POINT_SUFFIX" : "1A"
		}
	  }, {
		"DPA" : {
		  "UPRN" : "100071428504",
		  "UDPRN" : "431892",
		  "ADDRESS" : "2, RICHMOND PLACE, BIRMINGHAM, B14 7ED",
		  "BUILDING_NUMBER" : "2",
		  "THOROUGHFARE_NAME" : "RICHMOND PLACE",
		  "POST_TOWN" : "BIRMINGHAM",
		  "POSTCODE" : "B14 7ED",
		  "RPC" : "1",
		  "X_COORDINATE" : 407823.0,
		  "Y_COORDINATE" : 281663.0,
		  "STATUS" : "APPROVED",
		  "LOGICAL_STATUS_CODE" : "1",
		  "CLASSIFICATION_CODE" : "RD04",
		  "CLASSIFICATION_CODE_DESCRIPTION" : "Terraced",
		  "LOCAL_CUSTODIAN_CODE" : 4605,
		  "LOCAL_CUSTODIAN_CODE_DESCRIPTION" : "BIRMINGHAM",
		  "COUNTRY_CODE" : "E",
		  "COUNTRY_CODE_DESCRIPTION" : "This record is within England",
		  "POSTAL_ADDRESS_CODE" : "D",
		  "POSTAL_ADDRESS_CODE_DESCRIPTION" : "A record which is linked to PAF",
		  "BLPU_STATE_CODE" : "2",
		  "BLPU_STATE_CODE_DESCRIPTION" : "In use",
		  "TOPOGRAPHY_LAYER_TOID" : "osgb1000020370591",
		  "LAST_UPDATE_DATE" : "10/02/2016",
		  "ENTRY_DATE" : "16/04/2001",
		  "BLPU_STATE_DATE" : "29/04/2013",
		  "LANGUAGE" : "EN",
		  "MATCH" : 1.0,
		  "MATCH_DESCRIPTION" : "EXACT",
		  "DELIVERY_POINT_SUFFIX" : "1B"
		}
	  }, {
		"DPA" : {
		  "UPRN" : "100071428505",
		  "UDPRN" : "431893",
		  "ADDRESS" : "3, RICHMOND PLACE, BIRMINGHAM, B14 7ED",
		  "BUILDING_NUMBER" : "3",
		  "THOROUGHFARE_NAME" : "RICHMOND PLACE",
		  "POST_TOWN" : "BIRMINGHAM",
		  "POSTCODE" : "B14 7ED",
		  "RPC" : "1",
		  "X_COORDINATE" : 407822.0,
		  "Y_COORDINATE" : 281667.0,
		  "STATUS" : "APPROVED",
		  "LOGICAL_STATUS_CODE" : "1",
		  "CLASSIFICATION_CODE" : "RD04",
		  "CLASSIFICATION_CODE_DESCRIPTION" : "Terraced",
		  "LOCAL_CUSTODIAN_CODE" : 4605,
		  "LOCAL_CUSTODIAN_CODE_DESCRIPTION" : "BIRMINGHAM",
		  "COUNTRY_CODE" : "E",
		  "COUNTRY_CODE_DESCRIPTION" : "This record is within England",
		  "POSTAL_ADDRESS_CODE" : "D",
		  "POSTAL_ADDRESS_CODE_DESCRIPTION" : "A record which is linked to PAF",
		  "BLPU_STATE_CODE" : "2",
		  "BLPU_STATE_CODE_DESCRIPTION" : "In use",
		  "TOPOGRAPHY_LAYER_TOID" : "osgb1000020370590",
		  "LAST_UPDATE_DATE" : "10/02/2016",
		  "ENTRY_DATE" : "16/04/2001",
		  "BLPU_STATE_DATE" : "29/04/2013",
		  "LANGUAGE" : "EN",
		  "MATCH" : 1.0,
		  "MATCH_DESCRIPTION" : "EXACT",
		  "DELIVERY_POINT_SUFFIX" : "1D"
		}
	  }, {
		"DPA" : {
		  "UPRN" : "100071428506",
		  "UDPRN" : "431894",
		  "ADDRESS" : "4, RICHMOND PLACE, BIRMINGHAM, B14 7ED",
		  "BUILDING_NUMBER" : "4",
		  "THOROUGHFARE_NAME" : "RICHMOND PLACE",
		  "POST_TOWN" : "BIRMINGHAM",
		  "POSTCODE" : "B14 7ED",
		  "RPC" : "1",
		  "X_COORDINATE" : 407823.0,
		  "Y_COORDINATE" : 281674.0,
		  "STATUS" : "APPROVED",
		  "LOGICAL_STATUS_CODE" : "1",
		  "CLASSIFICATION_CODE" : "RD04",
		  "CLASSIFICATION_CODE_DESCRIPTION" : "Terraced",
		  "LOCAL_CUSTODIAN_CODE" : 4605,
		  "LOCAL_CUSTODIAN_CODE_DESCRIPTION" : "BIRMINGHAM",
		  "COUNTRY_CODE" : "E",
		  "COUNTRY_CODE_DESCRIPTION" : "This record is within England",
		  "POSTAL_ADDRESS_CODE" : "D",
		  "POSTAL_ADDRESS_CODE_DESCRIPTION" : "A record which is linked to PAF",
		  "BLPU_STATE_CODE" : "2",
		  "BLPU_STATE_CODE_DESCRIPTION" : "In use",
		  "TOPOGRAPHY_LAYER_TOID" : "osgb1000020370589",
		  "LAST_UPDATE_DATE" : "10/02/2016",
		  "ENTRY_DATE" : "16/04/2001",
		  "BLPU_STATE_DATE" : "29/04/2013",
		  "LANGUAGE" : "EN",
		  "MATCH" : 1.0,
		  "MATCH_DESCRIPTION" : "EXACT",
		  "DELIVERY_POINT_SUFFIX" : "1E"
		}
	  }, {
		"DPA" : {
		  "UPRN" : "100071428507",
		  "UDPRN" : "431895",
		  "ADDRESS" : "5, RICHMOND PLACE, BIRMINGHAM, B14 7ED",
		  "BUILDING_NUMBER" : "5",
		  "THOROUGHFARE_NAME" : "RICHMOND PLACE",
		  "POST_TOWN" : "BIRMINGHAM",
		  "POSTCODE" : "B14 7ED",
		  "RPC" : "1",
		  "X_COORDINATE" : 407823.0,
		  "Y_COORDINATE" : 281678.0,
		  "STATUS" : "APPROVED",
		  "LOGICAL_STATUS_CODE" : "1",
		  "CLASSIFICATION_CODE" : "RD04",
		  "CLASSIFICATION_CODE_DESCRIPTION" : "Terraced",
		  "LOCAL_CUSTODIAN_CODE" : 4605,
		  "LOCAL_CUSTODIAN_CODE_DESCRIPTION" : "BIRMINGHAM",
		  "COUNTRY_CODE" : "E",
		  "COUNTRY_CODE_DESCRIPTION" : "This record is within England",
		  "POSTAL_ADDRESS_CODE" : "D",
		  "POSTAL_ADDRESS_CODE_DESCRIPTION" : "A record which is linked to PAF",
		  "BLPU_STATE_CODE" : "2",
		  "BLPU_STATE_CODE_DESCRIPTION" : "In use",
		  "TOPOGRAPHY_LAYER_TOID" : "osgb1000020370588",
		  "LAST_UPDATE_DATE" : "10/02/2016",
		  "ENTRY_DATE" : "16/04/2001",
		  "BLPU_STATE_DATE" : "29/04/2013",
		  "LANGUAGE" : "EN",
		  "MATCH" : 1.0,
		  "MATCH_DESCRIPTION" : "EXACT",
		  "DELIVERY_POINT_SUFFIX" : "1F"
		}
  	} ]
}
`

func main() {
	port := env.Get("PORT", "8080")

	http.HandleFunc("/search/places/v1/postcode", func(w http.ResponseWriter, r *http.Request) {
		postcode := r.URL.Query().Get("postcode")
		log.Println("postcode searched:", postcode)

		w.Write([]byte(resultsJsonString))
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
