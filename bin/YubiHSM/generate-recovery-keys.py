#! /usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

import base64
import getpass

from yubihsm import YubiHsm
from yubihsm.defs import CAPABILITY, ALGORITHM
from yubihsm.objects import AsymmetricKey
from yubihsm import eddsa

authkeyID = 101  # For consistency, always create key number 101
password = getpass.getpass()

hsm = YubiHsm.connect("http://localhost:12345/connector/api")
session = hsm.create_session_derived(authkeyID, password)

# Generate recovery service keypairs on the YubiHSM for creating signatures. Always
# generate two sets of 64 keys, numbered 5001 through 5064 and 6001 through 6064.

for keynum in range(10001, 10065):
    key = AsymmetricKey.generate(  # Generate a new key object in the YubiHSM.
        session,  # Secure YubiHsm session to use.
        keynum,  # Object ID
        "ndau key " + str(keynum),  # Label for the object.
        1,  # Domain(s) for the object.
        CAPABILITY.SIGN_EDDSA,  # Capabilities for the ojbect.
        ALGORITHM.EC_ED25519,  # Algorithm for the key.
    )
    pub_key = eddsa.serialize_ed25519_public_key(key.get_public_key())
    print(
        "Public key number "
        + str(keynum)
        + ": "
        + base64.standard_b64encode(pub_key).decode()
    )

session.close()
hsm.close()
