workspace {

    model {
        // Users
        actor = person "Actor" "Actor interacting with a Lasting Power of Attorney."
        solicitor = person "Solicitor Software" "Solicitor interacting with a Lasting Power of Attorney."

        enterprise "Modernising Lasting Power of Attorney" {
            // Software Systems
            makeSoftwareSystem = softwareSystem "Modernising Lasting Power of Attorney" "Digital Lasting Power of Attorney infrastructure" {
                // Components
                mlpaOnlineContainer = container "Make a Lasting Power of Attorney Online" "Allows users to draft a Lasting Power of Attorney online." "Go, HTML, CSS, JS" "Web Browser" {
                    mlpaOnlineContainer_webapp = component "App" "Provides and delivers static content, business logic, routing, third party access and database access" "Go, HTML, CSS, JS" "Web Browser"
                    mlpaOnlineContainer_database = component "Database" "Stores actor information, Draft LPA details, access logs, etc." "DynamoDB" "Database"
                    mlpaOnlineContainer_databaseMonitoringTelemetry = component "Monitoring and Telemetery" "Cloudwatch logs, X-Ray and RUM" "AWS Cloudwatch" "Database"
                }

                mlpaDraftingService = container "LPA Drafting Service" "Stores and manages data for pre-registration." "API Gateway, Go, DynamoDB" "Container" {
                    mlpaDraftingServiceAPI = component "API" "Managing LPA data" "API Gateway, Go" "Component"
                    mlpaDraftingServiceSiriusAPI = component "Sirius API" "Managing Case Worker specific access" "API Gateway, Go" "Component"
                    mlpaDraftingServiceDatabase = component "Draft LPA Database" "Stores Draft LPA data." "DynamoDB" "Database"
                    mlpaDraftingServiceApp = component "App" "Manages data events and business logic." "Go" "Component"
                }
                mlpaSupporterAPI = container "Public LPA Support API" "Allows external companies to add submit LPAs." "API Gateway, Go" "Container"
                mlpaLPAIDAPI = container "LPA ID Service" "Manages the LPA IDs." "API Gateway, Go" "Container"
                
                mlpaOpgRegisterDatabase = container "Registered LPA Data Store" "Stores immutable LPA data with high availablility, security and auditing." "AuroraDB" "Database" {
                    mlpaOpgRegisterDatabase_database = component "Database" "Stores final Register LPA Data." "AuroraDB" "Database"
                    mlpaOpgRegisterDatabase_databaseMonitoringTelemetry = component "Monitoring and Telemetery" "Cloudwatch logs and X-Ray" "AWS Cloudwatch" "Database"
                }

                mlpaOpgRegisterService = container "Registered LPA Service" "Highly available API for reading and searching the LPA Register " "API Gateway" "Container" {
                    mlpaOpgRegisterService_ReadAPIGateway = component "Registered LPA Read API" "Highly available API for reading and searching the LPA Register " "API Gateway" "Container"
                    mlpaOpgRegisterService_WriteAPIGateway = component "Registered LPA Write API" "Interface to writing to the Registered LPA Database." "API Gateway, Go" "Container"
                    mlpaOpgRegisterService_ReadReplicaDatabase = component "Read Replica LPA Database" "Cached version of Registered LPA Data." "AuroraDB" "Database"
                    mlpaOpgRegisterService_ReadReplicaMonitoringTelemetry = component "Monitoring and Telemetery" "Cloudwatch logs and X-Ray" "AWS Cloudwatch" "Database"
                }

                mlpaPaperIngestionAPI = container "LPA Paper Ingestion Service" "Handles the ingestion of the Paper Journey." "API Gateway, Go" "Container"
                mlpaSiriusPublicAPI = container "Sirius Public API" "Interaction point between Sirius Case Management and other services." "API Gateway, Go" "Existing System"

                mlpaInternalPaymentService = container "OPG Internal Payment Service" "Handles GOV.UK Pay and Remissions and Exemptions information between all services and Sirius." "API Gateway, Go" "Existing System"

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
        mlpaOPGAuthService = softwareSystem "OPG Authentication Service" "User facing central Authentication service." "Container"
        externalSoftwareSystems = softwareSystem "External Services" "GOV.UK Notify, Pay, One Login, Yoti, Ordanance Survey" "Existing System"
        externalOPGSoftwareSystems = softwareSystem "OPG Services" "Court Ordered Severances" "Existing System"
        externalScanningSoftware = softwareSystem "Scanning Software" "TBC" "Existing System"
        mlpaUaLPA = softwareSystem "Use an LPA" "Use/View an LPA Service." "Web Browser Existing System"

        actor -> makeSoftwareSystem "interacts with"
        makeSoftwareSystem -> externalSoftwareSystems "interacts with"
        externalOPGSoftwareSystems -> mlpaSiriusCaseManagement "sends data to"
        externalScanningSoftware -> mlpaPaperIngestionAPI "sends scanned LPA Data to"

        mlpaSupporterAPI -> mlpaDraftingServiceAPI "makes calls to"
        solicitor -> mlpaSupporterAPI "interacts with"

        mlpaOnlineContainer -> mlpaDraftingServiceAPI "makes calls to"
        mlpaDraftingServiceAPI -> mlpaLPAIDAPI "gets LPA Code from"
        mlpaSiriusCaseManagement -> mlpaLPAIDAPI "gets LPA Code from"

        mlpaDraftingServiceSiriusAPI -> mlpaOpgRegisterService_WriteAPIGateway "writes validated data to"
        mlpaDraftingServiceSiriusAPI -> mlpaSiriusPublicAPI "writes case management data to and read data from"
        
        mlpaOnlineContainer -> mlpaOPGAuthService "authenticates with"
        mlpaUaLPA -> mlpaOPGAuthService "authenticates with"

        mlpaUaLPA -> mlpaSiriusPublicAPI "read data from"
        mlpaUaLPA -> mlpaOpgRegisterService_ReadAPIGateway "read data from"
        
        mlpaOpgRegisterService_WriteAPIGateway -> mlpaOpgRegisterDatabase_database "writes data to"
        mlpaOpgRegisterService_ReadAPIGateway -> mlpaOpgRegisterService_ReadReplicaDatabase "reads data from"
        mlpaOpgRegisterService_ReadReplicaDatabase -> mlpaOpgRegisterService_ReadReplicaMonitoringTelemetry "interacts with"
        mlpaOpgRegisterService_ReadReplicaDatabase -> mlpaOpgRegisterDatabase_database "syncs with"

        mlpaOpgRegisterDatabase_databaseMonitoringTelemetry -> mlpaOpgRegisterDatabase_database "interacts with"

        mlpaSiriusCaseManagement -> mlpaInternalPaymentService "writes and read data from"
        mlpaSiriusPublicAPI -> mlpaSiriusCaseManagement "writes and read data from"
        mlpaPaperIngestionAPI -> mlpaSiriusPublicAPI "reads data from"
        mlpaPaperIngestionAPI -> mlpaDraftingServiceAPI "writes data to"

        mlpaSiriusCaseManagement -> mlpaOpgRegisterService_ReadAPIGateway "reads data from"

        mlpaOnlineContainer_webapp -> mlpaOnlineContainer_database "reads and writes data to"
        mlpaOnlineContainer_database -> mlpaOnlineContainer_databaseMonitoringTelemetry "interacts with"
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

        component mlpaOpgRegisterDatabase "mlpaOpgRegisterDatabaseComponents" {
            include *
            autoLayout
        }

        component mlpaOnlineContainer "mlpaOnlineContainerComponents" {
            include *
            autoLayout
        }

        component mlpaOpgRegisterService "mlpaOpgRegisterWriteAPIComponents" {
            include *
            autoLayout
        }

        component mlpaSiriusCaseManagement "SiriusCaseManagementComponents" {
            include *
            autoLayout
        }

        component mlpaDraftingService "StagingApiComponents" {
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
