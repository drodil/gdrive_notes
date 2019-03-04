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
* Configuration of the tool
* Backup/reload notes to and from Google Drive

## Installation

Dependencies are handled with [golang/dep](https://github.com/golang/dep#installation) so please install it first

```bash
go get github.com/drodil/gdrive_notes
cd ~/go/src/github.com/drodil/gdrive_notes
dep ensure
go build
```

## Dependencies

* [golang/dep](https://github.com/golang/dep)
* [mitchellh/go-homedir](https://github.com/mitchellh/go-homedir)
* [googleapis/google-api-go-client](https://github.com/googleapis/google-api-go-client)
* [fatih/color](https://github.com/fatih/color)

## TODO/Future ideas

* Setting due date for notes
* Backup configuration to gdrive also
* Start working on note and track time to complete it
    * Should be able to stop work and track work time multiple times per note
    * Should be configurable to be used
* CLI GUI for handling notes
    * vim keymappings to navigate
    * Ordering of the notes
    * Finding notes
    * Marking notes done
    * Check https://github.com/nsf/termbox-go
* Add tests
* Support for "due today" or "due tomorrow" etc.
* More filters
    * ID range
    * Wildcard/regex searches for tags and content
* REST server for notes
* Background service for notifications
    * https://github.com/martinlindhe/notify

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
