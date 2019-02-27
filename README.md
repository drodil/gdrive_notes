# Google Drive TODO notes

App to handle your TODO notes and save them to Google Drive. CLI tool written with golang as a
training project.

## Dependencies

* [mitchellh/go-homedir](https://github.com/mitchellh/go-homedir)
* [googleapis/google-api-go-client](https://github.com/googleapis/google-api-go-client)

## TODO/Future ideas

* Adding of notes with $EDITOR instead quick add
* Editing of notes (content, priority, due date etc.)
* Configuration of the CLI application (~/.gdrive_notes/configuration.json)
* GUI for handling notes
    * vim keymappings to navigate
    * Ordering of the notes
    * Finding notes
    * Marking notes done
* Performance optimizations
    * Prevent unnecessary reloading from gdrive
* Web UI
* Android/iOS application
