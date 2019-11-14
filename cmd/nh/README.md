# `nh`: Noms History

`nh` is a tool to inspect the noms history, customized for the ndau use-case.
It allows for searches over higher-level concepts, such as accounts and addresses,
rather than noms' low-level `Value`s.

## Usage

By default, `nh` simply summarizes the database:

```sh
$ go run ./cmd/nh data/noms/
state summary:
    2497 accounts
       5 nodes
```

It is possible to emit data in JSON format:

```sh
$ ./nh data/noms/ --json
{"accounts":2497,"block height":56833,"nodes":5}
```

## Trace an account's history

This finds all deltas in an account's balance and keyset, and the block on which
it was created (if non-0). Note that noms history traversal is slow, so this
may take a while.

```sh
$ time ./nh data/noms/ --json --trace ndags3kdhtauecaek5gzrtwhr6q4jiex8jak3h659ukhphgw
{"balance":1,"block":19506,"prev.balance":99999988612}
{"balance":99999988612,"block":19500,"prev.balance":142053388430}
{"balance":142053388430,"block":19161,"prev.balance":142553888430}
{"balance":142553888430,"block":18297,"prev.balance":142095041050}
{"balance":142095041050,"block":14737,"prev.balance":142006926265}
{"balance":142006926265,"block":14131,"prev.balance":141931350247}
{"balance":141931350247,"block":13552,"prev.balance":141926888054}
{"balance":141926888054,"block":13513,"prev.balance":141629714486}
{"balance":141629714486,"block":11516,"prev.balance":141384215657}
{"balance":141384215657,"block":9785,"prev.balance":141388265657,"prev.validation_keys.0":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","prev.validation_keys.1":"npuba8jadtbbeac7cpavqfv555pi4wbdznse8ta5g4xdz6ehc35fmxaei7zc9j9sgxjcx893wh23","prev.validation_keys.qty":2,"validation_keys.0":"npuba4jaftckeebktqmbve29jktivd3ibzf8uyykh95xc6svicnicfgn73g6g39tgh2fxr5rnaaaaaaubm6ywqrfyd2jv53q5j2qku5iqupbphz8jfuqb3h4hq6uea4j8365agbjkksd","validation_keys.1":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","validation_keys.qty":3}
{"balance":141388265657,"block":9497,"prev.balance":140946456394,"prev.validation_keys.0":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","prev.validation_keys.1":"npuba8jadtbbeac7cpavqfv555pi4wbdznse8ta5g4xdz6ehc35fmxaei7zc9j9sgxjcx893wh23","prev.validation_keys.qty":2,"validation_keys.0":"npuba4jaftckeebktqmbve29jktivd3ibzf8uyykh95xc6svicnicfgn73g6g39tgh2fxr5rnaaaaaaubm6ywqrfyd2jv53q5j2qku5iqupbphz8jfuqb3h4hq6uea4j8365agbjkksd","validation_keys.1":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","validation_keys.qty":3}
{"balance":140946456394,"block":6518,"prev.balance":140946333065,"prev.validation_keys.0":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","prev.validation_keys.1":"npuba8jadtbbeac7cpavqfv555pi4wbdznse8ta5g4xdz6ehc35fmxaei7zc9j9sgxjcx893wh23","prev.validation_keys.qty":2,"validation_keys.0":"npuba4jaftckeebktqmbve29jktivd3ibzf8uyykh95xc6svicnicfgn73g6g39tgh2fxr5rnaaaaaaubm6ywqrfyd2jv53q5j2qku5iqupbphz8jfuqb3h4hq6uea4j8365agbjkksd","validation_keys.1":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","validation_keys.qty":3}
{"balance":140946333065,"block":6506,"prev.balance":140534478026,"prev.validation_keys.0":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","prev.validation_keys.1":"npuba8jadtbbeac7cpavqfv555pi4wbdznse8ta5g4xdz6ehc35fmxaei7zc9j9sgxjcx893wh23","prev.validation_keys.qty":2,"validation_keys.0":"npuba4jaftckeebktqmbve29jktivd3ibzf8uyykh95xc6svicnicfgn73g6g39tgh2fxr5rnaaaaaaubm6ywqrfyd2jv53q5j2qku5iqupbphz8jfuqb3h4hq6uea4j8365agbjkksd","validation_keys.1":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","validation_keys.qty":3}
{"balance":140534478026,"block":3628,"prev.balance":140083047311,"prev.validation_keys.0":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","prev.validation_keys.1":"npuba8jadtbbeac7cpavqfv555pi4wbdznse8ta5g4xdz6ehc35fmxaei7zc9j9sgxjcx893wh23","prev.validation_keys.qty":2,"validation_keys.0":"npuba4jaftckeebktqmbve29jktivd3ibzf8uyykh95xc6svicnicfgn73g6g39tgh2fxr5rnaaaaaaubm6ywqrfyd2jv53q5j2qku5iqupbphz8jfuqb3h4hq6uea4j8365agbjkksd","validation_keys.1":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","validation_keys.qty":3}
{"balance":140083047311,"block":111,"prev.balance":124893559000,"prev.validation_keys.0":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","prev.validation_keys.1":"npuba8jadtbbeac7cpavqfv555pi4wbdznse8ta5g4xdz6ehc35fmxaei7zc9j9sgxjcx893wh23","prev.validation_keys.qty":2,"validation_keys.0":"npuba4jaftckeebktqmbve29jktivd3ibzf8uyykh95xc6svicnicfgn73g6g39tgh2fxr5rnaaaaaaubm6ywqrfyd2jv53q5j2qku5iqupbphz8jfuqb3h4hq6uea4j8365agbjkksd","validation_keys.1":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","validation_keys.qty":3}
{"balance":140083047311,"block":0,"prev.balance":124893559000,"prev.validation_keys.0":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","prev.validation_keys.1":"npuba8jadtbbeac7cpavqfv555pi4wbdznse8ta5g4xdz6ehc35fmxaei7zc9j9sgxjcx893wh23","prev.validation_keys.qty":2,"validation_keys.0":"npuba4jaftckeebktqmbve29jktivd3ibzf8uyykh95xc6svicnicfgn73g6g39tgh2fxr5rnaaaaaaubm6ywqrfyd2jv53q5j2qku5iqupbphz8jfuqb3h4hq6uea4j8365agbjkksd","validation_keys.1":"npuba8jadtbbebp9pk68u3yx69ysrgaj6jcevbkwfqvduwakba8tnybtnww78jqighbyq9tgsmqg","validation_keys.qty":3}

real	1m30.469s
user	1m38.381s
sys	0m0.785s
```
