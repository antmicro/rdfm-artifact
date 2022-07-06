#!/bin/sh

set -e

# A simple shell script to verify that an Artifact is capeable of reading and writing
# an artifact on macOS (see https://tracker.mender.io/browse/MEN-2505).

touch rootfs.ext4

########## Step 0 - Help text produced #########
rdfm-artifact write | diff ma-write-help-text.golden -

########## Step 1 - Write an Artifact  ##########
rdfm-artifact write rootfs-image -t beaglebone -n release-1 -f rootfs.ext4 -o artifact.rdfm

########## Step 2 - Verify an Artifact ##########
rdfm-artifact validate artifact.rdfm > /dev/null

########## Step 3 - Read an Artifact   ##########
rdfm-artifact read artifact.rdfm | diff --ignore-matching-lines='modified:.*' ma-read-output.golden -

########## Step 4 - Clean up           ##########
rm rootfs.ext4 artifact.rdfm
