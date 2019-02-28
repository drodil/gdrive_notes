package main

import(
    "fmt"
    "strings"
    "strconv"

    "github.com/fatih/color"
)

type NotesPrinter struct {
    UseColor bool
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

    inst.UseColor = config.Color
    inst.PrintHeader = true
    inst.SkipDone = true
    inst.ShowDone = true
    inst.ShowCreated = true
    inst.ShowUpdated = false
    inst.ShowPriority = config.UsePriority
    inst.ShowDue = config.UseDue
    inst.MaxTitleLength = 30
    inst.TimeFormat = config.TimeFormat
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
    c := color.New(color.Bold).Add(color.FgHiCyan)
    if !p.UseColor {
        c.DisableColor()
    }

    c.Print("ID\t")
    if p.ShowDone {
        c.Print("DONE\t")
    }

    format := "%-" + strconv.Itoa(p.MaxTitleLength) + "v"
    c.Printf(format, "TITLE")

    if p.ShowPriority {
        c.Print("PRIO\t")
    }
    if p.ShowDue {
        c.Print("DUE\t")
    }
    if p.ShowCreated {
        c.Print("CREATED\t")
    }
    if p.ShowUpdated {
        c.Print("UPDATED\t")
    }

    c.Print("\n")
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
        c := p.getPriorityColor(n)
        c.Print(n.Priority)
    }

    if p.ShowDue {
        fmt.Print("\t")
        fmt.Print(n.Due.Format(p.TimeFormat))
    }

    if p.ShowCreated {
        fmt.Print("\t")
        fmt.Print(n.Created.Format(p.TimeFormat))
    }

    if p.ShowUpdated {
        fmt.Print("\t")
        fmt.Print(n.Updated.Format(p.TimeFormat))
    }
}

func (p *NotesPrinter) getPriorityColor(n *Note) (*color.Color) {
    c := color.New()
    if !p.UseColor {
        c.DisableColor()
    }

    switch(n.Priority) {
        case 0:
            fallthrough
        case 1:
            c.Add(color.FgHiGreen)
            break
        case 2:
            fallthrough
        case 3:
            c.Add(color.FgHiYellow)
            break
        case 4:
            fallthrough
        case 5:
            c.Add(color.FgHiRed)
            break
    }
    return c
}

func (p *NotesPrinter) PrintFullNote(n *Note) {
    c := color.New(color.FgHiGreen).Add(color.Underline)
    c.Printf("NOTE %v\n", n.Id)
    if p.ShowPriority {
        fmt.Print("Priority: ")
        c = p.getPriorityColor(n)
        c.Printf("%v\n", n.Priority)
    }

    if p.ShowDue {
        fmt.Println("Due: " + n.Due.Format(p.TimeFormat))
    }
    fmt.Println("Created: " + n.Created.Format(p.TimeFormat))
    fmt.Println("Updated: " + n.Updated.Format(p.TimeFormat))
    fmt.Println("----------------")
    fmt.Print(n.Content)
    fmt.Println("----------------")
}
