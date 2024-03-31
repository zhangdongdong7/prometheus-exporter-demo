> 想必大家对于黑盒监控都不陌生，我们经常使用blackbox_exporter来进行黑盒监控，在K8s中进行黑盒监控可以参考这里。

既然已经有成熟的工具，为何自己还要再来尝试开发一个？

我说是为了学习，你信吗？

既然是为了学习，整体逻辑就不用太复杂，主要需要实现以下功能：

* 可以通过配置文件的方式增加监控项
* 吐出Prometheus可收集指标
* 支持tcp和http探测
* 支持配置检测频率
## 写在前面
在正式开始之前，先简单介绍一下Prometheus以及Prometheus Exporter。

Prometheus是CNCF的一个开源监控工具，是近几年非常受欢迎的开源项目之一。在云原生场景下，经常使用它来进行指标监控。

Prometheus支持4种指标类型：

Counter（计数器）：只增不减的指标，比如请求数，每来一个请求，该指标就会加1。
Gauge（仪表盘）：动态变化的指标，比如CPU，可以看到它的上下波动。
Histogram（直方图）：数据样本分布情况的指标，它将数据按Bucket进行划分，并计算每个Bucket内的样本的一些统计信息，比如样本总量、平均值等。
Summary（摘要）：类似于Histogram，也用于表示数据样本的分布情况，但同时展示更多的统计信息，如样本数量、总和、平均值、上分位数、下分位数等。
在实际使用中，常常会将这些指标组合起来使用，以便能更好的观测系统的运行状态和性能指标。

这些指标从何而来？

Prometheus Exporter就是用来收集和暴露指标的工具，通常情况下是Prometheus Exporter收集并暴露指标，然后Prometheus收集并存储指标，使用Grafana或者Promethues UI可以查询并展示指标。

Prometheus Exporter主要包含两个重要的组件：

Collector：收集应用或者其他系统的指标，然后将其转化为Prometheus可识别收集的指标。
Exporter：它会从Collector获取指标数据，并将其转成为Prometheus可读格式。
那Prometheus Exporter是如何生成Prometheus所支持的4种类型指标（Counter、Gauge、Histogram、Summary）的呢？

Prometheus提供了客户端包github.com/prometheus/client_golang，通过它可以声明不通类型的指标

通过上面的介绍，对于怎么创建一个Prometheus Exporter是不是有了初步的了解？主要可分为下面几步

定义一个Exporter结构体，用于存放描述信息
实现Collector接口
实例化exporter
注册指标
暴露指标

## 现在开始
有了一定的基本知识后，我们开始开发自己的Exporter。

我们再来回顾一下需要实现的功能：

* 可以通过配置文件的方式增加监控项
* 吐出Prometheus可收集指标
* 支持tcp和http探测
* 支持配置检测频率

（1）我们的采集对象是通过配置文件加载的，所以我们可以先确定配置文件的格式，我希望的是如下格式：
```
- url: "http://www.baidu.com"  
  name: "百度测试"  
  protocol: "http"
  check_interval: 2s
- url: "localhost:2222"  
  name: "本地接口2222检测"  
  protocol: "tcp"
```

（2）定义接口探测的Collector接口，实现Promethues Collector接口
（3）实现Prometheus Collector接口的Describe和Collect方法
（4）实现http和tcp检测方法
（5）创建main方法，完成开发