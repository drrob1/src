package main

// I can't get this import to work.

func main() {
	var opts struct {
		// slice of bool will append true every time the option appears, like -vvv will be {true,true,true}
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information."`

		// automatic marshalling to desired type
		Offset uint `long:"offset" description:"Offset."`

		// example of a callback that's called each time the option is found
		Call func(string) `short:"c" description:"Call phone number"`

		// Example of a required flag
		Name string `short:"n" long:"name" description:"A name" required:"true"`

		// Example of a flag restricted to a predefined set of strings
		Animal string `long:"animal" choice:"cat" choice:"dog" choice:"moose"`

		// Example of a value name

	}

}
