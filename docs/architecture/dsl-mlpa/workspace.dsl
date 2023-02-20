workspace {

    model {
        // Users
        actor = person "Actor" "Actor interacting with a Lasting Power of Attorney."

        enterprise "Modernising Lasting Power of Attorney" {
            // Software Systems
            makeSoftwareSystem = softwareSystem "Modernising Lasting Power of Attorney" "Digital Lasting Power of Attorney infrastructure" {
                // Components
                mlpaOnlineContainer = container "Make a Lasting Power of Attorney Online" "Allows users to draft a Lasting Power of Attorney online." "Go, HTML, CSS, JS" "Web Browser"
                mlpaStaging = container "Staging API/Database" "Stores and manages data for pre-registration." "API Gateway, Go, DynamoDB" "Container"
                mlpaLPAIDAPI = container "LPA ID API" "Manages the LPA IDs." "API Gateway, Go" "Container"
                mlpaOpgRegisterDatabase = container "Registered LPA Data Store" "Stores immutable LPA data with high availablility, security and auditing." "AuroraDB" "Database"
                mlpaOpgRegisterWriteAPI = container "Registered LPA Write API" "API for writing registered LPA data." "API Gateway, Go" "Container"
                mlpaOpgRegisterReadAPI = container "Registered LPA Read API" "Highly available API for reading and searching the LPA Register " "API Gateway" "Container"

                mlpaPaperIngestionAPI = container "LPA Paper Ingestion API" "Handles the ingestion of the Paper Journey." "API Gateway, Go" "Container"
                mlpaSiriusPublicAPI = container "Sirius Public API" "Interaction point between Sirius Case Management and other services." "API Gateway, Go" "Existing System"

                mlpaUaLPA = container "Use an LPA" "Use an LPA Service." "PHP, HTML, CSS, JS" "Web Browser Existing System"
                mlpaVaLPA = container "View an LPA" "View an LPA Service." "PHP, HTML, CSS, JS" "Web Browser Existing System"
                
                mlpaSiriusCaseManagement = container "Sirius Case Management" "Case Management for case working LPAs." "Go, HTML, CSS, JS" "Component" {
                    mlpaSiriusInternalAPI = component "Sirius Internal API" "" "API Gateway, PHP" "Existing System"
                    mlpaSiriusDatabase = component "Sirius Database" "Stores Case Management data." "AuroraDB" "Database Existing System"
                    mlpaSiriusMSPreRegistrationCaseManagement = component "Sirius Pre-Registered Case Management" "Sirius Microservice for Pre-Registered LPAs." "Go, HTML, CSS, JS"
                    mlpaSiriusMSRegisteredCaseManagement = component "Sirius Registered Case Management" "Sirius Microservice for Registered LPAs." "Go, HTML, CSS, JS"
                    mlpaSiriusMSPaperCaseManagement = component "Sirius Paper Channel Case Management" "Sirius Microservice for Paper Channel LPAs." "Go, HTML, CSS, JS"
                }
            }
        }

        // External Systems
        externalSoftwareSystems = softwareSystem "External Services" "GOV.UK Notify, Pay, One Login, Yoti, Ordanance Survey" "Existing System"
        externalScanningSoftware = softwareSystem "Scanning Software" "TBC" "Existing System"

        actor -> makeSoftwareSystem "interacts with"
        makeSoftwareSystem -> externalSoftwareSystems "interacts with"
        externalScanningSoftware -> mlpaPaperIngestionAPI "sends scanned LPA Data to"

        mlpaOnlineContainer -> mlpaStaging "makes calls to"
        mlpaLPAIDAPI -> mlpaStaging "gets LPA Code from"
        //mlpaLPAIDAPI -> mlpaSiriusCaseManagement "gets LPA Code from"
        mlpaStaging -> mlpaOpgRegisterWriteAPI "writes validated data to"
        mlpaStaging -> mlpaSiriusPublicAPI "writes case management data to and read data from"
        
        mlpaUaLPA -> mlpaSiriusPublicAPI "read data from"
        mlpaVaLPA -> mlpaSiriusPublicAPI "read data from"
        mlpaVaLPA -> mlpaOpgRegisterReadAPI "read data from"
        
        mlpaOpgRegisterWriteAPI -> mlpaOpgRegisterDatabase "interacts with"
        mlpaOpgRegisterReadAPI -> mlpaOpgRegisterDatabase "interacts with"

        mlpaSiriusPublicAPI -> mlpaSiriusCaseManagement "writes and read data from"
        mlpaPaperIngestionAPI -> mlpaSiriusPublicAPI "reads data from"
        mlpaPaperIngestionAPI -> mlpaStaging "writes data to"

        //mlpaSiriusMSRegisteredCaseManagement -> mlpaSiriusCaseManagement "writes and reads data from"
        //mlpaSiriusMSPreRegistrationCaseManagement -> mlpaSiriusCaseManagement "writes and reads data from"
        //mlpaSiriusMSPaperCaseManagement -> mlpaSiriusCaseManagement "writes and reads data from"

        //mlpaSiriusCaseManagement -> mlpaSiriusAPI "writes and reads data from"
        mlpaSiriusCaseManagement -> mlpaOpgRegisterWriteAPI "writes and reads data from"
        mlpaSiriusCaseManagement -> mlpaOpgRegisterReadAPI "writes and reads data from"
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

        component mlpaSiriusCaseManagement "Components" {
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
            element "Database Existing System" {
                background #999999
                color #ffffff
                shape Cylinder
            }
            element "Web Browser Existing System" {
                background #999999
                color #ffffff
                shape WebBrowser
            }
        }
    }
}
