# lega-commander
[![Build Status](https://github.com/elixir-oslo/lega-commander/workflows/Go/badge.svg)](https://github.com/elixir-oslo/lega-commander/actions)
[![GoDoc](https://godoc.org/github.com/elixir-oslo/lega-commander?status.svg)](https://pkg.go.dev/github.com/elixir-oslo/lega-commander?tab=subdirectories)
[![CodeFactor](https://www.codefactor.io/repository/github/elixir-oslo/lega-commander/badge)](https://www.codefactor.io/repository/github/elixir-oslo/lega-commander)
[![Go Report Card](https://goreportcard.com/badge/github.com/elixir-oslo/lega-commander)](https://goreportcard.com/report/github.com/elixir-oslo/lega-commander)
[![codecov](https://codecov.io/gh/elixir-oslo/lega-commander/branch/master/graph/badge.svg)](https://codecov.io/gh/elixir-oslo/lega-commander)
[![DeepSource](https://static.deepsource.io/deepsource-badge-light.svg)](https://deepsource.io/gh/elixir-oslo/lega-commander/?ref=repository-badge)

## Installation / Update

### Linux
```
curl -fsSL https://raw.githubusercontent.com/elixir-oslo/lega-commander/master/install.sh | sudo sh
```

### MacOS
```
curl -fsSL https://raw.githubusercontent.com/elixir-oslo/lega-commander/master/install.sh | sh
```

### Windows
Go to the [releases page](https://github.com/elixir-oslo/lega-commander/releases) and download the latest binary manually.

## Configuration
Before using the app, make sure all the environment variables required for authentication are set:

```
export CENTRAL_EGA_USERNAME=...
export CENTRAL_EGA_PASSWORD=...
export ELIXIR_AAI_TOKEN=...
```

NB: `ELIXIR_AAI_TOKEN` has an expiration time of nearly two hours, so one would need to re-obtain and re-set it upon expiration.

Also, the tool is pre-configured to work with Norwegian Federated EGA instance: https://ega.elixir.no 
If you want to specify another instance, you can set `LOCAL_EGA_INSTANCE_URL` environment variable. 

## Usage

```
$ lega-commander
lega-commander [inbox | outbox | resumables | upload | download] <args>

 inbox:
  -l, --list    Lists uploaded files
  -d, --delete= Deletes uploaded file by name

 outbox:
  -l, --list  Lists exported files

 resumables:
  -l, --list    Lists resumable uploads
  -d, --delete= Deletes resumable upload by ID

 upload:
  -f, --file=FILE    File or folder to upload
  -r, --resume       Resumes interrupted upload

 download:
  -f, --file= File to download
```
