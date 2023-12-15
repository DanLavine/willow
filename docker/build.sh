#/bin/bash

# Get the directory for the current script
cd "$(dirname "$0")"
SCRIPT_DIR="$(pwd)"

__usage="Usage: `basename $0` [OPTIONS]

Build any images for the Willow service and name them '[service name]:local-latest'

OPTIONS:
  -h                                          Print the help output
  -build [all, willow, limiter, locker]       Build a specific docker image. If this flag is not provided the default 'all' will be used
"

BUILD_OPTIONS="all"

# Get the 
while [[ $# -gt 0 ]]; do
  case $1 in
  -h)
    echo "$__usage"
    exit 0
    ;;
  -build)
    shift

    if [[ "$1" == "all" ]]; then
      BUILD_OPTIONS="all"
    elif [[ "$1" == "willow" ]]; then
      BUILD_OPTIONS="willow"
    elif [[ "$1" == "limiter" ]]; then
      BUILD_OPTIONS="limiter"
    elif [[ "$1" == "locker" ]]; then
      BUILD_OPTIONS="locker"
    else
      echo "Unknown build option '$1'. Valid options [all, willow, limiter, locker]"
      exit 1
    fi
    ;;
  *)
    echo "Unkown flag $1"
    exit 1
    ;;
  esac

  shift
done


pushd $SCRIPT_DIR/.. > /dev/null
  case "$BUILD_OPTIONS" in
    "all") 
      docker build -f docker/Dockerfile.willow -t willow:local-latest .
      docker build -f docker/Dockerfile.limiter -t limiter:local-latest .
      docker build -f docker/Dockerfile.locker -t locker:local-latest .
      ;;
    "willow")
      docker build -f docker/Dockerfile.willow -t willow:local-latest .
      ;;
    "limiter")
      docker build -f docker/Dockerfile.limiter -t limiter:local-latest .
      ;;
    "locker")
      docker build -f docker/Dockerfile.locker -t locker:local-latest .
  esac
popd > /dev/null