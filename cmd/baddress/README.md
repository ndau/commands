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

Wrapping the appropriate language SDK can produce high-level wrapper functions. This package includes one such:

```sh
$ go run ./cmd/baddress/ check ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc7
ndaa59e7jxegzrjeeyiw3374spk7b93g3x7eb4kkm34tefc7 exists
```

Software in other languges which needs this functionality should leverage the AWS SDK for that language to write an equivalent high-level function.
