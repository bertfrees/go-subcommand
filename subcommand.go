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
	innerFlagsLong  map[string]*Flag
	innerFlagsShort map[string]*Flag
	fn              func(string)
}

//Parser is a command itself and contains other commands. It's the
//data structure and its name should be the executable name.
type Parser struct {
	Command
	Commands map[string]*Command
}

//NewParser constructs a parser for program name given
func NewParser(program string) *Parser {
	parser := new(Parser)
	parser.innerFlagsShort = make(map[string]*Flag)
	parser.innerFlagsLong = make(map[string]*Flag)
	parser.Commands = make(map[string]*Command)
	parser.Name = program
	return parser
}

//AddCommand inserts a new subcommand to the parser
func (p *Parser) AddCommand(name string, fn func(string)) (command *Command, err error) {
	if _, exists := p.Commands[name]; exists {
		return nil, fmt.Errorf("Command '%s' already exists ", name)
	}
	//create the command
	command = new(Command)
	command.Name = name
	command.innerFlagsShort = make(map[string]*Flag)
	command.innerFlagsLong = make(map[string]*Flag)
	command.fn = fn
	//add it to the p
	p.Commands[name] = command
	return command, nil
}

//AddFlag adds a new switch or option to the command.
//The distinction between switches and options is made from the long definition:
//
// "--option OPTION" expects a value after the flag
// "--switch " does not expect a value
//
//The short definition has no length restriction but it should be significantly shorter that its long counterpart
//Example:
//command.AddFlag("--thing THING","-t",thingProcessor)//option
//command.AddFlag("--tacata","-ta",isTacata)//switch
func (c *Command) AddFlag(long string, short string, description string, fn func(string)) (flag *Flag, err error) {
	flag, err = buildFlag(long, short, description, fn)
	if err != nil {
		return nil, err
	}
	if _, exists := c.innerFlagsLong[flag.Long]; exists {
		return nil, fmt.Errorf("Flag '%s' already exists ", long)
	}
	c.innerFlagsLong[flag.Long] = flag

	if _, exists := c.innerFlagsShort[flag.Short]; exists {
		return nil, fmt.Errorf("Flag '%s' already exists ", long)
	}
	c.innerFlagsShort[flag.Short] = flag
	return flag, nil

}

//flagged is convenience interface for treating commands and parsers equally
type flagged interface {
	getShortFlag(name string) (flag *Flag, ok bool)
	getLongFlag(name string) (flag *Flag, ok bool)
	getFlags() []Flag
	getName() string
}

//getShortFlag returns the flag for the given short definition or false if it's not present
func (c *Command) getShortFlag(name string) (flag *Flag, ok bool) {
	flag, ok = c.innerFlagsShort[name]
	return
}

//getLongFlag returns the flag for the long definition or false if it's not present
func (c *Command) getLongFlag(name string) (flag *Flag, ok bool) {
	flag, ok = c.innerFlagsLong[name]
	return
}

//getName returns the c name
func (c *Command) getName() string {
	return c.Name
}

//getFlags returns a slice containing the c's flags
func (c *Command) getFlags() []Flag {
	//return c.Name
	flags := make([]Flag, 0)
	for _, val := range c.innerFlagsLong {
		flags = append(flags, *val)
	}
	return flags
}

//Parse parses the arguments executing the associated functions for each command and flag. It returns the left overs if some non-option strings were not processed. Errors are returned in case an unknown flag is found or a mandatory flag was not supplied.
func (p *Parser) Parse(args []string) (leftOvers []string, err error) {
	leftOvers = make([]string, 0)
	visited := make([]Flag, 0)
	var currentCommand flagged = p
	//go comsuming options commands and sub-options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") { //flag
			var opt *Flag
			var ok bool
			if strings.HasPrefix(arg, "--") {
				opt, ok = currentCommand.getLongFlag(arg)
			} else {
				opt, ok = currentCommand.getShortFlag(arg)
			}
			//not present
			if !ok {
				err = fmt.Errorf("%v is not a valid flag for %v", arg, currentCommand.getName())
				//show help?
				return
			}
			if opt.Type == Option { //option
				if i+1 >= len(args) {
					err = fmt.Errorf("No value for option %v", arg)
					return
				}
				i++
				opt.fn(args[i])
			} else { //switch
				opt.fn("")
			}
			//add to visited options
			visited = append(visited, *opt)
		} else {
			//_,isParser:=currentCommand.(Parser)
			if ok, flag := checkVisited(visited, currentCommand.getFlags()); !ok {
				err = fmt.Errorf("%v was not found and is mandatory for %v", flag, currentCommand)
				return
			}
			cmd, ok := p.Commands[arg]
			if ok {
				currentCommand = cmd
				cmd.fn(arg)
			} else {
				leftOvers = append(leftOvers, arg)
			}

		}

	}

	return leftOvers, nil
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
func (f Flag) Must(isIt bool) {
	f.Mandatory = isIt
}

//parse flag type
func getFlagType(flag string) (fType FlagType, err error) {
	//is a switch?
	parts := strings.Split(flag, " ")
	l := len(parts)
	switch {
	case len(parts[0]) == 0:
		return -1, fmt.Errorf("Flag is empty")
	case l == 1:
		return Switch, err
	case l == 2:
		return Option, err
	default:
		return -1, fmt.Errorf("Flag '%s' has more than 2  words", flag)
	}
}

//parse long definition
func getFlagLonfDefinition(flag string) (name string, err error) {
	name = strings.Split(flag, " ")[0]
	if !strings.HasPrefix(name, "--") {
		return "", fmt.Errorf("Flag '%s' has to start by --", flag)
	}
	return name, nil
}

//parse short definition
func getFlagShortDefinition(flag string) (name string, err error) {
	parts := strings.Split(flag, " ")
	if len(parts) > 1 {
		return "", fmt.Errorf("Short flag must have only one word %s", flag)
	} else if !strings.HasPrefix(parts[0], "-") {
		return "", fmt.Errorf("Flag '%s' has to start by -", flag)
	}
	return parts[0], nil
}

//builds the flag struct
func buildFlag(long string, short string, desc string, fn func(string)) (flag *Flag, err error) {
	flag = new(Flag)
	flag.Type, err = getFlagType(long)
	if err != nil {
		return nil, err
	}

	flag.Long, err = getFlagLonfDefinition(long)
	if err != nil {
		return nil, err
	}
	flag.Short, err = getFlagShortDefinition(short)
	flag.fn = fn
	flag.Mandatory = false
	return flag, nil
}
