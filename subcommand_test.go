package subcommand

import (
	"errors"
	"testing"
)

var emptyFn = func(name, value string) error { return nil }
var emptyFnMult = func(command string, values ...string) error { return nil }

//build option
func TestParserOption(t *testing.T) {
	parser := NewParser("test")
	option := parser.AddOption("option", "o", "This is an option", emptyFn)

	if _, exists := parser.innerFlagsLong[option.Long]; !exists {
		t.Error("option is not present in the long names")
	}

	if _, exists := parser.innerFlagsShort[option.Short]; !exists {
		t.Error("option is not present in the short names")
	}
}

func TestBuildFlagOk(t *testing.T) {
	f := buildFlag("option", "o", "", emptyFn, Option)
	f2 := buildFlag("switch", "s", "", emptyFn, Switch)
	if f.Type != Option {
		t.Error("Option type not properly set")
	}
	if f2.Type != Switch {
		t.Error("Switch type not properly set")
	}
	if f.Long != "option" {
		t.Error("Option long type not properly set")
	}
	if f.Short != "o" {
		t.Error("Option short type not properly set")
	}
	if f.fn == nil {
		t.Error("Option fn not properly set")
	}
	if f.Mandatory {
		t.Error("Option mandatory not properly set")
	}
	f = buildFlag("option", "", "", emptyFn, Option)
	f2 = buildFlag("switch", "", "", emptyFn, Switch)

	if f.Type != Option {
		t.Error("Option type not properly set (empty short)")
	}
	if f2.Type != Switch {
		t.Error("Switch type not properly set (empty short)")
	}
}

func TestBuildFlagInvalidLong(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Not panicked with wrong long definition")
		}
	}()
	buildFlag("option OPTION", "o", "", emptyFn, Option)
}

func TestBuildFlagInvalidShort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Not panicked with wrong short definition")
		}
	}()
	buildFlag("option", "o o", "", emptyFn, Option)
}

func TestEmptyLong(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Not panicked with empty long definition")
		}
	}()
	buildFlag("", "o", "", emptyFn, Option)
}

func TestAddCommand(t *testing.T) {
	name := "com"
	parser := NewParser("test")
	command := parser.AddCommand(name, "", emptyFnMult)
	if command.Name != name {
		t.Errorf("Command name are not equals %v!=%v", command.Name, name)
	}
	if _, exists := parser.Commands[name]; !exists {
		t.Error("command not inserted")
	}

}

func TestAddCommandTwice(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Not panicked after inserting command twice")
		}
	}()
	name := "com"
	parser := NewParser("test")
	parser.AddCommand(name, "", emptyFnMult)
	parser.AddCommand(name, "", emptyFnMult)
}

func TestParseGlobalOption(t *testing.T) {
	parser := NewParser("test")
	processed := false
	parser.AddOption("option", "o", "This is an option", func(name, val string) error {
		if val == "value" && name == "option" {
			processed = true
		}
		return nil
	})
	parser.Parse([]string{"--option", "value"})
	if !processed {
		t.Error("Option wasn't processed")
	}

}

func TestParseGlobalOptionError(t *testing.T) {
	parser := NewParser("test")
	parser.AddOption("option", "o", "This is an option", func(name, val string) error {
		return errors.New("ERROR!")
	})
	_, err := parser.Parse([]string{"--option", "value"})
	if err == nil {
		t.Error("Error not thrown")
	}

}

func TestParseGlobalOptionShort(t *testing.T) {
	parser := NewParser("test")
	processed := false
	parser.AddOption("option", "o", "This is an option", func(name, val string) error {
		if val == "value" && name == "option" {
			processed = true
		}
		return nil
	})
	parser.Parse([]string{"-o", "value"})
	if !processed {
		t.Error("Option wasn't processed")
	}

}

