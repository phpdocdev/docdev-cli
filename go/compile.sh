#!/usr/bin/env bash

package_name=$(go list)

#the full list of the platforms: https://golang.org/doc/install/source#environment
platforms=(
"darwin/amd64"
"darwin/arm64"
"linux/amd64"
"linux/arm64"
"windows/amd64"
 )

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name="../build/"$package_name'-'$GOOS'-'$GOARCH
    #output_name="../build/"$package_name
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    version=$(cat ../VERSION)
    # version=$(echo "$lastversion+0.1"|bc -l)

    (
        env GOOS=$GOOS GOARCH=$GOARCH GODEBUG=netdns=cgo+2 CGO_ENABLED=0 go build -ldflags "-s -w -X 'main.Version=v$version'" -o $output_name .;
        upx --best --lzma $output_name
    )
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done