package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

/*
#define _GNU_SOURCE
#include <errno.h>
#include <fcntl.h>
#include <libgen.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mount.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

extern bool advance_arg(char *buf, char *cur, ssize_t size, bool required);
extern void error(char *msg);
extern void attach_userns(int pid);
extern int dosetns(int pid, char *nstype);

int mkdir_p(const char *dir, mode_t mode)
{
	const char *tmp = dir;
	const char *orig = dir;
	char *makeme;

	do {
		dir = tmp + strspn(tmp, "/");
		tmp = dir + strcspn(dir, "/");
		makeme = strndup(orig, dir - orig);
		if (*makeme) {
			if (mkdir(makeme, mode) && errno != EEXIST) {
				fprintf(stderr, "failed to create directory '%s': %s\n", makeme, strerror(errno));
				free(makeme);
				return -1;
			}
		}
		free(makeme);
	} while(tmp != dir);

	return 0;
}

void ensure_dir(char *dest) {
	struct stat sb;
	if (stat(dest, &sb) == 0) {
		if ((sb.st_mode & S_IFMT) == S_IFDIR)
			return;
		if (unlink(dest) < 0) {
			fprintf(stderr, "Failed to remove old %s: %s\n", dest, strerror(errno));
			_exit(1);
		}
	}
	if (mkdir(dest, 0755) < 0) {
		fprintf(stderr, "Failed to mkdir %s: %s\n", dest, strerror(errno));
		_exit(1);
	}
}

void ensure_file(char *dest) {
	struct stat sb;
	int fd;

	if (stat(dest, &sb) == 0) {
		if ((sb.st_mode & S_IFMT) != S_IFDIR)
			return;
		if (rmdir(dest) < 0) {
			fprintf(stderr, "Failed to remove old %s: %s\n", dest, strerror(errno));
			_exit(1);
		}
	}

	fd = creat(dest, 0755);
	if (fd < 0) {
		fprintf(stderr, "Failed to mkdir %s: %s\n", dest, strerror(errno));
		_exit(1);
	}
	close(fd);
}

void create(char *src, char *dest) {
	char *dirdup;
	char *destdirname;

	struct stat sb;
	if (stat(src, &sb) < 0) {
		fprintf(stderr, "source %s does not exist\n", src);
		_exit(1);
	}

	dirdup = strdup(dest);
	if (!dirdup)
		_exit(1);

	destdirname = dirname(dirdup);

	if (mkdir_p(destdirname, 0755) < 0) {
		fprintf(stderr, "failed to create path: %s\n", destdirname);
		free(dirdup);
		_exit(1);
	}
	free(dirdup);

	switch (sb.st_mode & S_IFMT) {
	case S_IFDIR:
		ensure_dir(dest);
		return;
	default:
		ensure_file(dest);
		return;
	}
}

void forkdomount(char *buf, char *cur, ssize_t size, pid_t pid) {
	char *src, *dest, *opts;

	attach_userns(pid);

	if (dosetns(pid, "mnt") < 0) {
		fprintf(stderr, "Failed setns to container mount namespace: %s\n", strerror(errno));
		_exit(1);
	}

	advance_arg(buf, cur, size, true);
	src = cur;

	advance_arg(buf, cur, size, true);
	dest = cur;

	create(src, dest);

	if (access(src, F_OK) < 0) {
		fprintf(stderr, "Mount source doesn't exist: %s\n", strerror(errno));
		_exit(1);
	}

	if (access(dest, F_OK) < 0) {
		fprintf(stderr, "Mount destination doesn't exist: %s\n", strerror(errno));
		_exit(1);
	}

	// Here, we always move recursively, because we sometimes allow
	// recursive mounts. If the mount has no kids then it doesn't matter,
	// but if it does, we want to move those too.
	if (mount(src, dest, "none", MS_MOVE | MS_REC, NULL) < 0) {
		fprintf(stderr, "Failed mounting %s onto %s: %s\n", src, dest, strerror(errno));
		_exit(1);
	}

	_exit(0);
}

void forkdoumount(char *buf, char *cur, ssize_t size, pid_t pid) {
	if (dosetns(pid, "mnt") < 0) {
		fprintf(stderr, "Failed setns to container mount namespace: %s\n", strerror(errno));
		_exit(1);
	}

	advance_arg(buf, cur, size, true);
	if (access(cur, F_OK) < 0) {
		fprintf(stderr, "Mount path doesn't exist: %s\n", strerror(errno));
		_exit(1);
	}

	if (umount2(cur, MNT_DETACH) < 0) {
		fprintf(stderr, "Error unmounting %s: %s\n", cur, strerror(errno));
		_exit(1);
	}
	_exit(0);
}

void forkmount(char *buf, char *cur, ssize_t size) {
	char *command = NULL;
	char *rootfs = NULL;
	pid_t pid = 0;

	// Get the subcommand
	if (advance_arg(buf, cur, size, false)) {
		command = cur;
	}

	// Get the pid
	advance_arg(buf, cur, size, false);
	if (command == NULL || (strcmp(cur, "--help") == 0 || strcmp(cur, "--version") == 0 || strcmp(cur, "-h") == 0)) {
		return;
	}

	pid = atoi(cur);

	// Check that we're root
	if (geteuid() != 0) {
		fprintf(stderr, "Error: forkmount requires root privileges\n");
		_exit(1);
	}

	// Call the subcommands
	if (strcmp(command, "mount") == 0) {
		forkdomount(buf, cur, size, pid);
	} else if (strcmp(command, "umount") == 0) {
		forkdoumount(buf, cur, size, pid);
	}
}
*/
import "C"

type cmdForkmount struct {
	cmd    *cobra.Command
	global *cmdGlobal
}

func (c *cmdForkmount) Command() *cobra.Command {
	// Main subcommand
	cmd := &cobra.Command{}
	cmd.Use = "forkmount"
	cmd.Short = "Perform mount operations"
	cmd.Long = `Description:
  Perform mount operations

  This set of internal commands are used for all container mount
  operations.
`
	cmd.Hidden = true

	// mount
	cmdMount := &cobra.Command{}
	cmdMount.Use = "mount <PID> <source> <destination>"
	cmdMount.Args = cobra.ExactArgs(3)
	cmdMount.RunE = c.Run
	cmd.AddCommand(cmdMount)

	// umount
	cmdUmount := &cobra.Command{}
	cmdUmount.Use = "umount <PID> <path>"
	cmdUmount.Args = cobra.ExactArgs(2)
	cmdUmount.RunE = c.Run
	cmd.AddCommand(cmdUmount)

	c.cmd = cmd
	return cmd
}

func (c *cmdForkmount) Run(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("This command should have been intercepted in cgo")
}
