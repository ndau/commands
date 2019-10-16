#! /usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

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
