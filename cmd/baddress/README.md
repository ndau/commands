# `baddress`: populate a DynamoDB with bad addresses

We want a service where anyone anywhere can securely check an address to see if it is known to be compromised, or otherwise known to be a bad idea to use. This is a key-value-store problem, easily solved by DynamoDB, AWS's KVSaaS.

The point of `baddress` is to populate the DB. It can generate a large quantity of bad addresses automatically, or manually insert known-bad addresses.

This service depends on the AWS SDK, which can populate its keys and secrets from the environment or certain configuration files.

## Accessing the Data

### Low Level

This database is stored in DynamoDB. At a low level, this can be accessed via signed HTTP(S) POST requests. An overview of that API is [here](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Programming.LowLevelAPI.html) and a reference is [here](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/).

### Mid Level

At a medium level, Amazon provides SDKs in multiple languages which take care of much of the gruntwork of correctly forming these requests. At that level, the important data for connections is:

**Region**: `us-east-1`

**Table name**: `bad-addresses`

A positive existence query using the CLI:

```sh
$ aws dynamodb get-item --table-name bad-addresses --key '{"address":{"S":"ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc7"}}'
{
    "Item": {
        "derivation-path": {
            "S": "/44'/20036'/100/5"
        },
        "reason": {
            "S": "derives from 12-word phrase with identical words: hockey"
        },
        "address": {
            "S": "ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc7"
        }
    }
}
```

A negative existence query using the CLI produces no output:

```sh
$ aws dynamodb get-item --table-name bad-addresses --key '{"address":{"S":"ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc8"}}'
```

### High Level

There is a public lambda function which will return appropriate existence information without authentication requirements:

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
