// Package main is the main package of lega-commander command-line tool, containing "files", "resumables" and "uploads"
// commands implementations along with additional helper methods.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/elixir-oslo/lega-commander/files"
	"github.com/elixir-oslo/lega-commander/resuming"
	"github.com/elixir-oslo/lega-commander/streaming"
	"github.com/jessevdk/go-flags"
	aurora "github.com/logrusorgru/aurora/v3"
)

var (
	version = "dev"
	date    = "unknown"
)

const (
	inboxCommand      = "inbox"
	outboxCommand     = "outbox"
	resumablesCommand = "resumables"
	uploadCommand     = "upload"
	downloadCommand   = "download"
)

var inboxOptions struct {
	List   bool   `short:"l" long:"list" description:"Lists uploaded files"`
	Delete string `short:"d" long:"delete" description:"Deletes uploaded file by name"`
}

var inboxOptionsParser = flags.NewParser(&inboxOptions, flags.None)

var outboxOptions struct {
	List bool `short:"l" long:"list" description:"Lists exported files"`
}

var outboxOptionsParser = flags.NewParser(&outboxOptions, flags.None)

var resumablesOptions struct {
	List   bool   `short:"l" long:"list" description:"Lists resumable uploads"`
	Delete string `short:"d" long:"delete" description:"Deletes resumable upload by ID"`
}

var resumablesOptionsParser = flags.NewParser(&resumablesOptions, flags.None)

var uploadingOptions struct {
	FileName string `short:"f"  long:"file" description:"File or folder to upload" value-name:"FILE" required:"true"`
	Resume   bool   `short:"r" long:"resume" description:"Resumes interrupted upload"`
	Straight bool   `short:"b" long:"beta" description:"Upload the files without the proxy service;i.e. directly to tsd file api"`
}

var uploadingOptionsParser = flags.NewParser(&uploadingOptions, flags.None)

var downloadingOptions struct {
	FileName string `short:"f"  long:"file" description:"File to download\t[optional]"`
	Straight bool   `short:"b" long:"beta" description:"download the files without the proxy service;i.e. directly from tsd file api"`
}

var downloadingOptionsParser = flags.NewParser(&downloadingOptions, flags.None)

const (
	usageString        = "Usage:\n  lega-commander\n"
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
	fileManager, err := files.NewFileManager(nil)
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	commandName := args[1]
	switch commandName {
	case inboxCommand:
		_, err := inboxOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		if inboxOptions.List {
			fileList, err := fileManager.ListFiles(true)
			if err != nil {
                    if _, ok := err.(*files.FolderNotFoundError); ok {
                        log.Fatal(aurora.Red("Inbox Error: The user folder is empty or does not exist yet"))
                    }
                }
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
		} else if inboxOptions.Delete != "" {
			err = fileManager.DeleteFile(inboxOptions.Delete)
			if err != nil {
				log.Fatal(aurora.Red(err))
			} else {
				fmt.Println(aurora.Green("Success"))
			}
		} else {
			log.Fatal(aurora.Red("none of the flags are selected"))
		}
	case outboxCommand:
		_, err := inboxOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		if inboxOptions.List {
			fileList, err := fileManager.ListFiles(false)
			if err != nil {
                    if _, ok := err.(*files.FolderNotFoundError); ok {
                        log.Fatal(aurora.Red("Outbox Error: No data has been staged in the outbox yet"))
                    }
                }
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
		} else if inboxOptions.Delete != "" {
			err = fileManager.DeleteFile(inboxOptions.Delete)
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
		streamer, err := streaming.NewStreamer(nil, nil, nil, uploadingOptions.Straight)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		err = streamer.Upload(uploadingOptions.FileName, uploadingOptions.Resume, uploadingOptions.Straight)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
	case downloadCommand:
		_, err := downloadingOptionsParser.Parse()
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		streamer, err := streaming.NewStreamer(nil, nil, nil, downloadingOptions.Straight)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		if downloadingOptions.FileName == "" {
			fmt.Println(aurora.Blue("File to export is not specified. Downloading the whole outbox folder."))
			fileList, err := fileManager.ListFiles(false)
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
			for _, file := range *fileList {
				err = streamer.Download(file.FileName)
				if err != nil {
					log.Fatal(aurora.Red(err))
				}
			}
		} else {
			err = streamer.Download(downloadingOptions.FileName)
			if err != nil {
				log.Fatal(aurora.Red(err))
			}
		}
	default:
		log.Fatal(aurora.Red(fmt.Sprintf("command '%v' is not recognized", commandName)))
	}
}

func generateHelpMessage() string {
	header := "lega-commander [inbox | outbox | resumables | upload | download] <args>\n"

	buf := bytes.Buffer{}
	inboxOptionsParser.WriteHelp(&buf)
	inboxUsage := buf.String()
	inboxUsage = strings.Replace(inboxUsage, usageString, "", 1)
	inboxUsage = strings.Replace(inboxUsage, applicationOptions, " "+inboxCommand, 1)

	buf.Reset()
	outboxOptionsParser.WriteHelp(&buf)
	outboxUsage := buf.String()
	outboxUsage = strings.Replace(outboxUsage, usageString, "", 1)
	outboxUsage = strings.Replace(outboxUsage, applicationOptions, " "+outboxCommand, 1)

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

	buf.Reset()
	downloadingOptionsParser.WriteHelp(&buf)
	downloadingUsage := buf.String()
	downloadingUsage = strings.Replace(downloadingUsage, usageString, "", 1)
	downloadingUsage = strings.Replace(downloadingUsage, applicationOptions, " "+downloadCommand, 1)

	return header + inboxUsage + outboxUsage + resumablesUsage + uploadingUsage + downloadingUsage
}
