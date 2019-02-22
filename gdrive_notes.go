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

type Note struct {
    Name string `json:"name"`
    Content string `json:"content"`
    Priority uint `json:"priority"`
    Done bool `json:"done"`
}

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

func createNotesFile(srv *drive.Service) {
    file := &drive.File{Name: "notes.json", Parents: []string{"appDataFolder"}}
    srv.Files.Create(file).Do()
}

func getNotesFile(srv *drive.Service) (file *drive.File, err error) {
    request := srv.Files.List().PageSize(10)
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

func reloadFromDrive(srv *drive.Service, file *drive.File) (err error) {
    export := srv.Files.Get(file.Id)

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

    return nil
}

func handleArgs(args []string) (err error) {
    if len(args) == 0 {
        return errors.New("Insufficient parameters")
    }

    if args[0] == "ls" {
        fmt.Println("LS")
    }

    return nil
}

func printHelp() {
    fmt.Println("Google Drive Notes")
    fmt.Println("------------------")
}

func main() {
    args := os.Args[1:]
    err := handleArgs(args)
    if err != nil {
        printHelp()
        os.Exit(0)
    }

    err = createAppFolder()
    if err != nil {
        log.Fatalf("Could not create app folder: %v", err)
        os.Exit(1)
    }

    b, err := ioutil.ReadFile("credentials.json")
    if err != nil {
        log.Fatalf("Unable to read client secret file: %v", err)
        os.Exit(1)
    }

    // If modifying these scopes, delete your previously saved token.json.
    config, err := google.ConfigFromJSON(b, drive.DriveAppdataScope)
    if err != nil {
        log.Fatalf("Unable to parse client secret file to config: %v", err)
        os.Exit(1)
    }
    client := getClient(config)

    srv, err := drive.New(client)
    if err != nil {
        log.Fatalf("Unable to retrieve Drive client: %v", err)
        os.Exit(1)
    }

    notes, err := getNotesFile(srv)
    if err != nil || notes == nil {
        createNotesFile(srv)
        notes, err = getNotesFile(srv)
    }

    if err != nil {
        log.Fatalf("Failed to read and create notes file: %v", err)
        os.Exit(1)
    }

    err = reloadFromDrive(srv, notes)
    if err != nil {
        log.Fatalf("Could not reload from Drive: %v", err)
        os.Exit(1)
    }
}
