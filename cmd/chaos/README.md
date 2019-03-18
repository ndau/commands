# `chaos` Tool

## Overview

The chaos tool can be used for manipulating the chaos blockchain (and sometimes affecting the ndau blockchain with sidechain transactions).

## System Variables

Here are some examples for how to use the chaos tool to set system variables once you have the `bpc-operations` ndau account and the `sysvar` chaos identity set up.

For large system variables like `svi`, you will likely want to get the current value first, then manipulate it, then set it back into the blockchain.  It's also useful to get a system variable after it's been changed, to see that it wound up how you expected it.  This is the general way to get a system variable in a human-readable format:

```
./chaos get sysvar <variable> -m
```

Replace `<variable>` with the system variable you want to get.  For example, `TransactionFeeScript`.  You can pipe the results to `jq`, for example, if you'd like to sort the keys of the output json:

```
./chaos get sysvar <variable> -m | jq . -S
```

Here's how to set a system variable, given the (possibly modified) json from the `get` command:

```
./chaos set sysvar <variable> --value-json <value_json> --value-json-types <hints_json>
```

Here's a fairly complex example of how one would go about defining a new system variable in the `svi` table, and then setting the system variable itself:

```
# Set NDAUHOME depending on which network you want to affect.
export NDAUHOME=$HOME/.localnet/data/ndau-0

# Use the sysvar namespace that was present in the genesis.toml file.
NS_B64=A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR

# Which sysvar to add and then set.
SYSVAR=AccountAttributes

# Convert it to base64.
SYSVAR_B64=$(printf "%s" $SYSVAR | base64)

# Get the current svi table and modify it by adding the new sysvar to it.
SVI=$(./chaos get sysvar svi -m | jq -S -c --arg a $NS_B64 --arg b $SYSVAR --arg c $SYSVAR_B64 '. + {($b):{"Current":[$a,$c],"Future":[$a,$c],"ChangeOn":0}}')

# Special step for the svi table sysvar only.  Each sysvar has its own rules.
HINTS='{"ChangeOn":["uint64"]}'

# Set the svi sysvar back into the blockchain, now containing the new sysvar.
./chaos set sysvar svi --value-json $SVI --value-json-types $HINTS

# See that it was updated.
./chaos get sysvar svi -m | jq . -S

# Prepare the value of the new sysvar.
ATTRIBUTES='{"ndaegwggj8qv7tqccvz6ffrthkbnmencp9t2y4mn89gdq3yk":{"x":{}}}'

# Set the new sysvar itself.  No special type hints are needed for this sysvar.
./chaos set sysvar AccountAttributes --value-json $ATTRIBUTES
```

Below are specific examples for individual system variables.  There are a variety of different json formats it expects, as well as custom "json type hints" sometimes needed for each.

### AccountAttributes

```
VALUE='{"ndaegwggj8qv7tqccvz6ffrthkbnmencp9t2y4mn89gdq3yk":{"x":{}}}'
./chaos set sysvar AccountAttributes --value-json $VALUE
```

### CommandValidatorChangeAddress

```
VALUE='["ndnf9ffbzhyf8mk7z5vvqc4quzz5i2exp5zgsmhyhc9cuwr4"]'
./chaos set sysvar CommandValidatorChangeAddress --value-json $VALUE
```

### DefaultSettlementDuration

```
VALUE=172800000000
./chaos set sysvar DefaultSettlementDuration --value-json $VALUE
```

### EAIFeeTable

```
VALUE='[{"Fee":4000000,"To":["ndaea8w9gz84ncxrytepzxgkg9ymi4k7c9p427i6b57xw3r4"]},{"Fee":1000000,"To":["ndmmw2cwhhgcgk9edp5tiieqab3pq7uxdic2wabzx49twwxh"]},{"Fee":100000,"To":["ndakj49v6nnbdq3yhnf8f2j6ivfzicedvfwtunckivfsw9qt"]},{"Fee":100000,"To":["ndnf9ffbzhyf8mk7z5vvqc4quzz5i2exp5zgsmhyhc9cuwr4"]},{"Fee":9800000,"To":null}]'
./chaos set sysvar EAIFeeTable --value-json $VALUE
```

### LockedRateTable

```
VALUE='[[7776000000000,10000000000],[15552000000000,20000000000],[31536000000000,30000000000],[63072000000000,40000000000],[94608000000000,50000000000]]'
HINTS='{"": ["int64", "uint64"]}'
./chaos set sysvar LockedRateTable --value-json $VALUE --value-json-types $HINTS
```

