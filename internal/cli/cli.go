package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var errValidation = errors.New("validation error")

type Option interface {
	GetName() byte
	GetHelp() string
	GetMetavar() string
	SetValue(v string) error
	SetDefault()
	IsFlag() bool
}

type BoolOption struct {
	Name    byte
	Default bool
	Help    string
	value   bool
}

func (o *BoolOption) GetName() byte {
	return o.Name
}

func (o *BoolOption) GetHelp() string {
	rv := o.Help
	if o.Default {
		rv += " (default: set)"
	}
	return rv
}

func (o *BoolOption) GetMetavar() string {
	return ""
}

func (o *BoolOption) SetValue(v string) error {
	if len(v) == 0 {
		o.value = !o.Default
		return nil
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}

	o.value = b
	return nil
}

func (o *BoolOption) SetDefault() {
	o.value = o.Default
}

func (o *BoolOption) IsFlag() bool {
	return true
}

func (o *BoolOption) GetValue() bool {
	return o.value
}

type StringOption struct {
	Name    byte
	Default string
	Help    string
	Metavar string
	value   string
}

func (o *StringOption) GetName() byte {
	return o.Name
}

func (o *StringOption) GetHelp() string {
	rv := o.Help
	if o.Default != "" {
		rv += fmt.Sprintf(" (default: %q)", o.Default)
	}
	return rv
}

func (o *StringOption) GetMetavar() string {
	return o.Metavar
}

func (o *StringOption) SetValue(v string) error {
	o.value = v
	return nil
}

func (o *StringOption) IsFlag() bool {
	return false
}

func (o *StringOption) SetDefault() {
	o.value = o.Default
}

func (o *StringOption) GetValue() string {
	return o.value
}

type Argument struct {
	Name     string
	Help     string
	Required bool
	value    string
}

func (a *Argument) GetValue() string {
	return a.value
}

type Cli struct {
	Help      string
	Version   string
	Options   []Option
	Arguments []*Argument
	iOptions  []Option
	oHelp     *BoolOption
	oVersion  *BoolOption
}

func (c *Cli) init() {
	if c.iOptions != nil {
		return
	}

	c.oHelp = &BoolOption{
		Name:    'h',
		Default: false,
		Help:    "show this help message and exit",
	}
	c.iOptions = []Option{c.oHelp}

	if c.Version != "" {
		c.oVersion = &BoolOption{
			Name:    'v',
			Default: false,
			Help:    "show version and exit",
		}
		c.iOptions = append(c.iOptions, c.oVersion)
	}
}

func (c *Cli) parseOpt(name byte, opt []string) (bool, error) {
	var op Option
	for _, o := range append(c.iOptions, c.Options...) {
		if o != nil && o.GetName() == name {
			op = o
		}
	}
	if op == nil || len(opt) == 0 {
		return false, fmt.Errorf("%w: invalid option: -%c", errValidation, name)
	}

	if op.IsFlag() {
		op.SetValue("")
		if len(opt[0]) > 0 {
			n := opt[0][0]
			opt[0] = opt[0][1:]
			return c.parseOpt(n, opt)
		}
		return false, nil
	}

	if len(opt[0]) > 0 {
		return false, op.SetValue(opt[0])
	}

	if len(opt) != 2 {
		return false, fmt.Errorf("%w: missing option value: -%c", errValidation, name)
	}

	return true, op.SetValue(opt[1])
}

func (c *Cli) parse(argv []string) error {
	c.init()

	l := len(argv)
	if l < 1 {
		return errors.New("invalid number of command line arguments")
	}

	for _, opt := range append(c.iOptions, c.Options...) {
		if opt != nil {
			opt.SetDefault()
		}
	}

	iArg := 0

	for i := 1; i < l; i++ {
		arg := argv[i]

		if len(arg) > 1 && arg[0] == '-' {
			opt := []string{arg[2:]}
			if i+1 < l {
				opt = append(opt, argv[i+1])
			}
			inc, err := c.parseOpt(arg[1], opt)
			if err != nil {
				return err
			}
			if inc {
				i++
			}
			continue
		}

		if iArg < len(c.Arguments) {
			if a := c.Arguments[iArg]; a != nil {
				a.value = arg
				iArg++
			}
		}
	}

	for i := iArg; i < len(c.Arguments); i++ {
		if a := c.Arguments[i]; a != nil && a.Required {
			return fmt.Errorf("%w: missing required argument: %s", errValidation, a.Name)
		}
	}

	return nil
}

func (c *Cli) Parse() {
	err := c.parse(os.Args)

	if err == nil || errors.Is(err, errValidation) {
		if c.oHelp.GetValue() {
			c.usage(true, os.Stderr, os.Args)
			os.Exit(0)
		}

		if len(os.Args) > 0 && c.oVersion != nil && c.oVersion.GetValue() {
			fmt.Fprintf(os.Stderr, "%s %s", filepath.Base(os.Args[0]), c.Version)
			fmt.Fprintln(os.Stderr)
			os.Exit(0)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s", filepath.Base(os.Args[0]), err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr)
		c.usage(false, os.Stderr, os.Args)
		os.Exit(1)
	}
}

func (c *Cli) optUsage(opt Option) string {
	rv := fmt.Sprintf("-%c", opt.GetName())
	if !opt.IsFlag() {
		mv := "VALUE"
		if v := opt.GetMetavar(); v != "" {
			mv = strings.ToUpper(v)
		}
		rv += fmt.Sprintf(" %s", mv)
	}
	return rv
}

func (c *Cli) argUsage(arg *Argument) string {
	return strings.ToUpper(strings.Replace(arg.Name, "-", "_", -1))
}

func (c *Cli) usage(full bool, w io.Writer, argv []string) {
	c.init()

	argv0 := "prog"
	if len(argv) > 0 {
		argv0 = filepath.Base(argv[0])
	}

	titlePadding := len(argv0)

	if full {
		fmt.Fprintf(w, "usage:\n    %s", argv0)
		titlePadding += 4
	} else {
		fmt.Fprintf(w, "usage: %s", argv0)
		titlePadding += 7
	}

	fOpts := append(c.iOptions, c.Options...)
	iOpts := []int{}

	for i := len(fOpts) - 1; i >= 0; i-- {
		if fOpts[i] == nil {
			continue
		}
		found := false
		for _, n := range iOpts {
			if n == i {
				found = true
				break
			}
		}
		if !found {
			iOpts = append(iOpts, i)
		}
	}

	opts := []Option{}

	for i := len(iOpts) - 1; i >= 0; i-- {
		if o := fOpts[iOpts[i]]; o != nil {
			opts = append(opts, o)
		}
	}

	for _, opt := range opts {
		fmt.Fprintf(w, " [%s]", c.optUsage(opt))
	}

	for _, arg := range c.Arguments {
		if arg == nil {
			continue
		}
		if arg.Required {
			fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, " [")
		}
		fmt.Fprint(w, c.argUsage(arg))
		if !arg.Required {
			fmt.Fprint(w, "]")
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%*s - %s", titlePadding, " ", c.Help)
	fmt.Fprintln(w)

	if !full {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "arguments:")
	for _, arg := range c.Arguments {
		if arg == nil {
			continue
		}
		fmt.Fprintf(w, "    %-20s %s", c.argUsage(arg), arg.Help)
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "options:")
	for _, opt := range opts { // already filtered
		fmt.Fprintf(w, "    %-20s %s", c.optUsage(opt), opt.GetHelp())
		fmt.Fprintln(w)
	}
}

func (c *Cli) Usage(full bool) {
	c.usage(full, os.Stderr, os.Args)
}
