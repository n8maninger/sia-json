#!/bin/bash

PRIVKEY=$1

# create fresh dist folder
rm -rf dist
mkdir dist

set -e

for os in darwin linux windows; do
	platforms=( amd64 )

	# add the arm arch for linux
	if [ "$os" == "linux" ]; then
		platforms=( amd64 arm )
	fi

	for arch in ${platforms[@]}; do
		bin=siajson

		# different naming convention for windows
		if [ "$os" == "windows" ]; then
			bin=siajson.exe
		fi

		rm -f dist/$bin

		GOOS=${os} GOARCH=${arch} go build -ldflags "-extldflags '-static'" -o dist/$bin main.go

		cd dist
		name="siajson-${os}-${arch}.zip"

		zip -m $name $bin
		cd ..
	done
done
