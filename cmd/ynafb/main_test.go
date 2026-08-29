package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreateCommands(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup [][]string
		want  string
	}{
		{
			name: "budget",
			args: []string{"budget", "create", "Home Budget"},
			want: "created budget 1\n",
		},
		{
			name:  "account",
			setup: [][]string{{"budget", "create", "Home Budget"}},
			args:  []string{"account", "create", "1", "Checking"},
			want:  "created account 1\n",
		},
		{
			name:  "allocation",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "1", "Groceries"}},
			args:  []string{"allocation", "create", "1", "1", "2500"},
			want:  "created allocation 1\n",
		},
		{
			name:  "category",
			setup: [][]string{{"budget", "create", "Home Budget"}},
			args:  []string{"category", "create", "1", "Groceries"},
			want:  "created category 1\n",
		},
		{
			name:  "goal",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "1", "Savings"}},
			args:  []string{"goal", "create", "1", "Emergency Fund", "target", "2026-08-28", "null", "1", "5000"},
			want:  "created goal 1\n",
		},
		{
			name:  "payee",
			setup: [][]string{{"budget", "create", "Home Budget"}},
			args:  []string{"payee", "create", "1", "Market"},
			want:  "created payee 1\n",
		},
		{
			name:  "payee-default-split",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "1", "Checking"}, {"category", "create", "1", "Groceries"}, {"payee", "create", "1", "Market"}},
			args:  []string{"payee-default-split", "create", "1", "1", "1", "1", "2500", "0"},
			want:  "created payee-default-split 1\n",
		},
		{
			name:  "transaction",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "1", "Checking"}, {"payee", "create", "1", "Market"}},
			args:  []string{"transaction", "create", "2026-08-28", "1", "1", "false", "weekly shop"},
			want:  "created transaction 1\n",
		},
		{
			name:  "transaction-split",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "1", "Checking"}, {"category", "create", "1", "Groceries"}, {"payee", "create", "1", "Market"}, {"transaction", "create", "2026-08-28", "1", "1", "false", "weekly shop"}},
			args:  []string{"transaction-split", "create", "1", "1", "1", "1", "2500", "0"},
			want:  "created transaction-split 1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "ynafb.db")

			for _, setupArgs := range tt.setup {
				stdout, stderr, exitCode := invoke(t, dbPath, setupArgs...)
				if exitCode != 0 {
					t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(setupArgs, " "), exitCode, stdout, stderr)
				}
			}

			stdout, stderr, exitCode := invoke(t, dbPath, tt.args...)
			if exitCode != 0 {
				t.Fatalf("command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
			}

			if stdout != tt.want {
				t.Fatalf("unexpected stdout: got %q want %q", stdout, tt.want)
			}

			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
		})
	}
}

func TestAccountListTransactions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "1", "Checking"},
		{"account", "create", "1", "Savings"},
		{"category", "create", "1", "Groceries"},
		{"category", "create", "1", "Household"},
		{"payee", "create", "1", "Cafe"},
		{"payee", "create", "1", "Bank"},
		{"payee", "create", "1", "Market"},
		{"transaction", "create", "2026-08-28", "1", "1", "true", "coffee"},
		{"transaction-split", "create", "1", "1", "1", "1", "1000", "0"},
		{"transaction", "create", "2026-08-29", "1", "2", "false", "move money"},
		{"transaction-split", "create", "2", "2", "1", "1", "1500", "0"},
		{"transaction", "create", "2026-08-30", "1", "3", "false", "weekly shop"},
		{"transaction-split", "create", "3", "1", "1", "1", "2000", "0"},
		{"transaction-split", "create", "3", "1", "1", "2", "500", "0"},
		{"transaction", "create", "2026-08-31", "2", "3", "false", "should not appear"},
		{"transaction-split", "create", "4", "2", "2", "1", "9999", "0"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "list-transactions", "1")
	if exitCode != 0 {
		t.Fatalf("command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"ID  DATE        PAYEE   TARGET     OUTFLOW  INFLOW  RECONCILED  NOTE",
		"1   2026-08-28  Cafe    Groceries  1000     0       true        coffee",
		"2   2026-08-29  Bank    Savings    1500     0       false       move money",
		"3   2026-08-30  Market  split      2500     0       false       weekly shop",
		"                        Groceries  2000     0                   ",
		"                        Household  500      0                   ",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountListTransactionsEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "create", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("setup failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "account", "create", "1", "Checking")
	if exitCode != 0 {
		t.Fatalf("setup failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "account", "list-transactions", "1")
	if exitCode != 0 {
		t.Fatalf("command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := "ID  DATE  PAYEE  TARGET  OUTFLOW  INFLOW  RECONCILED  NOTE\n"
	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountListTransactionsMissingAccount(t *testing.T) {
	stdout, stderr, exitCode := invoke(t, filepath.Join(t.TempDir(), "ynafb.db"), "account", "list-transactions")
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}

	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	want := "error: account list-transactions requires [account]\n"
	if stderr != want {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, want)
	}
}

func TestUsageForMissingArguments(t *testing.T) {
	stdout, stderr, exitCode := invoke(t, filepath.Join(t.TempDir(), "ynafb.db"))
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}

	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	if !strings.Contains(stderr, "Usage:\n") {
		t.Fatalf("expected usage in stderr, got %q", stderr)
	}
}

func invoke(t *testing.T, dbPath string, args ...string) (string, string, int) {
	t.Helper()

	restore := chdirToRepoRoot(t)
	defer restore()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(append([]string{"--db", dbPath}, args...), &stdout, &stderr)
	return stdout.String(), stderr.String(), exitCode
}

func chdirToRepoRoot(t *testing.T) func() {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}

	return func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}
