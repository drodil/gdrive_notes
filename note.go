package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "errors"
    "time"
    "os/exec"
    "strings"
)

// Single NOTE functionality
type Note struct {
    Id uint `json:"id"`
    Content string `json:"content"`
    Priority uint `json:"priority"`
    Done bool `json:"done"`
    Created time.Time `json:"created"`
    Updated time.Time `json:"updated"`
    Due time.Time     `json:"due"`
}

func (n *Note) Print() {
    fmt.Print(n.Id)
    fmt.Print(" ")

    // TODO: Create configuration to allow user configure which columns to show
    if !n.Done {
        fmt.Print("[ ]")
    } else {
        fmt.Print("[x]")
    }

    fmt.Print(" ")
    parts := strings.Split(n.Content, "\n")
    preview := parts[0]
    if len(preview) > 30 {
        preview = preview[0:30] + "..."
    }
    fmt.Print(preview)
    fmt.Print("\t")
    fmt.Print(n.Priority)
    fmt.Print("\t")
    // TODO: Create configuration to allow user select format of dates
    fmt.Print(n.Created.Format("2006-01-02 15:04"))
}

func (n *Note) EditInEditor() (error) {
    editor, ok := os.LookupEnv("EDITOR")
    if !ok {
        return errors.New("You don't have EDITOR variable set!")
    }

    fpath := os.TempDir() + "/" + RandomString(10) + ".md"
    f, err := os.Create(fpath)
    if err != nil {
        return err
    }

    _, err = f.WriteString(n.Content)
    if err != nil {
        return err
    }

    f.Close()
    cmd := exec.Command(editor, fpath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Start()
    if err != nil {
        return err
    }

    err = cmd.Wait()
    if err != nil {
        return err
    }

    dat, err := ioutil.ReadFile(fpath)
    if err != nil {
        return err
    }

    n.Content = string(dat)

    return nil
}


