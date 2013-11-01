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

//Convinience type for funcions passed to commands
type CommandFunction func(string, ...string) error

//Convinience type for funcions passed flags
type FlagFunction func(string, string) error

//Command aggregates different flags under a common name. Every time a command is found during the parsing process the associated function is executed.
type Command struct {
	//Name
	Name            string
	Description     string //Command help line
	Params          string //Description of the parameters
	innerFlagsLong  map[string]*Flag
	innerFlagsShort map[string]*Flag
	fn              CommandFunction
	postFlagsFn     func() error
	parent          *Command
	arity           Arity
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

//Returns the command parent
func (c Command) Parent() *Command {
	return c.parent
}

//Parser contains other commands. It's the data structure and its name should be the program's name.
type Parser struct {
	Command
	Commands map[string]*Command
	help     Command
}

func newCommand(parent *Command, name string, description string, fn CommandFunction) *Command {
	return &Command{
		Name:            name,
		innerFlagsShort: make(map[string]*Flag),
		innerFlagsLong:  make(map[string]*Flag),
		fn:              fn,
		postFlagsFn:     func() error { return nil },
		Description:     description,
		parent:          parent,
		arity:           Arity{-1, "arg1,arg2,..."},
	}
}

//Sets the help command. There is one default implementation automatically added when the parser is created.
func (p *Parser) SetHelp(name string, description string, fn CommandFunction) *Command {
	command := newCommand(&p.Command, name, description, fn)
	p.help = *command
	return command

}

//First level execution when parsing. The passed function is exectued taking the leftovers until the first command
//./prog -switch left1 left2 command
//in this case name will be prog, and left overs left1 and left2
func (p *Parser) OnCommand(fn CommandFunction) {
	p.fn = fn
}

func (p *Parser) PostFlags(fn func() error) {
	p.postFlagsFn = fn
}

//Returns the help command
func (p Parser) Help() Command {
	return p.help
}

//NewParser constructs a parser for program name given
func NewParser(program string) *Parser {
	parser := &Parser{
		Command:  *newCommand(nil, program, "", func(string, ...string) error { return nil }),
		Commands: make(map[string]*Command),
	}
	parser.Command.arity = Arity{0, ""}
	parser.SetHelp("help", fmt.Sprintf("Type %v help [command] for detailed information about a command", program), defaultHelp(parser))
	return parser
}

func defaultHelp(p *Parser) CommandFunction {
	return func(help string, args ...string) error {
		if len(args) > 0 {
			if cmd, ok := p.Commands[args[0]]; ok {
				visitCommand(*cmd)
				return nil
			} else {
				fmt.Printf("help: command not found %v\n", args[0])
			}
		}
		visitParser(*p)
		return nil
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
func (p *Parser) AddCommand(name string, description string, fn CommandFunction) *Command {
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
func (c *Command) AddOption(long string, short string, description string, fn FlagFunction) *Flag {
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
func (c *Command) AddSwitch(long string, short string, description string, fn FlagFunction) *Flag {
	flag := buildFlag(long, short, description, fn, Switch)
	c.addFlag(flag)
	return flag
}

type Arity struct {
	Count       int
	Description string
}

//Set arity:
//-1 accepts infinite arguments.
//Other restricts the arity to the given num
func (c *Command) SetArity(arity int, description string) *Command {
	c.arity = Arity{arity, description}
	return c
}

func (c Command) Arity() Arity {
	return c.arity
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
	if flag.Short != "" {
		c.innerFlagsShort[flag.Short] = flag
	}

}

//Parse parses the arguments executing the associated functions for each command and flag.
//It returns the left overs if some non-option strings or commands  were not processed.
//Errors are returned in case an unknown flag is found or a mandatory flag was not supplied.
// The set of function calls to be performed are carried in order and once the parsing process is done
func (p *Parser) Parse(args []string) (leftOvers []string, err error) {
	err = p.parse(args, p.Command)
	if err != nil {
		return
	}
	return
}

//The actual parsing process
func (p *Parser) parse(args []string, currentCommand Command) (err error) {
	//TODO : rewrite the parsing algorithm to make it a bit more clean and clever...
	//visited flags
	var flagsToCall []flagCallable
	var leftOvers []string
	var nextCommandCall func() error
	i := 0
	//functions to call once the parsing process is over
	//go comsuming options commands and sub-options
	for ; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") { //flag
			var fCallable flagCallable
			fCallable, i, err = currentCommand.parseFlag(args, i)
			flagsToCall = append(flagsToCall, fCallable)
			if err != nil {
				return
			}

		} else { //command or leftover
			//call the flags
			if err = currentCommand.callFlags(flagsToCall); err != nil {
				return
			}

			cmd, isCommand := p.Commands[arg]
			//if its a command or help
			if isHelp := (arg == p.help.Name); (isCommand || isHelp) && currentCommand.Name != p.help.Name {
				nextCommandCall = func() error {
					i := i
					if isHelp {
						cmd = &(p.help)
					}
					//call with the rest of the args
					err := p.parse(args[i+1:], *cmd)
					if err != nil {
						return err
					}
					return nil
				}

				break
			} else {
				leftOvers = append(leftOvers, arg)
			}

		}

	}
	//call the flags
	if nextCommandCall == nil && len(leftOvers) == 0 {
		if err = currentCommand.callFlags(flagsToCall); err != nil {
			return
		}
	}
	//call current command
	if err = currentCommand.exec(leftOvers); err != nil {
		return
	}
	//look for next command
	if nextCommandCall != nil {
		return nextCommandCall()
	}
	return nil
}

//Execute the command function with leftovers as parameters
func (c Command) exec(leftOvers []string) error {
	arity := c.Arity().Count
	//check correct number of params
	if arity != -1 && arity != len(leftOvers) {
		return fmt.Errorf("Command %s accepts %v parameters but %v found (%v)",
			c.Name, arity, len(leftOvers), leftOvers)

	}
	if err := c.fn(c.Name, leftOvers...); err != nil {
		return err
	}
	return nil
}

//Call the each flag with the associated value
func (c Command) callFlags(flagsToCall []flagCallable) error {
	//check if we got all the mandatory flags
	if err := checkVisited(flagsToCall, c); err != nil {
		return err
	}
	//call flag functions
	for _, fc := range flagsToCall {
		if err := fc.fn(); err != nil {
			return err
		}

	}
	//call post flags
	return c.postFlagsFn()
}

//convinience lambda to pass the flag function around
func flagFunction(name, value string, fn FlagFunction) func() error {
	return func() error { return fn(name, value) }
}

//contains the flag and its fucntion ready to call
type flagCallable struct {
	fn   func() error
	flag Flag
}

//parses a flag and returns a flag callable to execute and the new position of the args iterator
func (c Command) parseFlag(args []string, pos int) (callable flagCallable, newPos int, err error) {
	arg := args[pos]
	newPos = pos
	var opt *Flag
	var ok bool
	var fn func() error
	//long or shor definition
	if strings.HasPrefix(arg, "--") {
		opt, ok = c.innerFlagsLong[arg[2:]]
	} else {
		opt, ok = c.innerFlagsShort[arg[1:]]
	}
	//not present
	if !ok {
		err = fmt.Errorf("%v is not a valid flag for %v", arg, c.Name)
		return
	}

	if opt.Type == Option { //option
		if pos+1 >= len(args) {
			err = fmt.Errorf("No value for option %v", arg)
			return
		}
		fn = flagFunction(opt.Long, args[pos+1], opt.fn)
		newPos = pos + 1
	} else { //switch
		fn = flagFunction(opt.Long, "", opt.fn)
	}
	callable = flagCallable{fn, *opt}
	return
}

//checks if the mandatory flags were visited
func checkVisited(visited []flagCallable, command Command) error {
	for _, flag := range command.Flags() {
		if flag.Mandatory {
			ok := false
			for _, vFlag := range visited {
				if vFlag.flag.Long == flag.Long {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("%v was not found and is mandatory for %v", flag, command)
			}
		}
	}
	return nil
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
	fn func(string, string) error
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
	return fmt.Sprintf("%s\t%s", f.FlagStringPrefix(), f.Description)
}

func (f Flag) FlagStringPrefix() string {
	var format string
	var prefix string
	shortFormat := "%v"
	if f.Short != "" {
		shortFormat = "-%v,"
	}
	if f.Type == Option {
		if f.Mandatory {
			format = "--%v %v"
		} else {
			format = "--%v [%v]"
		}
		prefix = fmt.Sprintf(shortFormat+format, f.Short, f.Long, strings.ToUpper(f.Long))
	} else {
		format = "--%v"
		prefix = fmt.Sprintf(shortFormat+format, f.Short, f.Long)
	}
	return prefix
}

//Checks that the definition is just one word
func checkDefinition(flag string) bool {
	parts := strings.Split(flag, " ")
	return len(parts) == 1
}

//builds the flag struct panicking if errors are encountered
func buildFlag(long string, short string, desc string, fn FlagFunction, kind FlagType) *Flag {
	long = strings.Trim(long, " ")
	short = strings.Trim(short, " ")
	if len(long) == 0 {
		panic("Long definition is empty")
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
