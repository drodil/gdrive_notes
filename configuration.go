package main

import(
    "os"
    "io/ioutil"
    "encoding/json"
)

type Configuration struct {
    Etag string `json:"etag"`
    config_file string
}


func (c *Configuration) Init() (error) {
    app_folder, err := CreateAppFolder()
    if err != nil {
        return err
    }

    c.config_file = app_folder + "/config.json"

    err = c.loadConfig()
    if(err != nil) {
        err = c.Save()
        if(err != nil) {
            return err
        }
    }
    return nil
}

func (c *Configuration) Save() (error) {
    jsonStr, err := json.Marshal(c)
    if err != nil {
        return err
    }

    f, err := os.Create(c.config_file)
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

func (c *Configuration) loadConfig() (error) {
    dat, err := ioutil.ReadFile(c.config_file)
    if err != nil {
        return err
    }

    if len(dat) == 0 {
        return nil
    }

    err = json.Unmarshal(dat, c)
    if err != nil {
        return err
    }

    return nil
}
