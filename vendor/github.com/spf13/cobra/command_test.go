package cobra

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// test to ensure hidden commands run as intended
func TestHiddenCommandExecutes(t *testing.T) {

	// ensure that outs does not already equal what the command will be setting it
	// to, if it did this test would not actually be testing anything...
	if outs == "hidden" {
		t.Errorf("outs should NOT EQUAL hidden")
	}

	cmdHidden.Execute()

	// upon running the command, the value of outs should now be 'hidden'
	if outs != "hidden" {
		t.Errorf("Hidden command failed to run!")
	}
}

// test to ensure hidden commands do not show up in usage/help text
func TestHiddenCommandIsHidden(t *testing.T) {
	if cmdHidden.IsAvailableCommand() {
		t.Errorf("Hidden command found!")
	}
}

func TestStripFlags(t *testing.T) {
	tests := []struct {
		input  []string
		output []string
	}{
		{
			[]string{"foo", "bar"},
			[]string{"foo", "bar"},
		},
		{
			[]string{"foo", "--bar", "-b"},
			[]string{"foo"},
		},
		{
			[]string{"-b", "foo", "--bar", "bar"},
			[]string{},
		},
		{
			[]string{"-i10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-i=10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"--int=100", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-ib", "echo", "-bfoo", "baz"},
			[]string{"echo", "baz"},
		},
		{
			[]string{"-i=baz", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--int=baz", "-bbar", "-i", "foo", "blah"},
			[]string{"blah"},
		},
		{
			[]string{"--cat", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"-c", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--persist", "bar"},
			[]string{"bar"},
		},
		{
			[]string{"-p", "bar"},
			[]string{"bar"},
		},
	}

	cmdPrint := &Command{
		Use:   "print [string to print]",
		Short: "Print anything to the screen",
		Long:  `an utterly useless command for testing.`,
		Run: func(cmd *Command, args []string) {
			tp = args
		},
	}

	var flagi int
	var flagstr string
	var flagbool bool
	cmdPrint.PersistentFlags().BoolVarP(&flagbool, "persist", "p", false, "help for persistent one")
	cmdPrint.Flags().IntVarP(&flagi, "int", "i", 345, "help message for flag int")
	cmdPrint.Flags().StringVarP(&flagstr, "bar", "b", "bar", "help message for flag string")
	cmdPrint.Flags().BoolVarP(&flagbool, "cat", "c", false, "help message for flag bool")

	for _, test := range tests {
		output := stripFlags(test.input, cmdPrint)
		if !reflect.DeepEqual(test.output, output) {
			t.Errorf("expected: %v, got: %v", test.output, output)
		}
	}
}

func Test_DisableFlagParsing(t *testing.T) {
	as := []string{"-v", "-race", "-file", "foo.go"}
	targs := []string{}
	cmdPrint := &Command{
		DisableFlagParsing: true,
		Run: func(cmd *Command, args []string) {
			targs = args
		},
	}
	osargs := []string{"cmd"}
	os.Args = append(osargs, as...)
	err := cmdPrint.Execute()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(as, targs) {
		t.Errorf("expected: %v, got: %v", as, targs)
	}
}

func TestCommandsAreSorted(t *testing.T) {
	EnableCommandSorting = true

	originalNames := []string{"middle", "zlast", "afirst"}
	expectedNames := []string{"afirst", "middle", "zlast"}

	var tmpCommand = &Command{Use: "tmp"}

	for _, name := range originalNames {
		tmpCommand.AddCommand(&Command{Use: name})
	}

	for i, c := range tmpCommand.Commands() {
		if expectedNames[i] != c.Name() {
			t.Errorf("expected: %s, got: %s", expectedNames[i], c.Name())
		}
	}

	EnableCommandSorting = true
}

func TestEnableCommandSortingIsDisabled(t *testing.T) {
	EnableCommandSorting = false

	originalNames := []string{"middle", "zlast", "afirst"}

	var tmpCommand = &Command{Use: "tmp"}

	for _, name := range originalNames {
		tmpCommand.AddCommand(&Command{Use: name})
	}

	for i, c := range tmpCommand.Commands() {
		if originalNames[i] != c.Name() {
			t.Errorf("expected: %s, got: %s", originalNames[i], c.Name())
		}
	}

	EnableCommandSorting = true
}

func TestFlagErrorFunc(t *testing.T) {

	cmd := &Command{
		Use: "print",
		RunE: func(cmd *Command, args []string) error {
			return nil
		},
	}
	expectedFmt := "This is expected: %s"

	cmd.SetFlagErrorFunc(func(c *Command, err error) error {
		return fmt.Errorf(expectedFmt, err)
	})
	cmd.SetArgs([]string{"--bogus-flag"})
	cmd.SetOutput(new(bytes.Buffer))

	err := cmd.Execute()

	expected := fmt.Sprintf(expectedFmt, "unknown flag: --bogus-flag")
	if err.Error() != expected {
		t.Errorf("expected %v, got %v", expected, err.Error())
	}
}

func TestTraverseWithParentFlags(t *testing.T) {
	cmd := &Command{
		Use:              "do",
		TraverseChildren: true,
	}
	cmd.Flags().String("foo", "", "foo things")
	cmd.Flags().BoolP("goo", "g", false, "foo things")

	sub := &Command{Use: "next"}
	sub.Flags().String("add", "", "add things")
	cmd.AddCommand(sub)

	c, args, err := cmd.Traverse([]string{"-g", "--foo", "ok", "next", "--add"})
	if err != nil {
		t.Fatalf("Expected no error: %s", err)
	}
	if len(args) != 1 && args[0] != "--add" {
		t.Fatalf("wrong args %s", args)
	}
	if c.Name() != sub.Name() {
		t.Fatalf("wrong command %q expected %q", c.Name(), sub.Name())
	}
}

func TestTraverseNoParentFlags(t *testing.T) {
	cmd := &Command{
		Use:              "do",
		TraverseChildren: true,
	}
	cmd.Flags().String("foo", "", "foo things")

	sub := &Command{Use: "next"}
	sub.Flags().String("add", "", "add things")
	cmd.AddCommand(sub)

	c, args, err := cmd.Traverse([]string{"next"})
	if err != nil {
		t.Fatalf("Expected no error: %s", err)
	}
	if len(args) != 0 {
		t.Fatalf("wrong args %s", args)
	}
	if c.Name() != sub.Name() {
		t.Fatalf("wrong command %q expected %q", c.Name(), sub.Name())
	}
}

func TestTraverseWithBadParentFlags(t *testing.T) {
	cmd := &Command{
		Use:              "do",
		TraverseChildren: true,
	}
	sub := &Command{Use: "next"}
	sub.Flags().String("add", "", "add things")
	cmd.AddCommand(sub)

	expected := "got unknown flag: --add"

	c, _, err := cmd.Traverse([]string{"--add", "ok", "next"})
	if err == nil || strings.Contains(err.Error(), expected) {
		t.Fatalf("Expected error %s got %s", expected, err)
	}
	if c != nil {
		t.Fatalf("Expected nil command")
	}
}

func TestTraverseWithBadChildFlag(t *testing.T) {
	cmd := &Command{
		Use:              "do",
		TraverseChildren: true,
	}
	cmd.Flags().String("foo", "", "foo things")

	sub := &Command{Use: "next"}
	cmd.AddCommand(sub)

	// Expect no error because the last commands args shouldn't be parsed in
	// Traverse
	c, args, err := cmd.Traverse([]string{"next", "--add"})
	if err != nil {
		t.Fatalf("Expected no error: %s", err)
	}
	if len(args) != 1 && args[0] != "--add" {
		t.Fatalf("wrong args %s", args)
	}
	if c.Name() != sub.Name() {
		t.Fatalf("wrong command %q expected %q", c.Name(), sub.Name())
	}
}
