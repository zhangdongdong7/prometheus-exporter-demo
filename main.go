package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

// InterfaceConfig 定义接口配置结构
type InterfaceConfig struct {
	Name          string        `yaml:"name"`
	URL           string        `yaml:"url"`
	Protocol      string        `yaml:"protocol"`
	CheckInterval time.Duration `yaml:"check_interval,omitempty"`
}

type HealthCollector struct {
	interfaceConfigs []InterfaceConfig
	healthStatus     *prometheus.Desc
}

// loadConfig 从配置文件加载接口配置
func loadConfig(configFile string) ([]InterfaceConfig, error) {
	config := []InterfaceConfig{}

	// 从文件加载配置
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// 解析配置文件
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 设置默认的检测时间间隔为1s
	for i := range config {
		if config[i].CheckInterval == 0 {
			config[i].CheckInterval = time.Second
		}
	}

	return config, nil
}

// NewHealthCollector 创建HealthCollector实例
func NewHealthCollector(configFile string) (*HealthCollector, error) {
	// 从配置文件加载接口配置
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}

	// 初始化HealthCollector
	collector := &HealthCollector{
		interfaceConfigs: config,
		healthStatus: prometheus.NewDesc(
			"interface_health_status",
			"Health status of the interfaces",
			[]string{"name", "url", "protocol"},
			nil),
	}

	return collector, nil
}

// Describe 实现Prometheus Collector接口的Describe方法
func (c *HealthCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.healthStatus
}

// Collect 实现Prometheus Collector接口的Collect方法
func (c *HealthCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	for _, iface := range c.interfaceConfigs {
		wg.Add(1)

		go func(iface InterfaceConfig) {
			defer wg.Done()

			// 检测接口健康状态
			healthy := c.checkInterfaceHealth(iface)

			// 创建Prometheus指标
			var metricValue float64
			if healthy {
				metricValue = 1
			} else {
				metricValue = 0
			}
			ch <- prometheus.MustNewConstMetric(
				c.healthStatus,
				prometheus.GaugeValue,
				metricValue,
				iface.Name,
				iface.URL,
				iface.Protocol,
			)
		}(iface)
	}

	wg.Wait()
}

// checkInterfaceHealth 检测接口健康状态
func (c *HealthCollector) checkInterfaceHealth(iface InterfaceConfig) bool {
	switch iface.Protocol {
	case "http":
		return c.checkHTTPInterfaceHealth(iface)
	case "tcp":
		return c.checkTCPInterfaceHealth(iface)
	default:
		return false
	}
}

// checkHTTPInterfaceHealth 检测HTTP接口健康状态
func (c *HealthCollector) checkHTTPInterfaceHealth(iface InterfaceConfig) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(iface.URL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// checkTCPInterfaceHealth 检测TCP接口健康状态
func (c *HealthCollector) checkTCPInterfaceHealth(iface InterfaceConfig) bool {
	conn, err := net.DialTimeout("tcp", iface.URL, 5*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

func main() {
	// 解析命令行参数
	configFile := flag.String("config", "", "Path to the config file")
	flag.Parse()

	if *configFile == "" {
		// 默认使用当前目录下的config.yaml
		*configFile = "config.yaml"
	}

	// 加载配置文件
	collector, err := NewHealthCollector(*configFile)
	if err != nil {
		fmt.Println("Failed to create collector:", err)
		return
	}

	// 注册HealthCollector
	prometheus.MustRegister(collector)

	// 启动HTTP服务，暴露Prometheus指标
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Failed to start HTTP server:", err)
		os.Exit(1)
	}
}
