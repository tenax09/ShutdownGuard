package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "ShutdownGuard"
const serviceDisplayName = "自动定时服务"
const serviceDesc = "自动服务：晚上9点到早上11点期间自动关机"

func main() {
	// 判断是否以 Windows 服务方式运行
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("无法判断是否为Windows服务: %v", err)
	}

	if isService {
		runService()
		return
	}

	// 命令行模式，处理参数
	if len(os.Args) < 2 {
		usage()
		return
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "install":
		err = installService()
	case "uninstall", "remove":
		err = removeService()
	case "start":
		err = startService()
	case "stop":
		err = stopService()
	default:
		usage()
		return
	}

	if err != nil {
		log.Fatalf("命令 %s 执行失败: %v", cmd, err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "自动关机服务（21:00 ~ 次日 11:00）\n\n")
	fmt.Fprintf(os.Stderr, "用法: %s <命令>\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "可用命令:\n")
	fmt.Fprintf(os.Stderr, "  install   - 安装为 Windows 服务（开机自启）\n")
	fmt.Fprintf(os.Stderr, "  uninstall - 卸载服务\n")
	fmt.Fprintf(os.Stderr, "  start     - 启动服务\n")
	fmt.Fprintf(os.Stderr, "  stop      - 停止服务\n")
}

// installService 将程序安装为 Windows 服务
func installService() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %v", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("服务 %s 已存在，请先卸载", serviceName)
	}

	s, err = m.CreateService(serviceName, exePath, mgr.Config{
		DisplayName: serviceDisplayName,
		Description: serviceDesc,
		StartType:   mgr.StartAutomatic, // 开机自动启动
	})
	if err != nil {
		return fmt.Errorf("创建服务失败: %v", err)
	}
	defer s.Close()

	fmt.Printf("服务 %s 安装成功（开机自启动）\n", serviceName)
	return nil
}

// removeService 卸载 Windows 服务
func removeService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("打开服务失败（可能未安装）: %v", err)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return fmt.Errorf("删除服务失败: %v", err)
	}

	fmt.Printf("服务 %s 已卸载\n", serviceName)
	return nil
}

// startService 启动服务
func startService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("打开服务失败: %v", err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("启动服务失败: %v", err)
	}

	fmt.Printf("服务 %s 已启动\n", serviceName)
	return nil
}

// stopService 停止服务
func stopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接服务管理器失败: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("打开服务失败: %v", err)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("停止服务失败: %v", err)
	}

	fmt.Printf("服务 %s 已停止\n", serviceName)
	return nil
}
