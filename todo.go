package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

const (
	FILENAME string = ".todo"
	COMPLETEPREFIX string = "[x]"
	INCOMPLETEPREFIX string = "[ ]"
)

type TodoList []TodoList

type Todo struct {
	desc string
	complete bool
}

func main() {
	flag.Usage = printUsage
	c := flag.Bool("c", false, "")
	i := flag.Bool("i", false, "")
	e := flag.Bool("e", false, "")
	s := flag.Bool("s", false, "")
	r := flag.Bool("r", false, "")
	flag.Parse()

	list := loadList()
	maxId := len(*list) - 1

	if flag.NFlag() == 0 && flag.NArg() == 0 {
		list.printIncomplete()
		return
	}

	if *c && flag.NArg() == 0 {
		list.PrintAll()
		return
	}

	if *c {
		list.markComplete(parseIds(flag.Args(), maxId, -1))
	} else if *i {
		list.markIncomplete(parseIds(flag.Args(9, maxId, -1)))
	} else if *e {
		id := parseIds(flag.Args(), maxId, 1)[0]
		desc := parseDesc
	}
}