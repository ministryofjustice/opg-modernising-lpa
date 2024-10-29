package fixtures

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type ShareCodeSender interface {
	SendCertificateProviderInvite(context context.Context, appData appcontext.Data, invite sharecode.CertificateProviderInvite) error
	SendAttorneys(context context.Context, appData appcontext.Data, donorProvided *lpadata.Lpa) error
	UseTestCode(shareCode string)
}

const (
	testEmail  = "simulate-delivered@notifications.service.gov.uk"
	testMobile = "07700900000"
)

type fixturesData struct {
	App        appcontext.Data
	Sub        string
	DonorEmail string
	Errors     validation.List
	Members    []Name
}

type Name struct {
	Firstnames, Lastname string
}

func (n Name) Email() string {
	return strings.ToLower(fmt.Sprintf("%s-%s@example.org", n.Firstnames, n.Lastname))
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

func makeAttorney(name Name) donordata.Attorney {
	return donordata.Attorney{
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

func makeTrustCorporation(name string) donordata.TrustCorporation {
	return donordata.TrustCorporation{
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

func makeDonor(email string) donordata.Donor {
	return donordata.Donor{
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
		ThinksCanSign:             donordata.Yes,
		CanSign:                   form.Yes,
		ContactLanguagePreference: localize.En,
		LpaLanguagePreference:     localize.En,
	}
}

func makeCertificateProvider() donordata.CertificateProvider {
	return donordata.CertificateProvider{
		UID:                actoruid.New(),
		FirstNames:         "Charlie",
		LastName:           "Cooper",
		Email:              testEmail,
		Mobile:             testMobile,
		Relationship:       lpadata.Personally,
		RelationshipLength: donordata.GreaterThanEqualToTwoYears,
		CarryOutBy:         lpadata.ChannelOnline,
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

func makePersonToNotify(name Name) donordata.PersonToNotify {
	return donordata.PersonToNotify{
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

func makeCorrespondent(name Name) donordata.Correspondent {
	return donordata.Correspondent{
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
		Email:        testEmail,
		Phone:        testMobile,
	}
}

func makeVoucher(name Name) donordata.Voucher {
	return donordata.Voucher{
		FirstNames: name.Firstnames,
		LastName:   name.Lastname,
		Email:      fmt.Sprintf("%s.%s@example.org", name.Firstnames, name.Lastname),
		Allowed:    true,
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

func createAttorney(ctx context.Context, shareCodeStore ShareCodeStore, attorneyStore AttorneyStore, actorUID actoruid.UID, isReplacement, isTrustCorporation bool, lpaOwnerKey dynamo.LpaOwnerKeyType, email string) (*attorneydata.Provided, error) {
	shareCode := random.String(16)
	shareCodeData := sharecodedata.Link{
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

func createCertificateProvider(ctx context.Context, shareCodeStore ShareCodeStore, certificateProviderStore CertificateProviderStore, actorUID actoruid.UID, lpaOwnerKey dynamo.LpaOwnerKeyType, email string) (*certificateproviderdata.Provided, error) {
	shareCode := random.String(16)
	shareCodeData := sharecodedata.Link{
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

func createVoucher(ctx context.Context, shareCodeStore ShareCodeStore, voucherStore *voucher.Store, donor *donordata.Provided) (*voucherdata.Provided, error) {
	shareCode := random.String(16)
	shareCodeData := sharecodedata.Link{
		PK:          dynamo.ShareKey(dynamo.VoucherShareKey(shareCode)),
		SK:          dynamo.ShareSortKey(dynamo.MetadataKey(shareCode)),
		ActorUID:    donor.Voucher.UID,
		LpaOwnerKey: donor.SK,
		LpaKey:      donor.PK,
	}

	err := shareCodeStore.Put(ctx, actor.TypeVoucher, shareCode, shareCodeData)
	if err != nil {
		return nil, err
	}

	return voucherStore.Create(ctx, shareCodeData, donor.Voucher.Email)
}
