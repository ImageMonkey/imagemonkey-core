#!/bin/bash

terser ../js/imagemonkey/views/annotation_helper.js ../js/imagemonkey/views/annotation.js -o ../js/imagemonkey/views/annotation.min.js
if [ $? -ne 0 ]; then
	echo "Couldn't minify js files...aborting"
	exit 1
fi

go build -o web web.go auth.go && ./web
