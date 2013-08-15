//Package subcommand is an option parsing utility a la git/mercurial/go loosely
//inspired by Ruby's OptionParser
package subcommand

import (
	"fmt"
	"strings"
)

//parser:=new(Parser).AddFlag("--size ","-s","size of what ever")
//parser.AddFlag("--cool" , "-c" ,"size", func (value String));
//parser.AddCommand("cosa").AddFlag("--cosa","-c"," cosa ");
//struct parser
//struct command
//struct flag

//FlagType defines the different flag types. Options have values associated to the flag, Switches have no value associated.
type FlagType int

const (
	Option FlagType = iota
	Switch
)

//Command aggregates different flags under a common name. Every time a command is found during the parsing process the associated function is executed.
type Command struct {
	//Name
	Name            string
	Description     string
	innerFlagsLong  map[string]*Flag
	innerFlagsShort map[string]*Flag
	fn              func(command string, leftOvers ...string)
	parent          *Command
}

//Access to flags
type Flagged interface {
	Flags() []Flag
}

//getFlags returns a slice containing the c's flags
func (c *Command) Flags() []Flag {
	//return c.Name
	flags := make([]Flag, 0)
	for _, val := range c.innerFlagsLong {
		flags = append(flags, *val)
	}
	return flags
}

//Parser contains other commands. It's the
//data structure and its name should be the executable name.
type Parser struct {
	Command
	Commands    map[string]*Command
	help        Command
	helpVisitor HelpVisitor
}

func newCommand(parent *Command, name string, description string, fn func(string, ...string)) *Command {
	return &Command{
		Name:            name,
		innerFlagsShort: make(map[string]*Flag),
		innerFlagsLong:  make(map[string]*Flag),
		fn:              fn,
		Description:     description,
		parent:          parent,
	}
}

func (p *Parser) SetHelp(name string, description string, fn func(string, ...string)) *Command {
	command := newCommand(&p.Command, name, description, fn)
	p.help = *command
	return command

}

func (p *Parser) SetHelpVisitor(v HelpVisitor) {
	p.helpVisitor = v

}

func (p Parser) Help() Command {
	return p.help
}

//NewParser constructs a parser for program name given
func NewParser(program string) *Parser {
	parser := &Parser{
		Command:  *newCommand(nil, program, "", nil),
		Commands: make(map[string]*Command),
	}
	parser.SetHelp("help", fmt.Sprintf("Type %v help [command] for detailed information about a command", program), defaultHelp(parser))
        visitor:=&HelpPrinter{}
        parser.SetHelpVisitor(visitor)
	return parser
}

func defaultHelp(p *Parser) func(string, ...string) {
	return func(help string, args ...string) {
		if len(args) > 0 {
			if cmd, ok := p.Commands[args[0]]; ok {
				p.helpVisitor.VisitCommand(*cmd)
				return
			} else {
				fmt.Printf("help: command not found %v\n", args[0])
			}
		}
		p.helpVisitor.VisitParser(*p)
	}
}

//AddCommand inserts a new subcommand to the parser
func (p *Parser) AddCommand(name string, description string, fn func(string, ...string)) *Command {
	if _, exists := p.Commands[name]; exists {
		panic(fmt.Sprintf("Command '%s' already exists ", name))
	}
	//create the command
	command := newCommand(&p.Command, name, description, fn)
	//add it to the parser
	p.Commands[name] = command
	return command
}

//Adds a new option to the command to be used as
// "--option OPTION" (expects a value after the flag)
//
//The short definition has no length restriction but it should be significantly shorter that its long counterpart
//Example:
//command.AddFlag("--thing THING","-t",thingProcessor)//option
//command.AddFlag("--tacata","-ta",isTacata)//switch
//func (c *Command) AddFlag(long string, short string, description string, fn func(string)) (flag *Flag, err error) {
//flag, err = buildFlag(long, short, description, fn)
//if err != nil {
//return nil, err
//}
//if _, exists := c.innerFlagsLong[flag.Long]; exists {
//return nil, fmt.Errorf("Flag '%s' already exists ", long)
//}
//c.innerFlagsLong[flag.Long] = flag

