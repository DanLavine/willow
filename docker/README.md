Building and Running Docker Images
----------------------------------

To easily build any of the Docker files in this directory, we can use the `build.sh` script.

Usage:
```
Usage: build.sh [OPTIONS]

OPTIONS:
  -h                                          Print the help output
  -build [all, willow, limiter, locker]       Build a specific docker image. If this flag is not provided the default 'all' will be used
```

## All Images

The default of `build.sh` will build all three image and can just be run with:
```
./build.sh
```