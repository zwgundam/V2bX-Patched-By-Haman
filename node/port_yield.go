package node

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

type YieldedService struct {
	Type string // "systemd" or "docker" or "pid"
	Name string
}

// 物理检测 80 端口占用者并自动让出/恢复
func yieldPort80() (*YieldedService, error) {
	// 使用 ss 查 80 端口占用 PID 和进程名
	cmd := exec.Command("ss", "-tulpn")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, nil // ss 命令不可用或未占用
	}

	lines := strings.Split(out.String(), "\n")
	var pid string
	var procName string

	for _, line := range lines {
		if strings.Contains(line, ":80 ") || strings.Contains(line, ":80\t") {
			// 解析 pid, 例如 users:(("nginx",pid=1234,fd=6))
			if idx := strings.Index(line, "users:(("); idx != -1 {
				sub := line[idx:]
				parts := strings.Split(sub, "\"")
				if len(parts) >= 2 {
					procName = parts[1]
				}
				if pidIdx := strings.Index(sub, "pid="); pidIdx != -1 {
					pidSub := sub[pidIdx+4:]
					pidParts := strings.Split(pidSub, ",")
					if len(pidParts) > 0 {
						pid = pidParts[0]
					}
				}
			}
			break
		}
	}

	if procName == "" && pid == "" {
		return nil, nil // 80 端口未被占用，正常直接申请
	}

	log.Infof("检测到端口 80 当前被进程 '%s' (PID: %s) 占用！启动智能让出机制...", procName, pid)

	// 优先检测是否为 systemd 服务 (如 nginx, caddy, apache2)
	if procName != "" {
		checkSvc := exec.Command("systemctl", "is-active", procName)
		if checkSvc.Run() == nil {
			log.Infof("正在临时挂起 systemd 服务 '%s' 以借用 80 端口申请证书...", procName)
			stopCmd := exec.Command("systemctl", "stop", procName)
			if err := stopCmd.Run(); err == nil {
				return &YieldedService{Type: "systemd", Name: procName}, nil
			}
		}
	}

	// 检测是否为 docker 容器 (docker-proxy)
	if procName == "docker-proxy" || strings.Contains(procName, "docker") {
		// 查找占用 80 端口的 docker 容器
		dockerCmd := exec.Command("docker", "ps", "-q", "--filter", "publish=80")
		var dOut bytes.Buffer
		dockerCmd.Stdout = &dOut
		if dockerCmd.Run() == nil && strings.TrimSpace(dOut.String()) != "" {
			containerID := strings.TrimSpace(strings.Split(dOut.String(), "\n")[0])
			log.Infof("正在临时挂起占用 80 端口的 Docker 容器 '%s'...", containerID)
			stopDocker := exec.Command("docker", "stop", containerID)
			if err := stopDocker.Run(); err == nil {
				return &YieldedService{Type: "docker", Name: containerID}, nil
			}
		}
	}

	return nil, fmt.Errorf("端口 80 被 '%s' 占用且无法自动停用，请手动释放端口", procName)
}

func restorePort80(svc *YieldedService) {
	if svc == nil {
		return
	}
	switch svc.Type {
	case "systemd":
		log.Infof("证书处理完成，正在恢复 systemd 服务 '%s'...", svc.Name)
		_ = exec.Command("systemctl", "start", svc.Name).Run()
	case "docker":
		log.Infof("证书处理完成，正在恢复 Docker 容器 '%s'...", svc.Name)
		_ = exec.Command("docker", "start", svc.Name).Run()
	}
}
