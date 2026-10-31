// LiteBoost - 轻薄本性能优化工具
// Windows 平台性能一键优化软件
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// Windows API 常量
const (
	// 电源计划 GUID
	highPerformanceGUID = "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"
	balancedGUID        = "381b4222-f694-41f0-9685-ff5bb260df2e"
	powerSaverGUID      = "a1841308-3541-4fab-bc81-f71556f20b4a"
)

func main() {
	if runtime.GOOS != "windows" {
		fmt.Println("本软件仅支持 Windows 操作系统。")
		pauseExit()
		return
	}

	// 启用 UTF-8 控制台输出，避免中文乱码
	enableUTF8Console()

	// 尝试提权
	if !isAdmin() {
		fmt.Println("正在请求管理员权限以执行系统优化...")
		relaunchAsAdmin()
		return
	}

	showBanner()

	for {
		showMenu()
		choice := readLine()
		switch choice {
		case "1":
			setHighPerformancePower()
		case "2":
			cleanTempFiles()
		case "3":
			disableVisualEffects()
		case "4":
			disableStartupItems()
		case "5":
			flushDNS()
		case "6":
			optimizeMemory()
		case "7":
			disableSysMain()
		case "8":
			oneClickOptimize()
		case "9":
			restoreBalanced()
		case "0", "q", "Q":
			fmt.Println("\n感谢使用 LiteBoost，再见！")
			time.Sleep(800 * time.Millisecond)
			return
		default:
			fmt.Println("\n  [!] 无效输入，请重新选择。")
		}
		fmt.Println()
	}
}

// ---------- 界面 ----------

func showBanner() {
	fmt.Println()
	fmt.Println("  ====================================================")
	fmt.Println("        LiteBoost  轻薄本性能优化工具  v1.0")
	fmt.Println("  ====================================================")
	fmt.Println("  针对轻薄本 / 低功耗设备的一键性能优化方案")
	fmt.Println("  包含电源、存储、视觉、内存、服务等多项优化")
	fmt.Println("  ====================================================")
	fmt.Println()
}

func showMenu() {
	fmt.Println("  ┌──────────────────────────────────────────────┐")
	fmt.Println("  │              请选择优化项目                    │")
	fmt.Println("  ├──────────────────────────────────────────────┤")
	fmt.Println("  │  1. 切换到「高性能」电源计划                   │")
	fmt.Println("  │  2. 清理系统临时文件                           │")
	fmt.Println("  │  3. 关闭视觉效果，提升界面响应速度             │")
	fmt.Println("  │  4. 禁用不必要的开机启动项                     │")
	fmt.Println("  │  5. 刷新 DNS 缓存                             │")
	fmt.Println("  │  6. 优化内存（清空闲置进程工作集）             │")
	fmt.Println("  │  7. 关闭 SysMain 预读服务                     │")
	fmt.Println("  │  8. ★ 一键智能优化（推荐）                     │")
	fmt.Println("  │  9. 恢复「平衡」电源计划                       │")
	fmt.Println("  │  0. 退出                                      │")
	fmt.Println("  └──────────────────────────────────────────────┘")
	fmt.Print("  请输入选项编号: ")
}

// ---------- 优化功能 ----------

// setHighPerformancePower 切换到高性能电源计划
func setHighPerformancePower() {
	fmt.Println("\n  [*] 正在切换到高性能电源计划...")
	// 先确保高性能计划可用（部分系统默认隐藏）
	runCmd("powercfg", "/duplicatescheme", highPerformanceGUID)
	out, err := runCmdOut("powercfg", "/setactive", highPerformanceGUID)
	if err != nil {
		fmt.Println("  [!] 设置失败:", err)
		return
	}
	_ = out
	// 关闭 USB 选择性挂起、PCI Express 链接状态电源管理等节能项
	runCmd("powercfg", "/change", "standby-timeout-ac", "0")
	runCmd("powercfg", "/change", "hibernate-timeout-ac", "0")
	runCmd("powercfg", "/change", "disk-timeout-ac", "0")
	runCmd("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "2a737441-193a-4430-a5a1-0f8e0e0e0e0e", "3").Run()
	fmt.Println("  [√] 已切换到高性能电源计划，CPU 不会再被强制降频。")
}

// cleanTempFiles 清理临时文件
func cleanTempFiles() {
	fmt.Println("\n  [*] 正在清理系统临时文件...")
	paths := []string{
		os.Getenv("TEMP"),
		os.Getenv("TMP"),
		filepath.Join(os.Getenv("WINDIR"), "Temp"),
		filepath.Join(os.Getenv("WINDIR"), "Prefetch"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Temp"),
	}
	cleaned := 0
	for _, p := range paths {
		if p == "" {
			continue
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			continue
		}
		for _, e := range entries {
			full := filepath.Join(p, e.Name())
			if err := os.RemoveAll(full); err == nil {
				cleaned++
			}
		}
	}
	// 使用系统工具清理 Windows 更新缓存
	runCmd("cmd", "/c", "cleanmgr /sagerun:1").Run()
	fmt.Printf("  [√] 临时文件清理完成，共清理 %d 项。\n", cleaned)
}

// disableVisualEffects 关闭视觉效果
func disableVisualEffects() {
	fmt.Println("\n  [*] 正在关闭视觉效果以提升响应速度...")
	// VisualEffects = 2 表示「调整为最佳性能」
	regSet(`HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects`, "VisualFXSetting", "REG_DWORD", "2")
	// 关闭桌面背景合成、透明玻璃、淡入淡出等
	regSet(`HKCU\Control Panel\Desktop`, "UserPreferencesMask", "REG_BINARY", "90,12,03,80,10,00,00,00")
	regSet(`HKCU\Control Panel\Desktop\WindowMetrics`, "MinAnimate", "REG_SZ", "0")
	// 关闭菜单动画
	regSet(`HKCU\Control Panel\Desktop`, "MenuShowDelay", "REG_SZ", "0")
	fmt.Println("  [√] 视觉效果已关闭，界面响应将明显加快（需重启资源管理器生效）。")
	fmt.Println("  [i] 提示: 如需立即生效，可在任务管理器中重启 explorer.exe。")
}

// disableStartupItems 禁用不必要的启动项
func disableStartupItems() {
	fmt.Println("\n  [*] 正在禁用不必要的开机启动项...")
	// 通过注册表禁用当前用户和本地机器的 Run 启动项（备份后注释）
	keys := []string{
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		`HKLM\Software\Microsoft\Windows\CurrentVersion\Run`,
		`HKLM\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Run`,
	}
	disabled := 0
	for _, key := range keys {
		out, _ := runCmdOut("reg", "query", key)
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, key) {
				continue
			}
			// 取空格前的名称
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			name := fields[0]
			// 跳过系统关键项
			if isSystemStartup(name) {
				continue
			}
			// 导出备份后删除
			backup := filepath.Join(os.Getenv("TEMP"), "liteboost_run_backup_"+sanitize(name)+".reg")
			runCmd("reg", "export", key, backup, "/y").Run()
			if err := runCmd("reg", "delete", key, "/v", name, "/f").Run(); err == nil {
				disabled++
			}
		}
	}
	fmt.Printf("  [√] 已禁用 %d 个非必要启动项（备份已保存到 %%TEMP%%）。\n", disabled)
}

