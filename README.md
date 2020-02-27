# cpaml
**C**opy-**p**asted **a**nd **m**odified strings **l**ookup

This lib may be used to lookup spammers and abusers on social networks
or dating websites. Typical use cases is when spammers use same copy-pasted
messages but change email or phone number after being banned:

_The quick brown fox jumps over the lazy dog then once again runs away and calls 1234567890_

And then:

_The quick brown fox jumps over the lazy dog then once again runs away and calls gmail@gmail.com_

So you want a precise match, allowing variations.

Unlike other approximate text match, this lib "similarity" is exact matched similarity.
40% similarity means 40% of text matched exactly.
So even 10% may be a sign of copy-pasted spam message and you may want to review it.
50% may be used safely to ban/delete/flag message automatically.   


Samples index is kept in memory, so this may not work for large
databases. ~25000 messages samples takes ~300Mb of RAM

Usage example:

```
    spamIndex := cpaml.Init(13)
    spamIndex.AddToSet( "480f89e6fc3ffdfbf7cb2c518ab45f54",
        "The quick brown fox jumps over the lazy dog then once again runs away and calls 1234567890")
    spamIndex.AddToSet( "ab0f8abe6fc3ffdfbf7cb2c518ab4fab",
        "The quick brown fox jumps over the lazy dog then once again runs away and calls gmal@gmail.com")

    id, sim := spamIndex.LookupSimilar("The quick brown fox jumps over the lazy dog then once again runs away and calls gmal@yahoo.com")
    
```

`LookupSimilar(t string)` will return best matched sample ID and calculated "similarity" (0-100)

PS. You may want to call garbage collector on each index sync: `runtime.GC()`