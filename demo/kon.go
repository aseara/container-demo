// main container demo
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

func run() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	cg()
	must(cmd.Run())
	exe("cgdelete -r -g cpu,memory,pids:aseara")
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/home/aseara/rootfs/ubuntu"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
	must(syscall.Unmount("/proc", 0))
}

func cg() {
	exe("cgcreate -g cpu,memory,pids:aseara")
	exe("cgset -r pids.max=20 aseara")
	exe("cgset -r memory.limit_in_bytes=10000000 aseara")
	exe("cgset -r cpu.cfs_period_us=100000 aseara")
	exe("cgset -r cpu.cfs_quota_us=20000 aseara")
	exe("cgclassify -g cpu,memory,pids:aseara " + strconv.Itoa(os.Getpid()))
}

func exe(command string) {
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
