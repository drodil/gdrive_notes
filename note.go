package main

import (
    "os"
    "io/ioutil"
    "errors"
    "time"
    "os/exec"
    "strings"
    "crypto/md5"
    "encoding/hex"
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
    Tags []string `json:"tags"`
}

// Returns title of the note
func (n *Note) GetTitle() (string) {
    parts := strings.Split(n.Content, "\n")
    return parts[0]
}

func (n *Note) HasTag(tag string) (bool) {
    for _, t := range n.Tags {
        if t == tag {
            return true
        }
    }
    return false
}

func (n *Note) AddTag(tag string) (bool) {
    if !n.HasTag(tag) {
        n.Tags = append(n.Tags, tag)
        return true
    }
    return false
}

func (n *Note) RemoveTag(tag string) (bool) {
    for i := 0; i < len(n.Tags); i++ {
        t := n.Tags[i]
        if t == tag {
            n.Tags = append(n.Tags[:i], n.Tags[i+1:]...)
            return true
        }
    }

    return false
}

func (n *Note) ClearTags() {
    n.Tags = n.Tags[:0]
}

func (n *Note) EditInEditor() (bool, error) {
    editor, ok := os.LookupEnv("EDITOR")
    if !ok {
        return false, errors.New("You don't have EDITOR variable set!")
    }

    fpath := os.TempDir() + "/" + RandomString(10) + ".md"
    f, err := os.Create(fpath)
    if err != nil {
        return false, err
    }

    startHash := n.getMD5()
    _, err = f.WriteString(n.Content)
    if err != nil {
        return false, err
    }

    f.Close()
    cmd := exec.Command(editor, fpath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err = cmd.Start()
    if err != nil {
        return false, err
    }

    err = cmd.Wait()
    if err != nil {
        return false, err
    }

    dat, err := ioutil.ReadFile(fpath)
    if err != nil {
        return false, err
    }

    updated := false
    n.Content = string(dat)
    if startHash != n.getMD5() {
        n.Updated = time.Now()
        updated = true
    }

    return updated, nil
}

func (n *Note) getMD5() (string) {
    hasher := md5.New()
    hasher.Write([]byte(n.Content))
    return hex.EncodeToString(hasher.Sum(nil))
}