//if _, exists := c.innerFlagsShort[flag.Short]; exists {
//return nil, fmt.Errorf("Flag '%s' already exists ", long)
//}
//c.innerFlagsShort[flag.Short] = flag
//return flag, nil

//}

func (c *Command) AddOption(long string, short string, description string, fn func(string)) *Flag {
	flag := buildFlag(long, short, description, fn, Option)
	c.addFlag(flag)
	return flag
}

func (c *Command) AddSwitch(long string, short string, description string, fn func(string)) *Flag {
	flag := buildFlag(long, short, description, fn, Switch)
	c.addFlag(flag)
	return flag
}

func (c *Command) addFlag(flag *Flag) {

	if _, exists := c.innerFlagsLong[flag.Long]; exists {
		panic(fmt.Errorf("Flag '%s' already exists ", flag.Long))
	}
	if _, exists := c.innerFlagsShort[flag.Short]; exists {
		panic(fmt.Errorf("Flag '%s' already exists ", flag.Short))
	}

	c.innerFlagsLong[flag.Long] = flag
	c.innerFlagsShort[flag.Short] = flag

}

//func (c *Command) String() string{
//_,err:=fmt.Fprintf(w,"%v\t %v",c.Name,c.Description)
//return err
//}

//func (c *Command) HelpVervose(w io.Writer) error{
//_,err:=fmt.Fprintf(w,"%v\t %v",c.Name,c.Description)
//if err!=nil{
//return err
//}
//for _,flag:= range c.getFlags(){
//flag.Help(w)
//}
//return err
//}

//Parse parses the arguments executing the associated functions for each command and flag. It returns the left overs if some non-option strings were not processed. Errors are returned in case an unknown flag is found or a mandatory flag was not supplied.
func (p *Parser) Parse(args []string) (leftOvers []string, err error) {
	//get the delayed functions to call
	//for every flag//command
	fns, leftOvers, err := p.parse(args)
	if err != nil {
		return
	}
	for _, fn := range fns {
		fn()
	}
	return leftOvers, nil
}

func (p *Parser) parse(args []string) (functions []func(), leftOvers []string, err error) {
	//visited flags
	var visited []Flag
	//functions to call once the parsing process is over
	//TODO: user p.Command instead of the useless iface..
	var currentCommand Command = p.Command
	//go comsuming options commands and sub-options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") { //flag
			var opt *Flag
			var ok bool
			if strings.HasPrefix(arg, "--") {
				opt, ok = currentCommand.innerFlagsLong[arg[2:]]
			} else {
				opt, ok = currentCommand.innerFlagsShort[arg[1:]]
			}
			//not present
			if !ok {
				err = fmt.Errorf("%v is not a valid flag for %v", arg, currentCommand.Name)
				return
			}

			if opt.Type == Option { //option
				if i+1 >= len(args) {
					err = fmt.Errorf("No value for option %v", arg)
					return
				}
				i++
				functions = append(functions, flagCaller(args[i], opt.fn))
			} else { //switch
				functions = append(functions, flagCaller("", opt.fn))
			}
			//add to visited options
			visited = append(visited, *opt)
		} else {
			//_,isParser:=currentCommand.(Parser)
			if ok, flag := checkVisited(visited, currentCommand.Flags()); !ok {
				err = fmt.Errorf("%v was not found and is mandatory for %v", flag, currentCommand)
				return
			}
			cmd, ok := p.Commands[arg]
			//if its a command or the help command
			if ok && currentCommand.Name!=p.help.Name{
				currentCommand = *cmd
				functions = append(functions, commandCaller(arg, &leftOvers, cmd.fn))
                        }else if arg == p.help.Name {
				currentCommand = p.help
				functions = append(functions, commandCaller(arg, &leftOvers, p.help.fn))
			} else {
				leftOvers = append(leftOvers, arg)
			}

		}

	}

	return
}

