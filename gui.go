package main

import (
    "fmt"
    "strings"
    "time"

    "github.com/jroimartin/gocui"
    "github.com/fatih/color"
    "github.com/nsf/termbox-go"
)

const (
    LIST_VIEW = "list"
    PREVIEW_VIEW = "preview"
    COMMAND_VIEW = "cmd"
    HELP_VIEW = "help"
)

type NotesGui struct {
    Notes *Notes
    Config *Configuration
    preview bool
    idx int
    cmd string
    statusString string
    showNoteContent bool
    SaveModifications bool
    unsavedModifications bool
}

func (n *NotesGui) Start() (error) {
    g, err := gocui.NewGui(gocui.OutputNormal)
    if err != nil {
        return err
    }

    defer g.Close()
    g.SetManagerFunc(n.layout)
    g.InputEsc = true
    g.Cursor = false
    g.Mouse = false

    // Moving around in list view
    err = g.SetKeybinding(LIST_VIEW, 'j', gocui.ModNone, n.increaseIndex)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, 'k', gocui.ModNone, n.decreaseIndex)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, 'G', gocui.ModNone, n.gotoBottom)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, 'e', gocui.ModNone, n.editNote)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, 'D', gocui.ModNone, n.deleteNote)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, 'a', gocui.ModNone, n.addNote)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, gocui.KeySpace, gocui.ModNone, n.toggleDone)
    if err != nil {
        return err
    }

    err = g.SetKeybinding(LIST_VIEW, gocui.KeyEnter, gocui.ModNone, n.toggleContent)
    if err != nil {
        return err
    }

    // Starts vim like command
    err = g.SetKeybinding(LIST_VIEW, ':', gocui.ModNone, n.startCommand)
    if err != nil {
        return err
    }

    // Backspace command
    err = g.SetKeybinding(COMMAND_VIEW, gocui.KeyBackspace, gocui.ModNone, n.backspaceCommand)
    if err != nil {
        return err
    }
    err = g.SetKeybinding(COMMAND_VIEW, gocui.KeyBackspace2, gocui.ModNone, n.backspaceCommand)
    if err != nil {
        return err
    }

    // Handles vim like command arguments for example ':q'
    for _, char := range "qwertyuiopasdfghjklzxcvbnm!" {
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

    _, err = g.SetView(LIST_VIEW, 0, 0, maxX/2-1, maxY-2)
    if err != nil && err != gocui.ErrUnknownView {
        return err
    }

    _, err = g.SetView(PREVIEW_VIEW, maxX/2, 0, maxX-2, maxY-2)
    if err != nil && err != gocui.ErrUnknownView {
        return err
    }

    v, err = g.View(PREVIEW_VIEW)
    if err != nil {
        return err
    }
    v.Wrap = true
    v.Autoscroll = true

    return nil
}

