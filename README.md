# Google Drive TODO notes

App to handle your TODO notes and save them to Google Drive. CLI tool written with golang as a
training project.

Features:

* Quick adding and removing notes
* Adding and editing notes with your $EDITOR in markdown
* Marking notes done
* Listing notes in table or with details
* Adding and removing tags for and from the notes
* Search from note content
* Search from note tags
* Ordering of notes
* Opening URLs in browser mentioned in the note
* Configuration of the tool
* Backup/reload notes to and from Google Drive
* CLI GUI
    * See [available commands](COMMANDS.md)

## Installation

Dependencies are handled with [golang/dep](https://github.com/golang/dep#installation) so please install it first

```bash
go get github.com/drodil/gdrive_notes
```

## Configuration

Configuration can be done with the following command

```bash
gdrive_notes config
```

Configuration is machine specific so if you for example use notes in home/work you can set default tags to identify
notes done in different locations. Also you might want to have due date and/or priority for some of the notes but not
necessarily want to see them in another.

## Dependencies

* [golang/dep](https://github.com/golang/dep)
* [mitchellh/go-homedir](https://github.com/mitchellh/go-homedir)
* [googleapis/google-api-go-client](https://github.com/googleapis/google-api-go-client)
* [fatih/color](https://github.com/fatih/color)
* [jroimartin/gocui](https://github.com/jroimartin/gocui)
* [mvdan/xurls](https://github.com/mvdan/xurls)

## TODO/Future ideas

* Setting due date for notes
* Due date format not to include minutes/seconds/etc.
* Start working on note and track time to complete it
    * Should be able to stop work and track work time multiple times per note
    * Should be configurable to be used
* Add tests
* Support for "due today" or "due tomorrow" etc.
* More filters
    * ID range
    * Wildcard/regex searches for tags and content
* REST server for notes
* Background service for notifications
    * https://github.com/martinlindhe/notify
* Use goroutines for loading/saving notes
    * To make the UI more responsive
* Allow checking if there is modifications before saving from Notes
* Watch GDrive file for modifications from other apps/clients

Additional apps using the same notes database:

* Web UI
* Android/iOS application
* Vim plugin

## Contributing

You are free to contribute with pull requests. Especially as this is my first golang project, please feel free to make
the code more readable and faster.

If you are interested in taking any of the app ideas, please feel free to do so.

## License

Under [MIT](LICENSE)
