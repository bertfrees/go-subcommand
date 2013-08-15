//Package subcommand is an option parsing utility a la git/mercurial/go loosely
//inspired by Ruby's OptionParser
package subcommand

import (
	"fmt"
	"strings"
)

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

//Parser contains other commands. It's the data structure and its name should be the program's name.
type Parser struct {
	Command
	Commands map[string]*Command
	help     Command
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

//Sets the help command. There is one default implementation automatically added when the parser is created.
func (p *Parser) SetHelp(name string, description string, fn func(string, ...string)) *Command {
	command := newCommand(&p.Command, name, description, fn)
	p.help = *command
	return command

}

//Returns the help command
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
	return parser
}

func defaultHelp(p *Parser) func(string, ...string) {
	return func(help string, args ...string) {
		if len(args) > 0 {
			if cmd, ok := p.Commands[args[0]]; ok {
				visitCommand(*cmd)
				return
			} else {
				fmt.Printf("help: command not found %v\n", args[0])
			}
		}
		visitParser(*p)
	}
}

//AddCommand inserts a new subcommand to the parser. The callback fn receives as first argument
//the command name followed by the left overs of the parsing process
//Example:
// command "hello" prints the non flags (options and switches) arguments.
// The associated callback should be something like
// func processCommand(commandName string,args ...string){
//      fmt.Printf("The command %v says:\n",commandName)
//      for _,arg:= rage args{
//              fmt.Printf("%v \n",arg)
//      }
//}
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

//Adds a new option to the command to be used as "--option OPTION" (expects a value after the flag) in the command line
//The short definition has no length restriction but it should be significantly shorter that its long counterpart
//The function fn receives the name of the option and its value
//Example:
//command.AddOption("path","p",setPath)//option
//[...]
// func setPath(option,value string){
//      printf("According the option %v the path is set to %v",option,value);
//}
func (c *Command) AddOption(long string, short string, description string, fn func(string, string)) *Flag {
	flag := buildFlag(long, short, description, fn, Option)
	c.addFlag(flag)
	return flag
}

//Adds a new switch to the command to be used as "--switch" (expects no value after the flag) in the command line
//The short definition has no length restriction but it should be significantly shorter that its long counterpart
//The function fn receives two string, the first is the switch name and the second is just an empty string
//Example:
//command.AddSwitch("verbose","v",setVerbose)//option
//[...]
// func setVerbose(switch string){
//      printf("I'm get to get quite talkative! I'm set to be %v ",switch);
//}
func (c *Command) AddSwitch(long string, short string, description string, fn func(string, string)) *Flag {
	flag := buildFlag(long, short, description, fn, Switch)
	c.addFlag(flag)
	return flag
}

//Adds a flag to the command
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

func (c *Command) String() string {
	return fmt.Sprintf("%v\t %v", c.Name, c.Description)
}

//Parse parses the arguments executing the associated functions for each command and flag.
//It returns the left overs if some non-option strings or commands  were not processed.
//Errors are returned in case an unknown flag is found or a mandatory flag was not supplied.
// The set of function calls to be performed are carried in order and once the parsing process is done
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

//The actual parsing process
func (p *Parser) parse(args []string) (functions []func(), leftOvers []string, err error) {
	//visited flags
	var visited []Flag
	//functions to call once the parsing process is over
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
				functions = append(functions, flagCaller(opt.Long, args[i+1], opt.fn))
				i++
			} else { //switch
				functions = append(functions, flagCaller(opt.Long, "", opt.fn))
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
			//if its a command
			if ok && currentCommand.Name != p.help.Name {
				currentCommand = *cmd
				functions = append(functions, commandCaller(arg, &leftOvers, cmd.fn))
			} else if arg == p.help.Name { //it's the help
				currentCommand = p.help
				functions = append(functions, commandCaller(arg, &leftOvers, p.help.fn))
			} else {
				leftOvers = append(leftOvers, arg)
			}

		}

	}

	return
}

func flagCaller(name, value string, fn func(string, string)) func() {
	return func() { fn(name, value) }
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
	fn func(string, string)
	//Says if the flag is optional or mandatory
	Mandatory bool
}

//Must sets the flag as mandatory. The parser will raise an error in case it isn't present in the arguments
//TODO make sure that switches are not allowed to get mandatory
func (f *Flag) Must(isIt bool) {
	f.Mandatory = isIt
}

//Gets a help friendly flag representation:
//-o,--option  OPTION           This option does this and that
//-s,--switch                   This is a switch
//-i,--ignoreme [IGNOREME]      Optional option
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

//builds the flag struct panicking if errors are encountered
func buildFlag(long string, short string, desc string, fn func(string, string), kind FlagType) *Flag {
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

//Help printing functions
func visitParser(p Parser) {
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

func visitCommand(c Command) {
	fmt.Printf("%v\t\t%v\n", c.Name, c.Description)
	fmt.Printf("\n")
	for _, flag := range c.Flags() {
		fmt.Printf("\t%v\n", flag)
	}
}
