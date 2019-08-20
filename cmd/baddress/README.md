# `baddress`: populate a DynamoDB with bad addresses

We want a service where anyone anywhere can securely check an address to see if it is known to be compromised, or otherwise known to be a bad idea to use. This is a key-value-store problem, easily solved by DynamoDB, AWS's KVSaaS.

The point of `baddress` is to populate the DB. It can generate a large quantity of bad addresses automatically, or manually insert known-bad addresses.

This service depends on the AWS SDK, which can populate its keys and secrets from the environment or certain configuration files.
