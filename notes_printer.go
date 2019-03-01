package main

import(
    "fmt"
    "strings"
    "strconv"
    "sort"
    "time"

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
    SortColumn string
    SortAsc bool
    SearchStr string
    PrioFilter uint
    TagFilter string
    PrintDetails bool
    idSize int
    doneSize int
    titleSize int
    prioSize int
    timeSize int
}

func NewNotesPrinter(config *Configuration) (NotesPrinter) {
    inst := NotesPrinter{}

    inst.UseColor = config.Color
    inst.PrintHeader = true
    inst.SkipDone = true
    inst.ShowDone = true
    inst.ShowCreated = false
    inst.ShowUpdated = false
    inst.ShowPriority = config.UsePriority
    inst.ShowDue = config.UseDue
    inst.TimeFormat = config.TimeFormat
    inst.SortColumn = "Id"
    inst.SortAsc = true
    inst.SearchStr = ""
    inst.TagFilter = ""
    inst.PrioFilter = 0
    inst.PrintDetails = false

    inst.idSize = 6
    inst.doneSize = 6
    inst.titleSize = 30
    inst.timeSize = 12
    inst.prioSize = 4

    return inst
}

func (p *NotesPrinter) calculateColumnWidths(n *Notes) {
    now := time.Now()
    p.timeSize = len(now.Format(p.TimeFormat)) + 2
    p.idSize = len(string(n.GetMaxId())) + 3

    w := GetScreenWidth() - 2
    w -= p.idSize
    if p.ShowDone {
        w -= p.doneSize
    }
    if p.ShowPriority {
        w -= p.prioSize
    }
    if p.ShowDue {
        w -= p.timeSize
    }
    if p.ShowCreated {
        w -= p.timeSize
    }
    if p.ShowUpdated {
        w -= p.timeSize
    }
    p.titleSize = w
}

func (p *NotesPrinter) Print(n *Notes) {
    p.calculateColumnWidths(n)
    sort.Slice(n.Notes, func(i, j int) bool {
        ret := false
        switch(p.SortColumn) {
            case "prio":
                p.ShowPriority = true
                ret = n.Notes[i].Priority < n.Notes[j].Priority
                break
            case "title":
                ret = n.Notes[i].GetTitle() < n.Notes[j].GetTitle()
                break
            case "due":
                p.ShowDue = true
                ret = n.Notes[i].Due.Unix() < n.Notes[j].Due.Unix()
                break
            case "created":
                p.ShowCreated = true
                ret = n.Notes[i].Created.Unix() < n.Notes[j].Created.Unix()
                break
            case "updated":
                p.ShowUpdated = true
                ret = n.Notes[i].Updated.Unix() < n.Notes[j].Updated.Unix()
                break
            default:
                ret = n.Notes[i].Id < n.Notes[j].Id
        }

        if !p.SortAsc {
            return !ret
        }
        return ret
    })

    if p.PrintDetails {
        p.PrintHeader = false
    }

    if p.PrintHeader {
        p.printHeader()
    }

    notesPrinted := false
    for _, note := range n.Notes {
        if p.SkipDone && note.Done {
            continue
        }
        if len(p.SearchStr) > 0 && !strings.Contains(note.Content, p.SearchStr) {
            continue
        }
        if p.ShowPriority && note.Priority < p.PrioFilter {
            continue
        }
        if len(p.TagFilter) > 0 && !note.HasTag(p.TagFilter) {
            continue
        }

        if p.PrintDetails {
           p.PrintFullNote(&note)
        } else {
            p.PrintNote(&note)
            fmt.Print("\n")
        }
        notesPrinted = true
    }

    if !notesPrinted {
        fmt.Println("No notes")
    }

    PrintVerticalLine()
}

func (p *NotesPrinter) printHeader() {
    c := color.New(color.Bold).Add(color.FgHiCyan)
    if !p.UseColor {
        c.DisableColor()
    }

    PrintVerticalLine()

    c.Printf(" %-" + strconv.Itoa(p.idSize) + "v", "ID")
    if p.ShowDone {
        c.Printf("%-6v", "DONE")
    }

    format := "%-" + strconv.Itoa(p.titleSize) + "v"
    c.Printf(format, "TITLE")

    if p.ShowPriority {
        c.Printf("%-" + strconv.Itoa(p.prioSize) + "v", "PRIO")
    }
    if p.ShowDue {
        c.Printf("%-" + strconv.Itoa(p.timeSize) + "v", "DUE")
    }
    if p.ShowCreated {
        c.Printf("%-" + strconv.Itoa(p.timeSize) + "v", "CREATED")
    }
    if p.ShowUpdated {
        c.Printf("%-" + strconv.Itoa(p.timeSize) + "v", "UPDATED")
    }

    c.Print("\n")
    PrintVerticalLine()
}

func (p *NotesPrinter) PrintNote(n *Note) {
    fmt.Printf(" %-" + strconv.Itoa(p.idSize) + "v", n.Id)

    if p.ShowDone {
        if !n.Done {
            fmt.Printf("%-" + strconv.Itoa(p.doneSize) + "v", "[ ]")
        } else {
            fmt.Printf("%-" + strconv.Itoa(p.doneSize) + "v", "[x]")
        }
    }

    parts := strings.Split(n.Content, "\n")
    preview := parts[0]
    if len(preview) > (p.titleSize - 3) {
        preview = preview[0:(p.titleSize-3)] + "..."
    }
    format := "%-" + strconv.Itoa(p.titleSize) + "v"

    fmt.Printf(format, preview)
    if p.ShowPriority {
        c := p.getPriorityColor(n)
        c.Printf("%-" + strconv.Itoa(p.prioSize) + "v", n.Priority)
    }

    if p.ShowDue {
        due := ""
        if !n.Due.IsZero() {
           due = n.Due.Format(p.TimeFormat)
        }
        fmt.Printf("%-" + strconv.Itoa(p.timeSize) + "v", due)
    }

    if p.ShowCreated {
        created := ""
        if !n.Created.IsZero() {
            created = n.Created.Format(p.TimeFormat)
        }
        fmt.Printf("%-" + strconv.Itoa(p.timeSize) + "v", created)
    }

    if p.ShowUpdated {
        updated := ""
        if !n.Updated.IsZero() {
            updated = n.Updated.Format(p.TimeFormat)
        }
        fmt.Printf("%-" + strconv.Itoa(p.timeSize) + "v", updated)
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
    PrintVerticalLine()
    c.Printf("NOTE %v\n\n", n.Id)
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
    fmt.Println("Tags: " + strings.Join(n.Tags, ", "))
    fmt.Print("\n")
    fmt.Print(n.Content)
    fmt.Print("\n")
    PrintVerticalLine()
}
