[![Go Report Card](https://goreportcard.com/badge/kubernetes-sigs/scheduler-plugins)](https://goreportcard.com/report/kubernetes-sigs/scheduler-plugins) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/LICENSE)

# Scheduler Plugins

基于[scheduler framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/)的调度程序插件仓库。
本项目定义了一个调度器annotation-scheduler,实现filter plugin接口。在Filter plugin接口中判断集群的node节点annotation/annotationmatch annotation是否为boo，若是，则优先调度到该节点。

## 安装

容器镜像可以在[docker hub](https://hub.docker.com/repositories/irealuna)中获得。
有两个镜像，一个用于kube调度器，另一个用于控制器。

```shell
docker pull irealuna/scheduler-plugins-kube-scheduler:$TAG
docker pull irealuna/scheduler-plugins-controller:$TAG
```
### 添加调度器
编写plugin的配置文件scheduler-config.yaml
```shell
cp ./scheduler-config.yaml /etc/kubernetes/scheduler-config.yaml
```
### 配置调度器
备份文件 kube-scheduler.yaml
```shell
cp /etc/kubernetes/manifests/kube-scheduler.yaml /etc/kubernetes/kube-scheduler.yaml
```
修改`/etc/kubernetes/manifests/kube-scheduler.yaml`文件来运行scheduler-plugins
配置文件可参考本项目中./kube-scheduler.yaml
```shell
cp ./kube-scheduler.yaml /etc/kubernetes/manifests/kube-scheduler.yaml
```
这个配置文件较原文件改变了这些地方:
1. kube-scheder启动参数中添加`--config`参数指定`kube-scheduler.yaml`
2. 删除重复的CLI参数(例如 `——leader-elect`)，因为它们可能已经在配置文件中定义
3. 更改`spec.containers.image`为我们的新创建的image
4. 添加hostPath volumes将kube-scheduler.yaml映射到容器中，以便在调度器启动时可读

```shell
<     - --config=/etc/kubernetes/scheduler-config.yaml
---
<     image: registry.k8s.io/scheduler-plugins/kube-scheduler:v0.25.7
>     image: registry.k8s.io/kube-scheduler:v1.25.7
---
<     - mountPath: /etc/kubernetes/scheduler.conf
<       name: kubeconfig
<       readOnly: true
<     - mountPath: /etc/kubernetes/scheduler-config.yaml
<       name: scheduler-config
<       readOnly: true
<     - mountPath: /etc/localtime
<       name: localtime
---
<     - hostPath:
<       path: /etc/kubernetes/scheduler.conf
<         type: FileOrCreate
<         name: kubeconfig
<     - hostPath:
<         path: /etc/kubernetes/scheduler-config.yaml
<         type: FileOrCreate
<       name: scheduler-config
<     - hostPath:
<         path: /etc/localtime
<         type: ""
<       name: localtime
```
修改在/etc/kubernetes/manifests 目录下的kube-scheduler.yaml，kubelet会重新启动kube-scheduler
## 测试
1.给node1添加测试的annotation:
```shell
kubectl annotate node k8s-node1 annotation/annotationmatch=boo
```
2.创建一个job `job-annotation-not-match.yaml`
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pod-annotationnomatch
spec:
  parallelism: 1
  template:
    metadata:
      annotations:
         annotation/annotationmatch: booxx   
    spec:
      schedulerName: annotation-scheduler
      containers:
      - name: pod-state
        image: busybox
        command: ["sh", "-c", "sleep 10"]
      restartPolicy: Never
  backoffLimit: 4
```
3.创建一个job `job-annotation-match.yaml`
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pod-annotationmatch
spec:
  parallelism: 1
  template:
    metadata:
      annotations:
         annotation/annotationmatch: boo   
    spec:
      schedulerName: annotation-scheduler
      containers:
      - name: pod-state
        image: busybox
        command: ["sh", "-c", "sleep 10"]
      restartPolicy: Never
  backoffLimit: 4
```
4.观察其调度状态
```shell
[root@k8s-master0 as-a-second-scheduler]# kubectl get po -owide
NAME                          READY   STATUS      RESTARTS   AGE   IP             NODE        NOMINATED NODE   READINESS GATES
pod-annotationmatch-ksqkm     0/1     Completed   0          73m   172.16.36.87   k8s-node1   <none>           <none>
pod-annotationnomatch-dspk9   0/1     Pending     0          73m   <none>         <none>      <none>           <none>
```
可见带有annotation/annotationmatch: boo的job被调度到了预期的node1上
