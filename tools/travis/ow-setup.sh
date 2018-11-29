#!/bin/bash

set -ex

# Build script for Travis-CI.

SCRIPTDIR=$(cd $(dirname "$0") && pwd)
HOMEDIR="$SCRIPTDIR/../../../"

# clone utilties repo. in order to run scanCode.py
cd ${HOMEDIR}

# shallow clone OpenWhisk repo.
git clone --depth 1 https://github.com/apache/incubator-openwhisk.git ow

cd ow
./tools/travis/setup.sh