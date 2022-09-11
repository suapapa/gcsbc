#!/bin/sh

gcsfuse --implicit-dirs \
  --key-file=/sa-key.json $BUCKET_NAME /bucket