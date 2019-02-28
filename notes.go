package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "errors"
    "time"

    "github.com/mitchellh/go-homedir"

    "golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
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
    fmt.Print(n.Content)
    fmt.Print("\t")
    fmt.Print(n.Priority)
    fmt.Print("\t")
    // TODO: Create configuration to allow user select format of dates
    fmt.Print(n.Created.Format("2006-01-02 15:04"))
}

// NOTES functionality
type Notes struct {
    Notes []Note
    gdrive *drive.Service
    app_folder string
    file *drive.File
    max_id uint
}

func (n *Notes) Init() (error) {
    err := n.createAppFolder();
    if err != nil {
        return err
    }

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

    // TODO: Only reload if it has changed from last load
    err = n.reloadFromDrive()
    if err != nil {
        return err
    }

    return nil
}

func (n *Notes) SaveNotes() (error) {
    // TODO: Save only if changed
    return n.syncNotesFile()
}

func (n *Notes) AddNote(note Note) {
    note.Id = n.max_id + 1
    n.Notes = append(n.Notes, note)
}

func (n *Notes) createAppFolder() (error) {
    home, err := homedir.Dir()
    if(err != nil) {
        return err
    }

    os.Mkdir(home + "/.gdrive_notes", 0600)
    n.app_folder = home + "/.gdrive_notes"
    return nil
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
    fmt.Printf("Go to the following link in your browser then type the "+
            "authorization code: \n%v\n", authURL)

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
    jsonStr, err := json.Marshal(n.Notes)
    if err != nil {
        return err
    }

    f, err := os.Create("/tmp/" + n.file.Name)
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

    reader, err := os.Open("/tmp/" + n.file.Name)
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
    request.Fields("nextPageToken, files(id, name)")
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
    export := n.gdrive.Files.Get(n.file.Id)

    res, experr := export.Download()
    if experr != nil {
        return experr
    }

    f, err := os.Create("/tmp/" + n.file.Name)
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

    return n.parseNotes()
}

func (n *Notes) parseNotes() (err error) {
    dat, err := ioutil.ReadFile("/tmp/" + n.file.Name)
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

    n.Notes = notesJSON
    if len(n.Notes) > 0 {
        n.max_id = n.Notes[len(n.Notes)-1].Id
    }

    return nil
}

func (n *Notes) setUpDrive() (error) {
    b, err := ioutil.ReadFile("credentials.json")
    if err != nil {
        return err
    }

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

