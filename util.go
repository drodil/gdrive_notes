package main

import (
    "os"

    "github.com/mitchellh/go-homedir"
)

// Utility functions
func CreateAppFolder() (string, error) {
    home, err := homedir.Dir()
    if(err != nil) {
        return "", err
    }

    os.Mkdir(home + "/.gdrive_notes", 0600)
    return home + "/.gdrive_notes", nil
}


