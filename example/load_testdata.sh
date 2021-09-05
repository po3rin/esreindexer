#!/usr/bin/env bash

ELASTICSEARCH_BASE_URL=localhost:9200

datafile="example/example.json"

echo "indexing testdata"

CODE=$(curl -X POST "${ELASTICSEARCH_BASE_URL}/example-v1/_bulk" -w %{http_code} -s --output /dev/null --header "Content-Type: application/json" --data-binary "@${datafile}")
if [[ $CODE -ne 200 ]]; then
    echo "ERROR: server returned HTTP code $CODE"
    exit 1
fi

echo "done"

