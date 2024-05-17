makeRegisterSoftwareSystem = softwareSystem "Make and Register a Lasting Power of Attorney Online" "Allows users to draft a Lasting Power of Attorney online." {
    makeRegisterSoftwareSystem_database = container "Database" "Stores actor information, Draft LPA details etc." "DynamoDB" "Database"
    makeRegisterSoftwareSystem_databaseMonitoringTelemetry = container "Monitoring and Telemetery" "Cloudwatch logs, X-Ray and RUM" "AWS Cloudwatch" "Database"
    makeRegisterSoftwareSystem_webapp = container "App" "Provides and delivers static content, business logic, routing, third party access and database access" "Go, HTML, CSS, JS" "Web Browser" {
        -> makeRegisterSoftwareSystem_database "Reads from and writes to"
        -> makeRegisterSoftwareSystem_databaseMonitoringTelemetry "Writes to"
    }
}

certificateProvider -> makeRegisterSoftwareSystem_webapp "Uses"
donor -> makeRegisterSoftwareSystem_webapp "Uses"
attorney -> makeRegisterSoftwareSystem_webapp "Uses"
organisation -> makeRegisterSoftwareSystem_webapp "Uses"
