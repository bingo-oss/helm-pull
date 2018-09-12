package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"helm-pull/helm"
	"helm-pull/puller"

	"github.com/spf13/cobra"
)

type pullCmd struct {
	repoName           string
	username           string
	password           string
	localFolder        string
	sourceImageRepo    string
	targetImageRepo    string
	useHTTP            bool
	insecureSkipVerify bool
}

var (
	globalUsage = `Pull charts into local folder. Will retry failed GET requests and only pull charts that do not already exist.
Examples:
  $ helm pull ./ repo --source-image-repo=registry.bingosoft.net --target-image-repo=hub.bingosoft.net --insecure=true
`
)

func newPullCmd(args []string) *cobra.Command {
	d := &pullCmd{}
	cmd := &cobra.Command{
		Use:   "helm pull [local_folder] [repo_name]",
		Short: "Pull charts into local folder",
		Long:  globalUsage,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("This command needs 2 arguments, Try '--help' for more information.")
			}
			d.localFolder = args[0]
			d.repoName = args[1]
			d.setFieldsFromEnv()
			err := d.pull()
			if err != nil {
				return err
			}
			err = d.replace()
			return err
		},
	}
	f := cmd.Flags()
	f.StringVarP(&d.sourceImageRepo, "source-image-repo", "", "", "Current image repository [$HELM_SOURCE_IMAGE_REPO]")
	f.StringVarP(&d.targetImageRepo, "target-image-repo", "", "", "Change to the new image repository [$HELM_TARGET_IMAGE_REPO]")
	f.BoolVarP(&d.insecureSkipVerify, "insecure", "", false, "Connect to server with an insecure way by skipping certificate verification [$HELM_REPO_INSECURE]")
	f.Parse(args)
	return cmd
}

func (its *pullCmd) setFieldsFromEnv() {
	if v, ok := os.LookupEnv("HELM_SOURCE_IMAGE_REPO"); ok && its.sourceImageRepo == "" {
		its.sourceImageRepo = v
	}
	if v, ok := os.LookupEnv("HELM_TARGET_IMAGE_REPO"); ok && its.targetImageRepo == "" {
		its.targetImageRepo = v
	}
	if v, ok := os.LookupEnv("HELM_REPO_USE_HTTP"); ok {
		its.useHTTP, _ = strconv.ParseBool(v)
	}
	if v, ok := os.LookupEnv("HELM_REPO_INSECURE"); ok {
		its.insecureSkipVerify, _ = strconv.ParseBool(v)
	}
}

func (its *pullCmd) pull() error {
	repo, err := helm.GetRepoByName(its.repoName)
	if err != nil {
		return err
	}

	return puller.Run(repo, its.localFolder, its.insecureSkipVerify)
}

func (its *pullCmd) replace() error {
	if its.sourceImageRepo == "" {
		return nil
	}

	fmt.Println(fmt.Sprintf("Replace %s to %s", its.sourceImageRepo, its.targetImageRepo))

	err := filepath.Walk(its.localFolder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(info.Name()) != ".tgz" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		gr, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gr.Close()

		tr := tar.NewReader(gr)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			fileName := filepath.Join(filepath.Dir(path), h.Name)
			os.MkdirAll(filepath.Dir(fileName), os.ModePerm)

			fw, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				return err
			}
			defer fw.Close()

			buf, err := ioutil.ReadAll(tr)
			if err != nil {
				return err
			}
			content := string(buf)
			content = strings.Replace(content, its.sourceImageRepo, its.targetImageRepo, -1)

			fw.WriteString(content)
		}

		fmt.Println("Replace", strings.Replace(path, its.localFolder+"/", "", -1), "complete")
		os.Remove(path)

		return nil
	})

	return err
}

func main() {
	cmd := newPullCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
