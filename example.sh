#!/bin/sh

./cmd/import-details2jsonl/import-details2jsonl \
	./... |
	jq -c
