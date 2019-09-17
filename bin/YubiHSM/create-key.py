#! /usr/bin/env python3

import sys
import base64
import getpass

from yubihsm import YubiHsm
from yubihsm.defs import CAPABILITY, ALGORITHM
from yubihsm.objects import AsymmetricKey
from yubihsm import eddsa

if len(sys.argv) != 2:
    print("Usage: python generate-key.py <key_number>")
    exit()

keynum = int(sys.argv[1])
authkeyID = 101
password = getpass.getpass()

hsm = YubiHsm.connect("http://localhost:12345/connector/api")
session = hsm.create_session_derived(authkeyID, password)

key = AsymmetricKey.generate(  # Generate a new key object in the YubiHSM.
    session,  # Secure YubiHsm session to use.
    keynum,  # Object ID
    "ndau key " + str(keynum),  # Label for the object.
    1,  # Domain(s) for the object.
    CAPABILITY.SIGN_EDDSA,  # Capabilities for the ojbect.
    ALGORITHM.EC_ED25519,  # Algorithm for the key.
)

pub_key = eddsa.serialize_ed25519_public_key(key.get_public_key())
print(f"Public key number {keynum}: {base64.standard_b64encode(pub_key).decode()}")

session.close()
hsm.close()
