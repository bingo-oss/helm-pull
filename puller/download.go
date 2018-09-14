package puller

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"helm-pull/helm"

	"github.com/cenkalti/backoff"
	"github.com/dustin/go-humanize"
)

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer
// interface and we can pass this into io.TeeReader() which will report progress on each
// write cycle.
type writeCounter struct {
	FilePath string
	Total    uint64
}

func (its *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	its.Total += uint64(n)
	its.PrintProgress()
	return n, nil
}

// PrintProgress Outputs the status of the download
func (its writeCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading %s ... %s complete", its.FilePath, humanize.Bytes(its.Total))
}

// Run executes the download
func Run(repo *helm.Repo, localFolder string, skipVerify bool) error {
	repo.URL = strings.TrimRight(repo.URL, "/")

	fmt.Println("Repo url:", repo.URL)
	fmt.Println("Local folder:", localFolder)

	client, err := newClient(
		repourl(repo.URL),
		username(repo.Username),
		password(repo.Password),
		insecureSkipVerify(skipVerify),
	)
	if err != nil {
		return err
	}

	index, err := helm.GetIndexByRepo(repo, getIndexDownloader(client))
	if err != nil {
		return err
	}

	fmt.Println("Download started")

	for _, entire := range index.Entries {
		for _, x := range entire {
			for _, chartUrl := range x.URLs {
				var fileName, filePath, projName string
				if strings.HasPrefix(chartUrl, "http://") || strings.HasPrefix(chartUrl, "https://") {
					parseChartUrl, _ := url.Parse(chartUrl)
					parseRepoUrl, _ := url.Parse(repo.URL)
					parseChartUrl.Scheme = parseRepoUrl.Scheme
					parseChartUrl.User = nil
					chartUrl = parseChartUrl.String()
					values := strings.Split(strings.Replace(strings.Replace(chartUrl, repo.URL+"/", "", -1), "charts/", "", -1), "/")
					if len(values) == 2 {
						projName = values[0]
						fileName = values[1]
					} else {
						fileName = values[0]
					}
					chartUrl = strings.Replace(chartUrl, repo.URL+"/", "", -1)
				}

				dirPath := localFolder
				dirPath = path.Join(dirPath, projName)
				os.MkdirAll(dirPath, os.ModePerm)

				filePath = dirPath + "/" + fileName

				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					f := func() error {
						out, err := os.Create(filePath + ".tmp")
						if err != nil {
							return err
						}
						defer out.Close()

						resp, err := client.downloadFile(chartUrl)
						if err != nil {
							return err
						}
						defer resp.Body.Close()

						counter := &writeCounter{FilePath: strings.Replace(filePath, localFolder+"/", "", -1)}
						_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
						if err != nil {
							return err
						}
						// The progress use the same line so print a new line once it's finished downloading
						fmt.Print("\n")

						out.Close()
						resp.Body.Close()

						err = os.Rename(filePath+".tmp", filePath)
						if err != nil {
							return err
						}

						return nil
					}
					exponentialBackOff := backoff.NewExponentialBackOff()
					exponentialBackOff.MaxElapsedTime = 30 * time.Second
					exponentialBackOff.Reset()
					err := backoff.Retry(f, exponentialBackOff)
					if err != nil {
						return err
					}

				} else if err != nil {
					return err
				}
			}
		}
	}

	fmt.Println("Download finished")

	return nil
}

func getIndexDownloader(c *client) helm.IndexDownloader {
	return func() ([]byte, error) {
		resp, err := c.downloadFile("index.yaml")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, getChartmuseumError(b, resp.StatusCode)
		}
		return b, nil
	}
}

func getChartmuseumError(b []byte, code int) error {
	var er struct {
		Error string `json:"error"`
	}
	err := json.Unmarshal(b, &er)
	if err != nil || er.Error == "" {
		return fmt.Errorf("%d: could not properly parse response JSON: %s", code, string(b))
	}
	return fmt.Errorf("%d: %s", code, er.Error)
}
