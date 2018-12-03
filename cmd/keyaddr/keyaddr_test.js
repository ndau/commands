const address = require('./keyaddr.js');

mk = address.NewPrivateMaster('asdfwdsdfdewidkdkffjfjfuggujfjug');
console.log(mk);
mk
  .then((k) => {
    console.log(k);
    pubkey = k.Public();
    console.log(pubkey);
  })
  .catch((e) => {
    console.log('master key error: ', e);
  });

mk
  .then((k) => {
    msg = 'CAFEF00DBAAD1DEA';
    k
      .Sign(msg)
      .then((sig) => {
        console.log(sig);
        pubkey = k.Public();
        sig
          .Verify(msg, k.key)
          .then((b) => {
            console.log('Verified: ', b);
          })
          .catch((e) => {
            console.log('NO verify: ', e);
          });
      })
      .catch((e) => console.log('error: ', e));
  })
  .catch((e) => console.log('error: ', e));

theBytes = '\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F';
theWords = [
  'abandon',
  'amount',
  'liar',
  'amount',
  'expire',
  'adjust',
  'cage',
  'candy',
  'arch',
  'gather',
  'drum',
  'bundle'
];

wordlist = address
  .wordsFromBytes('en', theBytes)
  .then((l) => {
    console.log(l);
    err = false;
    for (i = 0; i < theWords.length; i++) {
      if (l[i] != theWords[i]) {
        err = true;
      }
    }
    if (err) {
      console.log("wordsFromBytes didn't return expected result");
    }
  })
  .catch((e) => console.log('error: ', e));

bytes = address
  .bytesFromWords('en', theWords)
  .then((b) => {
    if (b != theBytes) {
      console.log("bytesFromWords didn't return the expected result");
    }
  })
  .catch((e) => console.log('error: ', e));
