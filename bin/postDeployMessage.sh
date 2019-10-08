#!/bin/bash

# This just echoes its entire command line to the deploy channel on slack.

SLACK_DEPLOY_WEBHOOK=https://hooks.slack.com/services/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
echo $@ |jq --raw-input '{text:.}' |curl -X POST -H 'Content-type: application/json' --data @- ${SLACK_DEPLOY_WEBHOOK}
