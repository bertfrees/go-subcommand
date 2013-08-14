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


func newCommand(name string, fn func(string)) *Command {
        return &Command{
                Name: name,
                innerFlagsShort: make(map[string]*Flag),
                innerFlagsLong:  make(map[string]*Flag),
                fn: fn,
        }
}

//NewParser constructs a parser for program name given
func NewParser(program string) *Parser {
        return &Parser{
                Command:  *newCommand(program, nil),
                Commands: make(map[string]*Command),
        }
}

//AddCommand inserts a new subcommand to the parser
func (p *Parser) AddCommand(name string, fn func(string)) *Command{
	if _, exists := p.Commands[name]; exists {
		panic(fmt.Sprintf("Command '%s' already exists ", name))
	}
	//create the command
        command:=newCommand(name,fn)
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

func  (c *Command) addFlag(flag *Flag) {

	if _, exists := c.innerFlagsLong[flag.Long]; exists {
		 panic(fmt.Errorf("Flag '%s' already exists ", flag.Long))
	}
	if _, exists := c.innerFlagsShort[flag.Short]; exists {
		 panic(fmt.Errorf("Flag '%s' already exists ", flag.Short))
	}

	c.innerFlagsLong[flag.Long] = flag
	c.innerFlagsShort[flag.Short] = flag

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
                                opt, ok = currentCommand.getLongFlag(arg[2:])
			} else {
                                opt, ok = currentCommand.getShortFlag(arg[1:])
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


//checks that the definition is just one word
func checkDefinition(flag string)  bool{
	parts := strings.Split(flag, " ")
	return len(parts) == 1
}

//builds the flag struct
func buildFlag(long string, short string, desc string, fn func(string),kind FlagType) *Flag {
        long=strings.Trim(long," ")
        short=strings.Trim(short," ")
        if len(long)==0{
                panic("Long definition is empty")
        }
        if len(short)==0{
                panic("Short definition is empty")
        }

        if ! checkDefinition(long){
                panic(fmt.Sprintf("Long definition %v has two words. Only one is accepted",long))
        }

        if ! checkDefinition(short){
                panic(fmt.Sprintf("Short definition %v has two words. Only one is accepted",long))
        }
        return  &Flag{
                Type:kind,
                Long: long,
                Short: short,
                fn : fn,
                Mandatory : false,
        }
}
