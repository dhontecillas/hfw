package bundler

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

const (
	upSuffix      string = ".up.sql"
	downSuffix    string = ".down.sql"
	migrationsDir string = "migrations"
)

// MigrationFile contains info to match existing migration
// files with those coming from the sources
type MigrationFile struct {
	Idx          int64  // the index inside the directory
	Base         string // the base name of the migration
	Suffix       string // .up.sql or .down.sql
	FullPathFile string // the path + filename file
	Hash         string // hash of the content
	// Depends      []string // allow for a migration on depending on another one
}

// MigrationPair contains information for a pair
// of 'up' and 'down' migrations.
type MigrationPair struct {
	Up   MigrationFile
	Down MigrationFile
}

// MigrationFiles is a map of the base name (without the
// prefix Id, nor the 'up.sql' / 'down.sql' extentions)
// to the pair of files that compose a migration
type MigrationFiles map[string]MigrationPair

// UpdateMigrations finds the existing migrations in dstDir,
// and checks their hashes for changes, adding the new migrations
// as required. UpdateMigrations changes the existing IDs for
// their timestamps, so migration can
func UpdateMigrations(dstDir string, scanDirs []string, l logs.Logger) error {
	existing, issues := ListExistingMigrations(dstDir, l)

	for k, v := range existing {
		l.InfoMsg("existing migration").Str("name", k).
			Str("up_hash", v.Up.Hash).
			Str("down_hash", v.Down.Hash).Send()
	}
	srcMigrations := make(MigrationFiles)
	for _, scanDir := range scanDirs {
		dirMigrations, dirIssues := CollectMigrations(scanDir, l)
		issues = append(issues, dirIssues...)
		for k, v := range dirMigrations {
			if _, ok := srcMigrations[k]; ok {
				l.WarnMsg("duplicate migration name").Str("name", k)
			}
			srcMigrations[k] = v
		}
	}

	if len(srcMigrations) == 0 {
		l.Warn("No migrations found")
	} else {
		for k, v := range srcMigrations {
			l.InfoMsg("found migration").Str("name", k).
				Str("up_hash", v.Up.Hash).
				Str("down_hash", v.Down.Hash).Send()
		}
	}

	now := time.Now().Unix()
	// copy the migrations to the destination
	var serial int64
	for srcBaseName, srcM := range srcMigrations {
		serial++
		ex, ok := existing[srcBaseName]
		if ok {
			if ex.Up.Hash != srcM.Up.Hash {
				issues = append(issues, fmt.Errorf("different Up Hashes for %s and %s",
					ex.Up.FullPathFile, srcM.Up.Hash))
			}
			if ex.Down.Hash != srcM.Down.Hash {
				issues = append(issues, fmt.Errorf("different Down Hashes for %s and %s",
					ex.Down.FullPathFile, srcM.Down.Hash))
			}
		} else {
			mIdx := (now+serial)*100000 + srcM.Up.Idx
			copyM := MigrationPair{
				Up: MigrationFile{
					Idx:    mIdx,
					Base:   srcM.Up.Base,
					Suffix: srcM.Up.Suffix,
					FullPathFile: fmt.Sprintf("%s/%016d_%s%s",
						dstDir, mIdx, srcM.Up.Base, srcM.Up.Suffix),
					Hash: srcM.Up.Hash,
				},
				Down: MigrationFile{
					Idx:    mIdx,
					Base:   srcM.Down.Base,
					Suffix: srcM.Down.Suffix,
					FullPathFile: fmt.Sprintf("%s/%016d_%s%s",
						dstDir, mIdx, srcM.Down.Base, srcM.Down.Suffix),
					Hash: srcM.Down.Hash,
				},
			}

			if err := CopyMigrationFiles(&srcM, &copyM); err != nil {
				issues = append(issues, err)
				continue
			}
			existing[srcBaseName] = copyM
		}
	}

	if len(issues) > 0 {
		l.WarnMsg("").Send()

		for _, iss := range issues {
			l.Err(iss, "migration issue")
		}
	}
	return nil
}

