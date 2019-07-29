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
                        \"S3_CONFIG_PATH\": \"claimer-conf-${net}.toml\"
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

    echo "  done!"
}

update_net testnet
update_net mainnet
