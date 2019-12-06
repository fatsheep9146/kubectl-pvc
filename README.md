# kubectl-captain


This is kubectl plugin for [captain](https://github.com/alauda/captain), currently it support the following commands:

* `kubectl captain upgrade`: upgrade a helmrequest
* `kubectl captain rollback`: rollback a helmrequest
* `kubectl captain import`: import a helmrelease to captain
* `kubectl captain create`: create a helmrequest
* `kubectl captain create-repo`: create a chartrepo


## Install

Download the latest build from the [releases](https://github.com/alauda/kubectl-captain/releases) page, and run

```bash
mv kubectl-captain /usr/local/bin
chmod +x /usr/local/bin/kubectl-captain
```

## Example

1. kubectl captain upgrade

`kubectl captain upgrade jenkins -n default --set global.images.jenkins.tag=1.6.0 -v 1.6.0`

This command upgrade a HelmRequest named `jenkins` in `default` namespace, set the chart version to `1.6.0`, and set it's image tag to 1.6.0

2. kubectl captain rollback

`kubectl captain rollback jenkins -n default`

This command rollback a HelmRequest to it's previous settings.

3. kubectl captain import 

`kubectl captain import wordpress -n default --repo=stable --repo-namespace=captain`

This command import an existing helm v2 release named `wordpress`, who's chart is belongs to a repo named `stable`. this command will try to 
create a ChartRepo resource for this repo first in the `captain` namespace if it not exist, afterwards it will create a HelmRequest resource
named `wordpress` in the `default` namespace. Captain will do the sync stuff. 

4. kubectl captain create-repo

`kubectl captain create-repo test-repo --url=https://alauda.github.io/captain-test-charts/ -n captain -w --timeout=30`

This command create a ChartRepo named `test-repo` in the `captain` namespace.


5. kubectl captain create

`kubectl captain create test-nginx --chart=stable/nginx-ingress --version=1.26.2 --set=a=b -w --timeout=30`

This command create a HelmRequest named `test-nginx`, using chart `stable/nginx-ingress`, and version `1.26.2`, and some values.

