package main

import (
    "os"
    "math/rand"
    "bufio"
    "errors"
    "fmt"
    "strings"
    "golang.org/x/crypto/ssh/terminal"

    "github.com/fatih/color"
    "github.com/mitchellh/go-homedir"
)

// Utility functions
func CreateAppFolder() (string, error) {
    home, err := homedir.Dir()
    if(err != nil) {
        return "", err
    }

    os.Mkdir(home + "/.gdrive_notes", 0770)
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

func Question(question string) (string, error) {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print(question)
    text, err := reader.ReadString('\n')
    if err != nil {
        return "", err
    }
    return strings.TrimSuffix(text, "\n"), nil
}

func YesNoQuestion(question string) (bool, error) {
    text, err := Question(question)
    if err != nil {
        return false, err
    }
    if text == "y" || text == "Y" || text == "yes" {
        return true, nil
    } else if text == "n" || text == "N" || text == "no" {
        return false, nil
    }
    return false, errors.New("Invalid input")
}

func GetScreenWidth() (int) {
    width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        return 0
    }
    return width
}

func PrintVerticalLine() {
    c := color.New(color.FgHiBlack)
    c.Println(strings.Repeat("─", GetScreenWidth()))
}

func GetPriorityColor(n *Note) (*color.Color) {
    c := color.New()
    switch(n.Priority) {
        case 0:
            fallthrough
        case 1:
            c.Add(color.FgHiGreen)
            break
        case 2:
            fallthrough
        case 3:
            c.Add(color.FgHiYellow)
            break
        case 4:
            fallthrough
        case 5:
            c.Add(color.FgHiRed)
            break
    }
    return c
}

