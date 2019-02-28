package main

import (
    "os"
    "math/rand"

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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(n int) (string) {
    b:= make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}
