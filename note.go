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

    "github.com/mvdan/xurls"
    "github.com/pkg/browser"
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

func (n *Note) GetStatusAndTitle() (string) {
    ret := "["
    if n.Done {
        ret += "x]"
    } else {
        ret += " ]"
    }
    ret += " " + n.GetTitle()
    return ret
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
    tagStr := strings.Trim(tag, " ")

    if !n.HasTag(tagStr) {
        n.Tags = append(n.Tags, tagStr)
        return true
    }
    return false
}

func (n *Note) RemoveTag(tag string) (bool) {
    tagStr := strings.Trim(tag, " ")
    for i := 0; i < len(n.Tags); i++ {
        t := n.Tags[i]
        if t == tagStr {
            n.Tags = append(n.Tags[:i], n.Tags[i+1:]...)
            return true
        }
    }

    return false
}

func (n *Note) ClearTags() bool {
    if len(n.Tags) == 0 {
        return false
    }
    n.Tags = n.Tags[:0]
    return true
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

    defer f.Close()
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

func (n *Note) MatchesSearch(str string) (bool) {
    if strings.Contains(strings.ToLower(n.Content), strings.ToLower(str)) {
        return true
    }
    tagsStr := strings.ToLower(strings.Join(n.Tags, " "))
    if strings.Contains(tagsStr, strings.ToLower(str)) {
        return true
    }
    return false
}

func (n *Note) GetUrls() ([]string) {
    return xurls.Strict().FindAllString(n.Content, 1)
}

func (n *Note) OpenUrls() (int) {
    urls := n.GetUrls()
    if len(urls) > 0 {
        browser.Stdout = ioutil.Discard
        for _, url := range urls {
            browser.OpenURL(url)
        }
    }
    return len(urls)
}

func (n *Note) getMD5() (string) {
    hasher := md5.New()
    hasher.Write([]byte(n.Content))
    return hex.EncodeToString(hasher.Sum(nil))
}
