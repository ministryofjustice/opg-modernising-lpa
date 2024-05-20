makeRegisterSoftwareSystem = softwareSystem "Make and Register a Lasting Power of Attorney Online" "Allows users to draft a Lasting Power of Attorney online." {
    makeRegisterSoftwareSystem_database = container "Database" "Stores actor information, Draft LPA details etc." "DynamoDB" "Database"
    makeRegisterSoftwareSystem_storage = container "Storage" "Stores remissions and Exemptions evidence" "S3" "Database" 
    makeRegisterSoftwareSystem_search = container "Search" "Stores data for quick paging" "Elastic Search" "Database"
    makeRegisterSoftwareSystem_databaseMonitoringTelemetry = container "Monitoring and Telemetery" "Cloudwatch logs, X-Ray and RUM" "AWS Cloudwatch" "Database"
    makeRegisterSoftwareSystem_antiVirus = container "Anti virus" "Scans inbound files for malware/virus and blocks." "Lambda"{
        -> makeRegisterSoftwareSystem_storage "Reads & Tags"
    }
    makeRegisterSoftwareSystem_replication = container "Replication" "Batch replicates to Sirius" "Lambda" {
         -> makeRegisterSoftwareSystem_storage "Reads"
    }
    makeRegisterSoftwareSystem_eventHandler = container "Event receipt" "Receives events" "Lambda" {
        -> makeRegisterSoftwareSystem_database "Updates"
    }
    makeRegisterSoftwareSystem_webapp = container "App" "Provides and delivers static content, business logic, routing, third party access and database access" "Go, HTML, CSS, JS" "Web Browser" {
        -> makeRegisterSoftwareSystem_database "Reads from and writes to"
        -> makeRegisterSoftwareSystem_databaseMonitoringTelemetry "Writes to"
        -> makeRegisterSoftwareSystem_storage "Writes to"
        -> makeRegisterSoftwareSystem_search "Reads"
        -> makeRegisterSoftwareSystem_eventHandler "Handles"
    }


    makeRegisterSoftwareSystem_alb = container "Load balancer" "Routes traffic" "ALB"{
        -> makeRegisterSoftwareSystem_webapp
    }
    makeRegisterSoftwareSystem_WAF =  container "Web application Firewall" "Filters traffic" "WAF"{
        -> makeRegisterSoftwareSystem_alb
    }

}

users = person "User" "Application user"
users -> makeRegisterSoftwareSystem_WAF "Requests"


certificateProvider -> makeRegisterSoftwareSystem "Uses"
donor -> makeRegisterSoftwareSystem "Uses"
attorney -> makeRegisterSoftwareSystem "Uses"
organisation -> makeRegisterSoftwareSystem "Uses"
