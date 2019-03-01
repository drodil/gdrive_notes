package main

import (
    "fmt"
    "log"
    "os"
    "errors"
    "strings"
    "strconv"
    "time"
)

func getNoteFromArg(arg string, n *Notes) (*Note) {
    id, err := strconv.ParseUint(arg, 0, 32)

    if err != nil {
        return nil
    }

    return n.FindNote(uint(id))
}

func handleListArgs(args []string, printer *NotesPrinter) {
    for i, arg := range args {
        if (arg == "--order" || arg == "-o") && len(args) > i + 1 {
            col := args[i+1]
            if strings.HasPrefix(col, "-") {
                printer.SortAsc = false
                col = col[1:]
            }
            printer.SortColumn = col
        }

        if (arg == "--search" || arg == "-s") && len(args) > i + 1 {
            printer.SearchStr = args[i+1]
        }

        if (arg == "--prio" || arg == "-p") && len(args) > i + 1 {
            prio, err := strconv.ParseUint(args[i+1], 0, 32)
            if err != nil {
                printer.PrioFilter = uint(prio)
            }
        }

        if (arg == "--tag" || arg == "-t") && len(args) > i + 1 {
            printer.TagFilter = args[i+1]
        }

        if arg == "-la" {
            printer.PrintDetails = true
        }
    }
}

func handleArgs(args []string, n *Notes, c *Configuration) (bool, error) {
    // TODO: Return nil instead error when GUI is available and start that
    if len(args) == 0 {
        return false, errors.New("Insufficient parameters")
    }

    command := args[0]
    args = args[1:]
    now := time.Now()

    switch command {
        // Quick add note
        case "qa":
            if len(args) < 1 {
                return false, errors.New("Missing note content")
            }

            content := strings.Join(args, " ")
            note := Note{Content: content, Priority: 0, Done: false, Created: now, Updated: now}
            id := n.AddNote(note)
            fmt.Printf("Added new note \"%v\" with id %v\n", note.GetTitle(), id)
            return true, nil

        case "a":
            fallthrough
        case "add":
            note := Note{Created: now, Updated: now, Priority: 0}
            updated, err := note.EditInEditor()
            if err != nil {
                return false, err
            }

            if updated {
                id := n.AddNote(note)
                fmt.Printf("Added new note \"%v\" with id %v\n", note.GetTitle(), id)
            }

            return updated, nil

        // Clear all notes
        case "clear":
            for {
                ret, err := YesNoQuestion("Are you sure you want to delete all notes [y/n]? ")
                if err == nil {
                    if ret {
                        n.Notes = n.Notes[:0]
                        fmt.Println("All notes have been deleted")
                        return true, nil
                    }
                    return false, nil
                }
            }

        // List all notes
        // TODO: Additional parameters for ordering etc.
        case "list":
            fallthrough
        case "ls":
            printer := NewNotesPrinter(c)
            printer.SkipDone = false
            handleListArgs(args, &printer)
            printer.Print(n)
            return false, nil

        // List only not done notes
        // TODO: Additional parameters for ordering etc.
        case "td":
            fallthrough
        case "todo":
            printer := NewNotesPrinter(c)
            handleListArgs(args, &printer)
            printer.ShowDone = false
            printer.SkipDone = true
            printer.Print(n)
            return false, nil

        // Mark done by ID
        case "done":
            fallthrough
        case "md":
            if len(args) < 1 {
                return false, errors.New("Give note id")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }

            note.Done = true
            note.Updated = now
            fmt.Printf("Note \"%v\" with id %v is now done\n", note.GetTitle(), note.Id)
            return true, nil

        case "e":
            fallthrough
        case "edit":
             if len(args) < 1 {
                return false, errors.New("Give note id")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }

            return note.EditInEditor()

        case "p":
            fallthrough
        case "prio":
            if len(args) < 2 {
                return false, errors.New("Give note id and new priority")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }

            prio, err := strconv.ParseUint(args[1], 0, 32)
            if err != nil {
                return false, errors.New("Invalid priority given")
            }
            note.Priority = uint(prio)
            return true, nil

        case "s":
            fallthrough
        case "show":
            if len(args) < 1 {
                return false, errors.New("Give note id")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }

            printer := NewNotesPrinter(c)
            printer.PrintFullNote(note)

            return false, nil

        case "rm":
            fallthrough
        case "remove":
            if len(args) < 1 {
                return false, errors.New("Give note id")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }

            id := note.Id
            title := note.GetTitle()
            err := n.DeleteNote(note.Id)
            if err != nil {
                return false, err
            }

            fmt.Printf("Removed note \"%v\" with id %v\n", title, id)
            return true, nil

        case "ct":
            fallthrough
        case "ctags":
            if len(args) < 1 {
                return false, errors.New("Give note id")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }
            note.ClearTags()
            fmt.Printf("Cleared all tags from note %v\n", note.Id)
            return true, nil

        case "t":
            fallthrough
        case "tag":
            if len(args) < 2 {
                return false, errors.New("Give note id and the tag")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }


            ret := note.AddTag(args[1])
            if ret {
                fmt.Printf("Added tag \"%v\" for note %v\n", args[1], note.Id)
            } else {
                fmt.Printf("Note %v already had tag \"%v\"\n", note.Id, args[1])
            }
            return ret, nil

        case "rt":
            fallthrough
        case "rtag":
            if len(args) < 2 {
                return false, errors.New("Give note id and the tag")
            }

            note := getNoteFromArg(args[0], n)
            if note == nil {
                return false, errors.New("Could not find note with id")
            }


            ret := note.RemoveTag(args[1])
            if ret {
                fmt.Printf("Removed tag \"%v\" for note %v\n", args[1], note.Id)
            } else {
                fmt.Printf("Note %v does not have tag \"%v\"\n", note.Id, args[1])
            }
            return ret, nil


        case "h":
            fallthrough
        case "help":
            printHelp(nil)
            return false, nil

        case "config":
            c.Configure()
            return false, nil
    }

    return false, errors.New("Invalid command")
}

