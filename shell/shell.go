package shell

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type ExecScriptOut struct {
	ScriptName string
	Err        string
	Out        string
	Combined   string
	ErrOut     string
}

// Holds the running pid of the process triggered by currently running script
//
// It could be a script from previous nextwork event.
// In that case current network event will kill it forcefully
// and update runningProcessPid with script from current network event
var runningProcessPid int

// Holds the last process killed forcefully
var killedProcessPid int

func ExecuteScriptOld(command string, envVars map[string]string, args ...string) *ExecScriptOut {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()

	for key, value := range envVars {
		keyvalue := fmt.Sprintf("%s=%s", key, value)
		cmd.Env = append(cmd.Env, keyvalue)
	}

	// Set output to Byte Buffers
	if cmd.Stdout != nil || cmd.Stderr != nil {
		return &ExecScriptOut{
			ScriptName: filepath.Base(command),
			Err:        "Stdout/StdErr already set"}
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	errString := ""
	if err != nil {
		errString = err.Error()
		// Add more information from error output in case of critical error
		if errb.String() != "" {
			errString = errString + "\n" + errb.String()
		}
	}
	return &ExecScriptOut{
		ScriptName: filepath.Base(command),
		Out:        outb.String(),
		ErrOut:     errb.String(),
		Combined:   outb.String() + "\n" + errb.String(),
		Err:        errString}
}

func setRunningProcessPid(pid int) {
	runningProcessPid = pid
}
func killPreviousRunningScript(pid int) {
	pgid, err := syscall.Getpgid(pid)
	if err == nil {
		// Kill the entire process group
		syscall.Kill(-pgid, syscall.SIGKILL)
	}
}

func ExecuteScript(command string, envVars map[string]string, args ...string) *ExecScriptOut {
	outputChan := make(chan *ExecScriptOut)
	pidChan := make(chan int)

	// Forcefully kill the script from previous network event if it's still running
	if runningProcessPid != 0 {
		killPreviousRunningScript(runningProcessPid)
		killedProcessPid = runningProcessPid
	}
	log.Println("Execute dispatch script " + command)

	go func() {
		var outb, errb bytes.Buffer

		cmd := exec.Command(command, args...)
		cmd.Env = os.Environ()

		for key, value := range envVars {
			keyvalue := fmt.Sprintf("%s=%s", key, value)
			cmd.Env = append(cmd.Env, keyvalue)
		}

		// Set output to Byte Buffers
		if cmd.Stdout != nil || cmd.Stderr != nil {
			outputChan <- &ExecScriptOut{
				ScriptName: filepath.Base(command),
				Err:        "Stdout/StdErr already set"}
			return
		}

		cmd.Stdout = &outb
		cmd.Stderr = &errb
		// Create a new process group to allow kill to kill
		// all the children process might start
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		createExecScriptOut := func(err error) *ExecScriptOut {
			errString := ""
			if err != nil {
				errString = err.Error()
			}
			// Add more information from error output in case of critical error
			if errb.String() != "" {
				errString = errString + "\n" + errb.String()
			}
			return &ExecScriptOut{
				ScriptName: filepath.Base(command),
				Out:        outb.String(),
				ErrOut:     errb.String(),
				Combined:   outb.String() + "\n" + errb.String(),
				Err:        errString}
		}

		err := cmd.Start()

		//  some fatal error on starting a script
		if err != nil {
			outputChan <- createExecScriptOut(err)
			return
		}
		pidChan <- cmd.Process.Pid
		err = cmd.Wait()
		// script execution error
		if err != nil {
			outputChan <- createExecScriptOut(err)
			return
		}
		outputChan <- createExecScriptOut(nil)
	}()
	var runningProcessPid int
	for {
		select {
		case pid := <-pidChan:
			runningProcessPid = pid
			setRunningProcessPid(pid)
		case output := <-outputChan:
			if runningProcessPid > 0 && runningProcessPid == killedProcessPid {
				output.Err = "Script was killed forcefully because next network event happen"
			}
			setRunningProcessPid(0)
			return output
		}
	}
}
