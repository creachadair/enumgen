#!/bin/sh
#
# Update generated test enumerations.
#
# Incorporate a hash of the generator in the output so that the tests can
# detect if the source was modified without updating the tests.
#
set -euo pipefail

readonly tool='github.com/creachadair/enumgen'
readonly gen='../gen.go'
readonly yaml='gentest.yml'

if ! which sha256sum >/dev/null ; then
    hash="$(cat $gen $yaml | shasum -a 256 | cut -d' ' -f1)"
else
    hash="$(cat $gen $yaml | sha256sum | cut -d' ' -f1)"
fi

rm -f -- enums.go
go run "$tool" -config "$yaml" -output enums.go
echo "
// GeneratorHash is used by the tests to verify that the testdata
// package is updated when the code generator changes.
const GeneratorHash = \"$hash\"" >> enums.go

