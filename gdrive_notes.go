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

func handleArgs(args []string, n *Notes) (bool, error) {
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
            for _, note := range n.Notes {
                note.Print()
                fmt.Print("\n")
            }
            return false, nil

        // List only not done notes
        // TODO: Additional parameters for ordering etc.
        case "td":
            fallthrough
        case "todo":
            for _, note := range n.Notes {
                if note.Done == true {
                    continue
                }
                note.Print()
                fmt.Print("\n")
            }

            return false, nil

        // Mark done by ID
        case "done":
            fallthrough
        case "md":
            if len(args) < 1 {
                return false, errors.New("Give note id")
            }

            id, err := strconv.ParseUint(args[0], 0, 32)

            if err != nil {
                return false, err
            }

            note := n.FindNote(uint(id))
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

            id, err := strconv.ParseUint(args[0], 0, 32)

            if err != nil {
                return false, err
            }

            note := n.FindNote(uint(id))
            if note == nil {
                return false, errors.New("Could not find note with id")
            }

            return note.EditInEditor()

       // TODO: Set priority of notes
    }

    return false, errors.New("Invalid command")
}

func printHelp() {
    fmt.Println("Google Drive Notes")
    fmt.Println("------------------")
    // TODO: Create help for commands
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

    update, err := handleArgs(os.Args[1:], &notes)
    if err != nil {
        printHelp()
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
