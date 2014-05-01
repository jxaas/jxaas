#!/bin/bash -e

aws s3 cp --acl public-read .build/jxaas.tar.gz s3://jxaas/jxaas.tar.gz