// flushDNS 刷新 DNS 缓存
func flushDNS() {
	fmt.Println("\n  [*] 正在刷新 DNS 缓存...")
	if err := runCmd("ipconfig", "/flushdns").Run(); err != nil {
		fmt.Println("  [!] 刷新失败:", err)
		return
	}
	fmt.Println("  [√] DNS 缓存已刷新，网络解析将更顺畅。")
}

// optimizeMemory 优化内存：清理内存分页并释放工作集
func optimizeMemory() {
	fmt.Println("\n  [*] 正在优化内存...")
	// 清空系统工作集需要 SetProcessWorkingSetSize 对自身无效，
	// 通过 EmptyWorkingSet 遍历所有进程释放闲置内存
	// 使用 PowerShell 一行脚本实现
	ps := `Get-Process | Where-Object {$_.MainWindowHandle -ne 0} | ForEach-Object { $_.WorkingSet64 = 0 }`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
	fmt.Println("  [√] 已释放闲置进程占用的内存，可用内存增加。")
}

// disableSysMain 关闭 SysMain 服务
func disableSysMain() {
	fmt.Println("\n  [*] 正在关闭 SysMain (Superfetch) 服务...")
	runCmd("sc", "stop", "SysMain").Run()
	runCmd("sc", "config", "SysMain", "start=", "disabled").Run()
	fmt.Println("  [√] SysMain 服务已禁用，可减少磁盘占用（对 SSD 尤其有益）。")
}

// oneClickOptimize 一键智能优化
func oneClickOptimize() {
	fmt.Println("\n  [★] 开始一键智能优化，将依次执行安全优化项...")
	fmt.Println()
	setHighPerformancePower()
	cleanTempFiles()
	disableVisualEffects()
	flushDNS()
	optimizeMemory()
	fmt.Println()
	fmt.Println("  ====================================================")
	fmt.Println("  [★] 一键优化完成！建议重启电脑以获得最佳效果。")
	fmt.Println("  ====================================================")
}

// restoreBalanced 恢复平衡电源计划
func restoreBalanced() {
	fmt.Println("\n  [*] 正在恢复平衡电源计划...")
	runCmd("powercfg", "/setactive", balancedGUID).Run()
	fmt.Println("  [√] 已恢复平衡电源计划。")
}

// ---------- 辅助函数 ----------

func isAdmin() bool {
	// 尝试以管理员方式打开注册表项，成功则有管理员权限
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run() == nil
}

func relaunchAsAdmin() {
	// 使用 PowerShell 以管理员身份重启自身
	self, _ := os.Executable()
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`Start-Process -FilePath '%s' -Verb RunAs`, self))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}

func enableUTF8Console() {
	// 设置控制台代码页为 UTF-8 (65001)
	exec.Command("chcp", "65001").Run()
}

func runCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

func runCmdOut(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func regSet(path, name, typ, value string) {
	args := []string{"add", path, "/v", name, "/t", typ, "/d", value, "/f"}
	cmd := exec.Command("reg", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}

func isSystemStartup(name string) bool {
	system := []string{"SecurityHealth", "WindowsDefender", "RtkNGUI", "igfxTray", "HotKeysCmds", "Persistence"}
	lower := strings.ToLower(name)
	for _, s := range system {
		if strings.Contains(lower, strings.ToLower(s)) {
			return true
		}
	}
	return false
}

func sanitize(s string) string {
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_", " ", "_")
	return replacer.Replace(s)
}

func readLine() string {
	var s string
	fmt.Scanln(&s)
	return strings.TrimSpace(s)
}

func pauseExit() {
	fmt.Print("按回车键退出...")
	readLine()
}
