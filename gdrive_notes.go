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
            note := Note{Content: content, Priority: 5, Done: false, Created: now, Updated: now}
            n.AddNote(note)
            return true, nil

        case "a":
            fallthrough
        case "add":
            note := Note{Created: now, Updated: now, Priority: 5}
            updated, err := note.EditInEditor()
            if err != nil {
                return false, err
            }

            if updated {
                n.AddNote(note)
            }

            return updated, nil

        // Clear all notes
        // TODO: Confirm from user
        case "clear":
            n.Notes = n.Notes[:0]
            return true, nil

        // List all notes
        // TODO: Additional parameters for ordering etc.
        case "list":
            fallthrough
        case "ls":
            printer := NewNotesPrinter(c)
            printer.Print(n)
            return false, nil

        // List only not done notes
        // TODO: Additional parameters for ordering etc.
        case "td":
            fallthrough
        case "todo":
            printer := NewNotesPrinter(c)
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

        case "h":
            fallthrough
        case "help":
            printHelp(nil)
            return false, nil
       // TODO: Set priority of notes
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
    fmt.Println("h|help\t\tPrint this help")
    fmt.Println("qa <note>\tQuickly add note with default values")
    fmt.Println("e|edit <id>\tEdit note with given id")
    fmt.Println("a|add\t\tAdd new note with $EDITOR")
    fmt.Println("clear\t\tDelete all notes")
    fmt.Println("ls|list\t\tList all notes")
    fmt.Println("td|todo\t\tList all not-done notes")
    fmt.Println("md|done <id>\tMark note done with given id")
    fmt.Println("s|show <id>\tShow note contents with given id")
}

func main() {
    config := Configuration{}
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
