package collectors

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseUsersFiltersHumanAccounts(t *testing.T) {
	input := strings.Join([]string{
		"root:x:0:0:root:/root:/bin/bash",
		"daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin",
		"service:x:999:999:service:/srv/service:/bin/bash",
		"alice:x:1000:1000:Alice:/home/alice:/bin/bash",
		"bob:x:1001:1001:Bob:/home/bob:/bin/false",
		"invalid-line",
		"# comment",
		"",
	}, "\n")

	got, err := parseUsers(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseUsers returned error: %v", err)
	}

	want := []User{
		{Username: "root", UID: "0", GID: "0", Home: "/root"},
		{Username: "alice", UID: "1000", GID: "1000", Home: "/home/alice"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseUsers() = %#v, want %#v", got, want)
	}
}

func TestIsHumanUserRejectsInvalidUID(t *testing.T) {
	if isHumanUser("not-a-number", "/bin/bash") {
		t.Fatal("invalid UID must not be considered a human user")
	}
}
