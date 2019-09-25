## Overview 

`kubectl-pvc` 插件，主要用于方便查看集群中 pvc(PersistentVolumeClaim) 资源的当前状态。

## Background

PersistentVolumeClaim API 当前已经是 kubernetes 社区中使用存储资源的最佳方案。但是在使用 PersistentVolumeClaim 过程中，由于 PersistentVolumeClaim 以及其相关资源 PersistentVolume 的管理逻辑非常复杂。牵扯到 kubernetes 中多个组件的协调工作。所以一旦 PersistentVolumeClaim 的使用出现异常（使用 pvc 的 pod 因为 pvc 初始化不成功无法启动），我们很难排查问题。所以推出这个工具，能够协助我们快速确定 PersistentVolumeClaim 出现的问题。

PersistentVolumeClaim/PersistentVolume 资源的生命周期大致分为四个阶段

- Provision: 创建 PersistentVolume 资源，以及这个 PersistentVolume 所采用的存储方案中的对应的存储资源
- Bind: 将最合适的 PerisistentVolume 和用户创建的 PerisistentVolumeClaim 的建立一一对应的关系。
- Attach: 如果有 pod 使用这个 PerisistentVolumeClaim，则把和这个 PerisistentVolumeClaim bind 在一起的 PerisistentVolume 资源和这个 pod 所在的 Node Attach 起来
- Mount: 最终完成这个 PerisistentVolume 在这台机器上的剩余初始化步骤，并且挂载到某个路径下，供后续启动的容器真正使用。

所以针对任何一个 pvc 如果借助我们的命令来展示其状态时，我们也会按照这几个阶段来进行展示

## Common Usage

### 1. 列出某个 persistentVolumeClaim 的当前状态

**列出一个 ceph rbd 的 persistentVolumeClaim 名字叫做 test-rbd，位于 kube-system 这个 namespace 下，被一个叫做 test-pod 所使用**

```
$ kubectl-pvc -n kube-system inspect test-rbd
DESIRED POD                    DESIRED NODE
test-pod                       calico-net2
PHASE       STATUS    DETAIL
Provision   success
Bind        success
Attach      success
Mount       success
```

上面显示的结果表示，这个 pvc 所有的相关初始化操作均已完成，并且使用他的 pod 也开始正常运行了

**列出一个 cephfs 的 persistentVolumeClaim 当前名字叫做 test-cephfs, 位于 default namespace 下面，被三个 pod 所共享**

```
$ kubectl-pvc inspect test-cephfs 
DESIRED POD                       DESIRED NODE
test-deploy-6445845799-c8cgq   	  calico-master3
test-deploy-6445845799-pd8s2      calico-master1
test-deploy-6445845799-29czt      calico-master2
PHASE       STATUS        DETAIL
Provision   success
Bind        success
Attach      success
Mount       partly fail   pods: [test-deploy-6445845799-c8cgq] are still not mounted as desired
```

上图结果表面有 3 个 pod 想要使用这个 pvc，并且目前 Mount 步骤只有一部分 pod 完成了，test-deploy-6445845799-c8cgq 这个 pod 的 mount 操作还没有完成

### 2. 列出某个 namespace 下面的所有 pvc

```
$ kubectl-pvc -n kube-system ls
NAME                                          VOLUME
csi-cephfs-pvc                                pvc-58f38e38-7091-11e9-a38c-6c92bf24e26f
rbd-pvc                                       pvc-dafe629c-708d-11e9-a38c-6c92bf24e26f
```

### 3. 列出某个 pod 使用的所有 pvc

```
$ kubectl-pvc ls -p test-deploy-6445845799-c8cgq
NAME                                          VOLUME
test-cephfs                                   pvc-b05be774-7e26-11e9-bc3e-6c92bf244689
csi-cephfs-pvc                                pvc-58f38e38-7091-11e9-a38c-6c92bf24e26f
```

## Installation

```
$ make 
$ cp _output/kubectl-pvc /usr/bin
```

## TODO

