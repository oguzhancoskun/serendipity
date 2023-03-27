#!/bin/bash

PROFILE="default"
sam build --use-container
AWS_PROFILE=${PROFILE} sam package --s3-bucket serendipity1 --output-template-file packaged.yaml && \
AWS_PROFILE=${PROFILE} sam deploy --template-file packaged.yaml  --stack-name serendipity3 --region eu-west-1 --capabilities CAPABILITY_IAM
