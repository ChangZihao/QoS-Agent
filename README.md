# QoS-Agent
QoS-Agent是QoS-Serverless的单节点，主要负责策略的管理和运行。QoS-Agent包含pqos和perf版本，分别对用master和perf两个分支。perf版本会占用大量count不建议长期使用。

# 接口
 * IP:9001/metrics
    * 获取所监控的pod的IPC，MISSES，MBR， MBL信息
 * ip:9001/start?pod=${podID}
    * 开始监控pod1的所有进程的数据
 * ip:9001/stop?pod=${podID}
    * 停止监控pod2的数据
 * ip:9001/control?pod=${podID}&resourceType=${type}&value=${value}
    * 资源调控接口