func (n *NotesGui) toggleContent(g *gocui.Gui, v *gocui.View) error {
    n.showNoteContent = !n.showNoteContent
    return n.update(g)
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

func (n *NotesGui) gotoBottom(g *gocui.Gui, v *gocui.View) error {
    if len(n.Notes.Notes) > 0 {
       n.idx = len(n.Notes.Notes) - 1
    }
    return n.update(g)
}

func (n *NotesGui) editNote(g *gocui.Gui, v *gocui.View) error {
    selected := &n.Notes.Notes[n.idx]
    modified, err := selected.EditInEditor()
    if err != nil {
        return err
    }
    termbox.Sync()
    if modified {
        n.unsavedModifications = true
    }
    return n.update(g)
}

func (n *NotesGui) addNote(g *gocui.Gui, v *gocui.View) error {
    now := time.Now()
    note := Note{Created: now, Updated: now, Priority: 0}
    modified, err := note.EditInEditor()
    if err != nil {
        return err
    }
    termbox.Sync()
    if modified {
        n.Notes.AddNote(note)
        n.unsavedModifications = true
    }
    return n.update(g)
}

func (n *NotesGui) deleteNote(g *gocui.Gui, v *gocui.View) error {
    selected := &n.Notes.Notes[n.idx]
    n.Notes.DeleteNote(selected.Id)
    n.decreaseIndex(g, v)
    n.unsavedModifications = true
    return n.update(g)
}

func (n *NotesGui) toggleDone(g *gocui.Gui, v *gocui.View) error {
    now := time.Now()
    selected := &n.Notes.Notes[n.idx]
    selected.Done = !selected.Done
    selected.Updated = now
    n.unsavedModifications = true
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

func (n *NotesGui) backspaceCommand(g *gocui.Gui, v *gocui.View) error {
    sz := len(n.cmd)
    if sz > 0 {
        n.cmd = n.cmd[:sz-1]
    }
    return n.update(g)
}

func (n *NotesGui) executeCommand(g *gocui.Gui, v *gocui.View) error {
    command := n.cmd[1:]
    n.cmd = ""
    switch(command) {
        case "q!":
            return gocui.ErrQuit
        case "q":
            if n.unsavedModifications {
                n.statusString = "You have unsaved modifications. To discard them use :q!"
                break
            }
            fallthrough
        case "wq":
            if n.unsavedModifications {
                err := n.Notes.SaveNotes()
                if err != nil {
                    return err
                }
            }
            return gocui.ErrQuit
        case "w":
            if n.unsavedModifications {
                err := n.Notes.SaveNotes()
                if err != nil {
                    return err
                }
                n.statusString = "Notes saved"
            }
            break
        case "h":
            n.cmd = ""
            return n.showHelp(g)

        default:
            if len(command) > 0 {
                n.statusString = "Invalid command. Use :h to show help"
            }
    }

    _, err := g.SetCurrentView(LIST_VIEW)
    if err != nil {
        return err
    }
    return n.update(g)
}

func (n *NotesGui) showHelp(g *gocui.Gui) error {
    maxX, maxY := g.Size()
    v, err := g.SetView(HELP_VIEW, 0, 0, maxX-1, maxY-1)
    if err != nil && err != gocui.ErrUnknownView {
        return err
    }
    g.SetViewOnTop(HELP_VIEW)

    err = g.SetKeybinding(HELP_VIEW, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
        g.DeleteView(HELP_VIEW)
        g.SetCurrentView(LIST_VIEW)
        n.update(g)
        return nil
    })

    if err != nil {
        return err
    }

    v.Clear()
    v.Title = "Google Drive notes help"
    fmt.Fprintln(v, ":h - Show this help")
    fmt.Fprintln(v, "<esc> - Quit this help")
    fmt.Fprintln(v, ":q - Quit")
    fmt.Fprintln(v, ":q! - Quit without saving")
    fmt.Fprintln(v, ":wq - Save and quit")
    fmt.Fprintln(v, "<j> / <k> - Move up and down")
    fmt.Fprintln(v, "a - Add new note")
    fmt.Fprintln(v, "D - Delete selected note")
    fmt.Fprintln(v, "e - Edit selected note")
    fmt.Fprintln(v, "<enter> - Show note details / content")
    fmt.Fprintln(v, "G - Go to bottom of the list")
    fmt.Fprintln(v, ":a <note> - Quick add note")

    g.SetCurrentView(HELP_VIEW)
    return nil
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
    g.Cursor = false
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
    var selected *Note
    if len(n.Notes.Notes) > 0 {
        selected = &n.Notes.Notes[n.idx]
        for i, note := range n.Notes.Notes {
            if i == n.idx {
                c := color.New(color.Bold).Add(color.BgWhite).Add(color.FgBlack)
                c.Fprintln(v, note.GetStatusAndTitle())
                continue
            }
            fmt.Fprintln(v, note.GetStatusAndTitle())
        }
    } else {
        fmt.Fprintln(v, "No notes")
    }

    pv, err := g.View(PREVIEW_VIEW)
    if err != nil {
        return err
    }

    // TODO: Separate updating different views in multiple functions
    // TODO: Add some color here
    pv.Clear()
    if selected != nil && !n.showNoteContent {
        pv.Title = "Details"
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
    } else if selected != nil {
        pv.Title = "Content"
        fmt.Fprint(pv, selected.Content)
    }

    cv, err := g.View(COMMAND_VIEW)
    if err != nil {
        return err
    }
    cv.Clear()
    if len(n.cmd) > 0 {
        fmt.Fprintln(cv, n.cmd)
    } else if len(n.statusString) > 0 {
        fmt.Fprintln(cv, n.statusString)
        n.statusString = ""
    } else {
        _, err := g.SetCurrentView(LIST_VIEW)
        if err != nil {
            return err
        }
    }

    return nil
}

func (n *NotesGui) quit(g *gocui.Gui, v *gocui.View) error {
    return gocui.ErrQuit
}