// ListExistingMigrations computes the hash for the 'up' and 'down'
// migration files.
func ListExistingMigrations(dstDir string, l logs.Logger) (MigrationFiles, []error) {
	l.Info(fmt.Sprintf("Listing existing migrations in : %s", dstDir))
	migrationFiles := make(MigrationFiles)
	issues := []error{}

	issues = CollectMigrationsFromDir(dstDir, migrationFiles, issues, l)
	return migrationFiles, issues
}

// ParseMigrationFile extracts the components of a migrations file pair
func ParseMigrationFile(path string, name string) (*MigrationFile, error) {
	idx := strings.Index(name, "_")
	if idx < 1 {
		return nil, fmt.Errorf("no migration idx part")
	}
	if idx+1 >= len(name) {
		return nil, fmt.Errorf("no base part")
	}
	migIdx, err := strconv.ParseInt(name[:idx], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad idx format: %s", name[:idx])
	}

	stripped := name[idx+1:]
	idx = strings.Index(stripped, ".")
	if idx < 1 {
		return nil, fmt.Errorf("cannot find suffix")
	}

	baseName := stripped[:idx]
	suffix := stripped[idx:]
	if suffix != upSuffix && suffix != downSuffix {
		return nil, fmt.Errorf("bad suffix: %s", suffix)
	}

	fullPath := fmt.Sprintf("%s/%s", path, name)
	_, err = os.Lstat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %s", err)
	}
	hash, err := ComputeFileHash(fullPath)
	if err != nil {
		return nil, err
	}
	return &MigrationFile{
		Idx:          migIdx,
		Base:         baseName,
		Suffix:       suffix,
		FullPathFile: fullPath,
		Hash:         hash,
	}, nil
}

func migrationCandidate(path string, fi os.FileInfo) (*MigrationPair, []error) {
	var issues []error = []error{}
	upName := fi.Name()
	upMF, _ := ParseMigrationFile(path, upName)
	if upMF == nil || upMF.Suffix != upSuffix {
		return nil, issues
	}

	downName := fmt.Sprintf("%s%s", upName[:len(upName)-len(upSuffix)], downSuffix)
	downMF, err := ParseMigrationFile(path, downName)
	if downMF == nil || downMF.Suffix != downSuffix {
		if err != nil {
			issues = append(issues, fmt.Errorf(
				"no DOWN migration for: %s -> %s", downName, err.Error()))
		} else {
			issues = append(issues, fmt.Errorf(
				"no DOWN migration for: %s", downName))
		}
	}
	return &MigrationPair{
		Up:   *upMF,
		Down: *downMF,
	}, issues
}

// CollectMigrations scans a directory and all its descendants looking for
// directories called `migrations`.
func CollectMigrations(scanDir string, l logs.Logger) (MigrationFiles, []error) {
	migrationFiles := make(MigrationFiles)
	issues := []error{}

	_ = filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "/"+migrationsDir) {
			CollectMigrationsFromDir(path, migrationFiles, issues, l)
		}
		return nil
	})
	return migrationFiles, issues
}

// CollectMigrationsFromDir scans a single direcotry
func CollectMigrationsFromDir(path string, migrations MigrationFiles, issues []error,
	l logs.Logger) []error {
	l.Info(fmt.Sprintf("collecting migrations from %s", path))
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return append(issues, err)
	}
	for _, fi := range files {
		m, mIssues := migrationCandidate(path, fi)
		if m != nil {
			if _, ok := migrations[m.Up.Base]; ok {
				issues = append(issues, fmt.Errorf("DUPLICATE FILE FOR %s", m.Up.Base))
			} else {
				migrations[m.Up.Base] = *m
			}
		}
		issues = append(issues, mIssues...)
	}
	return issues
}

// CopyMigrationFiles copies ap pair of migration files from one dir to another
func CopyMigrationFiles(src *MigrationPair, dst *MigrationPair) error {
	if err := CopyFile(src.Up.FullPathFile, dst.Up.FullPathFile); err != nil {
		return err
	}
	return CopyFile(src.Down.FullPathFile, dst.Down.FullPathFile)
}
