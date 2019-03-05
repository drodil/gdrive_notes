package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "errors"
    "sort"
    "strings"
    "time"

    "golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
)

// NOTES functionality
type Notes struct {
    notes []Note
    gdrive *drive.Service
    app_folder string
    file *drive.File
    max_id uint
    config *Configuration
}

func (n *Notes) Init(config *Configuration) (error) {
    n.config = config
    app_folder, err := CreateAppFolder();
    if err != nil {
        return err
    }
    n.app_folder = app_folder

    err = n.setUpDrive()
    if err != nil {
        return err
    }

    notes_file, err := n.getNotesFile()
    if err != nil || notes_file == nil {
        notes_file, err = n.createNotesFile()
        if err != nil {
            return err
        }
    }

    if err != nil {
        return err
    }

    n.file = notes_file

    err = n.reloadFromDrive()
    if err != nil {
        return err
    }

    return nil
}

func (n *Notes) SaveNotes() (error) {
    now := time.Now()
    for _, note := range n.notes {
        note.Updated = now
    }
    return n.syncNotesFile()
}

func (n *Notes) AddNote(note Note) (uint) {
    note.Id = n.max_id + 1
    note.Created = time.Now()

    // Add default tags
    for _, tag := range n.config.DefaultTags {
        note.AddTag(tag)
    }
    n.notes = append(n.notes, note)
    return note.Id
}

func (n *Notes) FindNote(id uint) (*Note) {
    for i, _ := range n.notes {
        note := &n.notes[i]
        if note.Id == id {
            return note
        }
    }

    return nil
}

func (n *Notes) GetMaxId() (uint) {
    var maxId uint
    for _, note := range n.notes {
        if note.Id > maxId {
            maxId = note.Id
        }
    }

    return maxId
}

func (n *Notes) DeleteNote(id uint) (error) {
    found := false
    for i := 0; i < len(n.notes); i++ {
        note := &n.notes[i]
        if note.Id == id {
            n.notes = append(n.notes[:i], n.notes[i+1:]...)
            i--
            found = true
        }
    }

    if !found {
        return errors.New("Could not find note with given id")
    }
    return nil
}

func (n *Notes) GetTags() (map[string]int) {
    ret := map[string]int{}
    for i := 0; i < len(n.notes); i++ {
        note := &n.notes[i]
        for _, tag := range note.Tags {
            _, ok := ret[tag]
            if ok {
                ret[tag] = ret[tag] + 1
            } else {
                ret[tag] = 1
            }
        }
    }
    return ret
}

func (n *Notes) GetTagKeys() ([]string) {
    tags := n.GetTags()
    keys := make([]string, 0, len(tags))
    for k := range tags {
        keys = append(keys, k)
    }
    sort.Slice(keys, func(i, j int) bool {
        return keys[i] > keys[j]
    })
    return keys
}

func (n *Notes) GetNotes() []*Note {
    var ret[]*Note
    for i, _ := range n.notes {
        ret = append(ret, &n.notes[i])
    }
    return ret
}

func (n *Notes) SearchNotes(str string, notes []*Note) []*Note {
    var ret []*Note
    for _, note := range notes {
        if note.MatchesSearch(str) {
            ret = append(ret, note)
        }
    }
    return ret
}

func (n *Notes) ClearNotes() {
    n.notes = n.notes[:0]
}

func (n *Notes) FilterDoneNotes(notes[] *Note) []*Note {
    var ret[]*Note
    for _, note := range notes {
        if note.Done {
            continue
        }
        ret = append(ret, note)
    }
    return ret
}

func (n *Notes) FilterNotesByTag(tag string, notes []*Note) []*Note {
    var ret []*Note
    for _, note := range notes {
        if note.HasTag(tag) {
            ret = append(ret, note)
        }
    }
    return ret
}

func (n *Notes) FilterNotesByPriority(prio uint, notes []*Note) []*Note {
    var ret []*Note
    for _, note := range notes {
        if note.Priority >= prio {
            ret = append(ret, note)
        }
    }
    return ret
}


func (n *Notes) OrderNotes(columns []string, notes[]*Note) {
    sort.Slice(notes, func(i, j int) bool {
        ret := false
        for _, col := range columns {
            asc := true
            if strings.HasPrefix(col, "-") {
                asc = false
                col = col[1:]
            }

            switch(col) {
                case "prio":
                    ret = notes[i].Priority < notes[j].Priority
                    break
                case "title":
                    ret = notes[i].GetTitle() < notes[j].GetTitle()
                    break
                case "due":
                    ret = notes[i].Due.Unix() < notes[j].Due.Unix()
                    break
                case "created":
                    ret = notes[i].Created.Unix() < notes[j].Created.Unix()
                    break
                case "updated":
                    ret = notes[i].Updated.Unix() < notes[j].Updated.Unix()
                    break
                case "id":
                    ret = notes[i].Id > notes[j].Id
                    break
            }

            if asc {
                ret = !ret
            }

            if ret {
                break
            }
        }

        return ret
    })
}

