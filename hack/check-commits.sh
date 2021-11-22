#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

remote=$(git remote -v | (grep "github\.com.rh-ecosystem-edge/NVIDIA-gpu-add-on-init-container.git" || true) | head -n 1 | awk '{ print $1 }')
if [ -z $remote ]; then
    echo "could not find remote for github.com/rh-ecosystem-edge/NVIDIA-gpu-add-on-init-container"
    exit 1
fi
main_branch="$remote/main"
current_branch="$(git rev-parse --abbrev-ref HEAD)"

revs=$(git rev-list "${main_branch}".."${current_branch}")

for commit in ${revs};
do
    commit_message=$(git cat-file commit ${commit} | sed '1,/^$/d')
    tmp_commit_file="$(mktemp)"
    echo "${commit_message}" > ${tmp_commit_file}
    ${__dir}/check-commit-message.sh "${tmp_commit_file}"
done


exit 0
