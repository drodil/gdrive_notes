package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "errors"

    "github.com/mitchellh/go-homedir"

    "golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
)

var app_folder = ""
var gdrive *drive.Service = nil

type Note struct {
    Name string `json:"name"`
    Content string `json:"content"`
    Priority uint `json:"priority"`
    Done bool `json:"done"`
}

var notes []Note

func createAppFolder() (error) {
    home, err := homedir.Dir()
    if(err != nil) {
        return err
    }

    os.Mkdir(home + "/.gdrive_notes", 0600)
    app_folder = home + "/.gdrive_notes"
    return nil
}

func getClient(config *oauth2.Config) *http.Client {
    tokFile := app_folder + "/token.json"
    tok, err := tokenFromFile(tokFile)
    if err != nil {
        tok = getTokenFromWeb(config)
        saveToken(tokFile, tok)
    }
    return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
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

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    tok := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(tok)
    return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        log.Fatalf("Unable to cache oauth token: %v", err)
    }
    defer f.Close()
    json.NewEncoder(f).Encode(token)
}

func createNotesFile() (file *drive.File, err error) {
    new_file := &drive.File{Name: "notes.json", Parents: []string{"appDataFolder"}}
    ret, err := gdrive.Files.Create(new_file).Do()
    if err != nil {
        return nil, err
    }
    return ret, nil
}

func saveNotes(file *drive.File) (err error) {
    jsonStr, err := json.Marshal(notes)
    if err != nil {
        return err
    }

    f, err := os.Create("/tmp/" + file.Name)
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

func syncNotesFile(file *drive.File) (err error) {
    err = saveNotes(file)
    if err != nil {
        return err
    }

    reader, err := os.Open("/tmp/" + file.Name)
    if err != nil {
        return err
    }

    update := gdrive.Files.Update(file.Id, &drive.File{})
    _, err = update.Media(reader).Do()
    if err != nil {
        return err
    }

    return nil
}

func getNotesFile() (file *drive.File, err error) {
    request := gdrive.Files.List().PageSize(10)
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

func reloadFromDrive(file *drive.File) (err error) {
    export := gdrive.Files.Get(file.Id)

    res, experr := export.Download()
    if experr != nil {
        return experr
    }

    f, err := os.Create("/tmp/" + file.Name)
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

    return parseNotes(file)
}

func parseNotes(file *drive.File) (err error) {
    dat, err := ioutil.ReadFile("/tmp/" + file.Name)
    if err != nil {
        return err
    }

    if len(dat) == 0 {
        return syncNotesFile(file)
    }

    notesJSON := make([]Note, 0)
    err = json.Unmarshal(dat, &notesJSON)
    if err != nil {
        return err
    }

    notes = notesJSON
    fmt.Print(notes)

    return nil
}

func setUpDrive() (error) {
    b, err := ioutil.ReadFile("credentials.json")
    if err != nil {
        return err
    }

    config, err := google.ConfigFromJSON(b, drive.DriveAppdataScope)
    if err != nil {
        return err
    }
    client := getClient(config)

    srv, err := drive.New(client)
    if err != nil {
        return err
    }

    gdrive = srv

    return nil
}

func handleArgs(args []string) (err error) {
    if len(args) == 0 {
        return errors.New("Insufficient parameters")
    }

    if args[0] == "add" {
        note := Note{Name:"test", Content:"plaa", Priority: 1, Done: false}
        notes = append(notes, note)
    }

    return nil
}

func printHelp() {
    fmt.Println("Google Drive Notes")
    fmt.Println("------------------")
}

func main() {
    args := os.Args[1:]
    err := createAppFolder()
    if err != nil {
        log.Fatalf("Could not create app folder: %v", err)
        os.Exit(1)
    }

    err = setUpDrive()
    if err != nil {
        log.Fatalf("Could not setup Drive sync: %v", err)
        os.Exit(1)
    }

    notes, err := getNotesFile()
    if err != nil || notes == nil {
        notes, err = createNotesFile()
        if err != nil {
            log.Fatalf("Could not create file to Drive: %v", err)
            os.Exit(1)
        }
    }

    if err != nil {
        log.Fatalf("Failed to read and create notes file: %v", err)
        os.Exit(1)
    }

    err = reloadFromDrive(notes)
    if err != nil {
        log.Fatalf("Could not reload from Drive: %v", err)
        os.Exit(1)
    }

    err = handleArgs(args)
    if err != nil {
        printHelp()
        os.Exit(0)
    }

    err = syncNotesFile(notes)
    if err != nil {
        log.Fatalf("Could not sync notes to Drive: %v", err)
        os.Exit(1)
    }
}
