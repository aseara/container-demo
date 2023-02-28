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
	exec1("cgdelete -r -g cpu,memory,pids:aseara")
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/home/aseara/rootfs/ubuntu"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	exec2(os.Args[2:]...)
	must(syscall.Unmount("/proc", 0))
}

func cg() {
	exec1("cgcreate -g cpu,memory,pids:aseara")
	exec1("cgset -r pids.max=20 aseara")
	exec1("cgset -r memory.limit_in_bytes=10000000 aseara")
	exec1("cgset -r cpu.cfs_period_us=100000 aseara")
	exec1("cgset -r cpu.cfs_quota_us=20000 aseara")
	exec1("cgclassify -g cpu,memory,pids:aseara " + strconv.Itoa(os.Getpid()))
}

func exec1(command string) {
	args := strings.Split(command, " ")
	exec2(args...)
}

func exec2(args ...string) {
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
