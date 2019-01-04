# Gmail Cleaner command line

Simple app to mass deletion or trashing email writen in Go language

## Requirement

* [Go](https://golang.org/)
* Credentials for using GmailApi. Follow beginning of [this guide](https://developers.google.com/gmail/api/quickstart/go#step_1_turn_on_the)

## Install

### Manually

> You need [Dep](https://github.com/golang/dep) to install dependencies

```bash
git clone https://github.com/SkYNewZ/gmail-cleaner.git
cd gmail-cleaner
dep ensure
```

You can now run it by doing `go run main.go` or build `go build -o gmail-cleaner` and execute `./gmail-cleaner`

### Automatically

```bash
go get -u github.com/SkYNewZ/gmail-cleaner
```

## Usage

```bash
$ gmail-cleaner --help
Usage:
  main [OPTIONS]

Application Options:
  -s, --search=           Search criteria
  -d, --delete            Delete messages ?
      --credentials-file= Credentials file path as json for using GmailAPI (default: credentials.json)

Help Options:
  -h, --help              Show this help message
```

> For example, if your `credentials.json` is located into your `$HOME`, run `gmail-cleaner --credentials-file ~/credentials.json ...`

## Example

```bash
# Trash all messages from "github.com"
gmail-cleaner --search "from:(github.com)" --credentials-file ~/credentials.json

# Delete **permanently** messages with subject "[GitHub] A third-party OAuth application has been added to your account"
gmail-cleaner --search "subject:([GitHub] A third-party OAuth application has been added to your account)" --delete --credentials-file ~/Downloads/credentials.json
```

> For more details about the Gmail search engine, checkout [docs](https://support.google.com/mail/answer/7190)