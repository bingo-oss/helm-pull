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
	local              string
	replaceOld         string
	replaceNew         string
	insecureSkipVerify bool
}

var (
	globalUsage = `Pull charts into local folder. Will retry failed GET requests and only pull charts that do not already exist.
Examples:
	$ helm pull ./ repo													# Only pull
	$ helm pull ./ repo "registry.bingosoft.net" "hub.bingosoft.net"	# Pull and replace
`
)

func newPullCmd(args []string) *cobra.Command {
	d := &pullCmd{}
	cmd := &cobra.Command{
		Use:   "helm pull [local_folder] [repo_name] [replace_old] [replace_new]",
		Short: "Pull charts into local folder",
		Long:  globalUsage,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 && len(args) != 4 {
				return errors.New("This command needs 2 or 4 arguments, Try '--help' for more information.")
			}
			d.local = args[0]
			d.repoName = args[1]
			if len(args) == 4 {
				d.replaceOld = args[2]
				d.replaceNew = args[3]
			}
			d.setFieldsFromEnv()
			err := d.pullCharts()
			if err != nil {
				return err
			}
			return d.replaceCharts()
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&d.insecureSkipVerify, "insecure", "", false, "Connect to server with an insecure way by skipping certificate verification [$HELM_REPO_INSECURE]")
	f.Parse(args)
	return cmd
}

func (its *pullCmd) setFieldsFromEnv() {
	if v, ok := os.LookupEnv("HELM_REPO_INSECURE"); ok {
		its.insecureSkipVerify, _ = strconv.ParseBool(v)
	}
}

func (its *pullCmd) pullCharts() error {
	repo, err := helm.GetRepoByName(its.repoName)
	if err != nil {
		return err
	}

	return puller.Run(repo, its.local, its.insecureSkipVerify)
}

func (its *pullCmd) replaceCharts() error {
	flag := "f"
	if its.replaceOld != "" {
		flag = "d"
		fmt.Println(fmt.Sprintf("Replace `%s` to `%s` started", its.replaceOld, its.replaceNew))
	}

	err := filepath.Walk(its.local, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			repoName := info.Name()
			if path == its.local {
				repoName = ""
			}
			cmdShell := `
function push_all()
{
	folder=$(dirname $(readlink -f "$0"))
	if [ ! -n $1 ]; then
 		echo "Please enter the repo name, Example: ./push.sh bingo"
		return;
	fi
	for file in ` + "`ls $folder`;" + `
	do
		if [ -` + flag + ` "$folder/$file" ] ; then
			helm push "$folder/$file" "$1"
		fi
	done
}
push_all ${1:-` + repoName + `}`
			shellPath := path + "/push.sh"
			if _, err := os.Stat(shellPath); os.IsNotExist(err) {
				file, err := os.OpenFile(shellPath, os.O_CREATE, os.ModePerm)
				if err != nil {
					return err
				}
				defer file.Close()
				file.WriteString(cmdShell)
				file.Close()
			}
			return nil
		}

		if its.replaceOld == "" {
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

		dirPath := filepath.Join(filepath.Dir(path), info.Name())
		dirPath = strings.TrimRight(dirPath, ".tgz")
		os.RemoveAll(dirPath)
		os.MkdirAll(dirPath, os.ModePerm)

		tr := tar.NewReader(gr)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			fileName := filepath.Join(dirPath, strings.TrimLeft(h.Name, strings.Split(h.Name, "/")[0]+"/"))
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

			if its.replaceOld != "" {
				content = strings.Replace(content, its.replaceOld, its.replaceNew, -1)
			}

			fw.WriteString(content)
			fw.Close()
		}

		if its.replaceOld != "" {
			fmt.Println("Replacing", strings.Replace(path, its.local+"/", "", -1), "complete")
		}

		gr.Close()
		file.Close()
		err = os.Remove(path)
		if err != nil {
			return err
		}

		return nil
	})

	if flag == "f" {
		fmt.Println("Replace finished")
	}

	return err
}

func main() {
	cmd := newPullCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
