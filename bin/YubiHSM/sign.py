#! /usr/bin/env python3

import sys
import base64
import getpass

from yubihsm import YubiHsm
from yubihsm.objects import AsymmetricKey

if len(sys.argv) != 3:
    print("Usage: python sign.py <datafile> <key_number>")
    exit()

infile = open(sys.argv[1], "r")
keynum = int(sys.argv[2])
authkeyID = 101
password = getpass.getpass()

encoded_data = infile.read()
encoded_bytes = base64.standard_b64decode(encoded_data)

hsm = YubiHsm.connect("http://localhost:12345/connector/api")
session = hsm.create_session_derived(authkeyID, password)
key = AsymmetricKey(session, keynum)
sig = key.sign_eddsa(encoded_bytes)

encoded_sig = base64.standard_b64encode(sig).decode()
print(encoded_sig)

session.close()
hsm.close()
