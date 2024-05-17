workspace {

    model {
        !include https://raw.githubusercontent.com/ministryofjustice/opg-technical-guidance/main/dsl/poas/persons.dsl
        !include makeRegisterSoftwareSystem.dsl

        // External Systems
        notifyExternalSoftwareSystem = softwareSystem "GOV.UK Notify" "Handles SMS, Email and Letters." "Existing System"
        payExternalSoftwareSystem = softwareSystem "GOV.UK Pay" "Handles Payments for Donors." "Existing System"
        oneLoginExternalSoftwareSystem = softwareSystem "GOV.UK One Login" "Handles Authentication and Identification of Actors." "Existing System"
        osExternalSoftwareSystem = softwareSystem "Ordanance survey" "Used for postcode lookup." "Existing System"
        //other OPG systems
        opgSiriusCaseManagement = softwareSystem "Sirius Case Management" "Handles OPG casework tasks." "Existing system"
        opgLPAStore = softwareSystem "LPA Store" "Stores completed LPA." "Existing system"

        makeRegisterSoftwareSystem_webapp -> notifyExternalSoftwareSystem "Sends communication with"
        makeRegisterSoftwareSystem_webapp -> payExternalSoftwareSystem "Handles payment with"
        makeRegisterSoftwareSystem_webapp -> oneLoginExternalSoftwareSystem "Authenticates users with"
        makeRegisterSoftwareSystem_webapp -> osExternalSoftwareSystem "Looks up addressed with"
        makeRegisterSoftwareSystem_webapp -> opgSiriusCaseManagement "Sends events and documents to"
        makeRegisterSoftwareSystem_webapp -> opgLPAStore "Sends completed LPA to"
    }

    views {
        systemContext makeRegisterSoftwareSystem "SystemContext" {
            include *
            autoLayout
        }

        container makeRegisterSoftwareSystem {
            include *
            autoLayout
        }

        theme default

        styles {
            element "Existing System" {
                background #999999
                color #ffffff
            }
            element "Web Browser" {
                shape WebBrowser
            }
            element "Database" {
                shape Cylinder
            }
        }
    }
}
