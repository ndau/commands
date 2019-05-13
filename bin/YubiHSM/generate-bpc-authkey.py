#! /usr/bin/env python3

import base64
import getpass

from yubihsm import YubiHsm
from yubihsm.defs import CAPABILITY, ALGORITHM, OBJECT
from yubihsm.objects import AuthenticationKey, AsymmetricKey
from yubihsm import eddsa

default_authkey = 1  # Default key set up after hardware reset

# Only used to set up an authentication key for a newly-reset YubiHSM. Replaces the
# default authentication key 1 (with the password "password") and creates an
# eddsa signing key.

hsm = YubiHsm.connect("http://localhost:12345/connector/api")
session = hsm.create_session_derived(default_authkey, "password")

authkeynum = 101  # For consistency, always create key number 101

authpass = getpass.getpass("New authentication key password:")
authpassverify = getpass.getpass("Verify authentication key password: ")
if authpass != authpassverify:
    print("Passwords do not match")
    exit(-1)

capabilities = (
    CAPABILITY.SIGN_EDDSA
    + CAPABILITY.GENERATE_ASYMMETRIC_KEY
    + CAPABILITY.EXPORT_WRAPPED
    + CAPABILITY.IMPORT_WRAPPED
    + CAPABILITY.GET_LOG_ENTRIES
    + CAPABILITY.WRAP_DATA
    + CAPABILITY.UNWRAP_DATA
    + CAPABILITY.DELETE_ASYMMETRIC_KEY
    + CAPABILITY.DELETE_ASYMMETRIC_KEY
    + CAPABILITY.DELETE_AUTHENTICATION_KEY
)

# Generate a new authentication key on the YubiHSM for creating signatures

authkey = AuthenticationKey.put_derived(  # Generate a new key object in the YubiHSM.
    session,  # Secure YubiHsm session to use.
    authkeynum,  # Object ID, 0 to get one assigned.
    "ndau BPC authentication key",  # Label for the object.
    1,  # Domain(s) for the object.
    capabilities,  # Standard capabilities
    capabilities,  # Delegated capabilities
    authpass,  # Authentication password
)

print("Created " + str(authkey))
session.close()

# Log in with new authentication key, create EDDSA asymmetric key, delete
# default authentication key

signkeynum = 1001  # ID used for first BPC signing key
session = hsm.create_session_derived(authkeynum, authpass)
signkey = AsymmetricKey.generate(
    session,
    signkeynum,
    "ndau BPC signing key",
    1,
    CAPABILITY.SIGN_EDDSA,
    ALGORITHM.EC_ED25519,
)

print("Created " + str(signkey))
print(
    "Public key: "
    + str(
        base64.b64encode(eddsa.serialize_ed25519_public_key(signkey.get_public_key()))
    )
)

# Delete the default authentication key

session.get_object(default_authkey, OBJECT.AUTHENTICATION_KEY).delete()
session.close()
hsm.close()
exit(0)
