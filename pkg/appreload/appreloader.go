// Package appreload allows to detect changes in the source code for
// a Go application and rebuild the executable. In case of success it
// can kill the previous running version, and launch the newly created
// executable.
package appreload

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"

	"github.com/dhontecillas/hfw/pkg/bundler"
	"github.com/dhontecillas/hfw/pkg/obs"
)

// AppUpdaterConf has the config vars to set up an AppUpdater instance:
//   - ExecDir: the main sources directory for the executable (is also used
//     to generate the temp binary file name)
//   - PkgsDirs: the list of directories to observe for changes
//   - BuildFileExts: file extensions that should trigger a new binary build
//   - ResFilesExts: file extensions that should trigger an update of the resource files
type AppUpdaterConf struct {
	ExecDir       string
	PkgsDirs      []string
	BuildFileExts []string
	ResFileExts   []string
	BundleDir     string
}

// AppUpdater keeps tracks of changes in an app, and updated the files
// automatically. It also rebuilds and relaunches the app in case
// source files changed.
type AppUpdater struct {
	ins        *obs.Insighter
	conf       *AppUpdaterConf
	w          *watcher.Watcher
	runningCmd *exec.Cmd
}

// NewAppUpdater creates a new AppUpdater struct to keep track of changes
func NewAppUpdater(conf *AppUpdaterConf, ins *obs.Insighter) *AppUpdater {
	return &AppUpdater{
		ins:  ins,
		conf: conf,
	}
}

// Launch starts watching directories for changes, taking action for
// those file extensions that are listed in the build and resources
// list in the AppUpdaterConf
func (a *AppUpdater) Launch() {
	a.w = watcher.New()
	a.w.SetMaxEvents(20)

	if err := a.w.AddRecursive(a.conf.ExecDir); err != nil {
		a.ins.L.Err(err, fmt.Sprintf("cannot add recursive dir %s : %s",
			a.conf.ExecDir, err.Error()))
	}
	for _, pkgDir := range a.conf.PkgsDirs {
		if err := a.w.AddRecursive(pkgDir); err != nil {
			a.ins.L.Err(err, fmt.Sprintf("cannot add recursive dir %s : %s",
				a.conf.ExecDir, err.Error()))
		}
	}

	/* we do not use a filter hook, instead, we filter when we receive
	the event (see procRegularFile function).

	Otherwise we would use something like:
		r := regexp.MustCompile(strings.Join(a.conf.BuildFileExts, "|"))
		a.w.AddFilterHook(watcher.RegexFilterHook(r, true))
	*/

	// before watching, build the executable
	a.RebuildAndRelaunch()

	go a.procWatcherEvents()
	if err := a.w.Start(time.Second); err != nil {
		a.ins.L.Err(err, fmt.Sprintf("cannot start watcher: %s", err.Error()))
	}
	a.w.Wait()
}

// Shutdown stops watching for file changes
func (a *AppUpdater) Shutdown() {
	a.w.Close()
}

func (a *AppUpdater) procWatcherEvents() {
	for {
		select {
		case event := <-a.w.Event:
			if event.Mode().IsRegular() {
				a.procRegularFile(&event)
			}
		case err := <-a.w.Error:
			a.ins.L.Err(err, "watching event")
		case <-a.w.Closed:
			return
		}
	}
}

func (a *AppUpdater) procRegularFile(event *watcher.Event) {
	nm := event.Name()
	for _, ext := range a.conf.BuildFileExts {
		if strings.HasSuffix(nm, ext) {
			a.ins.L.Info(fmt.Sprintf("rebuilding for %s", event.String()))
			a.RebuildAndRelaunch()
			return
		}
	}

	for _, ext := range a.conf.ResFileExts {
		if strings.HasSuffix(nm, ext) {
			a.UpdateResource(event)
			return
		}
	}
}

// RebuildAndRelaunch tries to rebuild an executable, and if
// succesful kills old one to launch the new one.
func (a *AppUpdater) RebuildAndRelaunch() {
	if a.buildExecutable() {
		a.killCurrentExecutable()
		a.launchExecutable()
	} else {
		a.ins.L.Err(fmt.Errorf("build failed"), "")
	}
}

func (a *AppUpdater) killCurrentExecutable() {
	if a.runningCmd == nil || a.runningCmd.Process == nil {
		return
	}
	if a.runningCmd.ProcessState != nil {
		// the process already finished
		return
	}

	// kill the proces, first try Signaling and then Kill ?
	if err := a.runningCmd.Process.Signal(os.Interrupt); err != nil {
		a.ins.L.Err(err, "SIGINT failed")
	}
	// wait a second for "clean
	time.Sleep(time.Second)
	if err := a.runningCmd.Process.Kill(); err != nil {
		a.ins.L.Err(err, "Process Kill failed")
	}
}

// buildExecutable re-compiles the executable and returns
// true if it finished without issues
func (a *AppUpdater) buildExecutable() bool {
	a.ins.L.Info("building executable")
	buildCmd := exec.Command("go", "build", "-o",
		fmt.Sprintf("./dev_%s", filepath.Base(a.conf.ExecDir)),
		a.conf.ExecDir)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		a.ins.L.Warn(fmt.Sprintf("Build failed: %s", err.Error()))
		return false
	}
	a.ins.L.Info("Build completed")
	return true
}

// launchExecutable launches new executable
func (a *AppUpdater) launchExecutable() bool {
	a.runningCmd = exec.Command(fmt.Sprintf("./dev_%s", filepath.Base(a.conf.ExecDir)))
	a.runningCmd.Stdout = os.Stdout
	a.runningCmd.Stderr = os.Stderr
	a.runningCmd.Stdin = os.Stdin
	if err := a.runningCmd.Start(); err != nil {
		return false
	}
	return true
}

// UpdateResource collects a new / updated resource file and puts
// in the bundle directory.
func (a *AppUpdater) UpdateResource(e *watcher.Event) {
	// we ignore moves / chown / deletes
	if e.Op != watcher.Create && e.Op != watcher.Write {
		return
	}
	if !e.Mode().IsRegular() {
		return
	}
	dataDirs := bundler.DataDirs()

	a.ins.L.Warn(fmt.Sprintf("update %s", e.Path))
	for dst, dir := range dataDirs {
		if idx := strings.Index(e.Path, dir); idx >= 0 {
			dstPath := filepath.Join("./bundle", dst, e.Path[idx+len(dir):])
			parentDir := filepath.Dir(dstPath)
			if err := os.MkdirAll(parentDir, 0775); err != nil {
				a.ins.L.Err(err, fmt.Sprintf("error creating parent dir %s: %s",
					parentDir, err.Error()))
				continue
			}
			if err := bundler.CopyFile(e.Path, dstPath); err != nil {
				a.ins.L.Err(err, fmt.Sprintf("error copying file %s to %s: %s",
					e.Path, dstPath, err.Error()))
			}
		}
	}
}
