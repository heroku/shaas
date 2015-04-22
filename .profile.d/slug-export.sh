#!/usr/bin/env bash

set -e

# runs on start up on heroku
# exports the slug for download at http://shaas.example.com/app/slug.tgz

tar cz --transform s,^./,./app/, --owner=root --absolute-names /app > /tmp/slug.tgz
mv /tmp/slug.tgz /app/slug.tgz
