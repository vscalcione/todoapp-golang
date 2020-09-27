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

	/** Stored in user's home directory **/

	FILENAME         string = ".todo"
	COMPLETEPREFIX   string = "[x] "
	INCOMPLETEPREFIX string = "[ ] "
)

type TodoList []Todo

type Todo struct {
	desc     string // todo description
	complete bool   // whether the todo was completed or not
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
	maxID := len(*list) - 1

	if flag.NFlag() == 0 && flag.NArg() == 0 {
		list.printIncomplete() // default when no args
		return
	}
	if *c && flag.NArg() == 0 {
		list.printAll() // -c
		return
	}

	if *c {
		list.markComplete(parseIds(flag.Args(), maxID, -1)) // -c id...

	} else if *i {
		list.markIncomplete(parseIds(flag.Args(), maxID, -1)) // -i id...

	} else if *e {
		id := parseIds(flag.Args(), maxID, 1)[0]
		desc := parseDesc(flag.Args()[1:])
		split := strings.Split(desc, "/")
		if len(split) == 4 && split[0] == "" && split[3] == "" {
			list.replace(id, split[1], split[2]) // -e id /sub/rep/
		} else {
			list.edit(id, desc) // -e id desc...
		}

	} else if *s {
		ids := parseIds(flag.Args(), maxID, 2)
		list.swap(ids[0], ids[1]) // -s id id

	} else if *r {
		if flag.NArg() == 0 {
			list.removeComplete() // -r
		} else {
			list.remove(parseIds(flag.Args(), maxID, -1)) // -r id...
		}

	} else {
		list.add(parseDesc(flag.Args())) // desc...
	}

	list.printIncomplete()
	list.save()
}

func printUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), `usage: %v [-c|-i|-e|-s|-r|-h] [id...] [desc...]
 command line todo list
 todos are stored at %v
   desc...          add new todo
   -c               print also completed todos
   -c id...         mark specified todos as complete
   -i id...         mark specified todos as incomplete
   -e id desc...    edit description of specified todo
   -e id /sub/rep/  replace substring sub with rep in description of specified todo
   -s id id         swap position of specified todos
   -r               remove completed todos
   -r id...         remove specified todos
   -h               show usage message
 repo: https://github.com/vscalcione/todoapp-go
 `, os.Args[0], buildFilepath())
}

func buildFilepath() string {
	usr, err := user.Current()
	if err != nil {
		die("unable to get user's home directory")
	}
	return usr.HomeDir + "/" + FILENAME
}

func parseIds(ss []string, maxID, expected int) []int {
	var ids []int
	for i, s := range ss {
		if expected >= 0 && i == expected {
			break
		}
		id, err := strconv.ParseInt(s, 10, 0)
		if err != nil || id < 0 || int(id) > maxID {
			die("invalid id \"%v\"", s)
		}
		ids = append(ids, int(id))
	}
	if expected >= 0 && len(ids) != expected {
		die("expected %v ids but got %v", expected, len(ids))
	}
	return ids
}

func parseDesc(ss []string) string {
	if len(ss) == 0 {
		die("missing description")
	}
	return strings.Join(ss, " ")
}

func loadList() *TodoList {
	f, err := os.OpenFile(buildFilepath(), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		die(err.Error())
	}
	defer f.Close()

	var list TodoList
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		desc := scanner.Text()
		// lines without PREFIX are assumed to be incomplete todos
		complete := strings.HasPrefix(desc, COMPLETEPREFIX)
		if complete || strings.HasPrefix(desc, INCOMPLETEPREFIX) {
			desc = desc[len(COMPLETEPREFIX):] // both PREFIX have same len
		}
		list = append(list, Todo{desc, complete})
	}
	return &list
}

func (list *TodoList) add(desc string) {
	*list = append(*list, Todo{desc, false})
}

func (list *TodoList) edit(id int, desc string) {
	(*list)[id].desc = desc
}

func (list *TodoList) markComplete(ids []int) {
	for _, id := range ids {
		(*list)[id].complete = true
	}
}

func (list *TodoList) markIncomplete(ids []int) {
	for _, id := range ids {
		(*list)[id].complete = false
	}
}

func (list *TodoList) printAll() {
	for id, todo := range *list {
		prefix := INCOMPLETEPREFIX
		if todo.complete {
			prefix = COMPLETEPREFIX
		}
		fmt.Printf("%2v %v%v\n", id, prefix, todo.desc)
	}
}

func (list *TodoList) printIncomplete() {
	for id, todo := range *list {
		if !todo.complete {
			fmt.Printf("%2v %v%v\n", id, INCOMPLETEPREFIX, todo.desc)
		}
	}
}

func (list *TodoList) remove(ids []int) {
	list.removeIf(func(id int, todo Todo) bool {
		for _, aid := range ids {
			if id == aid {
				return todo.complete || askYesNo("todo %v is incomplete. Remove it?", id)
			}
		}
		return false
	})
}

func (list *TodoList) removeComplete() {
	if askYesNo("remove all completed todos?") {
		list.removeIf(func(id int, todo Todo) bool {
			return todo.complete
		})
	}
}

func (list *TodoList) removeIf(predicate func(int, Todo) bool) {
	newLen := 0
	for id := 0; id < len(*list); id++ {
		if !predicate(id, (*list)[id]) {
			(*list)[newLen] = (*list)[id]
			newLen++
		}
	}
	*list = (*list)[:newLen]
}

func askYesNo(question string, a ...interface{}) bool {
	fmt.Printf(question, a...)
	fmt.Print(" [y/N]: ")
	var ans string
	_, err := fmt.Scanln(&ans)
	return err == nil && (strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes"))
}

func (list *TodoList) replace(id int, sub, rep string) {
	(*list)[id].desc = strings.Replace((*list)[id].desc, sub, rep, -1)
}

func (list *TodoList) swap(id1 int, id2 int) {
	tmp := (*list)[id1]
	(*list)[id1] = (*list)[id2]
	(*list)[id2] = tmp
}

func (list *TodoList) save() {
	f, err := os.Create(buildFilepath())
	if err != nil {
		die(err.Error())
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, todo := range *list {
		if todo.complete {
			w.WriteString(COMPLETEPREFIX)
		} else {
			w.WriteString(INCOMPLETEPREFIX)
		}
		w.WriteString(todo.desc)
		w.WriteByte('\n')
	}
	w.Flush()
}

func die(reason string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, reason, a...)
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}
