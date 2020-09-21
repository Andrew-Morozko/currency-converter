#!/bin/bash

export CGO_ENABLED=0
export GO111MODULE=on

package_name="$1"
platforms="${@:2}"

if [ -n "$VER" ]; then
    VERVAL="v${VER}_"
else
    VERVAL=""
fi

mkdir /tmp/build

for platform in $platforms; do
    platform_split=(${platform//\// })
    export GOOS=${platform_split[0]}
    export GOARCH=${platform_split[1]}

    output_name=$package_name

    OSNAME="$GOOS"
    case "$OSNAME" in
    windows)
        OSNAME="win"
        output_name+='.exe'
        ;;

    darwin)
        OSNAME="macos"
        ;;
    esac

    ARCHNAME="$GOARCH"
    case "$ARCHNAME" in
    amd64)
        ARCHNAME="x64"
        ;;
    386)
        ARCHNAME="x86"
        ;;
    arm)
        ARCHNAME="arm32"
        ;;
    arm64)
        ARCHNAME="arm64"
        ;;
    esac

    echo "Building ${package_name}_${VERVAL}${OSNAME}_${ARCHNAME}"

    go build -ldflags="-s -w -X main._DEBUG_STR=false -X main.version=$VER" -trimpath -a -o "/tmp/build/$output_name" github.com/Andrew-Morozko/currency-converter/cmd/
    zip --quiet --junk-paths -9 "/release/${package_name}_${VERVAL}${OSNAME}_${ARCHNAME}.zip" "/tmp/build/$output_name"
done

rm -rf /tmp/build
