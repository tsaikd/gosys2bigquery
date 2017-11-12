#!/bin/bash

function timestamp() {
	date "+%Y-%m-%dT%H:%M:%S%z"
}

function isDarwin() {
	[ "$(uname -s)" == "Darwin" ]
}

function cachetime() {
	if isDarwin ; then
		date -v-7d +%s
	else
		date +%s -d -7day
	fi
}

function modified() {
	if isDarwin ; then
		stat -f %m "$@"
	else
		stat -c %Y "$@"
	fi
}

set -e

PN="${BASH_SOURCE[0]##*/}"
PD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

renice 15 $$
pushd "${PD}/.." >/dev/null

orgname="tsaikd"
projname="gosys2bigquery"
repo="github.com/${orgname}/${projname}"
githash="$(git rev-parse HEAD | cut -c1-6)"
cachedir="/tmp/${orgname}-${projname}-cache"
buildtoolimg="golang:1.9"

if [ -d "${cachedir}" ] ; then
	if [ "$(modified "${cachedir}")" -lt "$(cachetime)" ] ; then
		rm -rf "${cachedir}" || true
		docker pull "${buildtoolimg}"
	fi
else
	docker pull "${buildtoolimg}"
fi

echo "[$(timestamp)] build ${projname} binary (${githash})"
docker run --rm \
	-e CGO_ENABLED=0 \
	-e GITHUB_TOKEN="${GITHUB_TOKEN}" \
	-w "/go/src/${repo}" \
	-v "${PWD}:/go/src/${repo}" \
	-v "${cachedir}/go/src:/go/src" \
	-v "${cachedir}/go/bin:/go/bin" \
	"${buildtoolimg}" \
	"./docker/build-in-docker.sh"

echo "[$(timestamp)] ${projname} finished (${githash})"

popd >/dev/null
