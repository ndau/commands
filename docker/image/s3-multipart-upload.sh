#!/usr/bin/env bash

# Copyright 2017 Jesse Wang
# Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
# The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

# This script requires: 
# - AWS CLI to be properly configured (https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html)
# - Account has s3:PutObject access for the target S3 bucket

# Usage:
# bash s3-multipart-upload.sh YOUR_FILE YOUR_BUCKET (OPTIONAL: PROFILE)
# bash s3-multipart-upload.sh files.zip my-files
# bash s3-multipart-upload.sh files.zip my-files second-profile

fileName=$1
bucket=$2
profile=${3-default}

#Set to 1 GiB as default, 5 GiB is the limit for AWS files
mbSplitSize=1
((partSize = $mbSplitSize * 1000000000))

# Get main file size
echo "Preparing $fileName for multipart upload"
fileSize=`wc -c $fileName | awk '{print $1}'`
((parts = ($fileSize+$partSize-1) / partSize))

# Get main file hash
mainMd5Hash=`openssl md5 -binary $fileName | base64`

# Make directory to store temporary parts
echo "Splitting $fileName into $parts temporary parts"
mkdir temp-parts
cd temp-parts
split -b $partSize ../$fileName

# Ensure we're using AWS S3 accelerated upload
aws configure set default.s3.use_accelerate_endpoint true

# Create mutlipart upload
echo "Initiating multipart upload for $fileName"
uploadId=`aws s3api create-multipart-upload --bucket $bucket --key $fileName --metadata md5=$mainMd5Hash --profile $profile | jq -r '.UploadId'`

# Generate fileparts.json file that will be used at the end of the multipart upload
jsonData="{\"Parts\":["
for file in *
  do 
    ((index++))		
    echo "Uploading part $index of $parts..."
    hashData=`openssl md5 -binary $file | base64`
    eTag=`aws s3api upload-part --bucket $bucket --key $fileName --part-number $index --body $file --upload-id $uploadId --profile $profile | jq -r '.ETag'`
    jsonData+="{\"ETag\":$eTag,\"PartNumber\":$index}"

    if (( $index == $parts )) 
      then
        jsonData+="]}"
      else
        jsonData+=","
    fi	
done
jq -n $jsonData > fileparts.json

# Complete multipart upload, check ETag to verify success 
mainEtag=`aws s3api complete-multipart-upload --multipart-upload file://fileparts.json --bucket $bucket --key $fileName --upload-id $uploadId --profile $profile | jq -r '.ETag'`
if [[ $mainEtag != "" ]]; 
  then 
    echo "Successfully uploaded: $fileName to S3 bucket: $bucket"
  else
    echo "Something went wrong! $fileName was not uploaded to S3 bucket: $bucket"
fi

# Clean up files
cd ..
rm -R temp-parts
