package subcommand

import (
	"testing"
)

var assert string = "expected %v but was %v"
var emptyFn = func(value string) {}

func assertEquals(t *testing.T, exp string, res string) {
	if exp != res {
		t.Error(assert, exp, res)
	}
}
func assertEqualsI(t *testing.T, exp int, res int) {
	if exp != res {
		t.Error(assert, exp, res)
	}

}
func assertNil(t *testing.T, err error) {
	if err != nil {
		t.Error("Error is not nil")
	}
}

//build option full
func TestParserFlagOptionFull(t *testing.T) {
	parser := NewParser("test")
	option, err := parser.AddFlag("--option OPT", "-o", "This is an option", emptyFn)

	assertNil(t, err)
	if option == nil {
		t.Error("option is nil")
	}
	if _, exists := parser.innerFlagsLong[option.Long]; !exists {
		t.Error("option is not present in the long names")
	}

	if _, exists := parser.innerFlagsShort[option.Short]; !exists {
		t.Error("option is not present in the short names")
	}
}

func TestBuildFlag(t *testing.T) {
	flag, err := buildFlag("--cosa cosa", "-c", "desc!", func(value string) {
		//nothing
	})
	assertNil(t, err)
	assertEqualsI(t, int(Option), int(flag.Type))
	assertEquals(t, "--cosa", flag.Long)
	assertEquals(t, "-c", flag.Short)
}

func TestGetFlagTypeOption(t *testing.T) {
	fType, err := getFlagType("--option OPTION")
	assertNil(t, err)
	assertEqualsI(t, int(Option), int(fType))
}

func TestGetFlagTypeSwitch(t *testing.T) {
	fType, err := getFlagType("--switch")
	assertNil(t, err)
	assertEqualsI(t, int(Switch), int(fType))
}

func TestGetFlagTypeEmpty(t *testing.T) {
	fType, err := getFlagType("")
	if err == nil {
		t.Error("No error thrown")
	}
	assertEqualsI(t, -1, int(fType))
}

func TestGetFlagTypeTooLong(t *testing.T) {
	fType, err := getFlagType("--cosa otra hahha")
	if err == nil {
		t.Error("No error thrown")
	}
	assertEqualsI(t, -1, int(fType))
}

func TestGetFlagNameOption(t *testing.T) {
	name, err := getFlagLonfDefinition("--cosa THINGY")
	assertNil(t, err)
	assertEquals(t, "--cosa", name)
}

func TestGetFlagNameSwitch(t *testing.T) {
	name, err := getFlagLonfDefinition("--cosa")
	assertNil(t, err)
	assertEquals(t, "--cosa", name)
}
func TestGetFlagNoPrefix(t *testing.T) {
	_, err := getFlagLonfDefinition("-cosa")
	if err == nil {
		t.Error("No error thrown")
	}
}
func TestGetFlagShortName(t *testing.T) {
	name, err := getFlagShortDefinition("-c")
	assertNil(t, err)
	assertEquals(t, "-c", name)
}

func TestGetFlagShortNameTooManyWords(t *testing.T) {
	_, err := getFlagShortDefinition("-c a")
	if err == nil {
		t.Error("No error thrown")
	}
}
func TestGetFlagShortNameNoPrefix(t *testing.T) {
	_, err := getFlagShortDefinition("#c")
	if err == nil {
		t.Error("No error thrown")
	}
}

func TestAddCommand(t *testing.T) {
	parser := NewParser("test")
	command, err := parser.AddCommand("com", emptyFn)
	assertNil(t, err)
	assertEquals(t, "com", command.Name)
	if _, exists := parser.Commands["com"]; !exists {
		t.Error("command not inserted")
	}

}

func TestAddCommandTwice(t *testing.T) {
	parser := NewParser("test")
	_, err := parser.AddCommand("com", emptyFn)
	_, err = parser.AddCommand("com", emptyFn)
	if err == nil {
		t.Error("No error thrown")
	}

}
func TestParseGlobalOption(t *testing.T) {
	parser := NewParser("test")
	processed := false
	parser.AddFlag("--option OPT", "-o", "This is an option", func(val string) {
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
	parser.AddFlag("--option OPT", "-o", "This is an option", func(val string) {
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
	parser.AddFlag("--switch", "-s", "This is a switch", func(string) {
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
	parser.AddFlag("--switch", "-s", "This is a switch", func(string) {
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
	parser.AddFlag("--option ", "-o", "This is an option", emptyFn)
	_, err := parser.Parse([]string{"--option"})
	if err == nil {
		t.Error("No error thrown")
	}
}

func TestParseCommand(t *testing.T) {
	parser := NewParser("test")
	proc := false
	parser.AddCommand("command", func(string) {
		proc = true
	})
	parser.Parse([]string{"command"})
	if !proc {
		t.Error("Command wasn't processed")
	}
}

func TestParseUnknown(t *testing.T) {
	parser := NewParser("test")
	parser.AddCommand("command", func(string) {
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
	parser.AddFlag("--switch", "-s", "This is a global switch", func(string) {
		shouldnt = true
	})
	cmd, _ := parser.AddCommand("command", func(string) {
	})
	cmd.AddFlag("--switch", "-s", "This is a command switch", func(string) {
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