func TestParseGlobalSwitch(t *testing.T) {
	parser := NewParser("test")
	processed := false
	parser.AddSwitch("switch", "s", "This is a switch", func(name, val string) error {
		if name == "switch" {
			processed = true
		}
		return nil
	})
	parser.Parse([]string{"--switch", "value"})
	if !processed {
		t.Error("Switch wasn't processed")
	}

}

func TestParseGlobalSwitchError(t *testing.T) {
	parser := NewParser("test")
	parser.AddSwitch("switch", "s", "This is a switch", func(name, val string) error {
		return errors.New("Error")
	})
	_, err := parser.Parse([]string{"--switch", "value"})
	if err == nil {
		t.Error("Error not thrown")
	}

}
func TestParseGlobalSwitchShort(t *testing.T) {

	parser := NewParser("test")
	processed := false
	parser.AddSwitch("switch", "s", "This is a switch", func(name, val string) error {
		if name == "switch" {
			processed = true
		}
		return nil
	})
	parser.Parse([]string{"-s", "value"})
	if !processed {
		t.Error("Switch wasn't processed")
	}

}

func TestParseGlobalNoOptionFound(t *testing.T) {
	parser := NewParser("test")
	_, err := parser.Parse([]string{"--nanana", "value"})
	if err == nil {
		t.Error("No error thrown")
	}
}

func TestParseGlobalOptionEmpty(t *testing.T) {
	parser := NewParser("test")
	parser.AddOption("option", "o", "This is an option", emptyFn)
	_, err := parser.Parse([]string{"--option"})
	if err == nil {
		t.Error("No error thrown")
	}
}

func TestParseCommand(t *testing.T) {
	parser := NewParser("test")
	proc := false
	parser.AddCommand("command", "", func(string, ...string) error {
		proc = true
		return nil
	})
	parser.Parse([]string{"command"})
	if !proc {
		t.Error("Command wasn't processed")
	}
}

func TestParseCommandError(t *testing.T) {
	parser := NewParser("test")
	parser.AddCommand("command", "", func(string, ...string) error {
		return errors.New("Error")
	})
	_, err := parser.Parse([]string{"command"})
	if err == nil {
		t.Error("Error wasn't thrown")
	}
}

func TestParseUnknown(t *testing.T) {
	parser := NewParser("test")
	parser.AddCommand("command", "", func(string, ...string) error {
		return nil
	})
	leftOvers, _ := parser.Parse([]string{"paco", "pepe"})
	if len(leftOvers) != 2 {
		t.Errorf("the parsing leftovers size isn't 2 (%v)", leftOvers)
		return
	}

	if leftOvers[0] != "paco" {
		t.Error("First element  of the leftovers is wrong")
		return
	}
	if leftOvers[1] != "pepe" {
		t.Error("Second element of the leftovers is wrong")
		return
	}
}

func TestParseInnerFlagCommand(t *testing.T) {
	parser := NewParser("test")
	shouldnt := false
	proc := false
	parser.AddSwitch("switch", "s", "This is a global switch", func(string, string) error {
		shouldnt = true
		return nil
	})
	cmd := parser.AddCommand("command", "", func(string, ...string) error {
		return nil
	})
	cmd.AddSwitch("switch", "s", "This is a command switch", func(string, string) error {
		proc = true
		return nil
	})
	parser.Parse([]string{"command", "-s"})
	if !proc {
		t.Error("Switch wasn't processed")
	}
	if shouldnt {
		t.Error("Confusion between global and command flag")
	}
}

func TestParseMandatorySwitch(t *testing.T) {
	parser := NewParser("test")
	parser.AddSwitch("switch", "s", "This is a mandatory switch", func(string, string) error {
		return nil
	}).Must(true)
	_, err := parser.Parse([]string{""})
	if err == nil {
		t.Error("Mandatory switch didn't complain")
	}
}

