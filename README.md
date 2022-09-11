# GCSBC - Google Cloud Storage Bucket Cache

How to use Google Cloud Storage Bucket inside container and
run cache file server for the contents.

> 2022 edition with Go!

## Set up environment

Make `gcsbc.env` with following contents:

```bash
PROJECT_NAME=YOUR_GCP_PROJECT_NAME
BUCKET_NAME=YOUR_GCS_BUCKET_NAME
```

Apply it to current shell

```bash
source gcsbc.env
```

## GCP service account

Make service account and bind role:

```bash
$ gcloud iam service-accounts create \
    gcsbc-service-account --display-name "gcsbc"
$ gcloud iam roles create gcsbc \
    --project ${PROJECT_NAME} \
    --file gcsbc-roles.yaml
$ gcloud projects add-iam-policy-binding ${PROJECT_NAME} \
    --member=serviceAccount:gcsbc-service-account@${PROJECT_NAME}.iam.gserviceaccount.com \
    --role=projects/${PROJECT_NAME}/roles/gcsbc \
    --condition=None
```

Generate key for the service account and set it to k8s secret:

```bash
$ gcloud iam service-accounts keys create gcsbc-key.json \
    --iam-account=gcsbc-service-account@${PROJECT_NAME}.iam.gserviceaccount.com
```

## Test with docker

Build image:

```bash
docker build -t gcsbc:test .
```

Run container:

```bash
docker run -it --rm \
  --cap-add SYS_ADMIN --device /dev/fuse \
  -v `realpath gcsbc-key.json`:/sa-key.json \
  -e BUCKET_NAME=${BUCKET_NAME} \
  -p 8080:8080 \
  --entrypoint=/bin/sh \
  gcsbc:test
```

Mount bucket and run cache filer server (inside container):

```bash
# post_start.sh
# /app -r /bucket/
```

Check bucket contents accessable from host browser.

- http://127.0.0.1:8080/PATH/TO/CONTENTS

Unmount bucket and exit (inside container)

```bash
# pre_stop.sh
# Press Ctrl-D
```

## Deploy GCSBC to k8s (GKE)

Push the image:

```bash
docker tag gcsbc:test gcr.io/${PROJECT_NAME}/gcsbc:latest
docker push gcr.io/${PROJECT_NAME}/gcsbc:latest
```

Make ga-key.to secret:

```bash
k create secret generic sa-key --from-file=gcsbc-key.json
```

Deploy:

```bash
k apply -f deploy-gcsbc.yaml
```

## 참조
- [storage permissions](https://cloud.google.com/storage/docs/access-control/iam-permissions)
- [mount gcs to k8s](https://pliutau.com/mount-gcs-bucket-k8s/)
- [gcsfuse](https://github.com/GoogleCloudPlatform/gcsfuse)
- [gcloud auth activate-service-account](https://cloud.google.com/sdk/gcloud/reference/auth/activate-service-account)