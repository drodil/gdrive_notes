package main

import (
    "fmt"
    "strings"

    "github.com/jroimartin/gocui"
    "github.com/fatih/color"
)

const (
    LIST_VIEW = "list"
    PREVIEW_VIEW = "preview"
    COMMAND_VIEW = "cmd"
)

type NotesGui struct {
    Notes *Notes
    Config *Configuration
    preview bool
    idx int
    cmd string
    errString string
}

func (n *NotesGui) Start() (error) {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        return err
    }

    defer g.Close()
    g.SetManagerFunc(n.layout)
    g.InputEsc = true

    // Moving around in list view
    err = g.SetKeybinding(LIST_VIEW, 'j', gocui.ModNone, n.increaseIndex)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, 'k', gocui.ModNone, n.decreaseIndex)
    if err != nil {
        return err
    }

    // Starts vim like command
    err = g.SetKeybinding(LIST_VIEW, ':', gocui.ModNone, n.startCommand)
    if err != nil {
        return err
    }

    // Handles vim like command arguments for example ':q'
    for _, char := range "qwertyuiopasdfghjklzxcvbnm" {
        f := func(char rune) func(*gocui.Gui, *gocui.View) error {
            return func(g *gocui.Gui, v *gocui.View) error {
                n.cmd += string(char)
                return n.update(g)
            }
        }

        err = g.SetKeybinding(COMMAND_VIEW, char, gocui.ModNone, f(char))
        if err != nil {
            return err
        }
    }

    // Handles enter on commands
    err = g.SetKeybinding(COMMAND_VIEW, gocui.KeyEnter, gocui.ModNone, n.executeCommand)
    if err != nil {
        return err
    }

    // Exits command mode
    err = g.SetKeybinding(COMMAND_VIEW, gocui.KeyEsc, gocui.ModNone, n.cancelCommand)
    if err != nil {
        return err
    }

    // TODO: Ordering
    // TODO: Filtering
    // TODO: Moving faster around
    // TODO: Showing the note in the PREVIEW_VIEW
    // TODO: Showing the note in full screen (or in $EDITOR)

    g.Update(n.update)
    err = g.MainLoop()
    if err != nil && err != gocui.ErrQuit {
        return err
    }

    return nil
}

func (n *NotesGui) layout(g *gocui.Gui) (error) {
    // TODO: Add some colors
    maxX, maxY := g.Size()
    _, err := g.SetView(COMMAND_VIEW, 0, maxY-2, maxX, maxY)
    if err != nil && err != gocui.ErrUnknownView {
        return err
    }

    v, err := g.View(COMMAND_VIEW)
    if err != nil {
        return err
    }
    v.Frame = false

    _, err = g.SetView(LIST_VIEW, 0, 0, (2*maxX/3)-1, maxY-2)
    if err != nil && err != gocui.ErrUnknownView {
        return err
    }

    _, err = g.SetView(PREVIEW_VIEW, 2*maxX/3, 0, maxX-2, maxY-2)
    if err != nil && err != gocui.ErrUnknownView {
        return err
    }

    return nil
}

func (n *NotesGui) increaseIndex(g *gocui.Gui, v *gocui.View) error {
    n.idx++
    if n.idx >= len(n.Notes.Notes) {
        n.idx = 0
    }
    return n.update(g)
}

func (n *NotesGui) decreaseIndex(g *gocui.Gui, v *gocui.View) error {
    n.idx--
    if n.idx < 0 {
        n.idx = len(n.Notes.Notes) - 1
    }
    return n.update(g)
}

func (n *NotesGui) startCommand(g *gocui.Gui, v *gocui.View) error {
    n.cmd = ":"
    _, err := g.SetCurrentView(COMMAND_VIEW)
    if err != nil {
        return err
    }
    return n.update(g)
}

func (n *NotesGui) executeCommand(g *gocui.Gui, v *gocui.View) error {
    command := n.cmd[1:]
    n.cmd = ""
    switch(command) {
        case "q":
            fallthrough
        case "wq":
            // TODO: Separate plain q and qw to write changes to the gdrive
            return gocui.ErrQuit
        // TODO: Handle help
        // TODO: Handle adding
        // TODO: Handle editing
        // TODO: Handle deleting
        default:
            if len(command) > 0 {
                n.errString = "Invalid command. Use :h to show help"
            }
    }

    _, err := g.SetCurrentView(LIST_VIEW)
    if err != nil {
        return err
    }
    return n.update(g)
}

func (n *NotesGui) cancelCommand(g *gocui.Gui, v *gocui.View) error {
    n.cmd = ""
    _, err := g.SetCurrentView(LIST_VIEW)
    if err != nil {
        return err
    }
    return n.update(g)
}

func (n *NotesGui) update(g *gocui.Gui) error {
    if len(n.cmd) == 0 {
        _, verr := g.SetCurrentView(LIST_VIEW)
        if verr != nil {
            return verr
        }
    }

    v, err := g.View(LIST_VIEW)
    if err != nil {
        return err
    }
    v.Clear()
    v.Title = "Notes"
    selected := n.Notes.Notes[n.idx]
    for i, note := range n.Notes.Notes {
        if i == n.idx {
            c := color.New(color.Bold).Add(color.BgWhite).Add(color.FgBlack)
            c.Fprintln(v, note.GetTitle())
            continue
        }
        fmt.Fprintln(v, note.GetTitle())
    }

    pv, err := g.View(PREVIEW_VIEW)
    if err != nil {
        return err
    }

    // TODO: Separate updating different views in multiple functions
    // TODO: Add some color here
    pv.Clear()
    fmt.Fprintln(pv, "ID:       ", selected.Id)
    if n.Config.UsePriority {
        fmt.Fprintln(pv, "Priority: ", selected.Priority)
    }

    if n.Config.UseDue {
        if !selected.Due.IsZero() {
            fmt.Fprintln(pv, "Due:       ", selected.Due.Format(n.Config.TimeFormat))
        }
    }

    if len(selected.Tags) > 0 {
        fmt.Fprintln(pv, "Tags:     ", strings.Join(selected.Tags, ", "))
    }

    fmt.Fprintln(pv, "Created:  ", selected.Created.Format(n.Config.TimeFormat))
    fmt.Fprintln(pv, "Updated:  ", selected.Updated.Format(n.Config.TimeFormat))

    cv, err := g.View(COMMAND_VIEW)
    if err != nil {
        return err
    }
    cv.Clear()
    if len(n.cmd) > 0 {
        fmt.Fprintln(cv, n.cmd)
    } else if len(n.errString) > 0 {
        fmt.Fprintln(cv, n.errString)
        n.errString = ""
    }

    return nil
}

func (n *NotesGui) quit(g *gocui.Gui, v *gocui.View) error {
    return gocui.ErrQuit
}
