#!/usr/bin/env bash

if [ $# -eq 0 ]
  then
    echo "Please provide workspaces to be protected."
fi

if [ "$1" == "-h" ]; then
  echo "Usage: $(basename "$0") [workspaces separated by a space]"
  exit 0
fi

export TF_EXIT_CODE="0"

in_use_workspaces=( "$@" )
reserved_workspaces=( "default" "production" "preproduction" "ur" "demo" "weblate" )

protected_workspaces=( "${in_use_workspaces[@]} ${reserved_workspaces[@]}" )
all_workspaces=$(terraform workspace list|sed 's/*//g')

for workspace in $all_workspaces
do
  case "${protected_workspaces[@]}" in
    *$workspace*)
      echo "protected workspace: $workspace"
      ;;
    *)
      echo "cleaning up workspace $workspace..."
      terraform workspace select "$workspace"
      terraform destroy -auto-approve
      if ! terraform destroy -auto-approve; then
        TF_EXIT_CODE=1
      fi
      echo "deleting opensearch index..."
      response=$(awscurl \
        "${DEVELOPMENT_OPENSEARCH_COLLECTION_ENDPOINT}/lpas_v2_$workspace" \
        --request DELETE \
        --region eu-west-1 \
        --service aoss)
        if [[ $response == *'"acknowledged":true'* ]]; then
          echo "Request successful."
        elif [[ $response == *'"status":404'* ]]; then
          echo "Request successful but index not found."
        else
          exit 1
        fi
      echo "deleting containter insights log group..."
      export AWS_REGION=eu-west-1
      aws logs delete-log-group --log-group-name /aws/ecs/containerinsights/"$workspace"/performance
      terraform workspace select default
      terraform workspace delete "$workspace"
      ;;
  esac
done

if [[ $TF_EXIT_CODE == "1" ]]; then
  exit 1
fi
