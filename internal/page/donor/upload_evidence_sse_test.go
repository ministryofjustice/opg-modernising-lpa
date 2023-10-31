package donor

//func TestUploadEvidenceSSE(t *testing.T) {
//	w := httptest.NewRecorder()
//	r, _ := http.NewRequest(http.MethodGet, "/", nil)
//
//	donorStore := newMockDonorStore(t)
//	donorStore.
//		On("Get", r.Context()).
//		Return(&page.Lpa{Evidence: page.Evidence{Documents: []page.Document{
//			{Key: "a-key", Scanned: time.Now()},
//			{Key: "another-key"},
//		}}}, nil).Once()
//	donorStore.
//		On("Get", r.Context()).
//		Return(&page.Lpa{Evidence: page.Evidence{Documents: []page.Document{
//			{Key: "a-key", Scanned: time.Now()},
//			{Key: "another-key", Scanned: time.Now()},
//		}}}, nil).Once()
//
//	err := UploadEvidenceSSE(donorStore, 4*time.Millisecond, 2*time.Millisecond)(testAppData, w, r, &page.Lpa{Evidence: page.Evidence{Documents: []page.Document{
//		{Key: "a-key", Scanned: time.Now()},
//		{Key: "another-key"},
//	}}})
//	resp := w.Result()
//
//	bodyBytes, _ := io.ReadAll(resp.Body)
//
//	assert.Nil(t, err)
//	assert.Equal(t, http.StatusOK, resp.StatusCode)
//	assert.Equal(t, "event: message\ndata: {\"fileTotal\": 2, \"scannedTotal\": 1}\n\nevent: message\ndata: {\"fileTotal\": 2, \"scannedTotal\": 2}\n\nevent: message\ndata: {\"closeConnection\": \"1\"}\n\n", string(bodyBytes))
//}
//
//func TestUploadEvidenceSSEOnDonorStoreError(t *testing.T) {
//	w := httptest.NewRecorder()
//	r, _ := http.NewRequest(http.MethodGet, "/", nil)
//
//	donorStore := newMockDonorStore(t)
//	donorStore.
//		On("Get", r.Context()).
//		Return(&page.Lpa{Evidence: page.Evidence{Documents: []page.Document{
//			{Key: "a-key", Scanned: time.Now()},
//			{Key: "another-key"},
//		}}}, expectedError)
//
//	err := UploadEvidenceSSE(donorStore, 4*time.Millisecond, 2*time.Millisecond)(testAppData, w, r, &page.Lpa{Evidence: page.Evidence{Documents: []page.Document{
//		{Key: "a-key", Scanned: time.Now()},
//		{Key: "another-key"},
//	}}})
//	resp := w.Result()
//
//	bodyBytes, _ := io.ReadAll(resp.Body)
//
//	assert.Equal(t, expectedError, err)
//	assert.Equal(t, http.StatusOK, resp.StatusCode)
//	assert.Equal(t, "event: message\ndata: {\"closeConnection\": \"1\"}\n\n", string(bodyBytes))
//}
