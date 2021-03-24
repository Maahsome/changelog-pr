#!/bin/bash

declare SEMVER="$1"

# Set sed command to gsed if running on mac
if [[ `uname` = "Darwin" ]]; then
    GSED=$(which gsed)
    if [[ "${GSED}" == "gsed not found" ]]; then
        echo "This script requires GNU-sed, please install with 'brew install gnu-sed'"
        exit 1
    fi
    sed_command="gsed"
    GDATE=$(which gdate)
    if [[ "${GDATE}" == "gdate not found" ]]; then
        echo "This script requires GNU-sed, please install with 'brew install gnu-sed'"
        exit 1
    fi
    date_command="gdate"

else
    sed_command="sed"
    date_command="date"
fi

BUILD_DATE=$($date_command --utc +%FT%T.%3NZ)
GIT_TAG=$(git rev-parse HEAD)
GIT_REF="refs/tags/${SEMVER}"
echo "Build Date: ${BUILD_DATE}"
echo "SemVer: ${SEMVER}"
echo "GIT TAG: ${GIT_TAG}"
echo "GIT REF: ${GIT_REF}"

go build -ldflags "-X changelog-pr/cmd.semVer=${SEMVER} -X 'changelog-pr/cmd.buildDate=${BUILD_DATE}' -X changelog-pr/cmd.gitCommit=${GIT_TAG} -X changelog-pr/cmd.gitRef=${GIT_REF}"
