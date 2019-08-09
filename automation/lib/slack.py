#!/usr/bin/env python3

import os
import requests


def post_to_slack(message):
    """
    Post the given message to the Oneiro #deploys Slack channel.
    Must have SLACK_KEY environment variable set.
    """

    slack_key_name = "SLACK_DEPLOYS_KEY"
    try:
        slack_key_value = os.environ[slack_key_name]
    except:
        slack_key_value = ""
    if len(slack_key_value) == 0:
        print(f"Unable to post to slack without {slack_key_name} env var: '{message}'")
        return

    url = f"https://hooks.slack.com/services/{slack_key_value}"
    body = {"text": message}
    r = requests.post(url, json=body)
    if r.status_code == 200 and r.content.decode("utf-8") == "ok":
        print(f"Posted to slack: '{message}'")
    else:
        print(
            f"Got {r.status_code} when posting to slack because {r.reason}: '{message}'"
        )
