#!/bin/bash

if ! git diff --quiet HEAD; then
    git diff HEAD | cat
    echo "Repo is dirty, see above for diff"
    echo "This happens when the build process modifies checked-in files. The build process runs the linter, so you may want to try running 'make lint'"
    exit 1
fi

NOT_TRACKED_FILES=$(git ls-files . --exclude-standard --others | tr -d '[:space:]')
if ! [[ -z ${NOT_TRACKED_FILES} ]]; then
    echo "${NOT_TRACKED_FILES}"
    echo "Repo is dirty, see above for diff"
    echo "This happens when the build process creates files that are not checked-in."
    exit 1
fi
