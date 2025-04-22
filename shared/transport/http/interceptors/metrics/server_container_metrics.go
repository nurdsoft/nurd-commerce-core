package metrics

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"
)

var (
	cGroupMemLimitPaths = []string{
		"/sys/fs/cgroup/memory/memory.limit_in_bytes",
		"/sys/fs/cgroup/memory.max",
	}

	cGroupMemUsagePaths = []string{
		"/sys/fs/cgroup/memory/memory.usage_in_bytes",
		"/sys/fs/cgroup/memory.current",
	}
)

func readCgroupFile(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	str := strings.TrimSpace(string(data))
	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func getMemoryLimit() (uint64, error) {
	for _, path := range cGroupMemLimitPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			// Read memory limit
			memLimit, err := readCgroupFile(path)
			if err != nil {
				return 0, fmt.Errorf("error reading memory limit: %v", err)
			}
			return memLimit, nil
		}
	}
	return 0, fmt.Errorf("no suitable memory file found for memory limit")
}

func getMemoryUsage() (uint64, error) {
	for _, path := range cGroupMemUsagePaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			//Read memory usage
			memUsage, err := readCgroupFile(path)
			if err != nil {
				return 0, fmt.Errorf("error reading memory usage: %v", err)
			}
			return memUsage, nil
		}
	}

	return 0, fmt.Errorf("no suitable memory file found for memory usage")
}

func containerMetricsReporter(m *ServerMetrics) {
	go func() {
		for {

			memLimit, err := getMemoryLimit()
			if err != nil {
				log.Error(err)
				return
			}

			memUsage, err := getMemoryUsage()
			if err != nil {
				log.Error(err)
				return
			}

			if memLimit != 0 {
				m.serverContainerMemTotalGauge.Set(float64(memLimit / 1024 / 1024))
				if memUsage != 0 {
					m.serverContainerMemUsedGauge.Set(float64(memUsage / 1024 / 1024))
					m.serverContainerMemPercentageUsedGauge.Set(float64(memUsage) / float64(memLimit) * 100)
				}
			}

			// Sleep for a while before collecting stats again
			time.Sleep(2 * time.Second)
		}
	}()
}
