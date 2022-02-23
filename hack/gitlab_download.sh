#!/bin/sh

####
# SPDX-License-Identifier: CC BY-SA 4.0
# This script was based on https://stackoverflow.com/revisions/60568288/2
#
# Script adopted to fit our needs. Mostly regarding default values and file/folder paths

GITLAB_API_URL=https://gitlab.com/api/v4
export GITLAB_TOKEN=${GITLAB_TOKEN:-}
PROJECT=${PROJECT:-"nvidia/kubernetes/gpu-operator"}
BRANCH=${BRANCH:-"master"}
PROJECT_ENC=$(echo -n ${PROJECT} | jq -sRr @uri)
export WORKING_DIR=${WORKING_DIR:-.}

function fetchFile() {
  FILE=$1
  echo "Fetching file $FILE"
  FILE_ENC=$(echo -n ${FILE} | jq -sRr @uri)

  curl -s --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "${GITLAB_API_URL}/projects/${PROJECT_ENC}/repository/files/${FILE_ENC}?ref=${BRANCH}" -o /tmp/file.info
  if [ "$(dirname $FILE)" != "." ]; then
    mkdir -p $(dirname $FILE)
  fi
  cat /tmp/file.info | jq -r '.content' | tr -d "\n" | jq -sRr '@base64d' > $FILE
  rm /tmp/file.info
}

function fetchDir() {
  DIR=$1
  cd $WORKING_DIR
  echo "Fetching dir $DIR to $WORKING_DIR"
  mkdir -p $WORKING_DIR/$DIR
  FILES=$(curl -s --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" "${GITLAB_API_URL}/projects/${PROJECT_ENC}/repository/tree?ref=master&per_page=100&recursive=true&path=${DIR}" | jq -r '.[] | select(.type == "blob") | .path')
  for FILE in $FILES; do
    fetchFile $FILE
  done
}

fetchDir $1
