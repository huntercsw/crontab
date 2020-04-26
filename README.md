# 整体架构



- 无状态的Master集群将任务保存到ETCD，所有Worker从ETCD同步任务
- Master从ETCD中查询任务
- Worker将日志保存到MongoDB，Master从MongoDB查询日志

![1587532180500](C:\Users\Inno\AppData\Roaming\Typora\typora-user-images\1587532180500.png)

- 利用ETCD同步全量任务列表到所有Worker节点
- 每个worker独立调度全量任务，无需与Master产生直接的RPC
- 每个worker利用分布式锁抢占，解决并发调度相同任务的问题



# Master

- 任务管理接口（http）
  - 增、删、改、查
- 任务日志接口（http）
  - 查看任务历史日志
- 任务控制接口（http）
  - 强制接收任务



# Worker

- 同步任务，监听ETCD中/cron/jobs/目录的变化
  - 监听协程：
    - 监听/cron/jobs/和/cron/job/control/cancel/目录的变化
    - 将变化事件推送给调度协程，更新内存中的任务信息
- 任务调度，基于CRON表达式，触发过期的任务
  - 调度协程：
    - 监听任务变更事件，更新内存中维护的内存列表
    - 检查任务的CRON表达式，扫描到期任务，交给执行协程
    - 监听任务控制事件，强制中断正在执行的子进程
    - 监听任务执行结果，更新内存中的任务执行状态，记录任务日志
- 任务执行，并发执行多任务，分布式锁实现任务强制
  - 执行协程：
    - 在ETCD中抢占分布式乐观锁
    - 抢占成功，通过Command类执行Shell任务
    - 捕获Command输出，等待子进程结束，将执行结果投递给调度协程
- 日志保存，捕获任务输出，保存到MongoDB
  - 日志协程：
    - 监听调度协程发来的日志，放到一个batch中
    - 对新的batch启动一个定时器，超时自动提交
    - 如果在定时内，batch满了，自动提交，取消提交定时器



# ETCD结构：

- ```
  key: /cron/jobs/jobName
  calue: "{name: "", command: "", cronExpress: "", status: false, hostName: ""}"
  ```

- ```
  key: /cron/job/cancel/jobName
  calue: ""
  ```

- ```
  /cron/job/lock/jobName
  ```


# MongoDB结构：

- ```json
  {
  	jobName: "",
  	command: "",
  	stdOut: "",
  	stdErr: "",
  	startTime: ,
  	endTime: ,
  }
  ```
