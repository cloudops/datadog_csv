#!/usr/bin/env bash

gox -output="bin/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch="windows/amd64 darwin/amd64"