#!/bin/bash

VERSION_TYPE=$1
GIT_USER=$2
GIT_EMAIL=$3
GIT_USERNAME=$4
GIT_PAT=$5

# get highest tag number, and add v1.0.0 if doesn't exist
git fetch -t 2>/dev/null
CURRENT_VERSION=`git tag -l | tail -1 2>/dev/null`
echo "Current Version: $CURRENT_VERSION"

if [[ $CURRENT_VERSION == '' ]]
then
  CURRENT_VERSION='v1.0.0'
fi
echo "Incremented Version: $CURRENT_VERSION"

# replace . with space so can split into an array
CURRENT_VERSION_PARTS=(${CURRENT_VERSION//./ })

# get number parts
VNUM1=${CURRENT_VERSION_PARTS[0]}
echo "VNUM1:$VNUM1"
VNUM2=${CURRENT_VERSION_PARTS[1]}
VNUM3=${CURRENT_VERSION_PARTS[2]}

if [[ $VERSION_TYPE == 'major' ]]
then
  VNUM1=$(($VNUM1+1))
  echo "VNUM1:$VNUM1"
elif [[ $VERSION_TYPE == 'minor' ]]
then
  VNUM2=$(($VNUM2+1))
elif [[ $VERSION_TYPE == 'patch' ]]
then
  VNUM3=$(($VNUM3+1))
else
  echo "No version type (https://semver.org/) or incorrect type specified, try: -v [major, minor, patch]"
  exit 1
fi

# create new tag
NEW_TAG="$VNUM1.$VNUM2.$VNUM3"
echo "($VERSION_TYPE) updating $CURRENT_VERSION to $NEW_TAG"

# get current hash and see if it already has a tag
GIT_COMMIT=`git rev-parse HEAD`
NEEDS_TAG=`git describe --contains $GIT_COMMIT 2>/dev/null`

# only tag if no tag already
if [ -z "$NEEDS_TAG" ]; then
  echo "Tagged with $NEW_TAG"
  git tag $NEW_TAG
  git config --global url."https://$GIT_USERNAME:$GIT_PAT@github.com".insteadOf https://github.com
  git config --global user.email "$GIT_EMAIL"
  git config --global user.name "$GIT_USER"
  git config --global --add safe.directory '*'
  git push origin $NEW_TAG
else
  echo "Already a tag on this commit"
fi

exit 0
