#!/bin/bash

DIR="${DIR:-.}"
while getopts d: option; do
case "${option}" in
d) DIR=${OPTARG};;
esac
done

# set the linter var before we enable -e, since the var could be empty

# Use gometalinter HEAD until the following patch hits a release:
# https://github.com/alecthomas/gometalinter/pull/505
LINTER=$(which gometalinter)
set -e

if [[ -z ${LINTER} ]]; then
    go get -u github.com/alecthomas/gometalinter
    LINTER=$(which gometalinter)
    ${LINTER} --install
fi

DIRECTORY=$(dirname $0)
# echo loading config from "${PWD}/${DIRECTORY}/lintconfig_base.json"
# echo invoking "${LINTER} --config=${PWD}/${DIRECTORY}/lintconfig_base.json --vendor $DIR/..."
${LINTER} --config=${PWD}/${DIRECTORY}/lintconfig_base.json --vendor $DIR/...
echo Done linting
