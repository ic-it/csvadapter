#!/bin/bash

# By @ic-it
# semver.sh - A simple script to manipulate semantic version strings
# Usage: semver.sh <action> <part> <version>
# Where:
#   <action> is one of: get, set, bump, inc, dec
#   <part> is one of: major, minor, patch, build
#   <version> is the version string to operate on

# Parse arguments
action=$1
part=$2
version=$3

# Write usage message
usage() {
  echo "semver.sh - A simple script to manipulate semantic version strings"
  echo "Usage: semver.sh <action> <part> <version>"
  echo "Where:"
  echo "  <action> is one of: get, set, bump, inc, dec"
  echo "  <part> is one of: major, minor, patch, build"
  echo "  <version> is the version string to operate on"
}

# Validate arguments
if [ -z "$action" ] || [ -z "$part" ] || [ -z "$version" ]; then
  usage
  exit 1
fi

# Extract parts
major=$(echo $version | cut -d. -f1)
minor=$(echo $version | cut -d. -f2)
patch=$(echo $version | cut -d. -f3 | cut -d+ -f1)
build=$(echo $version | cut -d+ -f2)

# Perform action
case $action in
  get)
    case $part in
      major) echo $major ;;
      minor) echo $minor ;;
      patch) echo $patch ;;
      build) echo $build ;;
      *) usage; exit 1 ;;
    esac
    ;;
  set)
    case $part in
      major) echo "$2.$minor.$patch" ;;
      minor) echo "$major.$2.$patch" ;;
      patch) echo "$major.$minor.$2" ;;
      build) echo "$major.$minor.$patch+$2" ;;
      *) usage; exit 1 ;;
    esac
    ;;
  bump)
    case $part in
      major) echo "$((major + 1)).0.0" ;;
      minor) echo "$major.$((minor + 1)).0" ;;
      patch) echo "$major.$minor.$((patch + 1))" ;;
      *) usage; exit 1 ;;
    esac
    ;;
  inc)
    case $part in
      major) echo "$((major + 1)).$minor.$patch" ;;
      minor) echo "$major.$((minor + 1)).$patch" ;;
      patch) echo "$major.$minor.$((patch + 1))" ;;
      *) usage; exit 1 ;;
    esac
    ;;
  dec)
    case $part in
      major) echo "$((major - 1)).$minor.$patch" ;;
      minor) echo "$major.$((minor - 1)).$patch" ;;
      patch) echo "$major.$minor.$((patch - 1))" ;;
      *) usage; exit 1 ;;
    esac
    ;;
  *)
    usage
    exit 1
    ;;
esac

# Exit successfully
exit 0
