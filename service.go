package main

import (
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

// shutdownService 实现 svc.Handler 接口
type shutdownService struct{}

// Execute 是 Windows 服务的主循环
func (s *shutdownService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if shouldShutdown() {
				doShutdown()
				doShutdownAPI()
				doShutdownPowerShell()
			}

		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				return false, 0
			}
		}
	}
}

// shouldShutdown 综合判断是否需要关机
// 规则：在限制时间段内 -> 关机
func shouldShutdown() bool {
	hour := time.Now().Hour()
	return hour >= 21 || hour < 11
}

// doShutdown 方案1：直接调用 shutdown 命令
func doShutdown() {
	cmd := exec.Command("shutdown", "/s", "/t", "60", "/c", "当前处于限制使用时间段（21:00 - 11:00），电脑将在 60 秒后关机。")
	cmd.Run()
}

// doShutdownAPI 方案2：使用 Windows API 调用
func doShutdownAPI() {
	dll := syscall.NewLazyDLL("user32.dll")
	proc := dll.NewProc("ExitWindowsEx")
	// EWX_SHUTDOWN = 1, reason = SHTDN_REASON_MAJOR_APPLICATION
	proc.Call(1, 0)
}

// doShutdownPowerShell 方案3：使用 PowerShell 命令
func doShutdownPowerShell() {
	cmd := exec.Command("powershell", "-Command", "Stop-Computer -Force")
	cmd.Run()
}

// runService 以 Windows 服务模式运行
func runService() {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return
	}
	defer elog.Close()

	elog.Info(1, "服务 "+serviceName+" 正在启动")

	err = svc.Run(serviceName, &shutdownService{})
	if err != nil {
		elog.Error(1, "服务运行失败")
		return
	}

	elog.Info(1, "服务 "+serviceName+" 已停止")
}
