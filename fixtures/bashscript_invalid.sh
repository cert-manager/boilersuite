#!/usr/bin/env bash

set -eu -o pipefail

echo "this script tests longer shebangs for bash"
echo "it intentionally (incorrectly) missed out boilerplate"

# lines below are padding so that we don't reject the file for being too short;
# in this test, we want to ensure that the lack of prefix is reported

# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
# padding
