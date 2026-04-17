//go:build linux

package node

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MoeclubM/V2bX/api/panel"
	"golang.org/x/sys/unix"
)

func newMachineStatusFunc() (func() (*panel.MachineStatus, error), error) {
	lastTotal, lastIdle, err := readMachineCPUStat()
	if err != nil {
		return nil, err
	}
	return func() (*panel.MachineStatus, error) {
		total, idle, err := readMachineCPUStat()
		if err != nil {
			return nil, err
		}
		memTotal, memAvailable, swapTotal, swapFree, err := readMachineMemInfo()
		if err != nil {
			return nil, err
		}
		diskTotal, diskUsed, err := readMachineDiskUsage("/")
		if err != nil {
			return nil, err
		}
		cpuUsage := 0.0
		if total > lastTotal {
			deltaTotal := total - lastTotal
			deltaIdle := idle - lastIdle
			if deltaIdle > deltaTotal {
				deltaIdle = deltaTotal
			}
			cpuUsage = float64(deltaTotal-deltaIdle) * 100 / float64(deltaTotal)
		}
		lastTotal = total
		lastIdle = idle
		if memAvailable > memTotal {
			memAvailable = memTotal
		}
		if swapFree > swapTotal {
			swapFree = swapTotal
		}
		return &panel.MachineStatus{
			CPU: cpuUsage,
			Mem: panel.MachineResource{
				Total: memTotal,
				Used:  memTotal - memAvailable,
			},
			Swap: panel.MachineResource{
				Total: swapTotal,
				Used:  swapTotal - swapFree,
			},
			Disk: panel.MachineResource{
				Total: diskTotal,
				Used:  diskUsed,
			},
		}, nil
	}, nil
}

func readMachineCPUStat() (uint64, uint64, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return 0, 0, err
		}
		return 0, 0, fmt.Errorf("read /proc/stat failed")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, fmt.Errorf("unexpected /proc/stat cpu format")
	}
	var total uint64
	for i := 1; i < len(fields); i++ {
		value, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		total += value
	}
	idle, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	if len(fields) > 5 {
		iowait, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		idle += iowait
	}
	return total, idle, nil
}

func readMachineMemInfo() (uint64, uint64, uint64, uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer file.Close()
	values := make(map[string]uint64)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		values[strings.TrimSuffix(fields[0], ":")] = value * 1024
	}
	if err = scanner.Err(); err != nil {
		return 0, 0, 0, 0, err
	}
	memTotal, ok := values["MemTotal"]
	if !ok {
		return 0, 0, 0, 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
	}
	memAvailable, ok := values["MemAvailable"]
	if !ok {
		memAvailable = values["MemFree"] + values["Buffers"] + values["Cached"] + values["SReclaimable"]
		if shmem := values["Shmem"]; shmem > 0 && memAvailable > shmem {
			memAvailable -= shmem
		}
	}
	return memTotal, memAvailable, values["SwapTotal"], values["SwapFree"], nil
}

func readMachineDiskUsage(path string) (uint64, uint64, error) {
	fs := &unix.Statfs_t{}
	if err := unix.Statfs(path, fs); err != nil {
		return 0, 0, err
	}
	total := fs.Blocks * uint64(fs.Bsize)
	free := fs.Bfree * uint64(fs.Bsize)
	if free > total {
		free = total
	}
	return total, total - free, nil
}
