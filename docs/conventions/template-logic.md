# Template logic

Logic that varies content in page templates should be controlled in code rather than the template. For example:

```gotemplate
      {{ $detailsContent := "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay" }}

      {{ if .Lpa.CertificateProvider.Relationship.IsProfessionally }}
        {{ $detailsContent = "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessional" }}
      {{ end }}
```

should be set in a handler:

```go
data := &confirmYourDetailsData{
    DetailComponentContent: "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
}

if lpa.CertificateProvider.Relationship.IsProfessionally() {
    data.DetailComponentContent = "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessional"
}
```

This allows for easier testing in unit tests vs e2e tests and keeps logic in one place.
