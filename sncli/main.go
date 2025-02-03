	package main

	import (
	  "log"
	  "sncli/cmd"
	)

	func main() {
	  if err := cmd.Execute(); err != nil {
	    log.Fatal(err)
	  }
	}
