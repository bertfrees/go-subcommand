package subcommand

import (
	"testing"
)

var emptyFn = func(value string) {}
var emptyFnMult = func(command string, values ...string) {}

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

func TestEmptyShort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Not panicked with empty short definition")
		}
	}()
	buildFlag("option", "", "", emptyFn, Option)
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
	parser.AddOption("option", "o", "This is an option", func(val string) {
		if val == "value" {
			processed = true
		}
	})
	parser.Parse([]string{"--option", "value"})
	if !processed {
		t.Error("Option wasn't processed")
	}

}

func TestParseGlobalOptionShort(t *testing.T) {
	parser := NewParser("test")
	processed := false
	parser.AddOption("option", "o", "This is an option", func(val string) {
		if val == "value" {
			processed = true
		}
	})
	parser.Parse([]string{"-o", "value"})
	if !processed {
		t.Error("Option wasn't processed")
	}

}

func TestParseGlobalSwitch(t *testing.T) {
	parser := NewParser("test")
	processed := false
	parser.AddSwitch("switch", "s", "This is a switch", func(string) {
		processed = true
	})
	parser.Parse([]string{"--switch", "value"})
	if !processed {
		t.Error("Switch wasn't processed")
	}

}

func TestParseGlobalSwitchShort(t *testing.T) {

	parser := NewParser("test")
	processed := false
	parser.AddSwitch("switch", "s", "This is a switch", func(string) {
		processed = true
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
	parser.AddCommand("command", "", func(string, ...string) {
		proc = true
	})
	parser.Parse([]string{"command"})
	if !proc {
		t.Error("Command wasn't processed")
	}
}

func TestParseUnknown(t *testing.T) {
	parser := NewParser("test")
	parser.AddCommand("command", "", func(string, ...string) {
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
	parser.AddSwitch("switch", "s", "This is a global switch", func(string) {
		shouldnt = true
	})
	cmd := parser.AddCommand("command", "", func(string, ...string) {
	})
	cmd.AddSwitch("switch", "s", "This is a command switch", func(string) {
		proc = true
	})
	parser.Parse([]string{"command", "-s"})
	if !proc {
		t.Error("Switch wasn't processed")
	}
	if shouldnt {
		t.Error("Confusion between global and command flag")
	}
}

func TestParseCommandWithLefts(t *testing.T) {
	parser := NewParser("test")
	var name string
	var arg1 string
	var arg2 string

	parser.AddCommand("command", "", func(command string, args ...string) {
		name = command
		arg1 = args[0]
		arg2 = args[1]
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

/*func TestDefaultPrinter(t *testing.T) {*/
//parser := NewParser("test")
//parser.AddSwitch("switch", "s", "\tThis is a global switch", func(string) {
//})
////parser.AddOption("mandatory", "m", "This is a global mandatry option", func(string) {
////}).Must(true)
//parser.AddOption("option", "o", "This is a global option", func(string) {
//})
//cmd:=parser.AddCommand("command", "This is a global command", func(string, ...string) {})
//cmd.AddOption("comopt","o", "This is a command optoin", func(string) {})
////hPrinter:=&HelpPrinter{}
////hPrinter.VisitParser(*parser)
//parser.Parse([]string{"help","command"})
/*}*/
