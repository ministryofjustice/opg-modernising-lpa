package event

type Category string

const (
	CategoryDraftLPADeleted = Category("DraftLPADeleted")
	CategoryFunnelStartRate = Category("FunnelStartRate")
)

type Measure string

const (
	MeasureOnlineDonor               = Measure("ONLINEDONOR")
	MeasureOnlineAttorney            = Measure("ONLINEATTORNEY")
	MeasureOnlineCertificateProvider = Measure("ONLINECERTIFICATEPROVIDER")
)

type Metrics struct {
	Metrics []MetricWrapper `json:"metrics"`
}

type MetricWrapper struct {
	Metric Metric `json:"metric"`
}

type Metric struct {
	Project          string
	Category         string
	Subcategory      Category
	Environment      string
	MeasureName      Measure
	MeasureValue     string
	MeasureValueType string
	Time             string
}
