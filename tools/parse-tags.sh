#! /usr/bin/env bash

set -o pipefail
set -o errexit
set -o nounset

TAG_TO_PARSE="${1}"

if [[ ${TAG_TO_PARSE} =~ ^v[0-9]+\.[0-9]\.[0-9]+$ ]]; then
	awk -F'.' '{ print $1, $1 "." $2, $1 "." $2 "." $3 }' <<< "${TAG_TO_PARSE//v/}"
	exit 0
fi

echo "${TAG_TO_PARSE}"
