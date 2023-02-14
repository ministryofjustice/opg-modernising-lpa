workspace {

    model {
        // Users
        attorney = person "Attorney" "Attorney interacting with a Lasting Power of Attorney."
        donor = person "Donor" "Donor drafting the Lasting Power of Attorney."
        certificateProvider = person "Certificate Provider" "Certificate Provider interacting with a Lasting Power of Attorney."

        // Software Systems
        makeSoftwareSystem = softwareSystem "Make a Lasting Power of Attorney Online" "Allows users to draft a Lasting Power of Attorney online." {
            webapp = container "Web Application" "Provides and delivers static content, business logic, routing, third party access and database access" "Golang, HTML, CSS, JS" "Web Browser"
            database = container "Database"  "Stores actor information, Draft LPA details, access logs, etc." "DynamoDB" "Database"
        }

        // External Systems
        notifySoftwareSystem = softwareSystem "GOV.UK Notify" "Handles SMS, Email and Letters." "Existing System"
        paySoftwareSystem = softwareSystem "GOV.UK Pay" "Handles Payments for Donors." "Existing System"
        oneLoginSoftwareSystem = softwareSystem "GOV.UK One Login" "Handles Authentication and Identification of Actors." "Existing System"

        certificateProvider -> webapp "Uses"
        attorney -> webapp "Uses"
        donor -> webapp "Uses"

        webapp -> database "Reads from and writes to"

        webapp -> notifySoftwareSystem "Sends communication to"
        webapp -> paySoftwareSystem "Handles payment via"
        webapp -> oneLoginSoftwareSystem "Authenticates users via"
    }

    views {
        systemContext makeSoftwareSystem "SystemContext" {
            include *
            autoLayout
        }

        container makeSoftwareSystem {
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
