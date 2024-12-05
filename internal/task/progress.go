package task

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

type Localizer interface {
	Concat(list []string, joiner string) string
	Count(messageID string, count int) string
	Format(messageID string, data map[string]interface{}) string
	FormatCount(messageID string, count int, data map[string]any) string
	FormatDate(t date.TimeOrDate) string
	FormatTime(t time.Time) string
	FormatDateTime(t time.Time) string
	Possessive(s string) string
	SetShowTranslationKeys(s bool)
	ShowTranslationKeys() bool
	T(messageID string) string
}

type ProgressTracker struct {
	Localizer Localizer
}

type ProgressTask struct {
	Done  bool
	Label string
}

type Progress struct {
	Paid                      ProgressTask
	ConfirmedID               ProgressTask
	DonorSigned               ProgressTask
	CertificateProviderSigned ProgressTask
	AttorneysSigned           ProgressTask
	StatutoryWaitingPeriod    ProgressTask
	Registered                ProgressTask
}

func (p Progress) ToSlice() []ProgressTask {
	return []ProgressTask{
		p.Paid,
		p.ConfirmedID,
		p.DonorSigned,
		p.CertificateProviderSigned,
		p.AttorneysSigned,
		p.StatutoryWaitingPeriod,
		p.Registered,
	}
}

func (pt ProgressTracker) Progress(lpa *lpadata.Lpa) Progress {
	progress := Progress{
		Paid: ProgressTask{
			Done:  lpa.Paid,
			Label: pt.Localizer.T("lpaPaidFor"),
		},
		ConfirmedID: ProgressTask{
			Done: !lpa.Donor.IdentityCheck.CheckedAt.IsZero(),
		},
		DonorSigned: ProgressTask{
			Done: lpa.SignedForDonor(),
		},
		CertificateProviderSigned: ProgressTask{
			Done: !lpa.CertificateProvider.SignedAt.IsZero(),
		},
		AttorneysSigned: ProgressTask{
			Done:  lpa.AllAttorneysSigned(),
			Label: pt.Localizer.T("lpaSignedByAllAttorneys"),
		},
		StatutoryWaitingPeriod: ProgressTask{
			Done:  !lpa.StatutoryWaitingPeriodAt.IsZero(),
			Label: pt.Localizer.T("opgStatutoryWaitingPeriodBegins"),
		},
		Registered: ProgressTask{
			Done: !lpa.RegisteredAt.IsZero(),
		},
	}

	if lpa.IsOrganisationDonor {
		progress.ConfirmedID.Label = pt.Localizer.Format("donorsIdentityConfirmed",
			map[string]any{"DonorFullName": lpa.Donor.FullName()})
		progress.DonorSigned.Label = pt.Localizer.Format("lpaSignedByDonor",
			map[string]any{"DonorFullName": lpa.Donor.FullName()})
		if lpa.CertificateProvider.FirstNames == "" {
			progress.CertificateProviderSigned.Label = pt.Localizer.T("lpaCertificateProvided")
		} else {
			progress.CertificateProviderSigned.Label = pt.Localizer.Format("lpaCertificateProvidedBy",
				map[string]any{"CertificateProviderFullName": lpa.CertificateProvider.FullName()})
		}
		progress.Registered.Label = pt.Localizer.Format("donorsLpaRegisteredByOpg",
			map[string]any{"DonorFullName": lpa.Donor.FullName()})
	} else {
		progress.ConfirmedID.Label = pt.Localizer.T("yourIdentityConfirmed")
		progress.DonorSigned.Label = pt.Localizer.T("lpaSignedByYou")
		progress.CertificateProviderSigned.Label = pt.Localizer.T("lpaCertificateProvided")
		progress.Registered.Label = pt.Localizer.T("lpaRegisteredByOpg")
	}

	return progress
}
