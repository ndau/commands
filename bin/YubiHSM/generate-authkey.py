#! /usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

import getpass

from yubihsm import YubiHsm
from yubihsm.defs import CAPABILITY, OBJECT
from yubihsm.objects import AuthenticationKey

default_authkey = 1  # Default key set up after hardware reset
authkeynum = 101  # For consistency with BPC keys, always create key number 101
password = getpass.getpass()
passverify = getpass.getpass("Verify: ")
if password != passverify:
    print("Passwords do not match")
    exit(-1)

# Only used to set up an authentication key for a newly-reset YubiHSM.
# Replaces the default authentication key 1 with the password "password".

hsm = YubiHsm.connect("http://localhost:12345/connector/api")
session = hsm.create_session_derived(1, "password")

capabilities = (
    CAPABILITY.SIGN_EDDSA
    + CAPABILITY.GENERATE_ASYMMETRIC_KEY
    + CAPABILITY.EXPORT_WRAPPED
    + CAPABILITY.IMPORT_WRAPPED
    + CAPABILITY.GET_LOG_ENTRIES
    + CAPABILITY.WRAP_DATA
    + CAPABILITY.UNWRAP_DATA
    + CAPABILITY.DELETE_ASYMMETRIC_KEY
)

# Generate a private key on the YubiHSM for creating signatures:
authkey = AuthenticationKey.put_derived(  # Generate a new key object in the YubiHSM.
    session,  # Secure YubiHsm session to use.
    authkeynum,  # Object ID, 0 to get one assigned.
    "ndau Authentication key",  # Label for the object.
    1,  # Domain(s) for the object.
    capabilities,  # Standard capabilities
    capabilities,  # Delegated capabilities
    password,  # Authentication password
)

print("Created " + str(authkey))

originalkey = session.get_object(1, OBJECT.AUTHENTICATION_KEY)
originalkey.delete()

hsm.close()
exit(0)
