package internal

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void) {
	char *pid;
	pid = getenv("OB_PID");
	if (pid) {
		//fprintf(stdout, "got pid: `%s`\n", pid);
	} else {
		//fprintf(stdout, "no `OB_PID` env var, skip nsenter");
		return;
	}
	char *cmd;
	cmd = getenv("OB_CMD");
	if (cmd) {
		//fprintf(stdout, "got cmd: `%s`\n", cmd);
	} else {
		//fprintf(stdout, "no `OB_CMD` env var, skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			//fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(cmd);
	exit(0);
	return;
}
*/
import "C"
