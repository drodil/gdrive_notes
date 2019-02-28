package main

import(
    "fmt"
    "strings"
    "strconv"
)

type NotesPrinter struct {
    PrintHeader bool
    SkipDone bool
    ShowDone bool
    ShowCreated bool
    ShowUpdated bool
    ShowPriority bool
    ShowDue bool
    MaxTitleLength int
    TimeFormat string
}

func NewNotesPrinter(config *Configuration) (NotesPrinter) {
    inst := NotesPrinter{}

    inst.PrintHeader = true
    inst.SkipDone = true
    inst.ShowDone = true
    inst.ShowCreated = true
    inst.ShowUpdated = false
    inst.ShowPriority = true
    inst.ShowDue = false
    inst.MaxTitleLength = 30
    inst.TimeFormat = "02.01.2006 15:04"
    return inst
}

func (p *NotesPrinter) Print(n *Notes) {
    if p.PrintHeader {
        p.printHeader()
    }

    for _, note := range n.Notes {
        if p.SkipDone && note.Done {
            continue
        }
        p.PrintNote(&note)
        fmt.Print("\n")
    }
}

func (p *NotesPrinter) printHeader() {
    fmt.Print("ID\t")
    if p.ShowDone {
        fmt.Print("DONE\t")
    }

    format := "%-" + strconv.Itoa(p.MaxTitleLength) + "v"
    fmt.Printf(format, "TITLE")

    if p.ShowPriority {
        fmt.Print("PRIO\t")
    }
    if p.ShowDue {
        fmt.Print("DUE\t")
    }
    if p.ShowCreated {
        fmt.Print("CREATED\t")
    }
    if p.ShowUpdated {
        fmt.Print("UPDATED\t")
    }

    fmt.Print("\n")
}

func (p *NotesPrinter) PrintNote(n *Note) {
    fmt.Print(n.Id)
    fmt.Print("\t")

    if p.ShowDone {
        if !n.Done {
            fmt.Print("[ ]")
        } else {
            fmt.Print("[x]")
        }
        fmt.Print("\t")
    }

    fmt.Print(" ")
    parts := strings.Split(n.Content, "\n")
    preview := parts[0]
    if len(preview) > (p.MaxTitleLength - 3) {
        preview = preview[0:(p.MaxTitleLength-3)] + "..."
    }
    format := "%-" + strconv.Itoa(p.MaxTitleLength) + "v"

    fmt.Printf(format, preview)
    if p.ShowPriority {
        fmt.Print("\t")
        fmt.Print(n.Priority)
    }

    if p.ShowDue {
        fmt.Print("\t")
        // TODO: Create configuration to allow user select format of dates
        fmt.Print(n.Due.Format(p.TimeFormat))
    }

    if p.ShowCreated {
        fmt.Print("\t")
        // TODO: Create configuration to allow user select format of dates
        fmt.Print(n.Created.Format(p.TimeFormat))
    }

    if p.ShowUpdated {
        fmt.Print("\t")
        // TODO: Create configuration to allow user select format of dates
        fmt.Print(n.Updated.Format(p.TimeFormat))
    }
}
