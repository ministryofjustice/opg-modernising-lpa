package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type ShareCodeSender interface {
	SendCertificateProviderInvite(context context.Context, appData page.AppData, donorProvided page.CertificateProviderInvite) error
	SendAttorneys(context context.Context, appData page.AppData, donorProvided *lpastore.Lpa) error
	UseTestCode(shareCode string)
}

const (
	testEmail  = "simulate-delivered@notifications.service.gov.uk"
	testMobile = "07700900000"
)

type fixturesData struct {
	App        page.AppData
	Sub        string
	DonorEmail string
	Errors     validation.List
}

type Name struct {
	Firstnames, Lastname string
}

var (
	attorneyNames = []Name{
		{Firstnames: "Jessie", Lastname: "Jones"},
		{Firstnames: "Robin", Lastname: "Redcar"},
		{Firstnames: "Leslie", Lastname: "Lewis"},
		{Firstnames: "Ashley", Lastname: "Alwinton"},
		{Firstnames: "Frankie", Lastname: "Fernandes"},
	}
	replacementAttorneyNames = []Name{
		{Firstnames: "Blake", Lastname: "Buckley"},
		{Firstnames: "Taylor", Lastname: "Thompson"},
		{Firstnames: "Marley", Lastname: "Morris"},
		{Firstnames: "Alex", Lastname: "Abbott"},
		{Firstnames: "Billie", Lastname: "Blair"},
	}
	peopleToNotifyNames = []Name{
		{Firstnames: "Jordan", Lastname: "Jefferson"},
		{Firstnames: "Danni", Lastname: "Davies"},
		{Firstnames: "Bobbie", Lastname: "Bones"},
		{Firstnames: "Ally", Lastname: "Avery"},
		{Firstnames: "Deva", Lastname: "Dankar"},
	}
	invitedOrgMemberNames = []Name{
		{Firstnames: "Kamal", Lastname: "Singh"},
		{Firstnames: "Jo", Lastname: "Alessi"},
		{Firstnames: "Dan", Lastname: "Beaumont"},
		{Firstnames: "Nadia", Lastname: "Ksaiba"},
		{Firstnames: "Harry", Lastname: "Agius"},
	}
	orgMemberNames = []Name{
		{Firstnames: "Alice", Lastname: "Moxom"},
		{Firstnames: "Leon", Lastname: "Vynehall"},
		{Firstnames: "Derrick", Lastname: "Carter"},
		{Firstnames: "Luke", Lastname: "Solomon"},
		{Firstnames: "Josey", Lastname: "Rebelle"},
	}
	voucherName = Name{
		Firstnames: "Simone",
		Lastname:   "Sutherland",
	}
)

func makeAttorney(name Name) actor.Attorney {
	return actor.Attorney{
		UID:         actoruid.New(),
		FirstNames:  name.Firstnames,
		LastName:    name.Lastname,
		Email:       testEmail,
		DateOfBirth: date.New("2000", "1", "2"),
		Address: place.Address{
			Line1:      "2 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
			Country:    "GB",
		},
	}
}

func makeTrustCorporation(name string) actor.TrustCorporation {
	return actor.TrustCorporation{
		UID:           actoruid.New(),
		Name:          name,
		CompanyNumber: "555555555",
		Email:         testEmail,
		Address: place.Address{
			Line1:      "2 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
			Country:    "GB",
		},
	}
}

func makeDonor(email string) actor.Donor {
	return actor.Donor{
		UID:        actoruid.New(),
		FirstNames: "Sam",
		LastName:   "Smith",
		Address: place.Address{
			Line1:      "1 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
			Country:    "GB",
		},
		Email:                     email,
		DateOfBirth:               date.New("2000", "1", "2"),
		ThinksCanSign:             actor.Yes,
		CanSign:                   form.Yes,
		ContactLanguagePreference: localize.En,
		LpaLanguagePreference:     localize.En,
	}
}

