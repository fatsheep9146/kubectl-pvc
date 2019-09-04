# kubectl-captain


This is kubectl plugin for [captain](https://github.com/alauda/captain), currently it support two commands:

* `kubectl captain update`: update a helmrequest
* `kubectl captain rollback`: rollback a helmrequest


## install

Download the latest build from the [releases](https://github.com/alauda/kubectl-captain/releases) page, and run

```bash
mv kubectl-captain /usr/local/bin
chmod +x /usr/local/bin/kubectl-captain
```