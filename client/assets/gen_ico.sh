#!/bin/bash

filename=$(basename -- "$1")
extension="${filename##*.}"
filename="${filename%.*}"

rm -rf tmp
mkdir -p tmp

sizes=(16 32 48 128 256)
for size in ${sizes[@]}; do
    magick $1 -sample ${size}x${size} tmp/${size}.${extension}
done
files=( "${sizes[@]/%/.${extension}}" )
files=( "${files[@]/#/tmp/}" )
magick ${files[@]} -colors 256 ${filename}.ico

rm -rf tmp