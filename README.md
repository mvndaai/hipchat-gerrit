# hipchat-gerrit

Add reviewers to Gerrit then uses a hipchat webook to notify a group.


For this to work there needs to bash variables `gerritURL`, `gerritUsername`, `gerritPassword`, and `hipchatWebhookURL`.

I recommend adding a `creds.sh` file something like below.

```
#!/bin/bash

export gerritURL=""
export gerritUsername=""
export gerritPassword=""
export hipchatWebhookURL=""
```

Then just run the binary with two arguments
```
./hipchat-gerrit <gerrit review number> <gerrit user name/id/email>
```
