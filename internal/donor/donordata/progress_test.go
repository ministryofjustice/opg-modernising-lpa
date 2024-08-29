package donordata

//
//func TestProgressToSlice(t *testing.T) {
//	testcases := map[string]func(Progress) (Progress, []ProgressTask){
//		"donor": func(p Progress) (Progress, []ProgressTask) {
//			return p, []ProgressTask{
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"organisation": func(p Progress) (Progress, []ProgressTask) {
//			p.isOrganisation = true
//
//			return p, []ProgressTask{
//				p.Paid,
//				p.ConfirmedID,
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"donor notices sent": func(p Progress) (Progress, []ProgressTask) {
//			p.NoticesOfIntentSent.State = task.StateCompleted
//
//			return p, []ProgressTask{
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.NoticesOfIntentSent,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"organisation notices sent": func(p Progress) (Progress, []ProgressTask) {
//			p.isOrganisation = true
//			p.NoticesOfIntentSent.State = task.StateCompleted
//
//			return p, []ProgressTask{
//				p.Paid,
//				p.ConfirmedID,
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.NoticesOfIntentSent,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"donor with fee evidence pre signing": func(p Progress) (Progress, []ProgressTask) {
//			p.FeeEvidenceSubmitted.State = task.StateCompleted
//
//			return p, []ProgressTask{
//				p.FeeEvidenceSubmitted,
//				p.FeeEvidenceApproved,
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"donor with fee evidence pre signing communication sent": func(p Progress) (Progress, []ProgressTask) {
//			p.FeeEvidenceSubmitted.State = task.StateCompleted
//
//			p.FeeEvidenceNotification.State = task.StateCompleted
//			p.FeeEvidenceNotification.Completed = time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
//
//			return p, []ProgressTask{
//				p.FeeEvidenceSubmitted,
//				p.FeeEvidenceNotification,
//				p.FeeEvidenceApproved,
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"donor with fee evidence post signing communication sent": func(p Progress) (Progress, []ProgressTask) {
//			p.DonorSigned.State = task.StateCompleted
//			p.DonorSigned.Completed = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
//
//			p.FeeEvidenceSubmitted.State = task.StateCompleted
//
//			p.FeeEvidenceNotification.State = task.StateCompleted
//			p.FeeEvidenceNotification.Completed = time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
//
//			return p, []ProgressTask{
//				p.FeeEvidenceSubmitted,
//				p.DonorSigned,
//				p.FeeEvidenceNotification,
//				p.FeeEvidenceApproved,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"organisation with fee evidence pre signing": func(p Progress) (Progress, []ProgressTask) {
//			p.isOrganisation = true
//			p.FeeEvidenceSubmitted.State = task.StateCompleted
//
//			return p, []ProgressTask{
//				p.FeeEvidenceSubmitted,
//				p.FeeEvidenceApproved,
//				p.Paid,
//				p.ConfirmedID,
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"organisation with fee evidence pre signing communication sent": func(p Progress) (Progress, []ProgressTask) {
//			p.isOrganisation = true
//			p.FeeEvidenceSubmitted.State = task.StateCompleted
//
//			p.FeeEvidenceNotification.State = task.StateCompleted
//			p.FeeEvidenceNotification.Completed = time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
//
//			return p, []ProgressTask{
//				p.FeeEvidenceSubmitted,
//				p.FeeEvidenceNotification,
//				p.FeeEvidenceApproved,
//				p.Paid,
//				p.ConfirmedID,
//				p.DonorSigned,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//		"organisation with fee evidence post signing communication sent": func(p Progress) (Progress, []ProgressTask) {
//			p.isOrganisation = true
//			p.DonorSigned.State = task.StateCompleted
//			p.DonorSigned.Completed = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
//
//			p.FeeEvidenceSubmitted.State = task.StateCompleted
//
//			p.FeeEvidenceNotification.State = task.StateCompleted
//			p.FeeEvidenceNotification.Completed = time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
//
//			return p, []ProgressTask{
//				p.FeeEvidenceSubmitted,
//				p.DonorSigned,
//				p.FeeEvidenceNotification,
//				p.FeeEvidenceApproved,
//				p.Paid,
//				p.ConfirmedID,
//				p.CertificateProviderSigned,
//				p.AttorneysSigned,
//				p.LpaSubmitted,
//				p.StatutoryWaitingPeriod,
//				p.LpaRegistered,
//			}
//		},
//	}
//
//	for name, fn := range testcases {
//		t.Run(name, func(t *testing.T) {
//			progress, slice := fn(Progress{
//				FeeEvidenceSubmitted:      ProgressTask{State: task.StateNotStarted, Label: "Fee evidence submitted translation"},
//				FeeEvidenceNotification:   ProgressTask{State: task.StateNotStarted, Label: "Fee evidence notification translation"},
//				FeeEvidenceApproved:       ProgressTask{State: task.StateNotStarted, Label: "Fee evidence approved translation"},
//				Paid:                      ProgressTask{State: task.StateNotStarted, Label: "Paid translation"},
//				ConfirmedID:               ProgressTask{State: task.StateNotStarted, Label: "ConfirmedID translation"},
//				DonorSigned:               ProgressTask{State: task.StateInProgress, Label: "DonorSigned translation"},
//				CertificateProviderSigned: ProgressTask{State: task.StateNotStarted, Label: "CertificateProviderSigned translation"},
//				AttorneysSigned:           ProgressTask{State: task.StateNotStarted, Label: "AttorneysSigned translation"},
//				LpaSubmitted:              ProgressTask{State: task.StateNotStarted, Label: "LpaSubmitted translation"},
//				NoticesOfIntentSent:       ProgressTask{State: task.StateNotStarted, Label: "NoticesOfIntentSent translation"},
//				StatutoryWaitingPeriod:    ProgressTask{State: task.StateNotStarted, Label: "StatutoryWaitingPeriod translation"},
//				LpaRegistered:             ProgressTask{State: task.StateNotStarted, Label: "LpaRegistered translation"},
//			})
//
//			assert.Equal(t, slice, progress.ToSlice())
//		})
//	}
//}
//
//func TestProgressTrackerProgress(t *testing.T) {
//	lpaSignedAt := time.Now()
//	uid1 := actoruid.New()
//	uid2 := actoruid.New()
//	initialProgress := Progress{
//		FeeEvidenceSubmitted:      ProgressTask{State: task.StateNotStarted, Label: ""},
//		FeeEvidenceNotification:   ProgressTask{State: task.StateNotStarted, Label: ""},
//		Paid:                      ProgressTask{State: task.StateNotStarted, Label: ""},
//		ConfirmedID:               ProgressTask{State: task.StateNotStarted, Label: ""},
//		DonorSigned:               ProgressTask{State: task.StateInProgress, Label: "DonorSigned translation"},
//		CertificateProviderSigned: ProgressTask{State: task.StateNotStarted, Label: "CertificateProviderSigned translation"},
//		AttorneysSigned:           ProgressTask{State: task.StateNotStarted, Label: "AttorneysSigned translation"},
//		LpaSubmitted:              ProgressTask{State: task.StateNotStarted, Label: "LpaSubmitted translation"},
//		StatutoryWaitingPeriod:    ProgressTask{State: task.StateNotStarted, Label: "StatutoryWaitingPeriod translation"},
//		LpaRegistered:             ProgressTask{State: task.StateNotStarted, Label: "LpaRegistered translation"},
//	}
//
//	testCases := map[string]struct {
//		lpa              *lpadata.Lpa
//		donorTasks       Tasks
//		notifications    notification.Notifications
//		feeType          pay.FeeType
//		expectedProgress func() Progress
//		setupLocalizer   func(*task.mockLocalizer)
//	}{
//		"initial state": {
//			lpa: &lpadata.Lpa{
//				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			expectedProgress: func() Progress {
//				return initialProgress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					T("yourCertificateProviderHasDeclared").
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//			},
//		},
//		"initial state with certificate provider name": {
//			lpa: &lpadata.Lpa{
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			expectedProgress: func() Progress {
//				return initialProgress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("certificateProviderHasDeclared", map[string]interface{}{
//						"CertificateProviderFullName": "A B",
//					}).
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//			},
//		},
//		"fee evidence submitted - LPA not signed": {
//			lpa: &lpadata.Lpa{
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			feeType: pay.HalfFee,
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceApproved.State = task.StateInProgress
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateNotStarted
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("certificateProviderHasDeclared", map[string]interface{}{
//						"CertificateProviderFullName": "A B",
//					}).
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenSubmitted").
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenApproved").
//					Return("FeeEvidenceApproved translation")
//			},
//		},
//		"fee evidence notification sent - LPA not signed": {
//			lpa: &lpadata.Lpa{
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			feeType: pay.HalfFee,
//			notifications: notification.Notifications{
//				FeeEvidence: notification.Notification{Received: task.testNow},
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceNotification.State = task.StateCompleted
//				progress.FeeEvidenceNotification.Label = "FeeEvidenceNotification translation"
//				progress.FeeEvidenceNotification.Completed = task.testNow
//				progress.FeeEvidenceApproved.State = task.StateInProgress
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateNotStarted
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("certificateProviderHasDeclared", map[string]interface{}{
//						"CertificateProviderFullName": "A B",
//					}).
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenSubmitted").
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					T("yourFee").
//					Return("yourFee translation")
//				localizer.EXPECT().
//					FormatDate(task.testNow).
//					Return("formatted date")
//				localizer.EXPECT().
//					Format("weEmailedYouOnAbout", map[string]interface{}{
//						"On":    "formatted date",
//						"About": "yourFee translation",
//					}).
//					Return("FeeEvidenceNotification translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenApproved").
//					Return("FeeEvidenceApproved translation")
//			},
//		},
//		"fee evidence approved - LPA not signed": {
//			lpa: &lpadata.Lpa{
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			feeType: pay.HalfFee,
//			notifications: notification.Notifications{
//				FeeEvidence: notification.Notification{Received: task.testNow},
//			},
//			donorTasks: Tasks{PayForLpa: task.PaymentStateApproved},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceNotification.State = task.StateCompleted
//				progress.FeeEvidenceNotification.Label = "FeeEvidenceNotification translation"
//				progress.FeeEvidenceNotification.Completed = task.testNow
//				progress.FeeEvidenceApproved.State = task.StateCompleted
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateInProgress
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("certificateProviderHasDeclared", map[string]interface{}{
//						"CertificateProviderFullName": "A B",
//					}).
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenSubmitted").
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					T("yourFee").
//					Return("yourFee translation")
//				localizer.EXPECT().
//					FormatDate(task.testNow).
//					Return("formatted date")
//				localizer.EXPECT().
//					Format("weEmailedYouOnAbout", map[string]interface{}{
//						"On":    "formatted date",
//						"About": "yourFee translation",
//					}).
//					Return("FeeEvidenceNotification translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenApproved").
//					Return("FeeEvidenceApproved translation")
//			},
//		},
//		"fee evidence submitted - LPA signed": {
//			lpa: &lpadata.Lpa{
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				SignedAt:            task.testNow,
//			},
//			feeType: pay.HalfFee,
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceApproved.State = task.StateInProgress
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = task.testNow
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("certificateProviderHasDeclared", map[string]interface{}{
//						"CertificateProviderFullName": "A B",
//					}).
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenSubmitted").
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					T("yourLPAFeeEvidenceHasBeenApproved").
//					Return("FeeEvidenceApproved translation")
//			},
//		},
//		"lpa signed": {
//			lpa: &lpadata.Lpa{
//				Donor:     lpadata.Donor{FirstNames: "a", LastName: "b"},
//				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				SignedAt:  lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateInProgress
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					T("yourCertificateProviderHasDeclared").
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//			},
//		},
//		"certificate provider signed": {
//			lpa: &lpadata.Lpa{
//				Paid:                true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				SignedAt:            lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateInProgress
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					T("yourCertificateProviderHasDeclared").
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//			},
//		},
//		"attorneys signed": {
//			lpa: &lpadata.Lpa{
//				Paid:                true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}, {UID: uid2, SignedAt: lpaSignedAt.Add(time.Minute)}}},
//				SignedAt:            lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateInProgress
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					T("yourCertificateProviderHasDeclared").
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 2).
//					Return("AttorneysSigned translation")
//			},
//		},
//		"submitted": {
//			lpa: &lpadata.Lpa{
//				Paid:                true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
//				SignedAt:            lpaSignedAt,
//				Submitted:           true,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateCompleted
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					T("yourCertificateProviderHasDeclared").
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//			},
//		},
//		"perfect": {
//			lpa: &lpadata.Lpa{
//				Paid:                true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
//				SignedAt:            lpaSignedAt,
//				Submitted:           true,
//				PerfectAt:           lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateCompleted
//				progress.NoticesOfIntentSent.State = task.StateCompleted
//				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
//				progress.StatutoryWaitingPeriod.State = task.StateInProgress
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					T("yourCertificateProviderHasDeclared").
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//				localizer.EXPECT().
//					Format("weSentAnEmailYourLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
//					Return("NoticesOfIntentSent translation")
//				localizer.EXPECT().
//					FormatDate(lpaSignedAt).
//					Return("perfect-on")
//			},
//		},
//		"registered": {
//			lpa: &lpadata.Lpa{
//				Paid:                true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				SignedAt:            lpaSignedAt,
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}}},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Submitted:           true,
//				PerfectAt:           lpaSignedAt,
//				RegisteredAt:        lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateCompleted
//				progress.NoticesOfIntentSent.State = task.StateCompleted
//				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
//				progress.StatutoryWaitingPeriod.State = task.StateCompleted
//				progress.LpaRegistered.State = task.StateCompleted
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					T("yourCertificateProviderHasDeclared").
//					Return("CertificateProviderSigned translation")
//				localizer.EXPECT().
//					Count("attorneysHaveDeclared", 1).
//					Return("AttorneysSigned translation")
//				localizer.EXPECT().
//					Format("weSentAnEmailYourLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
//					Return("NoticesOfIntentSent translation")
//				localizer.EXPECT().
//					FormatDate(lpaSignedAt).
//					Return("perfect-on")
//			},
//		},
//	}
//
//	for name, tc := range testCases {
//		t.Run(name, func(t *testing.T) {
//			localizer := task.newMockLocalizer(t)
//			localizer.EXPECT().
//				T("youveSignedYourLpa").
//				Return("DonorSigned translation")
//			localizer.EXPECT().
//				T("weHaveReceivedYourLpa").
//				Return("LpaSubmitted translation")
//			localizer.EXPECT().
//				T("yourWaitingPeriodHasStarted").
//				Return("StatutoryWaitingPeriod translation")
//			localizer.EXPECT().
//				T("yourLpaHasBeenRegistered").
//				Return("LpaRegistered translation")
//
//			if tc.setupLocalizer != nil {
//				tc.setupLocalizer(localizer)
//			}
//
//			progressTracker := ProgressTracker{Localizer: localizer}
//
//			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.lpa, tc.donorTasks, tc.notifications, tc.feeType))
//		})
//	}
//}
//
//func TestLpaProgressAsSupporter(t *testing.T) {
//	dateOfBirth := date.Today()
//	lpaSignedAt := time.Now()
//	uid := actoruid.New()
//	initialProgress := Progress{
//		isOrganisation:            true,
//		FeeEvidenceSubmitted:      ProgressTask{State: task.StateNotStarted},
//		FeeEvidenceNotification:   ProgressTask{State: task.StateNotStarted},
//		Paid:                      ProgressTask{State: task.StateInProgress, Label: "Paid translation"},
//		ConfirmedID:               ProgressTask{State: task.StateNotStarted, Label: "ConfirmedID translation"},
//		DonorSigned:               ProgressTask{State: task.StateNotStarted, Label: "DonorSigned translation"},
//		CertificateProviderSigned: ProgressTask{State: task.StateNotStarted, Label: "CertificateProviderSigned translation"},
//		AttorneysSigned:           ProgressTask{State: task.StateNotStarted, Label: "AttorneysSigned translation"},
//		LpaSubmitted:              ProgressTask{State: task.StateNotStarted, Label: "LpaSubmitted translation"},
//		NoticesOfIntentSent:       ProgressTask{State: task.StateNotStarted},
//		StatutoryWaitingPeriod:    ProgressTask{State: task.StateNotStarted, Label: "StatutoryWaitingPeriod translation"},
//		LpaRegistered:             ProgressTask{State: task.StateNotStarted, Label: "LpaRegistered translation"},
//	}
//
//	testCases := map[string]struct {
//		lpa              *lpadata.Lpa
//		donorTasks       Tasks
//		notifications    notification.Notifications
//		feeType          pay.FeeType
//		expectedProgress func() Progress
//		setupLocalizer   func(*task.mockLocalizer)
//	}{
//		"initial state": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			expectedProgress: func() Progress {
//				return initialProgress
//			},
//		},
//		"fee evidence submitted - LPA not signed": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			feeType: pay.HalfFee,
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceApproved.State = task.StateInProgress
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateNotStarted
//				progress.Paid.State = task.StateNotStarted
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenSubmitted", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenApproved", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceApproved translation")
//				localizer.EXPECT().
//					Possessive("a b").
//					Return("Donor FullName Possessive")
//			},
//		},
//		"fee evidence notification sent - LPA not signed": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			feeType: pay.HalfFee,
//			notifications: notification.Notifications{
//				FeeEvidence: notification.Notification{Received: task.testNow},
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceNotification.State = task.StateCompleted
//				progress.FeeEvidenceNotification.Label = "FeeEvidenceNotification translation"
//				progress.FeeEvidenceNotification.Completed = task.testNow
//				progress.FeeEvidenceApproved.State = task.StateInProgress
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateNotStarted
//				progress.Paid.State = task.StateNotStarted
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenSubmitted", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenApproved", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceApproved translation")
//				localizer.EXPECT().
//					Possessive("a b").
//					Return("Donor FullName Possessive")
//				localizer.EXPECT().
//					FormatDate(task.testNow).
//					Return("Formatted date")
//				localizer.EXPECT().
//					T("theFee").
//					Return("Translated theFee")
//				localizer.EXPECT().
//					Format("weEmailedDonorNameOnAbout", map[string]interface{}{
//						"On":            "Formatted date",
//						"About":         "Translated theFee",
//						"DonorFullName": "a b",
//					}).
//					Return("FeeEvidenceNotification translation")
//			},
//		},
//		"fee evidence approved - LPA not signed": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//			},
//			feeType: pay.HalfFee,
//			notifications: notification.Notifications{
//				FeeEvidence: notification.Notification{Received: task.testNow},
//			},
//			donorTasks: Tasks{PayForLpa: task.PaymentStateApproved},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceNotification.State = task.StateCompleted
//				progress.FeeEvidenceNotification.Label = "FeeEvidenceNotification translation"
//				progress.FeeEvidenceNotification.Completed = task.testNow
//				progress.FeeEvidenceApproved.State = task.StateCompleted
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateNotStarted
//				progress.Paid.State = task.StateInProgress
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenSubmitted", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenApproved", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceApproved translation")
//				localizer.EXPECT().
//					Possessive("a b").
//					Return("Donor FullName Possessive")
//				localizer.EXPECT().
//					FormatDate(task.testNow).
//					Return("Formatted date")
//				localizer.EXPECT().
//					T("theFee").
//					Return("Translated theFee")
//				localizer.EXPECT().
//					Format("weEmailedDonorNameOnAbout", map[string]interface{}{
//						"On":            "Formatted date",
//						"About":         "Translated theFee",
//						"DonorFullName": "a b",
//					}).
//					Return("FeeEvidenceNotification translation")
//			},
//		},
//		"fee evidence submitted - LPA signed": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				SignedAt:            task.testNow,
//			},
//			feeType: pay.HalfFee,
//			expectedProgress: func() Progress {
//				progress := initialProgress
//
//				progress.FeeEvidenceSubmitted.State = task.StateCompleted
//				progress.FeeEvidenceSubmitted.Label = "FeeEvidenceSubmitted translation"
//				progress.FeeEvidenceApproved.State = task.StateInProgress
//				progress.FeeEvidenceApproved.Label = "FeeEvidenceApproved translation"
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = task.testNow
//				progress.Paid.State = task.StateNotStarted
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenSubmitted", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceSubmitted translation")
//				localizer.EXPECT().
//					Format("donorNamesLPAFeeEvidenceHasBeenApproved", map[string]interface{}{
//						"DonorFullNamePossessive": "Donor FullName Possessive",
//					}).
//					Return("FeeEvidenceApproved translation")
//				localizer.EXPECT().
//					Possessive("a b").
//					Return("Donor FullName Possessive")
//			},
//		},
//		"paid": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				Paid:                true,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateInProgress
//
//				return progress
//			},
//		},
//		"confirmed ID": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor: lpadata.Donor{
//					FirstNames:    "a",
//					LastName:      "b",
//					DateOfBirth:   dateOfBirth,
//					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
//				},
//				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				Paid:      true,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateCompleted
//				progress.DonorSigned.State = task.StateInProgress
//
//				return progress
//			},
//		},
//		"donor signed": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor: lpadata.Donor{
//					FirstNames:    "a",
//					LastName:      "b",
//					DateOfBirth:   dateOfBirth,
//					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
//				},
//				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				Paid:      true,
//				SignedAt:  lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateCompleted
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateInProgress
//
//				return progress
//			},
//		},
//		"certificate provider signed": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor: lpadata.Donor{
//					FirstNames:    "a",
//					LastName:      "b",
//					DateOfBirth:   dateOfBirth,
//					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
//				},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Paid:                true,
//				SignedAt:            lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateCompleted
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateInProgress
//
//				return progress
//			},
//		},
//		"attorneys signed": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor: lpadata.Donor{
//					FirstNames:    "a",
//					LastName:      "b",
//					DateOfBirth:   dateOfBirth,
//					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
//				},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Paid:                true,
//				SignedAt:            lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateCompleted
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateInProgress
//
//				return progress
//			},
//		},
//		"submitted": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor: lpadata.Donor{
//					FirstNames:    "a",
//					LastName:      "b",
//					DateOfBirth:   dateOfBirth,
//					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
//				},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Paid:                true,
//				SignedAt:            lpaSignedAt,
//				Submitted:           true,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateCompleted
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateCompleted
//
//				return progress
//			},
//		},
//		"perfect": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor: lpadata.Donor{
//					FirstNames:    "a",
//					LastName:      "b",
//					DateOfBirth:   dateOfBirth,
//					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
//				},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Paid:                true,
//				SignedAt:            lpaSignedAt,
//				Submitted:           true,
//				PerfectAt:           lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateCompleted
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateCompleted
//				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
//				progress.NoticesOfIntentSent.State = task.StateCompleted
//				progress.StatutoryWaitingPeriod.State = task.StateInProgress
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("weSentAnEmailTheLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
//					Return("NoticesOfIntentSent translation")
//				localizer.EXPECT().
//					FormatDate(lpaSignedAt).
//					Return("perfect-on")
//			},
//		},
//		"registered": {
//			lpa: &lpadata.Lpa{
//				IsOrganisationDonor: true,
//				Donor: lpadata.Donor{
//					FirstNames:    "a",
//					LastName:      "b",
//					DateOfBirth:   dateOfBirth,
//					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
//				},
//				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
//				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
//				Paid:                true,
//				SignedAt:            lpaSignedAt,
//				Submitted:           true,
//				PerfectAt:           lpaSignedAt,
//				RegisteredAt:        lpaSignedAt,
//			},
//			expectedProgress: func() Progress {
//				progress := initialProgress
//				progress.Paid.State = task.StateCompleted
//				progress.ConfirmedID.State = task.StateCompleted
//				progress.DonorSigned.State = task.StateCompleted
//				progress.DonorSigned.Completed = lpaSignedAt
//				progress.CertificateProviderSigned.State = task.StateCompleted
//				progress.AttorneysSigned.State = task.StateCompleted
//				progress.LpaSubmitted.State = task.StateCompleted
//				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
//				progress.NoticesOfIntentSent.State = task.StateCompleted
//				progress.StatutoryWaitingPeriod.State = task.StateCompleted
//				progress.LpaRegistered.State = task.StateCompleted
//
//				return progress
//			},
//			setupLocalizer: func(localizer *task.mockLocalizer) {
//				localizer.EXPECT().
//					Format("weSentAnEmailTheLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
//					Return("NoticesOfIntentSent translation")
//				localizer.EXPECT().
//					FormatDate(lpaSignedAt).
//					Return("perfect-on")
//			},
//		},
//	}
//
//	for name, tc := range testCases {
//		t.Run(name, func(t *testing.T) {
//			localizer := task.newMockLocalizer(t)
//			localizer.EXPECT().
//				Format(
//					"donorFullNameHasPaid",
//					map[string]interface{}{"DonorFullName": "a b"},
//				).
//				Return("Paid translation")
//			localizer.EXPECT().
//				Format(
//					"donorFullNameHasConfirmedTheirIdentity",
//					map[string]interface{}{"DonorFullName": "a b"},
//				).
//				Return("ConfirmedID translation")
//			localizer.EXPECT().
//				Format(
//					"donorFullNameHasSignedTheLPA",
//					map[string]interface{}{"DonorFullName": "a b"},
//				).
//				Return("DonorSigned translation")
//			localizer.EXPECT().
//				T("theCertificateProviderHasDeclared").
//				Return("CertificateProviderSigned translation")
//			localizer.EXPECT().
//				T("allAttorneysHaveSignedTheLpa").
//				Return("AttorneysSigned translation")
//			localizer.EXPECT().
//				T("opgHasReceivedTheLPA").
//				Return("LpaSubmitted translation")
//			localizer.EXPECT().
//				T("theWaitingPeriodHasStarted").
//				Return("StatutoryWaitingPeriod translation")
//			localizer.EXPECT().
//				T("theLpaHasBeenRegistered").
//				Return("LpaRegistered translation")
//
//			if tc.setupLocalizer != nil {
//				tc.setupLocalizer(localizer)
//			}
//
//			progressTracker := ProgressTracker{Localizer: localizer}
//
//			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.lpa, tc.donorTasks, tc.notifications, tc.feeType))
//		})
//	}
//}