func TestParseMandatoryOption(t *testing.T) {
	parser := NewParser("test")
	parser.AddOption("option", "o", "This is a mandatory option", func(string, string) error {
		return nil
	}).Must(true)
	_, err := parser.Parse([]string{"command"})
	if err == nil {
		t.Error("Mandatory option didn't complain")
	}
}
func TestParseMandatoryInnerOption(t *testing.T) {
	parser := NewParser("test")
	cmd := parser.AddCommand("command", "", func(string, ...string) error { return nil })
	cmd.AddOption("option", "o", "This is a mandatory option", func(string, string) error {
		return nil
	}).Must(true)
	_, err := parser.Parse([]string{"command"})
	if err == nil {
		t.Error("Mandatory inner option didn't complain")
	}
}

func TestParseMandatoryInnerSwitch(t *testing.T) {
	parser := NewParser("test")
	parser.AddSwitch("switch", "s", "This is a mandatory switch", func(string, string) error {
		return nil
	}).Must(true)
	_, err := parser.Parse([]string{"command"})
	if err == nil {
		t.Error("Mandatory switch didn't complain")
	}
}
func TestParseCommandWithLefts(t *testing.T) {
	parser := NewParser("test")
	var name string
	var arg1 string
	var arg2 string

	parser.AddCommand("command", "", func(command string, args ...string) error {
		name = command
		arg1 = args[0]
		arg2 = args[1]
		return nil
	})

	parser.Parse([]string{"command", "arg1", "arg2"})
	if name != "command" {
		t.Errorf("command name %v", name)
	}

	if arg1 != "arg1" {
		t.Errorf("arg1 != %v", arg1)
	}
	if arg2 != "arg2" {
		t.Errorf("arg2 != %v", arg2)
	}
}

func TestParseCommandWithLeftsMandatoryFlag(t *testing.T) {
	parser := NewParser("test")
	var name string
	var arg1 string
	var arg2 string
	visited := false
	cmd := parser.AddCommand("command", "", func(command string, args ...string) error {
		name = command
		arg1 = args[0]
		arg2 = args[1]
		return nil
	})
	cmd.AddOption("opt", "o", "Mandatory option", func(string, string) error {
		visited = true
		return nil
	}).Must(true)

	_, err := parser.Parse([]string{"command", "arg1", "arg2"})
	if err == nil {
		t.Error("Expected error not thrown")
	}

	_, err = parser.Parse([]string{"command", "-o", "val", "arg1", "arg2"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !visited {
		t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "visited", visited, !visited)
	}
	if name != "command" {
		t.Errorf("command name %v", name)
	}

	if arg1 != "arg1" {
		t.Errorf("arg1 != %v", arg1)
	}
	if arg2 != "arg2" {
		t.Errorf("arg2 != %v", arg2)
	}
}

func TestSetHelp(t *testing.T) {
	parser := NewParser("test")
	helped := false
	parser.SetHelp("canihazhelp", "", func(command string, args ...string) error {
		helped = true
		return nil
	})

	parser.Parse([]string{"canihazhelp", "arg1", "arg2"})
	if !helped {
		t.Error("Help didn't work")
	}

}
func TestOnCommand(t *testing.T) {
	parser := NewParser("test")
	onCommand := false
	parser.OnCommand(func() error {
		onCommand = true
		return nil
	})
	parser.AddCommand("command", "", func(string, ...string) error { return nil })
	_, err := parser.Parse([]string{"command"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if !onCommand {
		t.Error("On command didn't executed")
	}

}

//func TestDefaultPrinter(t *testing.T) {
//parser := NewParser("test")
//parser.AddSwitch("switch", "s", "\tThis is a global switch", func(string,string) {
//})
////parser.AddOption("mandatory", "m", "This is a global mandatry option", func(string) {
////}).Must(true)
//parser.AddOption("option", "", "This is a global option", func(string,string) {
//})
//cmd:=parser.AddCommand("command", "This is a global command", func(string, ...string) {})
//cmd.AddOption("comopt","", "This is a command optoin", func(string,string) {})
////hPrinter:=&HelpPrinter{}
////hPrinter.VisitParser(*parser)
//parser.Parse([]string{"help","command"})
/*}*/
