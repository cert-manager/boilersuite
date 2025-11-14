#!/usr/bin/env bash

# Copyright 2023 The cert-manager Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eu -o pipefail

BOILERSUITE=${1:-go run main.go}
FIXTURE_PATH=${2:-fixtures}

if [[ $BOILERSUITE = "" || $FIXTURE_PATH = "" ]]; then
	echo "usage: $0 <path-to-boilersuite> <path-to-fixtures>"
	exit 1
fi

logsfile=$(mktemp)

trap 'rm -f -- $logsfile' EXIT

$BOILERSUITE $FIXTURE_PATH &>$logsfile && exitcode=$? || exitcode=$?

if [[ $exitcode -eq 0 ]]; then
	echo "ERROR: expected boilersuite to fail but got a successful exit code"
	exit 1
fi

anyerrors=0

checkline() {
	rc=0

	grep -q "$1" $logsfile && rc=$? || rc=$?

	if [[ $rc -ne 0 ]]; then
		echo -e "ERROR: couldn't find required log line in output! wanted:\n > $1"
		anyerrors=1
	fi
}

checkline '"fixtures/Dockerfile": missing boilerplate'
checkline '"fixtures/Dockerfile.withsuffix": missing boilerplate'
checkline '"fixtures/bashscript_invalid.sh": missing boilerplate'
checkline '"fixtures/shscript_invalid.sh": missing boilerplate'
checkline '"fixtures/tooshort.py": missing boilerplate'
checkline 'at least one file had errors'

if [[ $anyerrors -ne 0 ]]; then
	echo "+++ at least one error was found in boilersuite output"
	echo "+++ full logs:"
	cat $logsfile
	exit 1
fi
