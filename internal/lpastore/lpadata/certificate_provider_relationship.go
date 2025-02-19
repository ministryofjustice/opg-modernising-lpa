package lpadata

//go:generate go tool enumerator -type CertificateProviderRelationship -linecomment -empty
type CertificateProviderRelationship uint8

const (
	Personally     CertificateProviderRelationship = iota + 1 // personally
	Professionally                                            // professionally
)
