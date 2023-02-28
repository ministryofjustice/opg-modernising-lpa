workspace {

    model {
        // Users
        attorney = person "Attorney" "Attorney interacting with a Lasting Power of Attorney."
        donor = person "Donor" "Donor drafting the Lasting Power of Attorney."
        certificateProvider = person "Certificate Provider" "Certificate Provider interacting with a Lasting Power of Attorney."

        enterprise "Modernising Lasting Power of Attorney" {
            // Software Systems
            makeSoftwareSystem = softwareSystem "Make a Lasting Power of Attorney Online" "Allows users to draft a Lasting Power of Attorney online." {
                webapp = container "App" "Provides and delivers static content, business logic, routing, third party access and database access" "Go, HTML, CSS, JS" "Web Browser"
                notificationManagerApp = container "Notification Manager" "Manages the sending of various communication touchpoints" "Go, HTML, CSS, JS" "Web Browser"
                database = container "Database" "Stores actor information, Draft LPA details, access logs, etc." "DynamoDB" "Database"
                databaseMonitoringTelemetry = container "Monitoring and Telemetery" "Cloudwatch logs, X-Ray and RUM" "AWS Cloudwatch" "Database"
            }

            makePaperSoftwareSystem = softwareSystem "Make a Lasting Power of Attorney Paper Journey" "Allows users to draft a Lasting Power of Attorney via paper." {
                paperApp = container "Unknown App" "TBC" "TBC" "Container"
            }
        }

        // External Systems
        notifyExternalSoftwareSystem = softwareSystem "GOV.UK Notify" "Handles SMS, Email and Letters." "Existing System"
        payExternalSoftwareSystem = softwareSystem "GOV.UK Pay" "Handles Payments for Donors." "Existing System"
        oneLoginExternalSoftwareSystem = softwareSystem "GOV.UK One Login" "Handles Authentication and Identification of Actors." "Existing System"
        yotiExternalSoftwareSystem = softwareSystem "Yoti" "Used for identity." "Existing System"
        osExternalSoftwareSystem = softwareSystem "Ordanance survey" "Used for identity." "Existing System"

        certificateProvider -> webapp "Uses"
        donor -> webapp "Uses"
        attorney -> webapp "Uses"
        certificateProvider -> makePaperSoftwareSystem "Uses"
        donor -> makePaperSoftwareSystem "Uses"
        attorney -> makePaperSoftwareSystem "Uses"

        webapp -> database "Reads from and writes to"
        webapp -> databaseMonitoringTelemetry "Writes to"

        webapp -> notifyExternalSoftwareSystem "Sends communication with"
        webapp -> payExternalSoftwareSystem "Handles payment with"
        webapp -> oneLoginExternalSoftwareSystem "Authenticates users with"
        webapp -> yotiExternalSoftwareSystem "Identifies users with"
        webapp -> osExternalSoftwareSystem "Looks up addressed with"
    }

    views {
        systemlandscape "SystemLandscape" {
            include *
            autoLayout
        }

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
