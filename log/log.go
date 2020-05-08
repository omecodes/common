package log

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
)

var debug bool

func init() {
	debug = os.Getenv("OE_DEBUG") == "1"
}

var logger Logger

var (
	serv, erro, info, success, deb, txt func(txt string) string
)

func getLogger() Logger {
	if logger == nil {
		if runtime.GOOS == `windows` {
			log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
			serv = func(txt string) string {
				return fmt.Sprintf("%s%-35s%s", "\x1b[34m", txt, "\x1b[0m")
			}
			success = func(txt string) string {
				return fmt.Sprintf("%s%-2s%s", "\x1b[32m", txt, "\x1b[0m")
			}
			info = func(txt string) string {
				return fmt.Sprintf("%s%-2s%s", "\x1b[90;37m", txt, "\x1b[0m")
			}
			erro = func(txt string) string {
				return fmt.Sprintf("%s%-2s%s", "\x1b[31m", txt, "\x1b[0m")
			}
			deb = func(txt string) string {
				return fmt.Sprintf("%s%-2s%s", "\x1b[35m", txt, "\x1b[0m")
			}
			txt = func(txt string) string {
				return fmt.Sprintf("%-50s", txt)
			}
		} else {

			red := color.New(color.FgRed).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			blue := color.New(color.FgBlue).SprintFunc()
			grey := color.New(color.FgHiCyan).SprintFunc()

			serv = func(txt string) string {
				return blue(fmt.Sprintf("%-35s", txt))
			}
			success = func(txt string) string {
				return green(fmt.Sprintf("%-2s", txt))
			}
			info = func(txt string) string {
				return grey(fmt.Sprintf("%-2s", txt))
			}
			erro = func(txt string) string {
				return red(fmt.Sprintf("%-2s", txt))
			}
			deb = func(txt string) string {
				return fmt.Sprintf("%-2s", txt)
			}
			txt = func(txt string) string {
				return fmt.Sprintf("%-50s", txt)
			}
		}
		logger = new(defaultLogger)
	}
	return logger
}

// Logger is a conveniance for logging
type Logger interface {
	su(from string, msg string, args ...interface{})
	in(from string, msg string, args ...interface{})
	de(from string, msg string, args ...interface{})
	er(from string, err error, args ...interface{})
	fa(from string, err error, args ...interface{})
	suf(from string, format string, args ...interface{})
	inf(from string, format string, args ...interface{})
	def(from string, format string, args ...interface{})
	erf(from string, err error, format string, args ...interface{})
	faf(from string, err error, format string, args ...interface{})
}

type defaultLogger struct{}

func (ul *defaultLogger) su(from string, msg string, args ...interface{}) {
	items := []interface{}{
		success("V"),
		serv(from),
		txt(msg),
	}
	if len(args) > 0 {
		items = append(items, args...)
	}
	log.Print(items...)
}
func (ul *defaultLogger) in(from string, msg string, args ...interface{}) {
	items := []interface{}{
		info("I"),
		serv(from),
		txt(msg),
	}
	if len(args) > 0 {
		items = append(items, args...)
	}
	log.Print(items...)
}
func (ul *defaultLogger) de(from string, msg string, args ...interface{}) {
	if debug {
		items := []interface{}{
			deb("D"),
			serv(from),
			txt(msg),
		}
		if len(args) > 0 {
			items = append(items, args...)
		}
		log.Print(items...)
	}
}
func (ul *defaultLogger) er(from string, err error, args ...interface{}) {
	items := []interface{}{
		erro(erro("!")),
		serv(from),
		err.Error(),
	}
	if len(args) > 0 {
		items = append(items, args...)
	}
	log.Print(items...)
}
func (ul *defaultLogger) fa(from string, err error, args ...interface{}) {
	items := []interface{}{
		erro("F"),
		serv(from),
		err.Error(),
	}
	if len(args) > 0 {
		items = append(items, args...)
	}
	log.Fatal(items...)
}
func (ul *defaultLogger) suf(from string, format string, args ...interface{}) {
	items := []interface{}{
		success("V"),
		serv(from),
		fmt.Sprintf(format, args...),
	}
	log.Print(items...)
}
func (ul *defaultLogger) inf(from string, format string, args ...interface{}) {
	items := []interface{}{
		info("I"),
		serv(from),
		fmt.Sprintf(format, args...),
	}
	log.Print(items...)
}
func (ul *defaultLogger) def(from string, format string, args ...interface{}) {
	if debug {
		items := []interface{}{
			deb("D"),
			serv(from),
			fmt.Sprintf(format, args...),
		}
		log.Print(items...)
	}
}
func (ul *defaultLogger) erf(from string, err error, format string, args ...interface{}) {
	items := []interface{}{
		erro("!"),
		serv(from),
		fmt.Sprintf(format, args...),
	}
	log.Print(items...)
}
func (ul *defaultLogger) faf(from string, err error, format string, args ...interface{}) {
	items := []interface{}{
		erro("F"),
		serv(from),
		fmt.Sprintf(format, args...),
	}
	log.Print(items...)
}

// Info displays informative message
func S(from string, msg string, args ...interface{}) {
	getLogger().su(from, msg, args...)
}

func I(from string, msg string, args ...interface{}) {
	getLogger().in(from, msg, args...)
}

// Debug used for showing more detailed activity
func D(from string, msg string, args ...interface{}) {
	getLogger().de(from, msg, args...)
}

// Error displays more detailed error message
func E(from string, err error, args ...interface{}) {
	getLogger().er(from, err, args...)
}

// Error displays more detailed error message
func F(from string, err error, args ...interface{}) {
	getLogger().fa(from, err, args...)
}

// Info displays informative message
func Suf(from string, format string, args ...interface{}) {
	getLogger().suf(from, format, args...)
}

func If(from string, format string, args ...interface{}) {
	getLogger().inf(from, format, args...)
}

// Debug used for showing more detailed activity
func Df(from string, format string, args ...interface{}) {
	getLogger().def(from, format, args...)
}

// Error displays more detailed error message
func Ef(from string, err error, format string, args ...interface{}) {
	getLogger().erf(from, err, format, args...)
}

// Error displays more detailed error message
func Ff(from string, err error, format string, args ...interface{}) {
	getLogger().faf(from, err, format, args...)
}
