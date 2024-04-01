package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	localDEBINIPath = `/etc/snclient/snclient_local.ini`
)

// this test requires:
// - snclient.deb
// it tries a installation and removes it afterwards
func TestDEBinstaller(t *testing.T) {
	// skip test unless requested, it will uninstall existing installations
	if os.Getenv("SNCLIENT_INSTALL_TEST") != "1" {
		t.Skipf("SKIPPED: pkg installer test requires env SNCLIENT_INSTALL_TEST=1")

		return
	}

	bin := getBinary()
	require.FileExistsf(t, "snclient.deb", "snclient.deb binary must exist")

	// install deb file
	runCmd(t, &cmd{
		Cmd:     "sudo",
		Args:    []string{"apt-get", "install", "-y", "./snclient.deb"},
		Timeout: DefaultCommandTimeout * 2,
	})

	requiredFiles := []string{
		"/usr/bin/snclient",
		"/usr/lib/snclient/node_exporter",
		"/etc/snclient/snclient.ini",
		"/etc/snclient/cacert.pem",
		"/etc/snclient/server.key",
		"/etc/snclient/server.crt",
		"/lib/systemd/system/snclient.service",
		"/etc/logrotate.d/snclient",
	}
	for _, file := range requiredFiles {
		require.FileExistsf(t, file, file+" has been installed")
	}

	runCmd(t, &cmd{
		Cmd:  "systemctl",
		Args: []string{"status", "snclient"},
		Like: []string{`/usr/bin/snclient`, `running`},
	})

	// add custom .ini
	runCmd(t, &cmd{
		Cmd:  "sudo",
		Args: []string{"touch", localDEBINIPath},
	})
	runCmd(t, &cmd{
		Cmd:  "sudo",
		Args: []string{"chmod", "666", localDEBINIPath},
	})
	writeFile(t, localDEBINIPath, localTestINI)
	writeFile(t, `snclient.ini`, localINI)

	runCmd(t, &cmd{
		Cmd:  "sudo",
		Args: []string{"systemctl", "restart", "snclient"},
	})

	waitUntilResponse(t, bin)

	// verify response
	runCmd(t, &cmd{
		Cmd:  bin,
		Args: []string{"run", "check_nsc_web", "-k", "-p", "test", "-u", "https://localhost:8443", "check_snclient_version"},
		Like: []string{`^SNClient\+ v`},
	})

	runCmd(t, &cmd{
		Cmd:  bin,
		Args: []string{"run", "check_nsc_web", "-k", "-p", "test", "-u", "https://localhost:8443", "check_uptime", "crit=uptime<2s", "warn=uptime<1s"},
		Like: []string{"OK - uptime"},
	})

	runCmd(t, &cmd{
		Cmd:  bin,
		Args: []string{"run", "check_nsc_web", "-k", "-p", "test", "-u", "https://localhost:8443", "check_cpu", "crit=load>101", "warn=load>101"},
		Like: []string{"OK - CPU load is ok."},
	})

	// cleanup
	os.Remove(localDEBINIPath)

	// uninstall pkg file
	runCmd(t, &cmd{
		Cmd:  "sudo",
		Args: []string{"apt-get", "remove", "-y", "--purge", "snclient"},
	})

	for _, file := range requiredFiles {
		assert.NoFileExistsf(t, file, file+" has been removed")
	}

	// remove remaining files
	os.Remove("snclient.ini")
}
