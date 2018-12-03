#!/bin/bash

# exit on failures
set -e

# get script's directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

escape_newlines() {
    echo "$1" | sed -e ':a' -e 'N' -e '$!ba' -e's/\n/\\n/g'
}

# subshell for cd
(
    cd "$DIR"

    # build for current platform and architecture
    go build .

    # generate api documentation
    api_doc="$(escape_newlines "$(./ndauapi -docs)")"
    tmpl="$(escape_newlines "$(cat ./README-template.md)")"

    # generate new readme with api documentation
    readme="${tmpl/api_replacement_token/$api_doc}"
    echo -e "$readme" > ./README.md
)
