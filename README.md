# kubectl-captain


This is kubectl plugin for [captain](https://github.com/alauda/captain), currently it support two commands:

* `kubectl captain upgrade`: upgrade a helmrequest
* `kubectl captain rollback`: rollback a helmrequest


## install

Download the latest build from the [releases](https://github.com/alauda/kubectl-captain/releases) page, and run

```bash
mv kubectl-captain /usr/local/bin
chmod +x /usr/local/bin/kubectl-captain
```

## Example

1. kubectl upgrade

`kubectl captain upgrade jenkins -n default --set global.images.jenkins.tag=1.6.0 -v 1.6.0`

This command upgrade a HelmRequest named `jenkins` in `default` namespace, set the chart version to `1.5.0`, and set it's image tag to 1.6.0

2. kubectl rollback

`kubectl captain rollback jenkins -n default`

This command rollback a HelmRequest to it's previous settings.
