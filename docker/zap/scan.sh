#!/usr/bin/env bash

set -e
set -o pipefail

# Setup
if [ -z "$SCAN_URL" ]; then
  echo "ERROR - SCAN_URL has not been set and is a required variable"
  exit 1
fi
echo "INFO - Current Configuration for Zap Scan"
echo "INFO - SCAN_URL set to:- $SCAN_URL"

if [ -z "$ACTIVE_SCAN" ]; then
  ACTIVE_SCAN="false"
fi
echo "INFO - ACTIVE_SCAN set to:- $ACTIVE_SCAN"

if [ -z "$ACTIVE_SCAN_TIMEOUT" ]; then
  ACTIVE_SCAN_TIMEOUT=300
fi
echo "INFO - ACTIVE_SCAN_TIMEOUT set to:- $ACTIVE_SCAN_TIMEOUT"

if [ -z "$SERVICE_NAME" ]; then
  SERVICE_NAME="Zap"
fi
# Clean up Service Name variable
SERVICE=$(echo $SERVICE | tr -dc '[:alnum:]\n\r')
echo "INFO - SERVICE_NAME set to:- $SERVICE_NAME"

disable_scanner_rules() {
    # Customise these for your service
    curl --fail "localhost:8090/JSON/ascan/action/disableScanners?ids=20015,40018,40019,40020,40021,40022,40024,40027,40043,40045,90034,40042"
}

active_scan(){
    echo "INFO - Starting Active Scan of $SERVICE_NAME"

    # Configure Zap to stop checking after 10 Alerts for a rule have been reported.
    curl --fail "localhost:8090/JSON/ascan/action/setOptionMaxAlertsPerRule?Integer=10"

    # Config Zap's number of parallel active scanners
    curl --fail "localhost:8090/JSON/ascan/action/setOptionHostPerScan?Integer=5"

    #####################################################################################################################
    #####################################################################################################################
    ###### Configure which active scan rules to disable if you want to be more selective in your scan to save time ######
    #####################################################################################################################
    #####################################################################################################################
    disable_scanner_rules

    # Start Active Scan
    curl --fail "localhost:8090/JSON/ascan/action/scan?url=$SCAN_URL"

    # Reset command timer
    SECONDS=0
    # Set defaults
    STATUS="RUNNING"
    CURRENT_RUN_TIME=$SECONDS
    # Let active scan run until it is finished or the timeout is reached
    until [ "$STATUS" == "\"FINISHED\"" ] || [ $CURRENT_RUN_TIME -ge $ACTIVE_SCAN_TIMEOUT ]
    do
        STATUS=$(curl --fail --silent "localhost:8090/JSON/ascan/view/scans" | jq '.scans[] | select(.id == "0").state')
        CURRENT_RUN_TIME=$SECONDS
        echo "INFO - Active Scan Status: $STATUS"
        echo "INFO - Scan has been running for $CURRENT_RUN_TIME seconds"
        sleep 10
    done
    if [ "$STATUS" == "\"FINISHED\"" ]
    then
        echo "INFO - Active Scan Completed"
    fi

    # If timeout is reached before the scan is finished stop the scan manually
    if [ "$STATUS" == "\"RUNNING\"" ]
    then
        echo "INFO - Active Scan is still running - stopping gracefully"
        curl --fail "localhost:8090/JSON/ascan/action/stopAllScans/"
        echo "INFO - Stopped all active scans"
    fi
}

filter_false_positives() {
    # This will globally report all alert results of this type as a false positive and exclude them from the reports
    # These Alert ID's can be found by viewing the ids of the /JSON/pscan/view/scanners/ and /JSON/ascan/view/scanners/ endpoints
    ALERTS=(
        "10031" # User Controllable HTML Element Attribute (Potential XSS)
        "6"     # Path Traversal
        "30002" # Format String Error
    )
    for ALERT in "${ALERTS[@]}"
    do
        # Add a Global Alert Filter
        echo "INFO - Excluding Rule $ALERT from Results"
        curl --fail "localhost:8090/JSON/alertFilter/action/addGlobalAlertFilter?ruleId="$ALERT"&newLevel=-1&enabled=true"
    done
    # Apply All Global Alert Filters
    echo "INFO - Applying all enabled Global Results Filters"
    curl --fail "localhost:8090/JSON/alertFilter/action/applyGlobal"
}

# Run default reports -  HTML Plus for Humans and Sarif JSON for Machines
run_reports() {
    echo "INFO - Running Reports for $SERVICE_NAME"
    curl --fail "localhost:8090/JSON/reports/action/generate?title="$SERVICE_NAME"Report&reportFileName="$SERVICE_NAME"Report&template=traditional-html-plus&reportDir=/zap/wrk&includedConfidences=Confirmed|High|Medium|Low"
    curl --fail "localhost:8090/JSON/reports/action/generate?title="$SERVICE_NAME"Report&reportFileName="$SERVICE_NAME"Report&template=sarif-json&reportDir=/zap/wrk&includedConfidences=Confirmed|High|Medium|Low"
}

#######################
#######################
##### Main script #####
#######################
#######################

# Passive scan is done as Cypress executes through Zap
# If Active Scan is enabled, run it
if [ "$ACTIVE_SCAN" == "true" ]
then
    active_scan
fi

##########################################################################################
##########################################################################################
##### Enable the below if you want to filter out results with the filtering function #####
##########################################################################################
##########################################################################################
filter_false_positives

# Run Reports
run_reports

echo "INFO - Zap Scan Complete!"
