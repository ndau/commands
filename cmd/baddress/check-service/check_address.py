import json

import boto3
from botocore.exceptions import ClientError

REGION = "us-east-1"
TABLE = "bad-addresses"

dynamodb = boto3.resource("dynamodb", region_name=REGION)
table = dynamodb.Table(TABLE)


def resp(code, body):
    return {"isBase64Encoded": False, "statusCode": code, "body": json.dumps(body)}


def lambda_handler(event, context):
    try:
        addr = json.loads(event["body"])["address"]
    except KeyError:
        return resp(400, {"err": 'key "address" not found in request body'})

    try:
        response = table.get_item(Key={"address": addr})
    except ClientError as e:
        return resp(500, {"err": e.response["Error"]["Message"]})

    body = {"exists": "Item" in response}
    if body["exists"]:
        body["data"] = response["Item"]
    return resp(200, body)
