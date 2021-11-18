#!/bin/bash
# ######################################################################################
# Copyright 2021 by tobi@backfrak.de. All
# rights reserved. Use of this source code is governed
# by a BSD-style license that can be found in the
# LICENSE file.
# ######################################################################################
# Script to build and run a docker container and do the following inside the container
# * clone the launchpad git repo
# * import the given samba_exporter github sources to the launchpad git repo
# * do the needed conversation steps, so debian package build can run
# * run debian binary package  build
# * run debian source package  build with tagging
# * commit the changes to the launchpad git repo
# * upload the debian source package to the launchpad ppa
# * push the launchpad git repo with tags
# ######################################################################################

# ################################################################################################################
# function definition
# ################################################################################################################
function print_usage()  {
    echo "Script to transfer a github tag to launchpad and publish the package in a ppa"
    echo ""
    echo "Usage: $0 tag <dry>"
    echo "-help     Print this help"
    echo "tag       The tag on the github repo to import, e. g. 0.7.8"
    echo "dry       Optional: Do not push the changes to launchpad git and not upload the sources to ppa"
    echo ""
    echo "The script expect the following environment variables to be set"
    echo "  LAUNCHPAD_SSH_ID_PUB        Public SSH key for the launchapd git repo"
    echo "  LAUNCHPAD_SSH_ID_PRV        Private SSH key for the launchapd git repo"
    echo "  LAUNCHPAD_GPG_KEY_PUB       Public GPG Key for the launchpad ppa"
    echo "  LAUNCHPAD_GPG_KEY_PRV       Private GPG Key for the launchpad ppa"
}

function buildAndRunDocker() {
    distVersion="$1"

    echo "Build the needed container from '$WORK_DIR/Dockerfile.${distVersion}', logging to $BRANCH_ROOT/logs/docker-build-${distVersion}.log"
    docker build --file "$WORK_DIR/Dockerfile.${distVersion}" --tag launchapd-publish-container-$distVersion . > $BRANCH_ROOT/logs/docker-build-${distVersion}.log 2>&1
    if [ "$?" != "0" ]; then 
        echo "Error during docker build"
        return 1
    fi
    echo "# ###################################################################"
    echo "Run the container"

    if [ "$dryRun" == "false" ]; then
        docker run --env LAUNCHPAD_SSH_ID_PUB="$LAUNCHPAD_SSH_ID_PUB" \
            --env LAUNCHPAD_SSH_ID_PRV="$LAUNCHPAD_SSH_ID_PRV"  \
            --env LAUNCHPAD_GPG_KEY_PUB="$LAUNCHPAD_GPG_KEY_PUB" \
            --env LAUNCHPAD_GPG_KEY_PRV="$LAUNCHPAD_GPG_KEY_PRV" \
            --mount type=bind,source="$DEB_PACKAGE_DIR",target="/build_results" \
            -i launchapd-publish-container-$distVersion \
            /bin/bash -c "/PublishLaunchpad.sh $tag"
    else
        docker run --env LAUNCHPAD_SSH_ID_PUB="$LAUNCHPAD_SSH_ID_PUB" \
            --env LAUNCHPAD_SSH_ID_PRV="$LAUNCHPAD_SSH_ID_PRV"  \
            --env LAUNCHPAD_GPG_KEY_PUB="$LAUNCHPAD_GPG_KEY_PUB" \
            --env LAUNCHPAD_GPG_KEY_PRV="$LAUNCHPAD_GPG_KEY_PRV" \
            --mount type=bind,source="$DEB_PACKAGE_DIR",target="/build_results" \
            -i launchapd-publish-container-$distVersion \
            /bin/bash -c "/PublishLaunchpad.sh $tag dry"
    fi

    if [ "$?" != "0" ]; then 
        echo "Error during docker run"
        return 1
    fi
    return 0
}

