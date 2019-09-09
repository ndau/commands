# Bad Address Check Service

The DynamoDB table allows anyone with an AWS Auth token to read the address check table. However, that's more formal than we want; we want any rando, anywhere on the internet, to be able to check whether or not an address is OK.

To that end, we create a Lambda function which does nothing but forward check queries and responses back and forth between The Internet and DynamoDB.

## Usage

```sh
$ http post https://dh3bwcunu1.execute-api.us-east-1.amazonaws.com/default/address-check address=ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc7 --print=HBhb
POST /default/address-check HTTP/1.1
Accept: application/json, */*
Accept-Encoding: gzip, deflate
Connection: keep-alive
Content-Length: 63
Content-Type: application/json
Host: dh3bwcunu1.execute-api.us-east-1.amazonaws.com
User-Agent: HTTPie/0.9.8

{
    "address": "ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc7"
}

HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 199
Content-Type: application/json
Date: Wed, 21 Aug 2019 09:47:30 GMT
X-Amzn-Trace-Id: Root=1-5d5d1332-8aa0067c063f67b01ee2d82e;Sampled=0
x-amz-apigw-id: ew_v1G5QIAMFnbA=
x-amzn-RequestId: b10b6326-c3f8-11e9-abbb-e374a253dd72

{
    "data": {
        "address": "ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc7",
        "derivation-path": "/44'/20036'/100/5",
        "reason": "derives from 12-word phrase with identical words: hockey"
    },
    "exists": true
}

$ http post https://dh3bwcunu1.execute-api.us-east-1.amazonaws.com/default/address-check address=ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc8 --print=HBhb
POST /default/address-check HTTP/1.1
Accept: application/json, */*
Accept-Encoding: gzip, deflate
Connection: keep-alive
Content-Length: 63
Content-Type: application/json
Host: dh3bwcunu1.execute-api.us-east-1.amazonaws.com
User-Agent: HTTPie/0.9.8

{
    "address": "ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc8"
}

HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 17
Content-Type: application/json
Date: Wed, 21 Aug 2019 09:47:40 GMT
X-Amzn-Trace-Id: Root=1-5d5d133c-f7ef3af46f1c6c24792000c8;Sampled=0
x-amz-apigw-id: ew_xaFscIAMF_GA=
x-amzn-RequestId: b71086d8-c3f8-11e9-9edb-11b522a6e291

{
    "exists": false
}
```
