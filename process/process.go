package process

import (
	"errors"
	"log"
	"os"
	"os/exec"
)

type Params struct {
	Name           string
	HideGUI        bool
	SingleInstance bool
	Executable     string
	Arguments      []string
	Env            []string
	WorkingDir     string
	InputFile      *os.File
	OutputLogFile  *os.File
	ErrorLogFile   *os.File
}

type ParamsWrapper interface {
	Wrap(*Params) (*Params, error)
}

type Process struct {
	cmd    *exec.Cmd
	params *Params
}

func (p *Process) Name() string {
	return ""
}

func (p *Process) Start() (err error) {
	if p.cmd == nil {
		p.cmd = exec.Command(p.params.Executable, p.params.Arguments...)
		p.cmd.Env = p.params.Env
		p.cmd.Dir = p.params.WorkingDir
		p.cmd.Stdin = p.params.InputFile
		p.cmd.Stdout = p.params.OutputLogFile
		p.cmd.Stderr = p.params.ErrorLogFile
		err = p.cmd.Start()
	} else {
		err = errors.New("process already started once")
	}
	return
}

func (p *Process) PID() int {
	if p.cmd.Process != nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p *Process) Stop() error {
	if p.cmd.Process == nil {
		return errors.New("process not even started")
	}
	return p.cmd.Process.Kill()
}

func (p *Process) Restart() (err error) {
	if err = p.Stop(); err != nil {
		log.Println("stop failed, cause: ", err)
	}
	return p.Start()
}

func (p *Process) Clean() error {
	return nil
}

func NewProcess(params *Params) *Process {
	return &Process{params: params}
}
