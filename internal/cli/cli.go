package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
)

var errValidation = errors.New("validation error")

type OptionCompletionFunc func(cur string) []string

type Option interface {
	GetName() byte
	GetHelp() string
	GetMetavar() string
	SetValue(v string) error
	SetDefault()
	IsFlag() bool
	IsSet() bool
	GetCompletionHandler() OptionCompletionFunc
}

type BoolOption struct {
	Name    byte
	Default bool
	Help    string
	value   bool
	isSet   bool
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
	o.value = !o.Default
	o.isSet = true
	return nil
}

func (o *BoolOption) SetDefault() {
	o.value = o.Default
}

func (o *BoolOption) IsFlag() bool {
	return true
}

func (o *BoolOption) IsSet() bool {
	return o.isSet
}

func (o *BoolOption) GetCompletionHandler() OptionCompletionFunc {
	return nil
}

func (o *BoolOption) GetValue() bool {
	return o.value
}

type StringOption struct {
	Name              byte
	Default           string
	Help              string
	Metavar           string
	CompletionHandler OptionCompletionFunc
	value             string
	isSet             bool
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
	o.isSet = true
	return nil
}

func (o *StringOption) IsFlag() bool {
	return false
}

func (o *StringOption) IsSet() bool {
	return o.isSet
}

func (o *StringOption) GetCompletionHandler() OptionCompletionFunc {
	return o.CompletionHandler
}

func (o *StringOption) SetDefault() {
	o.value = o.Default
}

func (o *StringOption) GetValue() string {
	return o.value
}

type ArgumentCompletionFunc func(prev string, cur string) []string

type Argument struct {
	Name              string
	Help              string
	Required          bool
	Remaining         bool
	CompletionHandler ArgumentCompletionFunc
	value             []string
	isSet             bool
}

func (a *Argument) GetValue() string {
	if l := len(a.value); l > 0 {
		return a.value[l-1]
	}
	return ""
}

func (a *Argument) GetValues() []string {
	return a.value
}

func (a *Argument) IsSet() bool {
	return a.isSet
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

func (c *Cli) getOption(name byte) Option {
	opts := append(c.iOptions, c.Options...)
	for i := len(opts) - 1; i >= 0; i-- {
		if o := opts[i]; o != nil && o.GetName() == name {
			return o
		}
	}
	return nil
}

func (c *Cli) completion() {
	c.init()

	compLine, found := os.LookupEnv("COMP_LINE")
	if !found || len(os.Args) != 4 {
		return
	}

	cur := os.Args[2]
	prev := os.Args[3]

	args, _ := shlex.Split(compLine)
	c.parse(args)

	comp := []string{}

	if strings.HasPrefix(cur, "-") {
		for _, o := range append(c.iOptions, c.Options...) {
			if n := fmt.Sprintf("-%c", o.GetName()); o != nil && strings.HasPrefix(n, cur) {
				comp = append(comp, n)
			}
		}
	}

	aPrev := ""
	if cur == "" || !strings.HasPrefix(cur, "-") {
		found := false
		if strings.HasPrefix(prev, "-") {
			if len(prev) == 2 {
				if o := c.getOption(prev[1]); o != nil && !o.IsFlag() {
					if h := o.GetCompletionHandler(); h != nil {
						comp = append(comp, h(cur)...)
					}
					found = true
				}
			}
		}
		if !found {
			for i, a := range c.Arguments {
				if a == nil {
					continue
				}
				if !a.isSet || a.GetValue() == cur {
					if a.CompletionHandler != nil {
						comp = append(comp, a.CompletionHandler(aPrev, cur)...)
					}
					break
				}
				if i == len(c.Arguments)-1 && a.Remaining && a.CompletionHandler != nil {
					comp = append(comp, a.CompletionHandler(aPrev, cur)...)
					break
				}
				aPrev = a.GetValue()
			}
		}
	}

	for _, c := range comp {
		fmt.Println(c)
	}

	os.Exit(0)
}

func (c *Cli) parseOpt(name byte, opt []string) (bool, error) {
	op := c.getOption(name)
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

		if lArg := len(c.Arguments); iArg < lArg || c.Arguments[lArg-1].Remaining {
			idx := iArg
			if idx >= lArg {
				idx = lArg - 1
			}

			if a := c.Arguments[idx]; a != nil {
				a.value = append(a.value, arg)
				a.isSet = true
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
	c.completion()

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

	for i, arg := range c.Arguments {
		if arg == nil {
			continue
		}
		if arg.Required {
			fmt.Fprint(w, " ")
		} else {
			fmt.Fprint(w, " [")
		}
		fmt.Fprint(w, c.argUsage(arg))
		if i == len(c.Arguments)-1 && arg.Remaining {
			fmt.Fprint(w, " ...")
		}
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