func (n *Notes) createNotesFile() (file *drive.File, err error) {
    new_file := &drive.File{Name: "notes.json", Parents: []string{"appDataFolder"}}
    ret, err := n.gdrive.Files.Create(new_file).Do()
    if err != nil {
        return nil, err
    }
    return ret, nil
}

func (n *Notes) getClient(config *oauth2.Config) *http.Client {
    tokFile := n.app_folder + "/token.json"
    tok, err := n.tokenFromFile(tokFile)
    if err != nil {
        tok = n.getTokenFromWeb(config)
        n.saveToken(tokFile, tok)
    }
    return config.Client(context.Background(), tok)
}

func (n *Notes) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Println("Google Drive setup")
    fmt.Printf("Go to the following link in your browser then type the "+
            "authorization code: \n\n%v\n", authURL)

    var authCode string
    if _, err := fmt.Scan(&authCode); err != nil {
        log.Fatalf("Unable to read authorization code %v", err)
    }

    tok, err := config.Exchange(context.TODO(), authCode)
    if err != nil {
        log.Fatalf("Unable to retrieve token from web %v", err)
    }
    return tok
}

func (n *Notes) tokenFromFile(file string) (*oauth2.Token, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    tok := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(tok)
    return tok, err
}

func (n *Notes) saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        log.Fatalf("Unable to cache oauth token: %v", err)
    }
    defer f.Close()
    json.NewEncoder(f).Encode(token)
}

func (n *Notes) saveNotes() (err error) {
    jsonStr, err := json.Marshal(n.notes)
    if err != nil {
        return err
    }

    f, err := os.Create(os.TempDir() + "/" + n.file.Name)
    if err != nil {
        return err
    }

    defer f.Close()
    _, err = f.Write(jsonStr)

    if err != nil {
        return err
    }

    f.Sync()
    return nil
}

func (n *Notes) syncNotesFile() (err error) {
    err = n.saveNotes()
    if err != nil {
        return err
    }

    reader, err := os.Open(os.TempDir() + "/" + n.file.Name)
    if err != nil {
        return err
    }

    update := n.gdrive.Files.Update(n.file.Id, &drive.File{})
    _, err = update.Media(reader).Do()
    if err != nil {
        return err
    }

    return nil
}

func (n *Notes) getNotesFile() (file *drive.File, err error) {
    request := n.gdrive.Files.List().PageSize(10)
    request.Spaces("appDataFolder")
    request.Fields("nextPageToken, files(id, name, md5Checksum)")
    r, err := request.Do()
    if err != nil {
        log.Fatalf("Unable to retrieve files: %v", err)
        return nil, err
    }

    for _, i := range r.Files {
        if i.Name == "notes.json" {
            return i, nil
        }
    }

    return nil, errors.New("Could not find notes file")
}

func (n *Notes) reloadFromDrive() (err error) {
    if n.file.Md5Checksum == n.config.Md5Checksum {
        parse_err := n.parseNotes()
        if parse_err != nil {
            return nil
        }
    }

    export := n.gdrive.Files.Get(n.file.Id)

    res, experr := export.Download()
    if experr != nil {
        return experr
    }

    f, err := os.Create(os.TempDir() + "/" + n.file.Name)
    if err != nil {
        return err
    }

    defer f.Close()
    body, readerr := ioutil.ReadAll(res.Body)
    if readerr != nil {
        return readerr
    }

    _, err = f.Write(body)

    res.Body.Close()
    if err != nil {
        return err
    }

    f.Sync()
    n.config.Md5Checksum = n.file.Md5Checksum
    n.config.Save()

    return n.parseNotes()
}

func (n *Notes) parseNotes() (err error) {
    dat, err := ioutil.ReadFile(os.TempDir() + "/" + n.file.Name)
    if err != nil {
        return err
    }

    if len(dat) == 0 {
        return nil
    }

    notesJSON := make([]Note, 0)
    err = json.Unmarshal(dat, &notesJSON)
    if err != nil {
        return err
    }

    n.notes = notesJSON
    if len(n.notes) > 0 {
        n.max_id = n.notes[len(n.notes)-1].Id
    }

    return nil
}

func (n *Notes) setUpDrive() (error) {
    b := []byte(`{
        "installed":{
            "client_id":"793575810882-hppntrbvumvbrlmggjpo73uce627rjiu.apps.googleusercontent.com",
            "project_id":"gdrive-notes",
            "auth_uri":"https://accounts.google.com/o/oauth2/auth",
            "token_uri":"https://oauth2.googleapis.com/token",
            "auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs",
            "client_secret":"nXKrGu3oISBhGQy0kwkJf393",
            "redirect_uris":[
                "urn:ietf:wg:oauth:2.0:oob",
                "http://localhost"
            ]
        }
    }`)
    config, err := google.ConfigFromJSON(b, drive.DriveAppdataScope)
    if err != nil {
        return err
    }
    client := n.getClient(config)

    srv, err := drive.New(client)
    if err != nil {
        return err
    }

    n.gdrive = srv

    return nil
}

