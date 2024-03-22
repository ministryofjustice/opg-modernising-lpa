package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrespondentFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Correspondent{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestParseCorrespondentShare(t *testing.T) {
	result, err := ParseCorrespondentShare([]string{"OPG", "CertificateProvider"})

	assert.Nil(t, err)
	assert.Equal(t, CorrespondentShareOPG|CorrespondentShareCertificateProvider, result)
	assert.False(t, result.Empty())

	result, err = ParseCorrespondentShare([]string{"OPG", "What"})

	assert.NotNil(t, err)
	assert.True(t, result.Empty())
}

func TestCorrespondentShareHas(t *testing.T) {
	assert.True(t, (CorrespondentShareOPG | CorrespondentShareAttorneys | CorrespondentShareCertificateProvider).HasOPG())
	assert.True(t, (CorrespondentShareOPG | CorrespondentShareAttorneys | CorrespondentShareCertificateProvider).HasAttorneys())
	assert.True(t, (CorrespondentShareOPG | CorrespondentShareAttorneys | CorrespondentShareCertificateProvider).HasCertificateProvider())

	assert.True(t, CorrespondentShareOPG.HasOPG())
	assert.False(t, CorrespondentShareAttorneys.HasOPG())
	assert.False(t, CorrespondentShareCertificateProvider.HasOPG())

	assert.False(t, CorrespondentShareOPG.HasAttorneys())
	assert.True(t, CorrespondentShareAttorneys.HasAttorneys())
	assert.False(t, CorrespondentShareCertificateProvider.HasAttorneys())

	assert.False(t, CorrespondentShareOPG.HasCertificateProvider())
	assert.False(t, CorrespondentShareAttorneys.HasCertificateProvider())
	assert.True(t, CorrespondentShareCertificateProvider.HasCertificateProvider())
}
