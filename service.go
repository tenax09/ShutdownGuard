package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
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

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	shutdownInitiated := false

	for {
		select {
		case <-ticker.C:
			// 判断是否需要关机：在限制时间段内 且 非 Administrator 登录
			if shouldShutdown() {
				if !shutdownInitiated {
					shutdownInitiated = true
					doShutdown()
				}
			} else {
				// 不满足关机条件，重置标志
				shutdownInitiated = false
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

// isInShutdownPeriod 判断当前小时是否处于关机时间段
// 关机时间段: 21:00 ~ 次日 11:00（即 hour >= 21 或 hour < 11）
func isInShutdownPeriod(hour int) bool {
	return hour >= 21 || hour < 11
}

// shouldShutdown 综合判断是否需要关机：在关机时间段内 且 不是 Administrator 登录
func shouldShutdown() bool {
	hour := time.Now().Hour()
	if !isInShutdownPeriod(hour) {
		return false
	}
	// Administrator 登录时不关机
	if isAdministratorLoggedIn() {
		return false
	}
	return true
}

// getActiveUsers 通过 "query user" 命令获取当前活跃的登录用户列表
func getActiveUsers() []string {
	cmd := exec.Command("query", "user")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 没有用户登录或命令不可用时返回空
		return nil
	}

	var users []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // 跳过标题行
		// 只关心 Active 状态的会话
		if !strings.Contains(line, "Active") || strings.Contains(line, "运行中") {
			continue
		}
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, ">") // 去掉当前会话标记
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			users = append(users, fields[0])
		}
	}
	return users
}

// isAdministratorLoggedIn 检查是否有 Administrator 用户当前处于活跃登录状态
func isAdministratorLoggedIn() bool {
	users := getActiveUsers()
	for _, user := range users {
		if strings.EqualFold(user, "administrator") {
			return true
		}
	}
	return false
}

// doShutdown 执行 Windows 关机命令
// 使用 60 秒倒计时，给用户一个保存工作的机会
func doShutdown() {
	cmd := exec.Command("shutdown", "/s", "/t", "60", "/c",
		"当前处于限制使用时间段（21:00 - 11:00），电脑将在 60 秒后关机。")
	err := cmd.Run()
	if err != nil {
		log.Printf("执行关机命令失败: %v", err)
	}
}

// runService 以 Windows 服务模式运行
func runService() {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("服务 %s 正在启动", serviceName))

	err = svc.Run(serviceName, &shutdownService{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("服务运行失败: %v", err))
		return
	}

	elog.Info(1, fmt.Sprintf("服务 %s 已停止", serviceName))
}

// runDebug 以调试模式在前台运行（方便测试）
func runDebug() {
	fmt.Println("========================================")
	fmt.Println("  ShutdownGuard - 调试模式")
	fmt.Println("  关机时间段: 21:00 ~ 次日 11:00")
	fmt.Println("  按 Ctrl+C 退出")
	fmt.Println("========================================")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	shutdownInitiated := false

	// 立即检查一次
	checkAndLog(&shutdownInitiated)

	for range ticker.C {
		checkAndLog(&shutdownInitiated)
	}
}

// checkAndLog 检查时间并输出日志（调试模式使用）
func checkAndLog(shutdownInitiated *bool) {
	now := time.Now()
	hour := now.Hour()
	users := getActiveUsers()
	adminLoggedIn := false
	for _, user := range users {
		if strings.EqualFold(user, "administrator") {
			adminLoggedIn = true
			break
		}
	}

	fmt.Printf("[%s] 时间检查 - 小时: %d, 活跃用户: %v, Administrator在线: %v",
		now.Format("2006-01-02 15:04:05"), hour, users, adminLoggedIn)

	if !isInShutdownPeriod(hour) {
		*shutdownInitiated = false
		fmt.Println(" → 不在限制时间段，正常运行")
	} else if adminLoggedIn {
		*shutdownInitiated = false
		fmt.Println(" → Administrator 已登录，跳过关机")
	} else if len(users) == 0 {
		*shutdownInitiated = false
		fmt.Println(" → 无用户登录，跳过关机")
	} else {
		if !*shutdownInitiated {
			*shutdownInitiated = true
			fmt.Println(" → 非管理员用户登录中，执行关机！")
			doShutdown()
		} else {
			fmt.Println(" → 关机命令已发出")
		}
	}
}
