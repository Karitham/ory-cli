// Tool for adding a license header to all supported files.

package headers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	goGitIgnore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"

	"github.com/ory/cli/cmd/dev/headers/comments"
)

// LICENSE defines the full license text.
const LICENSE_TEMPLATE = "Copyright © %d Ory Corp"

// LICENSE_TOKEN defines the token that identifies comments containing the license.
const LICENSE_TOKEN = "Copyright ©"

// file types that we don't want to add license headers to
var noLicenseHeadersFor = []comments.FileType{"md", "yml", "yaml"}

// AddLicenses adds or updates the Ory license header in all applicable files within the given directory.
func AddLicenses(dir string, year int, exclude []string) error {
	licenseText := fmt.Sprintf(LICENSE_TEMPLATE, year)
	gitIgnore, _ := goGitIgnore.CompileIgnoreFile(filepath.Join(dir, ".gitignore"))
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("cannot read directory %q: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		if gitIgnore != nil && gitIgnore.MatchesPath(info.Name()) {
			return nil
		}
		if !comments.SupportsFile(path) {
			return nil
		}
		if !fileTypeIsLicensed(path) {
			return nil
		}
		if isInExcludedFolder(path, exclude) {
			return nil
		}
		contentNoHeader, err := comments.FileContentWithoutHeader(path, LICENSE_TOKEN)
		if err != nil {
			return err
		}
		return comments.WriteFileWithHeader(path, licenseText, contentNoHeader)
	})
}

// isInExcludedFolder indicates whether the given path exists within the given list of folders
func isInExcludedFolder(path string, exclude []string) bool {
	for _, e := range exclude {
		if strings.HasPrefix(path, e) {
			return true
		}
	}
	return false
}

// indicates whether this tool is configured to add a license header to the file with the given path
func fileTypeIsLicensed(path string) bool {
	return !comments.ContainsFileType(noLicenseHeadersFor, comments.GetFileType(path))
}

var copyright = &cobra.Command{
	Use:   "license",
	Short: "Adds the license header to all known files in the current directory",
	Long: `Adds the license header to all known files in the current directory.

Does not add the license header to git-ignored files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine the current directory: %w", err)
		}
		year, _, _ := time.Now().Date()
		for e, excluded := range exclude {
			exclude[e] = filepath.Join(cwd, excluded)
		}
		return AddLicenses(cwd, year, exclude)
	},
}

func init() {
	Main.AddCommand(copyright)
	copyright.Flags().StringSliceVarP(&exclude, "exclude", "e", []string{}, "folders to exclude, provide comma-separated values or multiple instances of this flag")
}

// contains the folders to exclude
var exclude []string
