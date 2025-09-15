#! /bin/bash

PORT=8081

urlsFilePath=$1

if [ -z "$urlsFilePath" ]; then
  echo "Usage: $0 <urls-file-path>"
  exit 1
fi

curl --parallel --parallel-immediate --parallel-max 3 --config "$urlsFilePath"
