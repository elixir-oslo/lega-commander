// Package main is the main package of lega-uploader command-line tool, containing "files", "resumables" and "uploads"
// commands implementations along with additional helper methods.
package main

import (
	"bytes"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/logrusorgru/aurora"
	"github.com/uio-bmi/lega-uploader/files"
	"github.com/uio-bmi/lega-uploader/resuming"
	"github.com/uio-bmi/lega-uploader/uploading"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

var (
	version = "dev"
	date    = "unknown"
)

const (
	filesCommand      = "files"
	resumablesCommand = "resumables"
	uploadCommand     = "upload"
)

var filesOptions struct {
	List   bool   `short:"l" long:"list" description:"Lists uploaded files"`
	Delete string `short:"d" long:"delete" description:"Deletes uploaded file by name"`
}

var filesOptionsParser = flags.NewParser(&filesOptions, flags.None)

var resumablesOptions struct {
	List   bool   `short:"l" long:"list" description:"Lists resumable uploads"`
	Delete string `short:"d" long:"delete" description:"Deletes resumable upload by ID"`
}

var resumablesOptionsParser = flags.NewParser(&resumablesOptions, flags.None)

var uploadingOptions struct {
	FileName string `short:"f"  long:"file" description:"File or folder to upload" value-name:"FILE" required:"true"`
	Resume   bool   `short:"r" long:"resume" description:"Resumes interrupted upload"`
}

var uploadingOptionsParser = flags.NewParser(&uploadingOptions, flags.None)

const (
	usageString        = "Usage:\n  lega-uploader\n"
	applicationOptions = "Application Options"
)

func main() {
	args := os.Args
	if len(args) == 1 || args[1] == "-h" || args[1] == "--help" {
		fmt.Println(generateHelpMessage())
		os.Exit(0)
	}
	if args[1] == "-v" || args[1] == "--version" {
		fmt.Println(aurora.Blue(version))
		fmt.Println(aurora.Yellow(date))
		os.Exit(0)
	}
	commandName := args[1]
	switch commandName {
	case filesCommand:
		_, err := filesOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		fileManager, err := files.NewFileManager(nil)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		if filesOptions.List {
			fileList, err := fileManager.ListFiles()
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
			_, err = fmt.Fprintln(tw, aurora.Blue("File name\t File size"))
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
			for _, file := range *fileList {
				_, err = fmt.Fprintln(tw, aurora.Blue(file.FileName+"\t "+strconv.FormatInt(file.Size, 10)+" bytes"))
				if err != nil {
					log.Fatal(aurora.Red(err))
				}
			}
			err = tw.Flush()
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
		} else if filesOptions.Delete != "" {
			err = fileManager.DeleteFile(filesOptions.Delete)
			if err != nil {
				log.Fatal(aurora.Red(err))
			} else {
				fmt.Println(aurora.Green("Success"))
			}
		} else {
			log.Fatal(aurora.Red("none of the flags are selected"))
		}
	case resumablesCommand:
		_, err := resumablesOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		resumablesManager, err := resuming.NewResumablesManager(nil)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		if resumablesOptions.List {
			resumables, err := resumablesManager.ListResumables()
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
			_, err = fmt.Fprintln(tw, aurora.Blue("File name\t File size\t Resumable ID"))
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
			for _, resumable := range *resumables {
				_, err := fmt.Fprintln(tw, aurora.Blue(resumable.Name+"\t "+strconv.FormatInt(resumable.Size, 10)+" bytes"+"\t "+resumable.ID))
				if err != nil {
					log.Fatal(aurora.Red(err))
				}
			}
			err = tw.Flush()
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
		} else if resumablesOptions.Delete != "" {
			err = resumablesManager.DeleteResumable(resumablesOptions.Delete)
			if err != nil {
				log.Fatal(aurora.Red(err))
			} else {
				fmt.Println(aurora.Green("Success"))
			}
		} else {
			log.Fatal(aurora.Red("none of the flags are selected"))
		}
	case uploadCommand:
		_, err := uploadingOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		uploader, err := uploading.NewUploader(nil, nil, nil)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		err = uploader.Upload(uploadingOptions.FileName, uploadingOptions.Resume)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
	default:
		log.Fatal(aurora.Red(fmt.Sprintf("command '%v' is not recognized", commandName)))
	}
}

func generateHelpMessage() string {
	header := "lega-uploader [files | resumables | upload] <args>\n"

	buf := bytes.Buffer{}
	filesOptionsParser.WriteHelp(&buf)
	filesUsage := buf.String()
	filesUsage = strings.Replace(filesUsage, usageString, "", 1)
	filesUsage = strings.Replace(filesUsage, applicationOptions, " "+filesCommand, 1)

	buf.Reset()
	resumablesOptionsParser.WriteHelp(&buf)
	resumablesUsage := buf.String()
	resumablesUsage = strings.Replace(resumablesUsage, usageString, "", 1)
	resumablesUsage = strings.Replace(resumablesUsage, applicationOptions, " "+resumablesCommand, 1)

	buf.Reset()
	uploadingOptionsParser.WriteHelp(&buf)
	uploadingUsage := buf.String()
	uploadingUsage = strings.Replace(uploadingUsage, usageString, "", 1)
	uploadingUsage = strings.Replace(uploadingUsage, applicationOptions, " "+uploadCommand, 1)

	return header + filesUsage + resumablesUsage + uploadingUsage
}
