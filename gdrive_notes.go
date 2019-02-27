package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "errors"
    "strings"
    "strconv"

    "github.com/mitchellh/go-homedir"

    "golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
)

var app_folder = ""
var gdrive *drive.Service = nil
var max_id uint = 0

type Note struct {
    Id uint `json:"id"`
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
        return nil
    }

    notesJSON := make([]Note, 0)
    err = json.Unmarshal(dat, &notesJSON)
    if err != nil {
        return err
    }

    notes = notesJSON
    if len(notes) > 0 {
        max_id = notes[len(notes)-1].Id
    }

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

func printNote(note Note) {
    fmt.Print(note.Id)
    fmt.Print(" ")

    if !note.Done {
        fmt.Print("[ ]")
    } else {
        fmt.Print("[x]")
    }

    fmt.Print(" ")
    fmt.Print(note.Content)
    fmt.Print("\t")
    fmt.Print(note.Priority)
}

func handleArgs(args []string) (bool, error) {
    if len(args) == 0 {
        return false, errors.New("Insufficient parameters")
    }

    command := args[0]
    args = args[1:]

    if command == "qa" {
        if len(args) < 1 {
            return false, errors.New("Missing note content")
        }

        content := strings.Join(args, " ")
        note := Note{Id: max_id + 1, Content: content, Priority: 5, Done: false}
        notes = append(notes, note)
        return true, nil
    }

    if command == "clear" {
        notes = notes[:0]
        return true, nil
    }

    if command == "todo" {
        for _, note := range notes {
            printNote(note)
            fmt.Print("\n")
        }

        return false, nil
    }

    if command == "md" {
        if len(args) < 1 {
            return false, errors.New("Give note id")
        }

        id, err := strconv.ParseUint(args[0], 0, 32)

        if err != nil {
            return false, err
        }

        for i, _ := range notes {
            note := &notes[i]
            if note.Id == uint(id) {
                note.Done = true
                break
            }
        }

        return true, nil
    }

    return false, nil
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

    update, err := handleArgs(args)
    if err != nil {
        printHelp()
        os.Exit(0)
    }

    if update {
       err = syncNotesFile(notes)
        if err != nil {
            log.Fatalf("Could not sync notes to Drive: %v", err)
            os.Exit(1)
        }
    }
}
