// main.go
// Author: gregory at 24-01-2017

package main 

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"fmt"
)

type settings struct {
	Password	string	
	Admin		string 	
}

func main() {
	println("Hello, world!")
	rawSettings, err := ioutil.ReadFile("settings.conf");
	if err != nil {
		println("Error reading configuration file")
	}

	s := settings{}
	err = yaml.Unmarshal([]byte(rawSettings), &s)
	if err != nil {
		println("Error parsing configuration")
		fmt.Printf("%+v\n", err)
	}

    fmt.Printf("--- settings:\n%v\n\n", s)
    fmt.Printf("MONKEY BANANAN\n")

}