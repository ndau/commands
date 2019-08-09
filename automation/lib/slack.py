#!/usr/bin/env python3

import os
import requests


def post_to_slack(message):
    """
    Post the given message to the Oneiro #deploys Slack channel.
    Must have SLACK_KEY environment variable set.
    """

    try:
        slack_key = os.environ["SLACK_KEY"]
    except:
        slack_key = ""
    if len(slack_key) == 0:
        print(f"Unable to post to slack without SLACK_KEY env var: '{message}'")
        return

    url = f"https://hooks.slack.com/services/{slack_key}"
    body = {"text": message}
    r = requests.post(url, json=body)
    if r.status_code == 200:
        print(f"Posted to slack: '{message}'")
    else:
        print(
            f"Got {r.status_code} when posting to slack because {r.reason}: '{message}'"
        )
