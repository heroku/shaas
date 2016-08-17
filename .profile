#!/usr/bin/env bash

# runs on start up on heroku

set -e

# exports the slug for download at http://shaas.example.com/app/slug.tgz

slug_file=/app/slug.tgz
if [ ! -f $slug_file ]; then
    slug_tmp_file=/tmp/slug.tgz
    find /app -type f -print0 | \
        sort -z | \
        GZIP=-n tar cz -T - --null --transform s,^app/,./app/, --owner=root --group=root > $slug_tmp_file
    mv $slug_tmp_file $slug_file

    echo SHA256:$(shasum --algorithm 256 $slug_file | cut -f 1 -d ' ') | tr -d '\n' > ${slug_file}.sha256
fi