func flagCaller(value string, fn func(string)) func() {
	return func() { fn(value) }
}
func commandCaller(command string, leftOvers *[]string, fn func(string, ...string)) func() {
	return func() { fn(command, *leftOvers...) }
}

//checks if the mandatory flags were visited
func checkVisited(visited []Flag, commandFlags []Flag) (visted bool, err Flag) {
	for _, flag := range commandFlags {
		if flag.Mandatory {
			ok := false
			for _, vFlag := range visited {
				if vFlag.Long == flag.Long {
					ok = true
					break
				}
			}
			if !ok {
				return false, flag
			}
		}
	}
	return true, err
}

//Flag structure
type Flag struct {
	//long definition (--option OPTION)
	Long string
	//Short definition (-o )
	Short string
	//Description
	Description string
	//FlagType, option or switch
	Type FlagType
	//Function to call when the flag is found during the parsing process
	fn func(string)
	//Says if the flag is optional or mandatory
	Mandatory bool
}

//Must sets the flag as mandatory. The parser will raise an error in case it isn't present in the arguments
func (f *Flag) Must(isIt bool) {
	f.Mandatory = isIt
}

func (f Flag) String() string {
	var format string
	var help string
	if f.Type == Option {
		if f.Mandatory {
			format = "-%v, --%v %v\t%v"
		} else {
			format = "-%v, --%v [%v]\t%v"
		}
		help = fmt.Sprintf(format, f.Short, f.Long, strings.ToUpper(f.Long), f.Description)
	} else {
		format = "-%v, --%v \t%v"
		help = fmt.Sprintf(format, f.Short, f.Long, f.Description)
	}
	return help
}

//Checks that the definition is just one word
func checkDefinition(flag string) bool {
	parts := strings.Split(flag, " ")
	return len(parts) == 1
}

//builds the flag struct
func buildFlag(long string, short string, desc string, fn func(string), kind FlagType) *Flag {
	long = strings.Trim(long, " ")
	short = strings.Trim(short, " ")
	if len(long) == 0 {
		panic("Long definition is empty")
	}
	if len(short) == 0 {
		panic("Short definition is empty")
	}

	if !checkDefinition(long) {
		panic(fmt.Sprintf("Long definition %v has two words. Only one is accepted", long))
	}

	if !checkDefinition(short) {
		panic(fmt.Sprintf("Short definition %v has two words. Only one is accepted", long))
	}
	return &Flag{
		Type:        kind,
		Long:        long,
		Short:       short,
		fn:          fn,
		Description: desc,
		Mandatory:   false,
	}
}

//HelpVisitor to print help
type HelpVisitor interface {
	VisitParser(p Parser)
	VisitCommand(c Command)
}

type HelpPrinter struct{}

func (h *HelpPrinter) VisitParser(p Parser) {
	fmt.Printf("Usage: %v [global_options] command [arguments]\n", p.Name)
	fmt.Printf("\n")
	fmt.Printf("Global Options\n")
	fmt.Printf("--------------\n")
	fmt.Printf("\n")
	for _, flag := range p.Flags() {
		fmt.Printf("\t%v\n", flag)
	}
	fmt.Printf("Commands\n")
	fmt.Printf("--------\n")
	fmt.Printf("\n")
	for _, cmd := range p.Commands {
		fmt.Printf("\t%v\t\t%v\n", cmd.Name, cmd.Description)
	}

	fmt.Printf("\n")
	fmt.Printf("\t%v\t\t%v\n", p.help.Name, p.help.Description)
}

func (h *HelpPrinter) VisitCommand(c Command) {
        fmt.Printf("%v\t\t%v\n", c.Name, c.Description)
	fmt.Printf("\n")
        for _, flag := range c.Flags() {
                fmt.Printf("\t%v\n", flag)
        }
}
