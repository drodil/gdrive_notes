package main

import(
    "os"
    "io/ioutil"
    "encoding/json"
)

type Configuration struct {
    Md5Checksum string `json:"md5Checksum"`
    TimeFormat string `json:"time_format"`
    Color bool `json:"color"`
    UsePriority bool `json:"use_priority"`
    UseDue bool `json:"use_due"`
    config_file string
}

func NewConfiguration() (Configuration) {
    inst := Configuration{}

    inst.TimeFormat = "02.01.2006 15:04"
    inst.UsePriority = true
    inst.Color = true
    inst.UseDue = true

    return inst
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

func (c *Configuration) Configure() {
    for {
        color, err := YesNoQuestion("Use color output [y/n]? ")
        if err == nil {
            c.Color = color
            break
        }
    }

    for {
        due, err := YesNoQuestion("Use due dates for notes [y/n]? ")
        if err == nil {
            c.UseDue = due
            break
        }
    }

    for {
        prio, err := YesNoQuestion("Use priorities for notes [y/n]? ")
        if err == nil {
            c.UsePriority = prio
            break
        }
    }

    for {
        format, err := Question("Time format to use (empty as default dd.mm.YYYY HH:mm): ")
        if err == nil {
            c.TimeFormat = "02.01.2006 15:04"
            if len(format) > 1 {
                c.TimeFormat = format
            }
            break
        }
    }
    c.Save()
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
