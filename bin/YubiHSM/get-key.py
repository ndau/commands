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
from yubihsm import eddsa

if len(sys.argv) < 3:
    print("Usage: python get-key.py <start_key_number> <end_key_number>")
    exit()

connectorURL = "http://localhost:12345/connector/api"

authkeyID = 101
firstkey = int(sys.argv[1])
lastkey = int(sys.argv[2])
password = getpass.getpass()

hsm = YubiHsm.connect(connectorURL)
session = hsm.create_session_derived(authkeyID, password)

for keynum in range(firstkey, lastkey + 1):
    key = AsymmetricKey(session, keynum)
    pub_key = eddsa.serialize_ed25519_public_key(key.get_public_key())
    print(base64.standard_b64encode(pub_key).decode())

session.close()
hsm.close()