func makeCertificateProvider() actor.CertificateProvider {
	return actor.CertificateProvider{
		UID:                actoruid.New(),
		FirstNames:         "Charlie",
		LastName:           "Cooper",
		Email:              testEmail,
		Mobile:             testMobile,
		Relationship:       actor.Personally,
		RelationshipLength: actor.GreaterThanEqualToTwoYears,
		CarryOutBy:         actor.ChannelOnline,
		Address: place.Address{
			Line1:      "5 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
			Country:    "GB",
		},
	}
}

func makePersonToNotify(name Name) actor.PersonToNotify {
	return actor.PersonToNotify{
		UID:        actoruid.New(),
		FirstNames: name.Firstnames,
		LastName:   name.Lastname,
		Address: place.Address{
			Line1:      "4 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
			Country:    "GB",
		},
	}
}

func makeCorrespondent(name Name) actor.Correspondent {
	return actor.Correspondent{
		FirstNames: name.Firstnames,
		LastName:   name.Lastname,
		Address: place.Address{
			Line1:      "5 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
			Country:    "GB",
		},
		Organisation: "Ashfurlong and partners",
		WantAddress:  form.Yes,
		Share:        actor.CorrespondentShareAttorneys | actor.CorrespondentShareCertificateProvider,
		Email:        testEmail,
		Telephone:    testMobile,
	}
}

func makeUID() string {
	return strings.ToUpper("M-" + "FAKE" + "-" + random.String(4) + "-" + random.String(4))
}

func acceptCookiesConsent(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "cookies-consent",
		Value:  "accept",
		MaxAge: 365 * 24 * 60 * 60,
		Path:   "/",
	})
}

func createAttorney(ctx context.Context, shareCodeStore ShareCodeStore, attorneyStore AttorneyStore, actorUID actoruid.UID, isReplacement, isTrustCorporation bool, lpaOwnerKey dynamo.LpaOwnerKeyType, email string) (*actor.AttorneyProvidedDetails, error) {
	shareCode := random.String(16)
	shareCodeData := actor.ShareCodeData{
		PK:                    dynamo.ShareKey(dynamo.AttorneyShareKey(shareCode)),
		SK:                    dynamo.ShareSortKey(dynamo.MetadataKey(shareCode)),
		ActorUID:              actorUID,
		IsReplacementAttorney: isReplacement,
		IsTrustCorporation:    isTrustCorporation,
		LpaOwnerKey:           lpaOwnerKey,
	}

	attorneyType := actor.TypeAttorney
	if isReplacement {
		attorneyType = actor.TypeReplacementAttorney
	}

	err := shareCodeStore.Put(ctx, attorneyType, shareCode, shareCodeData)
	if err != nil {
		return nil, err
	}

	return attorneyStore.Create(ctx, shareCodeData, email)
}

func createCertificateProvider(ctx context.Context, shareCodeStore ShareCodeStore, certificateProviderStore CertificateProviderStore, actorUID actoruid.UID, lpaOwnerKey dynamo.LpaOwnerKeyType, email string) (*actor.CertificateProviderProvidedDetails, error) {
	shareCode := random.String(16)
	shareCodeData := actor.ShareCodeData{
		PK:          dynamo.ShareKey(dynamo.CertificateProviderShareKey(shareCode)),
		SK:          dynamo.ShareSortKey(dynamo.MetadataKey(shareCode)),
		ActorUID:    actorUID,
		LpaOwnerKey: lpaOwnerKey,
	}

	err := shareCodeStore.Put(ctx, actor.TypeCertificateProvider, shareCode, shareCodeData)
	if err != nil {
		return nil, err
	}

	return certificateProviderStore.Create(ctx, shareCodeData, email)
}

func makeVoucher(name Name) actor.Voucher {
	return actor.Voucher{
		FirstNames: name.Firstnames,
		LastName:   name.Lastname,
		Email:      fmt.Sprintf("%s.%s@example.org", name.Firstnames, name.Lastname),
		Allowed:    true,
	}
}
