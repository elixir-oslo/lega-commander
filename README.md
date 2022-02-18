# lega-commander
[![Build Status](https://github.com/elixir-oslo/lega-commander/workflows/Go/badge.svg)](https://github.com/elixir-oslo/lega-commander/actions)
[![GoDoc](https://godoc.org/github.com/elixir-oslo/lega-commander?status.svg)](https://pkg.go.dev/github.com/elixir-oslo/lega-commander?tab=subdirectories)
[![CodeFactor](https://www.codefactor.io/repository/github/elixir-oslo/lega-commander/badge)](https://www.codefactor.io/repository/github/elixir-oslo/lega-commander)
[![Go Report Card](https://goreportcard.com/badge/github.com/elixir-oslo/lega-commander)](https://goreportcard.com/report/github.com/elixir-oslo/lega-commander)
[![codecov](https://codecov.io/gh/elixir-oslo/lega-commander/branch/master/graph/badge.svg)](https://codecov.io/gh/elixir-oslo/lega-commander)
[![DeepSource](https://deepsource.io/gh/elixir-oslo/lega-commander.svg/?label=active+issues&show_trend=true)](https://deepsource.io/gh/elixir-oslo/lega-commander/?ref=repository-badge)

## Installation / Update

### Linux
run the command below on your terminal:
```
curl -fsSL https://raw.githubusercontent.com/elixir-oslo/lega-commander/master/install.sh | sudo sh
```

### MacOS
run the command below on your terminal:
```
curl -fsSL https://raw.githubusercontent.com/elixir-oslo/lega-commander/master/install.sh | sh
```

### Windows
Go to the [releases page](https://github.com/elixir-oslo/lega-commander/releases) and download the latest binary manually.

## Configuration
Before using the lega commander, make sure all the environment variables required for authentication are set:
>In linux and MacOS you can use below commands in commandlines  to set them:
>
>```
>export CENTRAL_EGA_USERNAME=...
>export CENTRAL_EGA_PASSWORD=...
>export ELIXIR_AAI_TOKEN=...
>(cont.)
>```

> In windows, the variables must be set in environmental variables list.(Explained [here](https://www.architectryan.com/2018/08/31/how-to-change-environment-variables-on-windows-10/))

Table below shows how there variables must be set:
| Enviromental variable name        | description
|-------------------                | -------------
|CENTRAL_EGA_USERNAME               | The user name that you received from [CEGA website](https://ega-archive.org/)
|CENTRAL_EGA_PASSWORD               | The password that you received from [CEGA website](https://ega-archive.org/)
|ELIXIR_AAI_TOKEN                   | The token that you received after login here:(https://ega.elixir.no/)
|TSD_API_VER                        | The version of tsd file api that this software communicates with (for example: `v1` or `v2`...)(The default value is `v1` which must be set)
|TSD_PROJ_NAME                      | The name of the project inside the tsd that the file(s) is/are being sent there (The default value is `p11` which must be set)
|TSD_SERV                           | The name of the service inside the tsd that the file(s) is/are being sent there (The default value is `ega` which must be set)



>for developers: the tool is pre-configured to work with 
 Norwegian Federated EGA instance: https://ega.elixir.no.
 If you want to specify another instance, you can set `LOCAL_EGA_INSTANCE_URL` environment variable. 


## Usage

> For the time being, all of **upload** and **download** commands **must** be run with `-p` argument.
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
  -f, --file=FILE or =FOLDER    File or folder to upload
  -r, --resume                  Resumes interrupted upload
  -p, --proxy                   Upload the files through the proxy service

 download:
  -f, --file= FILE or =FOLDER   File or folder to download
  -p, --proxy                   Download the files through the proxy service

```
### Example Usage
As an example, if we want to upload file named `sample-c4gh-file.c4gh` and in path of `/path/to/a/c4gh/file`
or `D:\path\to\a\c4gh\file`, we wil do it with commands below based on operating system:

**In linux or macos**:
```
lega-commander upload  -p  -f /path/to/a/c4gh/file/sample-c4gh-file.c4gh 
 ```
**In windows** (in the case that lega-commander binary [dowloaded from release page] is in `D:\users\lega-commander` folder):
```
D:\users\lega-commander upload  -p  -f D:\path\to\a\c4gh\file\sample-c4gh-file.c4gh 
 ```
or if we want to upload a folder with path of `/path/to/a/folder/containing/c4gh/files`
that contains c4gh files, we can use this example command:

```
lega-commander upload  -p  -f /path/to/a/folder/containing/c4gh/files
```
