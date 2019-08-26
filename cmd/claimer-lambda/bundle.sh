#!/usr/bin/env bash

set -e

if [ -n "$(git status --porcelain)" ]; then
    echo "branch is dirty; commit and try again"
    exit 1
fi
hash=$(git rev-parse --short HEAD)

echo "compiling..."
cd "$(go env GOPATH)/src/github.com/oneiro-ndev/commands"
GOOS=linux go build ./cmd/claimer-lambda
echo "zipping..."
zf="cmd/claimer-lambda/claimer-lambda.zip"
zip -9 "$zf" claimer-lambda
du -h "$zf"
rm claimer-lambda # get rid of linux binary

echo "uploading..."
# publish the latest code
data=$(
    aws lambda update-function-code \
        --function-name "claimer-service" \
        --zip-file "fileb://$zf"
)
code_sha=$(echo "$data" | jq -r .CodeSha256)
revision=$(echo "$data" | jq -r .RevisionId)

update_net() {
    net=$1

    echo "updating $net..."

    echo "  updating environment and description..."
    data=$(
        aws lambda update-function-configuration \
            --function-name "claimer-service" \
            --description "claim service revision=$hash net=$net" \
            --environment "{
                    \"Variables\": {
                        \"PORT\": \"80\",
                        \"S3_CONFIG_BUCKET\": \"claimer-service-config\",
                        \"S3_CONFIG_PATH\": \"claimer_conf_${net}.toml\",
                        \"CLAIMER_SYNC_MODE\": \"1\",
                        \"HONEYCOMB_KEY_ENCRYPTED\": \"AQICAHjdJG7IpPV3q1vuQqBwx7BqkzK7ZRcoefcuUK42fSelBgHpdbhjVe/afPtoAfFl7KxvAAAAfjB8BgkqhkiG9w0BBwagbzBtAgEAMGgGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMA7IIiEFoMuxrTNjCAgEQgDssejpKt4Mi2Ki/uRNqGpERi3pjvKP85Q7e/A/2sRjK3H6q+RO2a9hQKAmiqxBEKYd/duvSooEseX3+pA==\",
                        \"HONEYCOMB_DATASET\": \"$net\",
                        \"HONEYCOMB_AUTOFLUSH\": \"1\"
                    }
                }"
    )
    revision=$(echo "$data" | jq -r .RevisionId)

    echo "  publishing version..."
    data=$(
        aws lambda publish-version \
            --function-name "claimer-service" \
            --revision-id "$revision" \
            --code-sha256 "$code_sha"
    )
    version=$(echo "$data" | jq -r .Version)

    echo "  updating alias..."
    data=$(
        aws lambda update-alias \
            --function-name "claimer-service" \
            --name "$net" \
            --function-version "$version"
    )
    revision=$(echo "$data" | jq -r .RevisionId)

    echo "  adding stage permission..."
    data=$(
        aws lambda add-permission \
            --function-name "arn:aws:lambda:us-east-1:578681496768:function:claimer-service:${net}" \
            --source-arn "arn:aws:execute-api:us-east-1:578681496768:7ovwffck3i/*/POST/claim_winner" \
            --principal apigateway.amazonaws.com \
            --statement-id "$(uuidgen)" \
            --action lambda:InvokeFunction \
            --revision-id "$revision"
    )
    revision=$(echo "$data" | jq -r .RevisionId)


    echo "  done!"
}

update_net testnet
update_net mainnet
