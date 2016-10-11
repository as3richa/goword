#!/bin/sh
npm run-script build-dev-client && gofmt -w src/ && gb build && bin/server -debug
