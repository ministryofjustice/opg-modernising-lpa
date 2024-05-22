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
        opgSiriusCaseManagementSystem = softwareSystem "Sirius Case Management" "Handles OPG casework tasks." "Existing System"
        opgLPAStoreSystem = softwareSystem "LPA Store" "Stores completed LPA." "Existing System"
        opgLPAUIDServiceSystem = softwareSystem "UID Service" "Generates reference" "Existing System"

        makeRegisterSoftwareSystem_webapp -> notifyExternalSoftwareSystem "Sends communication with"
        makeRegisterSoftwareSystem_webapp -> payExternalSoftwareSystem "Handles payment with"
        makeRegisterSoftwareSystem_webapp -> oneLoginExternalSoftwareSystem "Authenticates users with"
        makeRegisterSoftwareSystem_webapp -> osExternalSoftwareSystem "Looks up addressed with"
        makeRegisterSoftwareSystem_webapp -> opgSiriusCaseManagementSystem "Sends events and documents to"
        makeRegisterSoftwareSystem_webapp -> opgLPAStoreSystem "Sends completed LPA to"
        makeRegisterSoftwareSystem_webapp -> opgLPAUIDServiceSystem "Requests Unique reference from"

        makeRegisterSoftwareSystem_replication -> opgSiriusCaseManagementSystem "Sends documents to"
        opgSiriusCaseManagementSystem -> makeRegisterSoftwareSystem_eventHandler "Sends events to"
        opgLPAStoreSystem -> makeRegisterSoftwareSystem_eventHandler "Sends event to"
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
