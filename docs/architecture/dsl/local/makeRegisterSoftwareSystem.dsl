makeRegisterSoftwareSystem = softwareSystem "Make and Register a Lasting Power of Attorney Online" "Allows users to draft a Lasting Power of Attorney online." {
    webapp = container "App" "Provides and delivers static content, business logic, routing, third party access and database access" "Go, HTML, CSS, JS" "Web Browser"
    database = container "Database" "Stores actor information, Draft LPA details, access logs, etc." "DynamoDB" "Database"
    databaseMonitoringTelemetry = container "Monitoring and Telemetery" "Cloudwatch logs, X-Ray and RUM" "AWS Cloudwatch" "Database"
}