func printHelp(err error) {
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println("------------------")
    fmt.Println("Google Drive Notes")
    fmt.Println("------------------")
    fmt.Println("")
    fmt.Println("Available commands are:")
    fmt.Println("")
    fmt.Println("GENERAL:")
    fmt.Println("h|help\t\t\tPrint this help")
    fmt.Println("config\t\t\tConfigure the look&feel")
    fmt.Println("")
    fmt.Println("ADDING / EDITING:")
    fmt.Println("qa <note>\t\tQuickly add note with default values")
    fmt.Println("e|edit <id>\t\tEdit note with given id")
    fmt.Println("a|add\t\t\tAdd new note with $EDITOR")
    fmt.Println("md|done <id>\t\tMark note done with given id")
    fmt.Println("p|prio <id> <prio>\tSet priority of the note")
    fmt.Println("t|tag <id> <tag>\tAdd tag for note with given id")
    fmt.Println("rt|rtag <id> <tag>\tRemote tag from note with given id")
    fmt.Println("ct|ctags <id>\t\tRemove all tags from note with given id")
    fmt.Println("")
    fmt.Println("DELETING:")
    fmt.Println("clear\t\t\tDelete all notes")
    fmt.Println("rm|remove <id>\t\tRemove note with given id")
    fmt.Println("")
    fmt.Println("SHOWING:")
    fmt.Println("ls|list\t\t\tList all notes")
    fmt.Println("td|todo\t\t\tList all not-done notes")
    fmt.Println("s|show <id>\t\tShow note contents with given id")
    fmt.Println("")
    fmt.Println("Additional parameters for listing:")
    fmt.Println("--order|-o <column>\tOrder by column. Has to be one of the following:")
    fmt.Println("\t\t\ttitle,prio,created,updated,due")
    fmt.Println("--search|-s <string>\tSearch for notes with given content")
    fmt.Println("--prio|-p <int>\tSearch for notes with this or greater priority")
    fmt.Println("--tag|-t <tag>\tSearch for notes with this tag")
    fmt.Println("-la\t\tPrint whole notes instead table")
}

func main() {
    config := NewConfiguration()
    err := config.Init()
    if err != nil {
        log.Fatalf("Could not set up configuration: %v", err)
        os.Exit(1)
    }

    notes := Notes{}
    err = notes.Init(&config)
    if err != nil {
        log.Fatalf("Could not set up Google Drive: %v", err)
        os.Exit(1)
    }

    update, err := handleArgs(os.Args[1:], &notes, &config)
    if err != nil {
        printHelp(err)
        os.Exit(0)
    }

    if update {
       err = notes.SaveNotes()
        if err != nil {
            log.Fatalf("Could not sync notes to Drive: %v", err)
            os.Exit(1)
        }
    }
}