# ################################################################################################################
# variable asigenment
# ################################################################################################################
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
BRANCH_ROOT="$SCRIPT_DIR/.."
WORK_DIR="$SCRIPT_DIR/LaunchpadPublish"
DEB_PACKAGE_DIR="$BRANCH_ROOT/bin/deb-packages"

# ################################################################################################################
# parameter and environment check
# ################################################################################################################

if [ "$1" == "-help" ]; then
    print_usage
    exit 0
fi  

if [ "$1" == "" ]; then
    echo "Error: No Tag given"
    print_usage
    exit 1
else 
    tag=$1
fi

if [ "$2" == "dry" ]; then
    dryRun="true"
    echo "It's a dry run! No changes will be uploaded or pushed to launchpad"
else
    dryRun="false"
fi

if [ "$LAUNCHPAD_SSH_ID_PUB" == "" ]; then
    echo "Error: Environment variables LAUNCHPAD_SSH_ID_PUB not set"
    print_usage
    exit 1
fi

if [ "$LAUNCHPAD_SSH_ID_PRV" == "" ]; then
    echo "Error: Environment variables LAUNCHPAD_SSH_ID_PRV not set"
    print_usage
    exit 1
fi


if [ "$LAUNCHPAD_GPG_KEY_PUB" == "" ]; then
    echo "Error: Environment variables LAUNCHPAD_GPG_KEY_PUB not set"
    print_usage
    exit 1
fi

if [ "$LAUNCHPAD_GPG_KEY_PRV" == "" ]; then
    echo "Error: Environment variables LAUNCHPAD_GPG_KEY_PRV not set"
    print_usage
    exit 1
fi


if [[ "$tag" =~ "-pre" ]]; then
    if [ "$dryRun" == "false" ]; then
        echo "Warinig: A pre release will be imported to launchpad!"
    else
        echo "Do a dry run with a pre release"
    fi
fi
# ################################################################################################################
# functional code
# ################################################################################################################
pushd "$WORK_DIR"

if [ -d "$DEB_PACKAGE_DIR" ]; then
    echo "Use existing $DEB_PACKAGE_DIR dir after cleanup"
    rm -rf $DEB_PACKAGE_DIR/*
else 
    echo "Create $DEB_PACKAGE_DIR dir"
    mkdir -p "$DEB_PACKAGE_DIR"
fi

if [ -d "$BRANCH_ROOT/logs" ]; then
    echo "Use existing $BRANCH_ROOT/logs dir"
    if ls $BRANCH_ROOT/logs/docker-build*.log 1> /dev/null 2>&1; then
        echo "Delete existing $BRANCH_ROOT/logs/docker-build*.log"
        rm -rf $BRANCH_ROOT/logs/docker-build*.log 
    fi

else 
    echo "Create $BRANCH_ROOT/logs dir"
    mkdir -p "$BRANCH_ROOT/logs"
fi

dockerError="false"
echo "Publish tag $tag on launchpad within a docker cotainer for focal"
echo "# ###################################################################"
buildAndRunDocker "focal"
if [ "$?" != "0" ]; then
    dockerError="true"
fi

echo "Publish tag $tag on launchpad within a docker cotainer for impish"
echo "# ###################################################################"
if [ "$dockerError" == "false" ];then 
    buildAndRunDocker "impish"
    if [ "$?" != "0" ]; then
        dockerError="true"
    fi
fi

echo "Publish tag $tag on launchpad within a docker cotainer for bullseye"
echo "# ###################################################################"
buildAndRunDocker "bullseye"
if [ "$?" != "0" ]; then
    dockerError="true"
fi

echo "Publish tag $tag on launchpad within a docker cotainer for bullseye"
echo "# ###################################################################"
buildAndRunDocker "buster"
if [ "$?" != "0" ]; then
    dockerError="true"
fi

echo "# ###################################################################"
echo "Delete the container image when done"    
docker rmi -f $(docker images --filter=reference="launchapd-publish*" -q) 

popd

if [ "$dockerError" == "true" ];then 
    echo "Error detected"
    exit 1
fi

exit 0
