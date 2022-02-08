package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/google/go-github/v42/github"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
)

var (
	releaseBinary = fmt.Sprintf("aws2fa-%s-amd64", runtime.GOOS)
	selfUpdateCmd = &cobra.Command{
		Use: "self-update",
		RunE: func(cmd *cobra.Command, args []string) error {
			latestRelease, asset, err := getLatestRelease()
			if err != nil {
				return err
			}

			if strings.Compare(version, *latestRelease.TagName) == 0 {
				fmt.Println("No newer versions available")

				return nil
			}

			tempFile, err := downloadAsset(asset)
			if err != nil {
				return err
			}

			return installNix(tempFile)
		},
	}
)

func init() {
	postUpdateCleanUp()
}

func postUpdateCleanUp() {
	oldBin := path.Join(binDir(), "aws2fa.old")
	if _, err := os.Stat(oldBin); os.IsNotExist(err) {
		return
	}

	if err := os.Remove(oldBin); err != nil {
		log("Unable to perform update clean up:", err)
	}
}

func binDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return path.Join(homeDir, ".bin")
}

func installNix(f *os.File) error {
	userBinDir := binDir()
	if _, err := os.Stat(userBinDir); os.IsNotExist(err) {
		log("Unable to discover home bin directory, attempting to create it at: ", userBinDir)
		if err := os.Mkdir(userBinDir, 0775); err != nil {
			return err
		}
	}

	binPath := path.Join(userBinDir, "aws2fa")
	if err := os.Rename(binPath, path.Join(userBinDir, "aws2fa.old")); err != nil {
		return err

	}
	if err := os.Rename(f.Name(), binPath); err != nil {
		return err
	}

	if err := os.Chmod(binPath, 0775); err != nil {
		return err
	}

	log("Installed binary at:", binPath)

	return f.Close()
}

func downloadAsset(asset *github.ReleaseAsset) (*os.File, error) {
	log("Starting download of", *asset.BrowserDownloadURL)
	resp, err := http.Get(*asset.BrowserDownloadURL)
	if err != nil {
		return nil, err
	}

	log("Retrieved new file with length:", resp.ContentLength)
	bar := pb.New64(resp.ContentLength)

	return writeReaderTo(resp.Body, bar)
}

func writeReaderTo(sourceStream io.ReadCloser, bar *pb.ProgressBar) (*os.File, error) {
	// open output file
	fo, err := ioutil.TempFile(os.TempDir(), "aws2fa")
	if err != nil {
		return nil, err
	}
	log("Created temp file at:", fo.Name())

	// close fo on exit and check for its returned error
	defer func() {
		if err := sourceStream.Close(); err != nil {
			panic(err)
		}
	}()

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := sourceStream.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			panic(err)
		}
	}

	bar.Finish()

	return fo, nil
}

func getLatestRelease() (*github.RepositoryRelease, *github.ReleaseAsset, error) {
	c := github.NewClient(nil)
	releases, _, err := c.Repositories.ListReleases(context.TODO(), "zquintana", "aws2fa", &github.ListOptions{})

	if err != nil {
		return nil, nil, err
	}

	for _, r := range releases {
		if len(r.Assets) < 2 {
			continue
		}

		assets := r.Assets
		for _, a := range assets {
			if 0 == strings.Compare(releaseBinary, *a.Name) {
				return r, a, nil
			}
		}
	}

	return nil, nil, errors.New("unable to discover a release for current target")
}
