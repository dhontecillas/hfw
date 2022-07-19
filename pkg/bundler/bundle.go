package bundler

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// StaticDir has the directory name where the static assets
	// should be placed.
	StaticDir string = "static"
	// AppDir contains the root app directory, where the output
	// binary should be placed.
	AppDir string = "app"
	// ConfDir is the directory where the configuration for the
	// app will be stored.
	ConfDir string = "app/config"
	// DBMigrationsDir is the directory where database migrations
	// should be placed.
	DBMigrationsDir string = "app/dbmigrations"
	// AppDataDir is the directory where additional files for the
	// app must be placed (the ones that won't be directly served
	// by any other app).
	AppDataDir string = "app/data"

	// AppDataHTMLTemplates is the directory where the html templates
	// should be placed.
	AppDataHTMLTemplates string = "app/data/html_templates"
	// AppDataNotificationTemplates is the directory for the email
	// notification templates.
	AppDataNotificationTemplates string = "app/data/notifications/templates"
)

// DataDirs creates a map that returns pairs of "data target dir" to
// the "source data dir".
func DataDirs() map[string]string {
	return map[string]string{
		StaticDir:                    "/static",
		AppDataHTMLTemplates:         "/html_templates",
		AppDataNotificationTemplates: "/notifications/templates",
	}
}

// collectWithSuffix copies to dstDir all found directories that matches
// the suffix in the scanDirs directories recursively
func collectWithSuffix(dstDir string, scanDirs []string, suffix string) error {
	for _, sd := range scanDirs {
		fullDirName, err := filepath.Abs(sd)
		if err != nil {
			return err
		}
		if strings.HasSuffix(sd, suffix) {
			if err := CopyDir(fullDirName, dstDir); err != nil {
				return fmt.Errorf("cannot copy dir %s to %s: %w",
					fullDirName, dstDir, err)
			}
			continue
		}
		err = filepath.Walk(fullDirName,
			func(path string, info os.FileInfo, err error) error {
				if strings.HasSuffix(path, suffix) && info.IsDir() {
					if err := CopyDir(path, dstDir); err != nil {
						return fmt.Errorf("cannot copy dir %s to %s: %w",
							fullDirName, dstDir, err)
					}
				}
				return nil
			})
		if err != nil {
			return err
		}
	}
	return nil
}

// CollectFiles searches for files that must be collected in
// diferent directories and copies them in the bundle
func CollectFiles(dstDir string, scanDirs []string) error {
	dataDirs := DataDirs()
	for dst, suffix := range dataDirs {
		d := filepath.Join(dstDir, dst)
		if err := collectWithSuffix(d, scanDirs, suffix); err != nil {
			return err
		}
	}
	return nil
}

// PrepareBundleDirs completely deletes the dstDir, and
// creates the dir structure to populate an app bundle.
func PrepareBundleDirs(dstDir string) (string, error) {
	absDstDir, err := filepath.Abs(dstDir)
	if err != nil {
		return "", err
	}
	parentDir := filepath.Dir(dstDir)

	if err := os.MkdirAll(absDstDir, 0775); err != nil {
		return "", err
	}

	// delete everything from existing dir and create a new one
	baseDir := filepath.Base(dstDir)
	if err := os.RemoveAll(dstDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(parentDir, baseDir), 0775); err != nil {
		return "", err
	}

	// create al required dirs
	allDirs := []string{
		StaticDir,
		AppDir,
		ConfDir,
		DBMigrationsDir,
		AppDataDir,
		AppDataHTMLTemplates,
		AppDataNotificationTemplates,
	}
	for _, dd := range allDirs {
		if err := os.MkdirAll(filepath.Join(absDstDir, dd), 0775); err != nil {
			return "", err
		}
	}
	return absDstDir, nil
}

// PrepareExecutables finds directories under the `cmd` dir
// and tries to compile each of those packages, placing the
// output binary under the app directory.
func PrepareExecutables(projDir string, dstDir string) error {
	cmdDir, err := filepath.Abs(filepath.Join(projDir, "cmd"))
	if err != nil {
		return err
	}
	cmdDirInfos, err := ioutil.ReadDir(cmdDir)
	if err != nil {
		return err
	}

	for _, finfo := range cmdDirInfos {
		if finfo.IsDir() {
			inPkg := fmt.Sprintf("./cmd/%s", finfo.Name())
			outExe := fmt.Sprintf("%s/app/%s", dstDir, finfo.Name())
			cmd := exec.Command("go", "build", "-o", outExe, inPkg)
			cmd.Dir = projDir
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

// CompressBundle uses a call to exec.Command to run the
// `tar` utility to compress the bundle.
func CompressBundle(dstDir string) error {
	outfile := filepath.Base(dstDir) + ".tgz"
	parentDst := filepath.Dir(dstDir)
	cmd := exec.Command("tar", "-zcvf", outfile, filepath.Base(dstDir))
	cmd.Dir = parentDst
	return cmd.Run()
}

// PrepareBundle collects all static files, data files to be used by
// the executable, and config file.
func PrepareBundle(projDir string, dstDir string, extraPkgsDirs []string,
	envName string) error {
	absProjDir, err := filepath.Abs(projDir)
	if err != nil {
		return err
	}

	absExtraPkgsDirs := make([]string, 0, len(extraPkgsDirs))
	for _, epd := range extraPkgsDirs {
		abp, err := filepath.Abs(epd)
		if err != nil {
			return err
		}
		absExtraPkgsDirs = append(absExtraPkgsDirs, abp)
	}

	absDstDir, err := PrepareBundleDirs(dstDir)
	if err != nil {
		return err
	}

	absExtraPkgsDirs = append(absExtraPkgsDirs,
		filepath.Join(absProjDir, "pkg"))

	err = CollectFiles(absDstDir, absExtraPkgsDirs)
	if err != nil {
		return err
	}

	// copy the dbmigrations only for the project dir:
	err = CopyDir(filepath.Join(absProjDir, "dbmigrations"),
		filepath.Join(dstDir, DBMigrationsDir))
	if err != nil {
		return err
	}

	dstConfFile := filepath.Join(dstDir, ConfDir, "config.yaml")
	var srcConfFile string
	if len(envName) > 0 {
		srcConfFile = filepath.Join(projDir,
			fmt.Sprintf("config/config.%s.yaml", strings.ToLower(envName)))
	} else {
		srcConfFile = filepath.Join(projDir, "config/config.yaml")
	}
	err = CopyFile(srcConfFile, dstConfFile)
	if err != nil {
		return err
	}

	err = PrepareExecutables(absProjDir, absDstDir)
	if err != nil {
		return err
	}

	return CompressBundle(absDstDir)
}