### MinDurationBetweenNodeRewardNominations

```
VALUE=86400000000
./chaos set sysvar MinDurationBetweenNodeRewardNominations --value-json $VALUE
```

### MinNodeRegistrationStakeAmount

```
VALUE=100000000000
./chaos set sysvar MinNodeRegistrationStakeAmount --value-json $VALUE
```

### NodeGoodnessFunction

```
VALUE='"oACI"'
./chaos set sysvar NodeGoodnessFunction --value-json $VALUE
```

### NodeRewardNominationTimeout

```
VALUE=30000000
./chaos set sysvar NodeRewardNominationTimeout --value-json $VALUE
```

### NominateNodeRewardAddress

```
VALUE='["ndnf9ffbzhyf8mk7z5vvqc4quzz5i2exp5zgsmhyhc9cuwr4"]'
./chaos set sysvar NominateNodeRewardAddress --value-json $VALUE
```

### ReleaseFromEndowmentAddress

```
VALUE='["ndmfgnz9qby6nyi35aadjt9nasjqxqyd4vrswucwfmceqs3y"]'
./chaos set sysvar ReleaseFromEndowmentAddress --value-json $VALUE
```

### TransactionFeeScript

```
VALUE='"oAAgiA=="'
./chaos set sysvar TransactionFeeScript --value-json $VALUE
```

### UnlockedRateTable

```
VALUE='[[2592000000000,20000000000],[5184000000000,30000000000],[7776000000000,40000000000],[10368000000000,50000000000],[12960000000000,60000000000],[15552000000000,70000000000],[18144000000000,80000000000],[20736000000000,90000000000],[23328000000000,100000000000]]'
HINTS='{"": ["int64", "uint64"]}'
./chaos set sysvar UnlockedRateTable --value-json $VALUE --value-json-types $HINTS
```

### svi

```
VALUE='{"CommandValidatorChangeAddress":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Q29tbWFuZFZhbGlkYXRvckNoYW5nZUFkZHJlc3M="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Q29tbWFuZFZhbGlkYXRvckNoYW5nZUFkZHJlc3M="]},"DefaultSettlementDuration":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RGVmYXVsdFNldHRsZW1lbnREdXJhdGlvbg=="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RGVmYXVsdFNldHRsZW1lbnREdXJhdGlvbg=="]},"EAIFeeTable":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RUFJRmVlVGFibGU="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RUFJRmVlVGFibGU="]},"LockedRateTable":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TG9ja2VkUmF0ZVRhYmxl"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TG9ja2VkUmF0ZVRhYmxl"]},"MinDurationBetweenNodeRewardNominations":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluRHVyYXRpb25CZXR3ZWVuTm9kZVJld2FyZE5vbWluYXRpb25z"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluRHVyYXRpb25CZXR3ZWVuTm9kZVJld2FyZE5vbWluYXRpb25z"]},"MinNodeRegistrationStakeAmount":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluTm9kZVJlZ2lzdHJhdGlvblN0YWtlQW1vdW50"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluTm9kZVJlZ2lzdHJhdGlvblN0YWtlQW1vdW50"]},"NodeGoodnessFunction":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZUdvb2RuZXNzRnVuY3Rpb24="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZUdvb2RuZXNzRnVuY3Rpb24="]},"NodeRewardNominationTimeout":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZVJld2FyZE5vbWluYXRpb25UaW1lb3V0"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZVJld2FyZE5vbWluYXRpb25UaW1lb3V0"]},"NominateNodeRewardAddress":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9taW5hdGVOb2RlUmV3YXJkQWRkcmVzcw=="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9taW5hdGVOb2RlUmV3YXJkQWRkcmVzcw=="]},"ReleaseFromEndowmentAddress":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","UmVsZWFzZUZyb21FbmRvd21lbnRBZGRyZXNz"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","UmVsZWFzZUZyb21FbmRvd21lbnRBZGRyZXNz"]},"TransactionFeeScript":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VHJhbnNhY3Rpb25GZWVTY3JpcHQ="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VHJhbnNhY3Rpb25GZWVTY3JpcHQ="]},"UnlockedRateTable":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VW5sb2NrZWRSYXRlVGFibGU="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VW5sb2NrZWRSYXRlVGFibGU="]}}'
HINTS='{"ChangeOn":["uint64"]}'
./chaos set sysvar svi --value-json $VALUE --value-json-types $HINTS
```
