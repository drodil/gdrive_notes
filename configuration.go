package main

import(
    "fmt"
    "os"
    "io/ioutil"
    "strings"
    "encoding/json"
    "strconv"
)

type Configuration struct {
    Md5Checksum string `json:"md5Checksum"`
    TimeFormat string `json:"time_format"`
    DueFormat string `json:"due_format"`
    Color bool `json:"color"`
    UsePriority bool `json:"use_priority"`
    UseDue bool `json:"use_due"`
    DefaultTags []string `json:"default_tags"`
    DefaultPriority uint `json:"default_priority"`
    DefaultCategory string `json:"default_category"`
    config_file string
}

func NewConfiguration() (Configuration) {
    inst := Configuration{}

    inst.TimeFormat = "02.01.2006 15:04"
    inst.UsePriority = true
    inst.Color = true
    inst.UseDue = true
    inst.DefaultPriority = 3
    inst.DefaultCategory = ""

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
    fmt.Println("Google Drive TODO notes configuration")
    PrintVerticalLine()

    fmt.Println("")
    fmt.Println("This is machine specific configuration for look&feel of the tool")

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
        format, err := Question("Time format to use (default dd.mm.YYYY HH:mm): ")
        if err == nil {
            c.TimeFormat = "02.01.2006 15:04"
            if len(format) > 1 {
                c.TimeFormat = format
            }
            break
        }
    }

    for {
        format, err := Question("Due date format to use (default dd.mm.YYYY): ")
        if err == nil {
           c.DueFormat = "02.01.2006"
           if len(format) > 1 {
                c.TimeFormat = format
           }
           break
        }
    }

    for {
        tagsStr, err := Question("Default tags (comma separated): ")
        if err == nil {
            if len(tagsStr) == 0 {
                break
            }

            tags := strings.Split(tagsStr, ",")
            c.DefaultTags = c.DefaultTags[:0]
            for _, tag := range tags {
                c.DefaultTags = append(c.DefaultTags, strings.Trim(tag, " "))
            }
            break
        }
    }

    for {
        prioStr, err := Question("Default priority for notes [0-5] (default 3): ")
        if err == nil {
            if len(prioStr) == 0 {
                c.DefaultPriority = 3
                break
            }

            i, err := strconv.ParseUint(prioStr, 10, 64)
            if err == nil && i <= 5{
                c.DefaultPriority = uint(i)
                break
            }
        }
    }

    for {
        catStr, err := Question("Default category (either empty, \"prio\" or \"due\"): ")
        if err == nil {
           if len(catStr) == 0 || catStr == "prio" || catStr == "due" {
                c.DefaultCategory = catStr
                break
           }
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
