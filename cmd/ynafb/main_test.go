package main

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/importer"
	"samuellando.com/YNAFB/internal/importer/testparser"
)

func init() {
	importer.Register(testparser.Parser{})
}

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
			args:  []string{"account", "create", "Checking"},
			want:  "created account 1\n",
		},
		{
			name:  "allocation",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Groceries"}},
			args:  []string{"allocation", "create", "2026-08", "Groceries", "25.00"},
			want:  "created allocation 1\n",
		},
		{
			name:  "category",
			setup: [][]string{{"budget", "create", "Home Budget"}},
			args:  []string{"category", "create", "Groceries"},
			want:  "created category 1\n",
		},
		{
			name:  "goal",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Savings"}},
			args:  []string{"goal", "create", "monthly", "2026-08", "null", "Savings", "50.00"},
			want:  "created goal 1\n",
		},
		{
			name:  "payee",
			setup: [][]string{{"budget", "create", "Home Budget"}},
			args:  []string{"payee", "create", "Market"},
			want:  "created payee 1\n",
		},
		{
			name:  "payee default-category",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"category", "create", "Groceries"}, {"payee", "create", "Market"}},
			args:  []string{"payee", "default-category", "create", "Market", "100", "--category", "Groceries"},
			want:  "created payee default-category 1\n",
		},
		{
			name:  "transaction",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"payee", "create", "Market"}},
			args:  []string{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "weekly shop"},
			want:  "created transaction 1\n",
		},
		{
			name:  "transaction category",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"category", "create", "Groceries"}, {"payee", "create", "Market"}, {"transaction", "create", "2026-08-28", "Checking", "Market", "2500", "0", "weekly shop"}},
			args:  []string{"transaction", "category", "create", "1", "25.00", "0.00", "--category", "Groceries"},
			want:  "created transaction category 1\n",
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

type updateVerify struct {
	args     []string
	contains string
}

type updateTestCase struct {
	name   string
	args   []string
	setup  [][]string
	want   string
	verify []updateVerify
}

func TestUpdateCommands(t *testing.T) {
	tests := []updateTestCase{
		{
			name:  "budget",
			setup: [][]string{{"budget", "create", "Home Budget"}},
			args:  []string{"budget", "update", "Home Budget", "Renamed"},
			want:  "updated budget \"Renamed\"\n",
			verify: []updateVerify{
				{args: []string{"budget", "list"}, contains: "Renamed"},
			},
		},
		{
			name:  "account",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}},
			args:  []string{"account", "update", "Checking", "Main"},
			want:  "updated account \"Main\"\n",
			verify: []updateVerify{
				{args: []string{"account", "list"}, contains: "Main"},
			},
		},
		{
			name:  "allocation",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Groceries"}, {"allocation", "create", "2026-08", "Groceries", "25.00"}},
			args:  []string{"allocation", "update", "2026-08", "Groceries", "30.00"},
			want:  "updated allocation 2026-08 \"Groceries\"\n",
			verify: []updateVerify{
				{args: []string{"allocation", "list"}, contains: "30.00"},
			},
		},
		{
			name:  "category",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Groceries"}},
			args:  []string{"category", "update", "Groceries", "Food"},
			want:  "updated category \"Food\"\n",
			verify: []updateVerify{
				{args: []string{"category", "list"}, contains: "Food"},
			},
		},
		{
			name:  "goal",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Savings"}, {"goal", "create", "monthly", "2026-08", "null", "Savings", "50.00"}},
			args:  []string{"goal", "update", "Savings", "save", "2026-08", "2026-12", "100.00"},
			want:  "updated goal for category \"Savings\"\n",
			verify: []updateVerify{
				{args: []string{"goal", "list"}, contains: "100.00"},
			},
		},
		{
			name:  "payee",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"payee", "create", "Market"}},
			args:  []string{"payee", "update", "Market", "Supermarket"},
			want:  "updated payee \"Supermarket\"\n",
			verify: []updateVerify{
				{args: []string{"payee", "list"}, contains: "Supermarket"},
			},
		},
		{
			name:  "payee default-category",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"category", "create", "Groceries"}, {"payee", "create", "Market"}, {"payee", "default-category", "create", "Market", "100", "--category", "Groceries"}},
			args:  []string{"payee", "default-category", "update", "1", "Market", "50", "--category", "Groceries"},
			want:  "updated payee default-category 1\n",
		},
		{
			name:  "transaction",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"payee", "create", "Market"}, {"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "weekly shop"}},
			args:  []string{"transaction", "update", "1", "2026-08-29", "Checking", "Market", "30.00", "0.00", "new note"},
			want:  "updated transaction 1\n",
			verify: []updateVerify{
				{args: []string{"transaction", "list"}, contains: "30.00"},
				{args: []string{"transaction", "list"}, contains: "2026-08-29"},
				{args: []string{"transaction", "list"}, contains: "new note"},
			},
		},
		{
			name:  "transaction category",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"category", "create", "Groceries"}, {"payee", "create", "Market"}, {"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "weekly shop"}, {"transaction", "category", "create", "1", "25.00", "0.00", "--category", "Groceries"}},
			args:  []string{"transaction", "category", "update", "1", "1", "30.00", "0.00", "--category", "Groceries"},
			want:  "updated transaction category 1\n",
			verify: []updateVerify{
				{args: []string{"account", "show", "Checking"}, contains: "30.00"},
			},
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

			for _, v := range tt.verify {
				checkOut, checkErr, checkExit := invoke(t, dbPath, v.args...)
				if checkExit != 0 {
					t.Fatalf("verify command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(v.args, " "), checkExit, checkOut, checkErr)
				}
				if !strings.Contains(checkOut, v.contains) {
					t.Fatalf("verify command %q: stdout %q does not contain %q", strings.Join(v.args, " "), checkOut, v.contains)
				}
			}
		})
	}
}

func TestUpdateCommandsFail(t *testing.T) {
	tests := []struct {
		name    string
		setup   [][]string
		args    []string
		wantErr string
	}{
		{
			name:    "budget",
			setup:   [][]string{{"budget", "create", "Home Budget"}},
			args:    []string{"budget", "update", "Nope", "Renamed"},
			wantErr: `unknown budget "Nope"`,
		},
		{
			name:    "account",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}},
			args:    []string{"account", "update", "Nope", "Main"},
			wantErr: `unknown account "Nope"`,
		},
		{
			name:    "allocation",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Groceries"}},
			args:    []string{"allocation", "update", "2026-08", "Groceries", "30.00"},
			wantErr: `no allocation for 2026-08 "Groceries"`,
		},
		{
			name:    "category",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Groceries"}},
			args:    []string{"category", "update", "Nope", "Food"},
			wantErr: `unknown category "Nope"`,
		},
		{
			name:    "goal",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Savings"}},
			args:    []string{"goal", "update", "Savings", "save", "2026-08", "2026-12", "100.00"},
			wantErr: `no goal for category "Savings"`,
		},
		{
			name:    "payee",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"payee", "create", "Market"}},
			args:    []string{"payee", "update", "Nope", "Supermarket"},
			wantErr: `unknown payee "Nope"`,
		},
		{
			name:    "payee default-category",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"category", "create", "Groceries"}, {"payee", "create", "Market"}},
			args:    []string{"payee", "default-category", "update", "99", "Market", "50", "--category", "Groceries"},
			wantErr: `unknown payee default-category 99`,
		},
		{
			name:    "transaction",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"payee", "create", "Market"}},
			args:    []string{"transaction", "update", "99", "2026-08-29", "Checking", "Market", "30.00", "0.00", "note"},
			wantErr: `unknown transaction 99`,
		},
		{
			name:    "transaction category",
			setup:   [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"category", "create", "Groceries"}, {"payee", "create", "Market"}},
			args:    []string{"transaction", "category", "update", "99", "1", "30.00", "0.00", "--category", "Groceries"},
			wantErr: `unknown transaction category 99`,
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

			assertCommandFails(t, dbPath, tt.wantErr, tt.args...)
		})
	}
}

func TestTransactionCreateAutoCreatesPayee(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "transaction", "create", "2026-08-28", "Checking", "Corner Store", "12.50", "0.00", "snacks")
	if exitCode != 0 {
		t.Fatalf("transaction create failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if stdout != "created transaction 1\n" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "transaction", "category", "create", "1", "12.50", "0.00", "--category", "Groceries")
	if exitCode != 0 {
		t.Fatalf("transaction category create failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "account", "show", "Checking")
	if exitCode != 0 {
		t.Fatalf("account show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Name: Checking",
		"Balance: -12.50",
		"Reconciled balance: 0.00",
		"",
		"ID  DATE        PAYEE         TARGET     OUTFLOW  INFLOW  RECONCILED  NOTE",
		"1   2026-08-28  Corner Store  Groceries  12.50    0.00    false       snacks",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountListTransactionsIncludesTransactionsWithoutCategories(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "ws-visa"},
		{"transaction", "create", "2026-08-28", "ws-visa", "Online Shop", "42.00", "0.00", "pending import"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "show", "ws-visa")
	if exitCode != 0 {
		t.Fatalf("account show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Name: ws-visa",
		"Balance: -42.00",
		"Reconciled balance: 0.00",
		"",
		"ID  DATE        PAYEE        TARGET  OUTFLOW  INFLOW  RECONCILED  NOTE",
		"1   2026-08-28  Online Shop          42.00    0.00    false       pending import",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountListTransactionsShowsMismatchedSingleCategorySeparately(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "42.00", "0.00", "import mismatch"},
		{"transaction", "category", "create", "1", "40.00", "0.00", "--category", "Groceries"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "show", "Checking")
	if exitCode != 0 {
		t.Fatalf("account show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Name: Checking",
		"Balance: -42.00",
		"Reconciled balance: 0.00",
		"",
		"ID  DATE        PAYEE        TARGET     OUTFLOW  INFLOW  RECONCILED  NOTE",
		"1   2026-08-28  Online Shop  category   42.00    0.00    false       import mismatch",
		"                             Groceries  40.00    0.00                ",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountImport(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "ws-visa"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	statementFile := filepath.Join(t.TempDir(), "statement.txt")
	if err := os.WriteFile(statementFile, []byte("ynafb-test-statement"), 0o644); err != nil {
		t.Fatalf("write statement file: %v", err)
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "import", "ws-visa", statementFile)
	if exitCode != 0 {
		t.Fatalf("import failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if stdout != "imported 2 transactions\n" {
		t.Fatalf("unexpected import stdout: %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "account", "show", "ws-visa")
	if exitCode != 0 {
		t.Fatalf("account show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if !strings.Contains(stdout, "Test Merchant") {
		t.Fatalf("expected Test Merchant in listing, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Incoming Transfer") {
		t.Fatalf("expected Incoming Transfer in listing, got:\n%s", stdout)
	}
}

func TestAccountListTransactions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"payee", "create", "Cafe"},
		{"payee", "create", "Bank"},
		{"payee", "create", "Market"},
		{"transaction", "create", "2026-08-28", "Checking", "Cafe", "10.00", "0.00", "coffee"},
		{"transaction", "category", "create", "1", "10.00", "0.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-29", "Checking", "Bank", "15.00", "0.00", "move money"},
		{"transaction", "category", "create", "2", "15.00", "0.00", "--other-account", "Savings"},
		{"transaction", "create", "2026-08-30", "Checking", "Market", "25.00", "0.00", "weekly shop"},
		{"transaction", "category", "create", "3", "20.00", "0.00", "--category", "Groceries"},
		{"transaction", "category", "create", "3", "5.00", "0.00", "--category", "Household"},
		{"transaction", "create", "2026-08-31", "Savings", "Market", "99.99", "0.00", "should not appear"},
		{"transaction", "category", "create", "4", "99.99", "0.00", "--category", "Groceries"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "show", "Checking")
	if exitCode != 0 {
		t.Fatalf("command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Name: Checking",
		"Balance: -50.00",
		"Reconciled balance: 0.00",
		"",
		"ID  DATE        PAYEE   TARGET     OUTFLOW  INFLOW  RECONCILED  NOTE",
		"3   2026-08-30  Market  category   25.00    0.00    false       weekly shop",
		"                        Groceries  20.00    0.00                ",
		"                        Household  5.00     0.00                ",
		"2   2026-08-29  Bank    @Savings   15.00    0.00    false       move money",
		"1   2026-08-28  Cafe    Groceries  10.00    0.00    false       coffee",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountListTransactionsShowsMirroredTransfersForOtherAccount(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "ws-visa"},
		{"account", "create", "merry"},
		{"category", "create", "groceries"},
		{"payee", "create", "superC"},
		{"transaction", "create", "2026-08-09", "ws-visa", "superC", "1.00", "0.00", ""},
		{"transaction", "category", "create", "1", "0.25", "0.00", "--category", "groceries"},
		{"transaction", "category", "create", "1", "0.75", "0.00", "--other-account", "merry"},
		{"transaction", "create", "2026-08-09", "ws-visa", "superC", "2.00", "0.00", ""},
		{"transaction", "category", "create", "2", "1.00", "0.00", "--category", "groceries"},
		{"transaction", "category", "create", "2", "1.00", "0.00", "--other-account", "merry"},
		{"transaction", "create", "2026-08-09", "ws-visa", "superC", "2.50", "0.00", ""},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "show", "merry")
	if exitCode != 0 {
		t.Fatalf("command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Name: merry",
		"Balance: 1.75",
		"Reconciled balance: 0.00",
		"",
		"ID  DATE        PAYEE   TARGET    OUTFLOW  INFLOW  RECONCILED  NOTE",
		"1   2026-08-09  superC  @ws-visa  0.00     0.75    false       ",
		"2   2026-08-09  superC  @ws-visa  0.00     1.00    false       ",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountReconcileMatchingBalance(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "25.00", "0.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-30", "Checking", "Market", "10.00", "0.00", "later"},
	})

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, "-25.00\n", "account", "reconcile", "Checking", "2026-08-28")
	if exitCode != 0 {
		t.Fatalf("reconcile failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	want := strings.Join([]string{
		"Balance as of 2026-08-28: -25.00",
		"Statement balance: reconciled 1 transactions through 2026-08-28 in \"Checking\"",
		"",
	}, "\n")
	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Reconciled balance: -25.00") {
		t.Fatalf("expected reconciled balance -25.00, got %q", show)
	}
	if !strings.Contains(show, "true") {
		t.Fatalf("expected on-date transaction reconciled, got %q", show)
	}
	if !strings.Contains(show, "2026-08-30  Market             10.00    0.00    false       later") {
		t.Fatalf("expected later transaction to remain unreconciled, got %q", show)
	}
}

func TestAccountReconcileMismatch(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "25.00", "0.00", "--category", "Groceries"},
	})

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, "-1.00\n", "account", "reconcile", "Checking", "2026-08-28")
	if exitCode == 0 {
		t.Fatalf("expected reconcile to fail, got stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "balance mismatch: statement -1.00 != calculated -25.00") {
		t.Fatalf("expected mismatch error, got %q", stderr)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Reconciled balance: 0.00") {
		t.Fatalf("expected no reconciliation on mismatch, got %q", show)
	}
	if !strings.Contains(show, "false") {
		t.Fatalf("expected transaction to remain unreconciled, got %q", show)
	}
}

func TestAccountReconcileSubsequentDates(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "25.00", "0.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-30", "Checking", "Market", "10.00", "0.00", "later"},
		{"transaction", "category", "create", "2", "10.00", "0.00", "--category", "Groceries"},
	})

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, "-35.00\n", "account", "reconcile", "Checking", "2026-08-30")
	if exitCode != 0 {
		t.Fatalf("reconcile failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "reconciled 2 transactions") {
		t.Fatalf("expected both transactions reconciled, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Reconciled balance: -35.00") {
		t.Fatalf("expected reconciled balance -35.00, got %q", show)
	}
}

func TestAccountReconcileIncomingTransferReconciledIndependently(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"transaction", "create", "2026-08-28", "Checking", "Bank", "15.00", "0.00", "move"},
		{"transaction", "category", "create", "1", "15.00", "0.00", "--other-account", "Savings"},
	})

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, "15.00\n", "account", "reconcile", "Savings", "2026-08-31")
	if exitCode != 0 {
		t.Fatalf("reconcile failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "Balance as of 2026-08-31: 15.00") {
		t.Fatalf("expected incoming transfer included in balance, got %q", stdout)
	}
	if !strings.Contains(stdout, "reconciled 1 transactions") {
		t.Fatalf("expected incoming transfer reconciled for receiving account, got %q", stdout)
	}

	savingsShow := showAccount(t, dbPath, "Savings")
	if !strings.Contains(savingsShow, "Reconciled balance: 15.00") {
		t.Fatalf("expected receiving account reconciled balance 15.00, got %q", savingsShow)
	}
	if !strings.Contains(savingsShow, "1   2026-08-28  Bank   @Checking  0.00     15.00   true        move") {
		t.Fatalf("expected incoming transfer reconciled on receiving account, got %q", savingsShow)
	}

	checkingShow := showAccount(t, dbPath, "Checking")
	if !strings.Contains(checkingShow, "1   2026-08-28  Bank   @Savings  15.00    0.00    false       move") {
		t.Fatalf("expected source transfer to remain unreconciled on its own account, got %q", checkingShow)
	}
}

func TestAccountReconcileCancel(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "shop"},
	})

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, "\n", "account", "reconcile", "Checking", "2026-08-28")
	if exitCode != 0 {
		t.Fatalf("cancel should not error, stdout=%q stderr=%q", stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "cancelled") {
		t.Fatalf("expected cancelled message, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Reconciled balance: 0.00") {
		t.Fatalf("expected no reconciliation after cancel, got %q", show)
	}
}

func TestTransactionCategoryCreateRejectsMultipleTargets(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"payee", "create", "Market"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "weekly shop"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "transaction", "category", "create", "1", "25.00", "0.00", "--category", "Groceries", "--other-account", "Savings")
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}

	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	want := "error: set exactly one of --category, --other-account, or --income\n"
	if stderr != want {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, want)
	}
}

func TestAccountListTransactionsEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "create", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("setup failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "account", "create", "Checking")
	if exitCode != 0 {
		t.Fatalf("setup failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "account", "show", "Checking")
	if exitCode != 0 {
		t.Fatalf("command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := "Name: Checking\nBalance: 0.00\nReconciled balance: 0.00\n\nID  DATE  PAYEE  TARGET  OUTFLOW  INFLOW  RECONCILED  NOTE\n"
	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestAccountListTransactionsMissingAccount(t *testing.T) {
	stdout, stderr, exitCode := invoke(t, filepath.Join(t.TempDir(), "ynafb.db"), "account", "show")
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}

	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	want := "error: account show requires [account]\n"
	if stderr != want {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, want)
	}
}

func TestBudgetFlagRequiredWhenMultipleBudgetsExist(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	for _, args := range [][]string{{"budget", "create", "Home Budget"}, {"budget", "create", "Travel Budget"}} {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "create", "Checking")
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}

	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	want := "error: multiple budgets exist; pass --budget [name]\n"
	if stderr != want {
		t.Fatalf("unexpected stderr: got %q want %q", stderr, want)
	}
}

func TestBudgetFlagSelectsBudget(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	for _, args := range [][]string{{"budget", "create", "Home Budget"}, {"budget", "create", "Travel Budget"}} {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "--budget", "Travel Budget", "account", "create", "Checking")
	if exitCode != 0 {
		t.Fatalf("command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if stdout != "created account 1\n" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "--budget", "Travel Budget", "account", "show", "Checking")
	if exitCode != 0 {
		t.Fatalf("list command failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if stdout != "Name: Checking\nBalance: 0.00\nReconciled balance: 0.00\n\nID  DATE  PAYEE  TARGET  OUTFLOW  INFLOW  RECONCILED  NOTE\n" {
		t.Fatalf("unexpected show stdout: %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestCommandRequiringBudgetFailsWhenNoBudgetsExist(t *testing.T) {
	stdout, stderr, exitCode := invoke(t, filepath.Join(t.TempDir(), "ynafb.db"), "account", "create", "Checking")
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}

	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	want := "error: no budgets exist; create one with `ynafb budget create [name]`\n"
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
	if !strings.Contains(stderr, "[--budget name]") {
		t.Fatalf("expected budget flag in usage, got %q", stderr)
	}
}

func TestListCommands(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"payee", "create", "Market"},
		{"payee", "create", "Cafe"},
		{"allocation", "create", "2026-08", "Groceries", "50.00"},
		{"allocation", "create", "2026-08", "Household", "25.00"},
		{"goal", "create", "monthly", "2020-01", "null", "Groceries", "1000.00"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "shop"},
		{"transaction", "create", "2026-08-29", "Savings", "Cafe", "10.00", "0.00", "coffee"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"budget", "list"}, "NAME         BALANCE  RECONCILED BALANCE\nHome Budget  -35.00   0.00\n"},
		{[]string{"account", "list"}, "NAME      BALANCE  RECONCILED BALANCE\nChecking  -25.00   0.00\nSavings   -10.00   0.00\n"},
		{[]string{"category", "list"}, "NAME\nGroceries\nHousehold\n"},
		{[]string{"payee", "list"}, "NAME\nCafe\nMarket\n"},
		{[]string{"allocation", "list"}, "MONTH    CATEGORY   AMOUNT\n2026-08  Groceries  50.00\n2026-08  Household  25.00\n"},
		{[]string{"goal", "list"}, "TYPE     CATEGORY   START    END  AMOUNT   MONTHLY\nmonthly  Groceries  2020-01       1000.00  1000.00\n"},
		{[]string{"transaction", "list"}, "ID  DATE        ACCOUNT   PAYEE   OUTFLOW  INFLOW  NOTE\n2   2026-08-29  Savings   Cafe    10.00    0.00    coffee\n1   2026-08-28  Checking  Market  25.00    0.00    shop\n"},
	}

	for _, tt := range tests {
		stdout, stderr, exitCode := invoke(t, dbPath, tt.args...)
		if exitCode != 0 {
			t.Fatalf("command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(tt.args, " "), exitCode, stdout, stderr)
		}
		if stdout != tt.want {
			t.Fatalf("command %q: got %q want %q", strings.Join(tt.args, " "), stdout, tt.want)
		}
		if stderr != "" {
			t.Fatalf("command %q: unexpected stderr %q", strings.Join(tt.args, " "), stderr)
		}
	}
}

func TestDeleteCommands(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"payee", "create", "Market"},
		{"allocation", "create", "2026-08", "Groceries", "50.00"},
		{"goal", "create", "monthly", "2026-09", "null", "Groceries", "1000.00"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "20.00", "0.00", "--category", "Groceries"},
		{"payee", "default-category", "create", "Market", "100", "--category", "Groceries"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"transaction", "category", "delete", "1"}, "deleted transaction category 1\n"},
		{[]string{"payee", "default-category", "delete", "1"}, "deleted payee default-category 1\n"},
		{[]string{"transaction", "delete", "1"}, "deleted transaction 1\n"},
		{[]string{"allocation", "delete", "2026-08", "Groceries"}, "deleted allocation 2026-08 \"Groceries\"\n"},
		{[]string{"goal", "delete", "Groceries"}, "deleted goal for category \"Groceries\"\n"},
		{[]string{"payee", "delete", "Market"}, "deleted payee \"Market\"\n"},
		{[]string{"category", "delete", "Groceries"}, "deleted category \"Groceries\"\n"},
		{[]string{"account", "delete", "Checking"}, "deleted account \"Checking\"\n"},
	}

	for _, tt := range tests {
		stdout, stderr, exitCode := invoke(t, dbPath, tt.args...)
		if exitCode != 0 {
			t.Fatalf("command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(tt.args, " "), exitCode, stdout, stderr)
		}
		if stdout != tt.want {
			t.Fatalf("command %q: got %q want %q", strings.Join(tt.args, " "), stdout, tt.want)
		}
		if stderr != "" {
			t.Fatalf("command %q: unexpected stderr %q", strings.Join(tt.args, " "), stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "transaction", "list")
	if exitCode != 0 {
		t.Fatalf("transaction list failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "ID  DATE  ACCOUNT  PAYEE  OUTFLOW  INFLOW  NOTE\n" {
		t.Fatalf("expected empty transaction list after deletes, got %q", stdout)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "budget", "delete", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("budget delete failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "deleted budget \"Home Budget\"\n" {
		t.Fatalf("unexpected budget delete stdout %q", stdout)
	}
}

func TestCascadeDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"payee", "create", "Market"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "25.00", "0.00", "--category", "Groceries"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "delete", "Checking")
	if exitCode != 0 {
		t.Fatalf("account delete failed: stdout=%q stderr=%q", stdout, stderr)
	}

	// Verify transaction is gone (cascade delete)
	stdout, stderr, exitCode = invoke(t, dbPath, "transaction", "list")
	if exitCode != 0 {
		t.Fatalf("transaction list failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "ID  DATE  ACCOUNT  PAYEE  OUTFLOW  INFLOW  NOTE\n" {
		t.Fatalf("expected cascaded transactions to be gone, got %q", stdout)
	}

	// Delete budget - this should cascade delete everything
	stdout, stderr, exitCode = invoke(t, dbPath, "budget", "delete", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("budget delete failed: stdout=%q stderr=%q", stdout, stderr)
	}

	// After budget deletion, verify budget is gone
	stdout, stderr, exitCode = invoke(t, dbPath, "budget", "list")
	if exitCode != 0 {
		t.Fatalf("budget list failed after delete: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "NAME  BALANCE  RECONCILED BALANCE\n" {
		t.Fatalf("expected no budgets after budget delete, got %q", stdout)
	}

	// After budget deletion, we can't list accounts (no budget context)
	// This is expected behavior - verify we can't list accounts
	stdout, stderr, exitCode = invoke(t, dbPath, "account", "list")
	if exitCode == 0 {
		t.Fatalf("account list should fail when no budget exists, got stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "no budgets exist") {
		t.Fatalf("expected 'no budgets exist' error, got stderr=%q", stderr)
	}
}

func TestBudgetDeleteCascadesGoalsAndAllocations(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"category", "create", "Groceries"},
		{"allocation", "create", "2026-08", "Groceries", "50.00"},
		{"goal", "create", "monthly", "2026-08", "null", "Groceries", "50.00"},
	})

	assertRowCount(t, dbPath, "goal", 1)
	assertRowCount(t, dbPath, "allocation", 1)

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "delete", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("budget delete failed: stdout=%q stderr=%q", stdout, stderr)
	}

	assertRowCount(t, dbPath, "goal", 0)
	assertRowCount(t, dbPath, "allocation", 0)
}

func assertRowCount(t *testing.T, dbPath, table string, want int) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var got int
	if err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s count = %d, want %d", table, got, want)
	}
}

func TestBudgetFlagUnknownBudget(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
	})

	assertCommandFails(t, dbPath, `unknown budget "Nope"`, "--budget", "Nope", "account", "list")
}

func TestUnknownResource(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	assertCommandFails(t, dbPath, `unsupported resource "bogus"`, "bogus", "foo")
}

func TestUnknownAction(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{{"budget", "create", "Home Budget"}})
	assertCommandFails(t, dbPath, `unsupported action "bogus" for resource "account"`, "account", "bogus")
}

func TestCommandArgValidation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{{"budget", "create", "Home Budget"}})

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"budget", "delete"}, "budget delete requires [name]"},
		{[]string{"budget", "show"}, "budget show requires [budget_name] or [budget_name] [month]"},
		{[]string{"budget", "show", "Home Budget", "2026-08", "extra"}, "budget show requires [budget_name] or [budget_name] [month]"},
		{[]string{"budget", "show", "Home Budget", "2026"}, "parse month: expected YYYY-MM"},
		{[]string{"account", "create"}, "account create requires [name]"},
		{[]string{"account", "show"}, "account show requires [account]"},
		{[]string{"account", "import", "Checking"}, "account import requires [account] [pdf]"},
		{[]string{"account", "categorize"}, "account categorize requires [account]"},
		{[]string{"account", "reconcile"}, "account reconcile requires [account] [date]"},
		{[]string{"account", "reconcile", "Checking"}, "account reconcile requires [account] [date]"},
		{[]string{"allocation", "create", "Groceries"}, "allocation create requires [month] [category] [amount]"},
		{[]string{"allocation", "delete"}, "allocation delete requires [month] [category]"},
		{[]string{"category", "create"}, "category create requires [name]"},
		{[]string{"category", "delete"}, "category delete requires [name]"},
		{[]string{"goal", "create", "monthly"}, "goal create requires [type] [start] [end|null] [category] [amount]"},
		{[]string{"goal", "delete"}, "goal delete requires [category]"},
		{[]string{"payee", "create"}, "payee create requires [name]"},
		{[]string{"payee", "delete"}, "payee delete requires [name]"},
		{[]string{"transaction", "create", "2026-08-28"}, "transaction create requires [date] [account] [payee] [total_out] [total_in] [note]"},
		{[]string{"transaction", "delete"}, "transaction delete requires [id]"},
		{[]string{"transaction", "category", "delete"}, "transaction category delete requires [id]"},
		{[]string{"transaction", "category"}, "transaction category requires a subcommand (create|delete)"},
		{[]string{"payee", "default-category"}, "payee default-category requires a subcommand (create|delete)"},
		{[]string{"payee", "default-category", "delete"}, "payee default-category delete requires [id]"},
	}

	for _, tt := range tests {
		assertCommandFails(t, dbPath, tt.want, tt.args...)
	}
}

func TestUnknownNameResolution(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"payee", "create", "Market"},
	})

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"account", "show", "Nope"}, `unknown account "Nope" in budget "Home Budget"`},
		{[]string{"account", "delete", "Nope"}, `unknown account "Nope" in budget "Home Budget"`},
		{[]string{"category", "delete", "Nope"}, `unknown category "Nope" in budget "Home Budget"`},
		{[]string{"payee", "delete", "Nope"}, `unknown payee "Nope" in budget "Home Budget"`},
		{[]string{"allocation", "create", "2026-08", "Nope", "100"}, `unknown category "Nope" in budget "Home Budget"`},
		{[]string{"allocation", "delete", "2026-08", "Nope"}, `unknown category "Nope" in budget "Home Budget"`},
		{[]string{"goal", "delete", "Nope"}, `unknown category "Nope" in budget "Home Budget"`},
		{[]string{"budget", "delete", "Nope"}, `unknown budget "Nope"`},
		{[]string{"budget", "show", "Nope"}, `unknown budget "Nope"`},
	}

	for _, tt := range tests {
		assertCommandFails(t, dbPath, tt.want, tt.args...)
	}
}

func TestTransactionCategoryValidation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "25.00", "0.00", ""},
	})

	assertCommandFails(t, dbPath, "set exactly one of --category, --other-account, or --income", "transaction", "category", "create", "1", "100", "0")
	assertCommandFails(t, dbPath, "set exactly one of --category, --other-account, or --income", "transaction", "category", "create", "1", "100", "0", "--category", "Groceries", "--other-account", "Savings")
	assertCommandFails(t, dbPath, "--category may only be set once", "transaction", "category", "create", "1", "100", "0", "--category", "Groceries", "--category", "Groceries")
	assertCommandFails(t, dbPath, `unknown category "Nope" in budget "Home Budget"`, "transaction", "category", "create", "1", "100", "0", "--category", "Nope")
	assertCommandFails(t, dbPath, `unknown account "Nope" in budget "Home Budget"`, "transaction", "category", "create", "1", "100", "0", "--other-account", "Nope")
	assertCommandFails(t, dbPath, `unsupported flag "--bogus"`, "transaction", "category", "create", "1", "100", "0", "--bogus", "x")
}

func TestPayeeDefaultCategoryValidation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"payee", "create", "Market"},
	})

	assertCommandFails(t, dbPath, "set exactly one of --category, --other-account, or --income", "payee", "default-category", "create", "Market", "100")
	assertCommandFails(t, dbPath, "set exactly one of --category, --other-account, or --income", "payee", "default-category", "create", "Market", "100", "--category", "Groceries", "--other-account", "Savings")
	assertCommandFails(t, dbPath, `unknown payee "Nope" in budget "Home Budget"`, "payee", "default-category", "create", "Nope", "100", "--category", "Groceries")
	assertCommandFails(t, dbPath, "parse percent", "payee", "default-category", "create", "Market", "abc", "--category", "Groceries")
}

func TestParseErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
	})

	assertCommandFails(t, dbPath, "invalid amount", "allocation", "create", "2026-08", "Groceries", "abc")
	assertCommandFails(t, dbPath, "parse total_out", "transaction", "create", "2026-08-28", "Checking", "Market", "abc", "0", "")
	assertCommandFails(t, dbPath, "expected RFC3339 or YYYY-MM-DD", "transaction", "create", "not-a-date", "Checking", "Market", "100", "0", "")
}

func TestDuplicateNameConstraints(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
	})
	assertCommandFails(t, dbPath, "UNIQUE constraint failed", "budget", "create", "Home Budget")

	runCommands(t, dbPath, [][]string{
		{"account", "create", "Checking"},
	})
	assertCommandFails(t, dbPath, "UNIQUE constraint failed", "account", "create", "Checking")

	runCommands(t, dbPath, [][]string{
		{"category", "create", "Groceries"},
	})
	assertCommandFails(t, dbPath, "UNIQUE constraint failed", "category", "create", "Groceries")

	runCommands(t, dbPath, [][]string{
		{"payee", "create", "Market"},
	})
	assertCommandFails(t, dbPath, "UNIQUE constraint failed", "payee", "create", "Market")

	runCommands(t, dbPath, [][]string{
		{"goal", "create", "monthly", "2026-08", "null", "Groceries", "50.00"},
	})
	assertCommandFails(t, dbPath, "UNIQUE constraint failed", "goal", "create", "save", "2026-08", "2026-12", "Groceries", "100.00")
}

func TestGoalWithEndDate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"category", "create", "Groceries"},
		{"goal", "create", "save", "2026-09", "2026-12", "Groceries", "1000.00"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "goal", "list")
	if exitCode != 0 {
		t.Fatalf("goal list failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "2026-12") {
		t.Fatalf("expected goal end date in list, got %q", stdout)
	}
}

func TestTransactionCreateDollarAmounts(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "12.50", "3.25", "both"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "transaction", "list")
	if exitCode != 0 {
		t.Fatalf("transaction list failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "12.50") || !strings.Contains(stdout, "3.25") {
		t.Fatalf("expected dollar amounts in listing, got %q", stdout)
	}
}

func TestTransactionCreateRFC3339Date(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28T10:00:00Z", "Checking", "Market", "1.00", "0.00", ""},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "transaction", "list")
	if exitCode != 0 {
		t.Fatalf("transaction list failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "2026-08-28") {
		t.Fatalf("expected RFC3339 date rendered as YYYY-MM-DD, got %q", stdout)
	}
}

func TestAccountShowInflowTransaction(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Employer", "0.00", "42.00", "paycheck"},
		{"transaction", "category", "create", "1", "0.00", "42.00", "--income"},
	})

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Income") {
		t.Fatalf("expected inflow target Income, got %q", show)
	}
	if !strings.Contains(show, "42.00") {
		t.Fatalf("expected inflow amount 42.00, got %q", show)
	}
}

func TestEmptyLists(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
	})

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"account", "list"}, "NAME  BALANCE  RECONCILED BALANCE\n"},
		{[]string{"category", "list"}, "NAME\n"},
		{[]string{"payee", "list"}, "NAME\n"},
		{[]string{"allocation", "list"}, "MONTH  CATEGORY  AMOUNT\n"},
		{[]string{"goal", "list"}, "TYPE  CATEGORY  START  END  AMOUNT  MONTHLY\n"},
		{[]string{"transaction", "list"}, "ID  DATE  ACCOUNT  PAYEE  OUTFLOW  INFLOW  NOTE\n"},
	}

	for _, tt := range tests {
		stdout, stderr, exitCode := invoke(t, dbPath, tt.args...)
		if exitCode != 0 {
			t.Fatalf("command %q failed: stdout=%q stderr=%q", strings.Join(tt.args, " "), stdout, stderr)
		}
		if stdout != tt.want {
			t.Fatalf("command %q: got %q want %q", strings.Join(tt.args, " "), stdout, tt.want)
		}
		if stderr != "" {
			t.Fatalf("command %q: unexpected stderr %q", strings.Join(tt.args, " "), stderr)
		}
	}
}

func TestAccountImportUnknownAccount(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "ws-visa"},
	})

	statementFile := filepath.Join(t.TempDir(), "statement.txt")
	if err := os.WriteFile(statementFile, []byte("ynafb-test-statement"), 0o644); err != nil {
		t.Fatalf("write statement file: %v", err)
	}

	assertCommandFails(t, dbPath, `unknown account "Nope" in budget "Home Budget"`, "account", "import", "Nope", statementFile)
}

func TestAccountImportUnsupportedFormat(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "ws-visa"},
	})

	garbageFile := filepath.Join(t.TempDir(), "garbage.txt")
	if err := os.WriteFile(garbageFile, []byte("this is not a statement"), 0o644); err != nil {
		t.Fatalf("write garbage file: %v", err)
	}

	assertCommandFails(t, dbPath, "unsupported statement format", "account", "import", "ws-visa", garbageFile)
}

func TestPayeeDefaultCategoryCreateOtherAccount(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"payee", "create", "Market"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "payee", "default-category", "create", "Market", "100", "--other-account", "Savings")
	if exitCode != 0 {
		t.Fatalf("create failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "created payee default-category 1\n" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
}

func TestPayeeDefaultCategoryCreateIncome(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"payee", "create", "Employer"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "payee", "default-category", "create", "Employer", "100", "--income")
	if exitCode != 0 {
		t.Fatalf("create failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "created payee default-category 1\n" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
}

func TestCategoryCreateIncomeReserved(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
	})

	assertCommandFails(t, dbPath, `category name "income" is reserved`, "category", "create", "income")
	assertCommandFails(t, dbPath, `category name "Income" is reserved`, "category", "create", "Income")
}

func TestCategoryCreateWithGroup(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"category", "create", "Groceries", "--group", "Fixed"},
		{"category", "create", "Utilities", "--group", "Fixed"},
		{"category", "create", "Fun"},
	})

	assertRowCount(t, dbPath, "category_group", 1)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var groupID int64
	if err := db.QueryRow("SELECT id FROM category_group WHERE name = 'Fixed'").Scan(&groupID); err != nil {
		t.Fatalf("lookup group: %v", err)
	}

	var grouped, ungrouped int64
	if err := db.QueryRow("SELECT COUNT(*) FROM category WHERE category_group = ?", groupID).Scan(&grouped); err != nil {
		t.Fatalf("count grouped: %v", err)
	}
	if grouped != 2 {
		t.Fatalf("grouped categories = %d, want 2", grouped)
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM category WHERE category_group IS NULL").Scan(&ungrouped); err != nil {
		t.Fatalf("count ungrouped: %v", err)
	}
	if ungrouped != 1 {
		t.Fatalf("ungrouped categories = %d, want 1", ungrouped)
	}
}

func TestCategoryCreateGroupValidation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
	})

	assertCommandFails(t, dbPath, "category create requires [name]", "category", "create")
	assertCommandFails(t, dbPath, "unsupported flag \"--bogus\"", "category", "create", "Groceries", "--bogus", "x")
	assertCommandFails(t, dbPath, "missing value for --group", "category", "create", "Groceries", "--group")
}

func TestGroupCommands(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"group", "create", "Fixed"},
		{"group", "create", "Variable"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "group", "list")
	if exitCode != 0 {
		t.Fatalf("group list failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stdout != "NAME\nFixed\nVariable\n" {
		t.Fatalf("group list: got %q want %q", stdout, "NAME\nFixed\nVariable\n")
	}
	if stderr != "" {
		t.Fatalf("group list: unexpected stderr %q", stderr)
	}

	runCommands(t, dbPath, [][]string{
		{"group", "update", "Fixed", "Essentials"},
	})

	assertRowCount(t, dbPath, "category_group", 2)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var name string
	if err := db.QueryRow("SELECT name FROM category_group WHERE name = 'Essentials'").Scan(&name); err != nil {
		t.Fatalf("lookup renamed group: %v", err)
	}
	if err := db.QueryRow("SELECT name FROM category_group WHERE name = 'Fixed'").Scan(&name); err != sql.ErrNoRows {
		t.Fatalf("old group name still present: %v", err)
	}

	runCommands(t, dbPath, [][]string{
		{"group", "delete", "Variable"},
	})

	assertRowCount(t, dbPath, "category_group", 1)
}

func TestGroupDeleteUngroupsCategories(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"group", "create", "Fixed"},
		{"category", "create", "Groceries", "--group", "Fixed"},
	})

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var groupID int64
	if err := db.QueryRow("SELECT id FROM category_group WHERE name = 'Fixed'").Scan(&groupID); err != nil {
		t.Fatalf("lookup group: %v", err)
	}

	runCommands(t, dbPath, [][]string{
		{"group", "delete", "Fixed"},
	})

	var count int64
	if err := db.QueryRow("SELECT COUNT(*) FROM category_group").Scan(&count); err != nil {
		t.Fatalf("count groups: %v", err)
	}
	if count != 0 {
		t.Fatalf("groups after delete = %d, want 0", count)
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM category WHERE category_group = ?", groupID).Scan(&count); err != nil {
		t.Fatalf("count grouped: %v", err)
	}
	if count != 0 {
		t.Fatalf("grouped categories after delete = %d, want 0", count)
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM category WHERE category_group IS NULL").Scan(&count); err != nil {
		t.Fatalf("count ungrouped: %v", err)
	}
	if count != 1 {
		t.Fatalf("ungrouped categories = %d, want 1", count)
	}
}

func TestGroupCommandValidation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
	})

	assertCommandFails(t, dbPath, "group create requires [name]", "group", "create")
	assertCommandFails(t, dbPath, "group update requires [name] [new_name]", "group", "update", "Fixed")
	assertCommandFails(t, dbPath, "group delete requires [name]", "group", "delete")
	assertCommandFails(t, dbPath, `unknown group "Nope" in budget "Home Budget"`, "group", "update", "Nope", "New")
	assertCommandFails(t, dbPath, `unknown group "Nope" in budget "Home Budget"`, "group", "delete", "Nope")
}

func TestAccountCategorizeIncomeSavesDefaultAndPrefills(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Employer", "0.00", "42.00", "paycheck"},
	})

	input := strings.Join([]string{
		"a",
		"income",
		"",
		"",
		"o",
		"y",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "→ added: Income") {
		t.Fatalf("expected income added, got %q", stdout)
	}
	if !strings.Contains(stdout, `saved default for "Employer"`) {
		t.Fatalf("expected default saved, got %q", stdout)
	}

	runCommands(t, dbPath, [][]string{
		{"transaction", "create", "2026-08-29", "Checking", "Employer", "0.00", "42.00", "second paycheck"},
	})

	second := runCategorize(t, dbPath, "o\n", "Checking")
	if !strings.Contains(second, "(pre-filled from payee default)") {
		t.Fatalf("expected pre-fill from income default, got %q", second)
	}
	if !strings.Contains(second, "Income") {
		t.Fatalf("expected income pre-fill, got %q", second)
	}
	if !strings.Contains(second, "matches existing default") {
		t.Fatalf("expected matches existing default, got %q", second)
	}
}

func TestInflowCategorizedToCategory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Reimbursements"},
		{"transaction", "create", "2026-08-28", "Checking", "Friend", "0.00", "25.00", "reimbursement"},
		{"transaction", "category", "create", "1", "0.00", "25.00", "--category", "Reimbursements"},
	})

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Reimbursements") {
		t.Fatalf("expected reimbursement categorized to a category, got %q", show)
	}
	if strings.Contains(show, "Income") {
		t.Fatalf("expected no income target for a category inflow, got %q", show)
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08")
	if exitCode != 0 {
		t.Fatalf("budget show failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "Income: 0.00") {
		t.Fatalf("expected income line 0.00, got %q", stdout)
	}
	if !strings.Contains(stdout, "Reimbursements") {
		t.Fatalf("expected reimbursement category in budget show, got %q", stdout)
	}
}

func TestAccountCategorize(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"transaction", "create", "2026-08-29", "Checking", "Grocery Mart", "42.00", "0.00", ""},
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "42.00", "0.00", "import mismatch"},
		{"payee", "default-category", "create", "Grocery Mart", "80", "--category", "Groceries"},
		{"payee", "default-category", "create", "Grocery Mart", "20", "--category", "Household"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	input := strings.Join([]string{
		"o",
		"a",
		"Groceries",
		"42.00",
		"",
		"o",
		"y",
	}, "\n")

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, input, "account", "categorize", "Checking")
	if exitCode != 0 {
		t.Fatalf("account categorize failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	if !strings.Contains(stdout, "(pre-filled from payee default)") {
		t.Fatalf("expected pre-fill note, got %q", stdout)
	}
	if !strings.Contains(stdout, "matches existing default") {
		t.Fatalf("expected matches existing default message, got %q", stdout)
	}
	if !strings.Contains(stdout, `saved default for "Online Shop"`) {
		t.Fatalf("expected saved default message, got %q", stdout)
	}
	if !strings.Contains(stdout, "Done — 2 categorized, 0 skipped") {
		t.Fatalf("expected Done summary, got %q", stdout)
	}

	show, stderr, exitCode := invoke(t, dbPath, "account", "show", "Checking")
	if exitCode != 0 {
		t.Fatalf("account show failed with exit code %d, stdout=%q stderr=%q", exitCode, show, stderr)
	}
	if !strings.Contains(show, "Groceries") {
		t.Fatalf("expected transactions categorized as Groceries, got %q", show)
	}
}

func TestAccountCategorizeCreatesCategory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "42.00", "0.00", ""},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	input := strings.Join([]string{
		"a",
		"Brand New",
		"y",
		"42.00",
		"",
		"o",
		"n",
	}, "\n")

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, input, "account", "categorize", "Checking")
	if exitCode != 0 {
		t.Fatalf("account categorize failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	if !strings.Contains(stdout, `created category "Brand New"`) {
		t.Fatalf("expected category created message, got %q", stdout)
	}

	show, stderr, exitCode := invoke(t, dbPath, "account", "show", "Checking")
	if exitCode != 0 {
		t.Fatalf("account show failed with exit code %d, stdout=%q stderr=%q", exitCode, show, stderr)
	}
	if !strings.Contains(show, "Brand New") {
		t.Fatalf("expected transaction categorized as Brand New, got %q", show)
	}
}

func TestAccountCategorizePrefillMatchesDefaultDoesNotPrompt(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "merry"},
		{"category", "create", "groceries"},
		{"transaction", "create", "2026-07-26", "Checking", "IGA #8644", "29.57", "0.00", ""},
		{"payee", "default-category", "create", "IGA #8644", "50", "--category", "groceries"},
		{"payee", "default-category", "create", "IGA #8644", "50", "--other-account", "merry"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, "o\n", "account", "categorize", "Checking")
	if exitCode != 0 {
		t.Fatalf("account categorize failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	if !strings.Contains(stdout, "matches existing default") {
		t.Fatalf("expected matches existing default message, got %q", stdout)
	}
	if strings.Contains(stdout, "Save as default") {
		t.Fatalf("did not expect save-as-default prompt for unchanged pre-fill, got %q", stdout)
	}
}

func TestAccountCategorizeUnknownAccountErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "42.00", "0.00", ""},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	input := strings.Join([]string{
		"a",
		"@Nope",
		"o",
		"q",
	}, "\n")

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, input, "account", "categorize", "Checking")
	if exitCode != 0 {
		t.Fatalf("account categorize failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if !strings.Contains(stdout, `unknown account "Nope"`) {
		t.Fatalf("expected unknown account error, got %q", stdout)
	}
}

func runCommands(t *testing.T, dbPath string, commands [][]string) {
	t.Helper()

	for _, args := range commands {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}
}

func runCategorize(t *testing.T, dbPath, input, account string) string {
	t.Helper()

	stdout, stderr, exitCode := invokeWithInput(t, dbPath, input, "account", "categorize", account)
	if exitCode != 0 {
		t.Fatalf("account categorize failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	return stdout
}

func assertCommandFails(t *testing.T, dbPath string, wantErr string, args ...string) {
	t.Helper()

	stdout, stderr, exitCode := invoke(t, dbPath, args...)
	if exitCode == 0 {
		t.Fatalf("command %q expected to fail, got stdout=%q", strings.Join(args, " "), stdout)
	}
	if !strings.Contains(stderr, wantErr) {
		t.Fatalf("command %q: expected stderr to contain %q, got %q", strings.Join(args, " "), wantErr, stderr)
	}
}

func showAccount(t *testing.T, dbPath, account string) string {
	t.Helper()

	stdout, stderr, exitCode := invoke(t, dbPath, "account", "show", account)
	if exitCode != 0 {
		t.Fatalf("account show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	return stdout
}

func TestAccountCategorizeSkipsFullyCategorized(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
		{"transaction", "category", "create", "1", "42.00", "0.00", "--category", "Groceries"},
	})

	stdout := runCategorize(t, dbPath, "", "Checking")
	if !strings.Contains(stdout, "No transactions need categorization") {
		t.Fatalf("expected no-transactions message, got %q", stdout)
	}
}

func TestAccountCategorizeInflowTransaction(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Employer", "0.00", "42.00", "paycheck"},
	})

	input := strings.Join([]string{
		"a",
		"Income",
		"",
		"",
		"o",
		"n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "Done — 1 categorized, 0 skipped") {
		t.Fatalf("expected Done summary, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Income") {
		t.Fatalf("expected inflow transaction categorized as Income, got %q", show)
	}
}

func TestAccountCategorizeExcludesMirroredTransfers(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"transaction", "create", "2026-08-28", "Savings", "Me", "0.00", "5.00", ""},
		{"transaction", "category", "create", "1", "0.00", "5.00", "--other-account", "Checking"},
	})

	stdout := runCategorize(t, dbPath, "", "Checking")
	if !strings.Contains(stdout, "No transactions need categorization") {
		t.Fatalf("expected mirrored transfer to be excluded, got %q", stdout)
	}
}

func TestAccountCategorizeAddTransfer(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"transaction", "create", "2026-08-28", "Checking", "Bank", "15.00", "0.00", "move"},
	})

	input := strings.Join([]string{
		"a",
		"@Savings",
		"",
		"",
		"o",
		"n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "→ added: @Savings") {
		t.Fatalf("expected transfer added message, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "@Savings") {
		t.Fatalf("expected transfer target @Savings, got %q", show)
	}
}

func TestAccountCategorizeEmptyAmountDefaultsToRemaining(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
	})

	input := strings.Join([]string{
		"a",
		"Groceries",
		"",
		"",
		"o",
		"n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "→ added: Groceries  42.00 out / 0.00 in") {
		t.Fatalf("expected defaulted remaining amount in add message, got %q", stdout)
	}
}

func TestAccountCategorizeCancelCategoryCreation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
	})

	input := strings.Join([]string{
		"a",
		"New Cat",
		"n",
		"s",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if strings.Contains(stdout, `created category "New Cat"`) {
		t.Fatalf("did not expect category to be created, got %q", stdout)
	}
	if !strings.Contains(stdout, `category "New Cat" not created`) {
		t.Fatalf("expected category not-created message, got %q", stdout)
	}
	if !strings.Contains(stdout, "→ skipped") {
		t.Fatalf("expected skip after cancelled create, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if strings.Contains(show, "New Cat") {
		t.Fatalf("did not expect New Cat to persist, got %q", show)
	}
}

func TestAllocationMonthBehavior(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"category", "create", "Groceries"},
		{"allocation", "create", "2026-08", "Groceries", "50.00"},
		{"allocation", "create", "2026-09", "Groceries", "75.00"},
	})

	// Same category + month must be rejected (UNIQUE budget+category+month).
	assertCommandFails(t, dbPath, "UNIQUE constraint failed", "allocation", "create", "2026-08", "Groceries", "60.00")

	stdout, stderr, exitCode := invoke(t, dbPath, "allocation", "list")
	if exitCode != 0 {
		t.Fatalf("allocation list failed: stdout=%q stderr=%q", stdout, stderr)
	}
	want := "MONTH    CATEGORY   AMOUNT\n2026-08  Groceries  50.00\n2026-09  Groceries  75.00\n"
	if stdout != want {
		t.Fatalf("allocation list: got %q want %q", stdout, want)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "allocation", "delete", "2026-08", "Groceries")
	if exitCode != 0 {
		t.Fatalf("allocation delete failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "deleted allocation 2026-08 \"Groceries\"\n" {
		t.Fatalf("unexpected delete stdout: %q", stdout)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "allocation", "list")
	if exitCode != 0 {
		t.Fatalf("allocation list failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "MONTH    CATEGORY   AMOUNT\n2026-09  Groceries  75.00\n" {
		t.Fatalf("expected only September allocation after delete, got %q", stdout)
	}
}

func TestBudgetShowMonths(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"allocation", "create", "2026-06", "Groceries", "50.00"},
		{"allocation", "create", "2026-07", "Groceries", "50.00"},
		{"transaction", "create", "2026-07-15", "Checking", "Market", "10.00", "0.00", ""},
		{"transaction", "create", "2026-08-15", "Checking", "Cafe", "5.00", "0.00", ""},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := "MONTH\n2026-06\n2026-07\n2026-08\n"
	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowMonth(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Emergency"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"allocation", "create", "2026-08", "Groceries", "50.00"},
		{"allocation", "create", "2026-08", "Household", "25.00"},
		{"transaction", "create", "2026-08-05", "Checking", "Market", "20.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "20.00", "0.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-10", "Checking", "Cafe", "5.00", "0.00", "coffee"},
		{"transaction", "category", "create", "2", "5.00", "0.00", "--category", "Household"},
		{"transaction", "create", "2026-08-15", "Checking", "Mart", "30.00", "0.00", ""},
		{"transaction", "category", "create", "3", "30.00", "0.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-20", "Checking", "Refund", "0.00", "10.00", ""},
		{"transaction", "category", "create", "4", "0.00", "10.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-25", "Checking", "Bank", "15.00", "0.00", "transfer"},
		{"transaction", "category", "create", "5", "15.00", "0.00", "--other-account", "Savings"},
		{"transaction", "create", "2026-08-28", "Checking", "Unknown", "7.00", "0.00", ""},
		{"transaction", "create", "2026-09-02", "Checking", "Market", "9.99", "0.00", ""},
		{"transaction", "category", "create", "7", "9.99", "0.00", "--category", "Groceries"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -82.00",
		"Income: 0.00",
		"Goals: 0.00",
		"Allocated: 75.00",
		"Spent: 45.00",
		"Remaining: 30.00",
		"Uncategorized: 7.00",
		"",
		"No group:",
		"CATEGORY   GOAL  ALLOCATED  SPENT  REMAINING  WARNING",
		"Emergency        0.00       0.00   0.00",
		"Groceries        50.00      40.00  10.00",
		"Household        25.00      5.00   20.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("budget show failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "MONTH\n" {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, "MONTH\n")
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08")
	if exitCode != 0 {
		t.Fatalf("budget show month failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "Available: 0.00\nIncome: 0.00\nGoals: 0.00\nAllocated: 0.00\nSpent: 0.00\nRemaining: 0.00\nUncategorized: 0.00\n\n" {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, "Available: 0.00\nIncome: 0.00\nGoals: 0.00\nAllocated: 0.00\nSpent: 0.00\nRemaining: 0.00\nUncategorized: 0.00\n\n")
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestGoalCreateValidation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"category", "create", "Groceries"},
	})

	assertCommandFails(t, dbPath, `goal create: unknown goal type "bogus" (expected monthly, save, or refill)`, "goal", "create", "bogus", "2026-08", "null", "Groceries", "5000")
	assertCommandFails(t, dbPath, "goal create: save goals require an end month", "goal", "create", "save", "2026-08", "null", "Groceries", "5000")
	assertCommandFails(t, dbPath, "goal create: end month must not be before start month", "goal", "create", "monthly", "2026-09", "2026-08", "Groceries", "5000")
	assertCommandFails(t, dbPath, "parse start: expected YYYY-MM", "goal", "create", "monthly", "not-a-month", "null", "Groceries", "5000")
	assertCommandFails(t, dbPath, "parse end: expected YYYY-MM", "goal", "create", "monthly", "2026-08", "2026-13", "Groceries", "5000")
	assertCommandFails(t, dbPath, `no goal for category "Groceries" in budget "Home Budget"`, "goal", "delete", "Groceries")
}

func TestGoalMonthlyValue(t *testing.T) {
	mustMonth := func(s string) time.Time {
		t.Helper()
		parsed, err := time.Parse("2006-01", s)
		if err != nil {
			t.Fatalf("parse month %q: %v", s, err)
		}
		return parsed
	}

	newGoal := func(type_, start, end string, amount int64) data.ListGoalsByBudgetRow {
		g := data.ListGoalsByBudgetRow{
			ID:           1,
			Type:         type_,
			CategoryID:   1,
			CategoryName: "Cat",
			Amount:       amount,
			Start:        mustMonth(start),
		}
		if end != "" {
			g.End = sql.NullTime{Time: mustMonth(end), Valid: true}
		}
		return g
	}

	alloc := map[int64]map[time.Time]int64{
		1: {
			mustMonth("2026-07"): 10000,
			mustMonth("2026-08"): 10000,
		},
	}

	tests := []struct {
		name  string
		goal  data.ListGoalsByBudgetRow
		now   time.Time
		spend map[int64]map[int64]int64
		plan  bool
		want  int64
	}{
		{"save no allocation yet", newGoal("save", "2026-08", "2026-12", 100000), mustMonth("2026-08"), nil, false, 20000},
		{"save counts only prior months", newGoal("save", "2026-07", "2026-12", 100000), mustMonth("2026-08"), nil, false, 18000},
		{"save fully funded", newGoal("save", "2026-07", "2026-12", 10000), mustMonth("2026-08"), nil, false, 0},
		{"save not started", newGoal("save", "2026-09", "2026-12", 100000), mustMonth("2026-08"), nil, false, 0},
		{"save ended", newGoal("save", "2026-07", "2026-08", 100000), mustMonth("2026-09"), nil, false, 0},
		{"save last month", newGoal("save", "2026-07", "2026-08", 100000), mustMonth("2026-08"), nil, false, 90000},
		{"monthly active", newGoal("monthly", "2026-07", "", 5000), mustMonth("2026-08"), nil, false, 5000},
		{"monthly not started", newGoal("monthly", "2026-09", "", 5000), mustMonth("2026-08"), nil, false, 0},
		{"monthly ended", newGoal("monthly", "2026-07", "2026-08", 5000), mustMonth("2026-09"), nil, false, 0},
		{"refill no history", func() data.ListGoalsByBudgetRow { g := newGoal("refill", "2026-08", "", 100000); g.CategoryID = 2; return g }(), mustMonth("2026-08"), nil, false, 100000},
		{"refill counts prior balance", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), nil, false, 90000},
		{"refill fully funded", newGoal("refill", "2026-07", "", 10000), mustMonth("2026-08"), nil, false, 0},
		{"refill subtracts spending", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), map[int64]map[int64]int64{1: {202607: 4000}}, false, 94000},
		{"refill overspent clamps at zero", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), map[int64]map[int64]int64{1: {202607: 20000}}, false, 100000},
		{"refill excludes current month", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), nil, false, 90000},
		{"refill not started", newGoal("refill", "2026-09", "", 100000), mustMonth("2026-08"), nil, false, 0},
		{"refill ended", newGoal("refill", "2026-07", "2026-08", 100000), mustMonth("2026-09"), nil, false, 0},
		{"refill plan ignores carryover", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), nil, true, 100000},
		{"refill plan ignores spending", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), map[int64]map[int64]int64{1: {202607: 4000}}, true, 100000},
		{"monthly plan unaffected", newGoal("monthly", "2026-07", "", 5000), mustMonth("2026-08"), nil, true, 5000},
		{"save plan unaffected", newGoal("save", "2026-07", "2026-12", 100000), mustMonth("2026-08"), nil, true, 18000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goalMonthlyValue(tt.goal, alloc, tt.spend, tt.now, tt.plan)
			if got != tt.want {
				t.Fatalf("goalMonthlyValue() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGoalWarning(t *testing.T) {
	mustMonth := func(s string) time.Time {
		t.Helper()
		parsed, err := time.Parse("2006-01", s)
		if err != nil {
			t.Fatalf("parse month %q: %v", s, err)
		}
		return parsed
	}

	newGoal := func(type_, start, end string, amount int64) data.ListGoalsByBudgetRow {
		g := data.ListGoalsByBudgetRow{
			ID:           1,
			Type:         type_,
			CategoryID:   1,
			CategoryName: "Cat",
			Amount:       amount,
			Start:        mustMonth(start),
		}
		if end != "" {
			g.End = sql.NullTime{Time: mustMonth(end), Valid: true}
		}
		return g
	}

	alloc := map[int64]map[time.Time]int64{
		1: {
			mustMonth("2026-07"): 10000,
			mustMonth("2026-08"): 10000,
		},
		2: {
			mustMonth("2026-07"): 100000,
		},
	}

	tests := []struct {
		name  string
		goal  data.ListGoalsByBudgetRow
		now   time.Time
		spend map[int64]map[int64]int64
		plan  bool
		want  string
	}{
		{"monthly fully funded", newGoal("monthly", "2026-07", "", 5000), mustMonth("2026-08"), nil, false, ""},
		{"monthly exactly funded", newGoal("monthly", "2026-07", "", 10000), mustMonth("2026-08"), nil, false, ""},
		{"monthly underfunded", newGoal("monthly", "2026-07", "", 15000), mustMonth("2026-08"), nil, false, "underfunded 50.00"},
		{"monthly not started", newGoal("monthly", "2026-09", "", 5000), mustMonth("2026-08"), nil, false, ""},
		{"monthly ended", newGoal("monthly", "2026-07", "2026-07", 15000), mustMonth("2026-08"), nil, false, ""},
		{"save on track", newGoal("save", "2026-07", "2026-12", 20000), mustMonth("2026-08"), nil, false, ""},
		{"save exactly on pace", newGoal("save", "2026-07", "2026-12", 60000), mustMonth("2026-08"), nil, false, ""},
		{"save behind", newGoal("save", "2026-07", "2026-12", 100000), mustMonth("2026-08"), nil, false, "behind 80.00"},
		{"save fully funded", newGoal("save", "2026-07", "2026-12", 10000), mustMonth("2026-08"), nil, false, ""},
		{"save not started", newGoal("save", "2026-09", "2026-12", 100000), mustMonth("2026-08"), nil, false, ""},
		{"save ended", newGoal("save", "2026-07", "2026-08", 100000), mustMonth("2026-09"), nil, false, ""},
		{"refill at target", newGoal("refill", "2026-07", "", 10000), mustMonth("2026-08"), nil, false, ""},
		{"refill below target", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), nil, false, "needs refill 900.00"},
		{"refill with spending", func() data.ListGoalsByBudgetRow { g := newGoal("refill", "2026-07", "", 100000); g.CategoryID = 2; return g }(), mustMonth("2026-08"), map[int64]map[int64]int64{2: {202607: 4000}}, false, "needs refill 40.00"},
		{"refill overspent clamps to full", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), map[int64]map[int64]int64{1: {202607: 20000}}, false, "needs refill 1000.00"},
		{"refill not started", newGoal("refill", "2026-09", "", 100000), mustMonth("2026-08"), nil, false, ""},
		{"refill ended", newGoal("refill", "2026-07", "2026-08", 100000), mustMonth("2026-09"), nil, false, ""},
		{"refill plan ignores carryover", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), nil, true, "needs refill 1000.00"},
		{"refill plan ignores spending", newGoal("refill", "2026-07", "", 100000), mustMonth("2026-08"), map[int64]map[int64]int64{1: {202607: 4000}}, true, "needs refill 1000.00"},
		{"monthly plan unaffected", newGoal("monthly", "2026-07", "", 15000), mustMonth("2026-08"), nil, true, "underfunded 50.00"},
		{"save plan unaffected", newGoal("save", "2026-07", "2026-12", 100000), mustMonth("2026-08"), nil, true, "behind 80.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := goalWarning(tt.goal, alloc, tt.spend, tt.now, tt.plan)
			if got != tt.want {
				t.Fatalf("goalWarning() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCategoryRemaining(t *testing.T) {
	mustMonth := func(s string) time.Time {
		t.Helper()
		parsed, err := time.Parse("2006-01", s)
		if err != nil {
			t.Fatalf("parse month %q: %v", s, err)
		}
		return parsed
	}

	rows := []data.ListBudgetMonthCategoriesRow{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	allocations := []data.ListAllocationsByBudgetRow{
		{Month: mustMonth("2026-07"), CategoryID: 1, Amount: 40000},
		{Month: mustMonth("2026-08"), CategoryID: 1, Amount: 10000},
		{Month: mustMonth("2026-07"), CategoryID: 2, Amount: 10000},
		{Month: mustMonth("2026-08"), CategoryID: 2, Amount: 5000},
		{Month: mustMonth("2026-07"), CategoryID: 3, Amount: 10000},
		{Month: mustMonth("2026-08"), CategoryID: 3, Amount: 10000},
	}

	spending := []data.ListCategoryMonthlySpendingByBudgetRow{
		{Category: sql.NullInt64{Int64: 1, Valid: true}, Month: 202607, Net: 5000},
		{Category: sql.NullInt64{Int64: 1, Valid: true}, Month: 202608, Net: 3000},
		{Category: sql.NullInt64{Int64: 2, Valid: true}, Month: 202608, Net: 20000},
		{Category: sql.NullInt64{Int64: 3, Valid: true}, Month: 202607, Net: 20000},
	}

	month := mustMonth("2026-08")
	end := month.AddDate(0, 1, 0)

	t.Run("normal rolls prior months forward", func(t *testing.T) {
		got := categoryRemaining(rows, allocations, spending, month, end, false)
		// cat 1: July 40000-5000=35000, Aug 10000-3000=7000; total 42000
		if got[1] != 42000 {
			t.Fatalf("cat 1 remaining = %d, want 42000", got[1])
		}
		// cat 2: July +10000, Aug 5000-20000=-15000 → 10000-15000 = -5000, target is last month so unclamped
		if got[2] != -5000 {
			t.Fatalf("cat 2 remaining = %d, want -5000", got[2])
		}
		// cat 3: July 10000-20000 = -10000, clamped to 0 before Aug, then +10000 = 10000
		if got[3] != 10000 {
			t.Fatalf("cat 3 remaining = %d, want 10000", got[3])
		}
	})

	t.Run("plan ignores carryover and spending, counts only target month allocation", func(t *testing.T) {
		got := categoryRemaining(rows, allocations, spending, month, end, true)
		// cat 1: Aug 10000 only (spending ignored)
		if got[1] != 10000 {
			t.Fatalf("plan cat 1 remaining = %d, want 10000", got[1])
		}
		// cat 2: Aug 5000 only
		if got[2] != 5000 {
			t.Fatalf("plan cat 2 remaining = %d, want 5000", got[2])
		}
		// cat 3: Aug 10000 only
		if got[3] != 10000 {
			t.Fatalf("plan cat 3 remaining = %d, want 10000", got[3])
		}
	})
}

func TestBudgetShowGoal(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Emergency"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"allocation", "create", "2026-08", "Groceries", "50.00"},
		{"allocation", "create", "2026-08", "Household", "25.00"},
		{"goal", "create", "monthly", "2026-07", "null", "Groceries", "50.00"},
		{"goal", "create", "save", "2026-08", "2026-12", "Household", "1000.00"},
		{"transaction", "create", "2026-08-05", "Checking", "Market", "20.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "20.00", "0.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-10", "Checking", "Cafe", "5.00", "0.00", "coffee"},
		{"transaction", "category", "create", "2", "5.00", "0.00", "--category", "Household"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -75.00",
		"Income: 0.00",
		"Goals: 250.00",
		"Allocated: 75.00",
		"Spent: 25.00",
		"Remaining: 50.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY   GOAL    ALLOCATED  SPENT  REMAINING  WARNING",
		"Emergency          0.00       0.00   0.00",
		"Groceries  50.00   50.00      20.00  30.00",
		"Household  200.00  25.00      5.00   20.00      behind 175.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowRefillGoal(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Repair"},
		{"goal", "create", "refill", "2026-07", "null", "Repair", "1000.00"},
		{"allocation", "create", "2026-07", "Repair", "400.00"},
		{"transaction", "create", "2026-07-20", "Checking", "Hardware", "50.00", "0.00", "fix"},
		{"transaction", "category", "create", "1", "50.00", "0.00", "--category", "Repair"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -400.00",
		"Income: 0.00",
		"Goals: 650.00",
		"Allocated: 0.00",
		"Spent: 0.00",
		"Remaining: 350.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY  GOAL    ALLOCATED  SPENT  REMAINING  WARNING",
		"Repair    650.00  0.00       0.00   350.00     needs refill 650.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowRefillGoalPlan(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Repair"},
		{"goal", "create", "refill", "2026-07", "null", "Repair", "1000.00"},
		{"allocation", "create", "2026-07", "Repair", "400.00"},
		{"transaction", "create", "2026-07-20", "Checking", "Hardware", "50.00", "0.00", "fix"},
		{"transaction", "category", "create", "1", "50.00", "0.00", "--category", "Repair"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08", "--plan")
	if exitCode != 0 {
		t.Fatalf("budget show --plan failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -50.00",
		"Income: 0.00",
		"Goals: 1000.00",
		"Allocated: 0.00",
		"Spent: 0.00",
		"Remaining: 0.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY  GOAL     ALLOCATED  SPENT  REMAINING  WARNING",
		"Repair    1000.00  0.00       0.00   0.00       needs refill 1000.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowUnsupportedFlag(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")
	runCommands(t, dbPath, [][]string{{"budget", "create", "Home Budget"}})
	assertCommandFails(t, dbPath, `unsupported flag "--bogus"`, "budget", "show", "Home Budget", "2026-08", "--bogus")
}

func TestBudgetShowPlanMixedGoals(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"category", "create", "Repair"},
		{"goal", "create", "monthly", "2026-07", "null", "Groceries", "50.00"},
		{"goal", "create", "save", "2026-08", "2026-12", "Household", "1000.00"},
		{"goal", "create", "refill", "2026-07", "null", "Repair", "1000.00"},
		{"allocation", "create", "2026-07", "Repair", "400.00"},
		{"allocation", "create", "2026-08", "Groceries", "30.00"},
		{"allocation", "create", "2026-08", "Repair", "100.00"},
		{"transaction", "create", "2026-07-20", "Checking", "Hardware", "50.00", "0.00", "fix"},
		{"transaction", "category", "create", "1", "50.00", "0.00", "--category", "Repair"},
		{"transaction", "create", "2026-08-10", "Checking", "Hardware", "25.00", "0.00", "more"},
		{"transaction", "category", "create", "2", "25.00", "0.00", "--category", "Repair"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08", "--plan")
	if exitCode != 0 {
		t.Fatalf("budget show --plan failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -205.00",
		"Income: 0.00",
		"Goals: 1250.00",
		"Allocated: 130.00",
		"Spent: 0.00",
		"Remaining: 130.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY   GOAL     ALLOCATED  SPENT  REMAINING  WARNING",
		"Groceries  50.00    30.00      0.00   30.00      underfunded 20.00",
		"Household  200.00   0.00       0.00   0.00       behind 200.00",
		"Repair     1000.00  100.00     0.00   100.00     needs refill 1000.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowPlanHidesMonthToDate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"allocation", "create", "2026-08", "Groceries", "100.00"},
		{"transaction", "create", "2026-08-05", "Checking", "Market", "40.00", "0.00", "shop"},
		{"transaction", "category", "create", "1", "40.00", "0.00", "--category", "Groceries"},
		{"transaction", "create", "2026-08-10", "Checking", "Employer", "0.00", "200.00", "paycheck"},
		{"transaction", "category", "create", "2", "0.00", "200.00", "--income"},
		{"transaction", "create", "2026-08-15", "Checking", "ATM", "50.00", "0.00", "cash"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: 50.00",
		"Income: 200.00",
		"Goals: 0.00",
		"Allocated: 100.00",
		"Spent: 40.00",
		"Remaining: 60.00",
		"Uncategorized: 50.00",
		"",
		"No group:",
		"CATEGORY   GOAL  ALLOCATED  SPENT  REMAINING  WARNING",
		"Groceries        100.00     40.00  60.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected normal stdout: got %q want %q", stdout, want)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08", "--plan")
	if exitCode != 0 {
		t.Fatalf("budget show --plan failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want = strings.Join([]string{
		"Available: 10.00",
		"Income: 0.00",
		"Goals: 0.00",
		"Allocated: 100.00",
		"Spent: 0.00",
		"Remaining: 100.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY   GOAL  ALLOCATED  SPENT  REMAINING  WARNING",
		"Groceries        100.00     0.00   100.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected plan stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowGoalVariants(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Fees"},
		{"category", "create", "Fun"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"category", "create", "Savings"},
		{"category", "create", "Travel"},
		{"allocation", "create", "2026-07", "Groceries", "100.00"},
		{"allocation", "create", "2026-07", "Household", "200.00"},
		{"allocation", "create", "2026-08", "Groceries", "50.00"},
		{"allocation", "create", "2026-08", "Household", "25.00"},
		{"goal", "create", "save", "2026-09", "2026-12", "Fees", "400.00"},
		{"goal", "create", "monthly", "2026-06", "2026-07", "Fun", "30.00"},
		{"goal", "create", "save", "2026-07", "2026-10", "Groceries", "400.00"},
		{"goal", "create", "save", "2026-07", "2026-10", "Household", "200.00"},
		{"goal", "create", "save", "2026-08", "2026-08", "Savings", "50.00"},
		{"goal", "create", "monthly", "2026-09", "null", "Travel", "25.00"},
		{"transaction", "create", "2026-08-05", "Checking", "Employer", "0.00", "42.00", "paycheck"},
		{"transaction", "category", "create", "1", "0.00", "42.00", "--income"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-08")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -333.00",
		"Income: 42.00",
		"Goals: 150.00",
		"Allocated: 75.00",
		"Spent: 0.00",
		"Remaining: 375.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY   GOAL    ALLOCATED  SPENT  REMAINING  WARNING",
		"Fees               0.00       0.00   0.00",
		"Fun                0.00       0.00   0.00",
		"Groceries  100.00  50.00      0.00   150.00     behind 50.00",
		"Household  0.00    25.00      0.00   225.00",
		"Savings    50.00   0.00       0.00   0.00       behind 50.00",
		"Travel             0.00       0.00   0.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowRollover(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"allocation", "create", "2026-06", "Groceries", "100.00"},
		{"transaction", "create", "2026-06-05", "Checking", "Market", "40.00", "0.00", ""},
		{"transaction", "category", "create", "1", "40.00", "0.00", "--category", "Groceries"},
		{"allocation", "create", "2026-07", "Groceries", "50.00"},
		{"transaction", "create", "2026-07-10", "Checking", "Market", "20.00", "0.00", ""},
		{"transaction", "category", "create", "2", "20.00", "0.00", "--category", "Groceries"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-07")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -150.00",
		"Income: 0.00",
		"Goals: 0.00",
		"Allocated: 50.00",
		"Spent: 20.00",
		"Remaining: 90.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY   GOAL  ALLOCATED  SPENT  REMAINING  WARNING",
		"Groceries        50.00      20.00  90.00",
		"Household        0.00       0.00   0.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowNoNegativeRollover(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"allocation", "create", "2026-06", "Groceries", "50.00"},
		{"transaction", "create", "2026-06-05", "Checking", "Market", "70.00", "0.00", ""},
		{"transaction", "category", "create", "1", "70.00", "0.00", "--category", "Groceries"},
		{"allocation", "create", "2026-07", "Groceries", "50.00"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget", "2026-06")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want := strings.Join([]string{
		"Available: -70.00",
		"Income: 0.00",
		"Goals: 0.00",
		"Allocated: 50.00",
		"Spent: 70.00",
		"Remaining: -20.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY   GOAL  ALLOCATED  SPENT  REMAINING  WARNING",
		"Groceries        50.00      70.00  -20.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "budget", "show", "Home Budget", "2026-07")
	if exitCode != 0 {
		t.Fatalf("budget show failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	want = strings.Join([]string{
		"Available: -120.00",
		"Income: 0.00",
		"Goals: 0.00",
		"Allocated: 50.00",
		"Spent: 0.00",
		"Remaining: 50.00",
		"Uncategorized: 0.00",
		"",
		"No group:",
		"CATEGORY   GOAL  ALLOCATED  SPENT  REMAINING  WARNING",
		"Groceries        50.00      0.00   50.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestBudgetShowMonthsIgnoresGoals(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"category", "create", "Groceries"},
		{"goal", "create", "monthly", "2026-08", "null", "Groceries", "50.00"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "budget", "show", "Home Budget")
	if exitCode != 0 {
		t.Fatalf("budget show failed: stdout=%q stderr=%q", stdout, stderr)
	}
	if stdout != "MONTH\n" {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, "MONTH\n")
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestGoalListOrderingAndTypes(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"category", "create", "Zebra"},
		{"category", "create", "Apple"},
		{"category", "create", "Mango"},
		{"category", "create", "Repair"},
		{"goal", "create", "Monthly", "2020-01", "null", "Mango", "10.00"},
		{"goal", "create", "SAVE", "2020-01", "2020-12", "Apple", "120.00"},
		{"goal", "create", "monthly", "2020-01", "null", "Zebra", "5.00"},
		{"goal", "create", "refill", "2020-01", "null", "Repair", "500.00"},
	})

	stdout, stderr, exitCode := invoke(t, dbPath, "goal", "list")
	if exitCode != 0 {
		t.Fatalf("goal list failed: stdout=%q stderr=%q", stdout, stderr)
	}

	want := strings.Join([]string{
		"TYPE     CATEGORY  START    END      AMOUNT  MONTHLY",
		"save     Apple     2020-01  2020-12  120.00  0.00",
		"monthly  Mango     2020-01           10.00   10.00",
		"refill   Repair    2020-01           500.00  500.00",
		"monthly  Zebra     2020-01           5.00    5.00",
		"",
	}, "\n")

	if stdout != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout, want)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestParseAmount(t *testing.T) {
	tests := []struct {
		input string
		want  int64
		err   bool
	}{
		{"42", 4200, false},
		{"42.00", 4200, false},
		{"$42.00", 4200, false},
		{"42.5", 4250, false},
		{"42.999", 4299, false},
		{"-4.20", -420, false},
		{"0", 0, false},
		{"0.00", 0, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := parseAmount(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("parseAmount(%q) expected error, got %d", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseAmount(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseAmount(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestAccountCategorizeDeleteExisting(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
		{"transaction", "category", "create", "1", "20.00", "0.00", "--category", "Groceries"},
		{"transaction", "category", "create", "1", "10.00", "0.00", "--category", "Household"},
	})

	input := strings.Join([]string{
		"d",
		"1",
		"a",
		"@Savings",
		"",
		"",
		"o",
		"n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "→ removed: Groceries") {
		t.Fatalf("expected removed message, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Household") {
		t.Fatalf("expected Household to remain, got %q", show)
	}
	if !strings.Contains(show, "@Savings") {
		t.Fatalf("expected @Savings added, got %q", show)
	}
	if strings.Contains(show, "Groceries") {
		t.Fatalf("did not expect deleted Groceries to persist, got %q", show)
	}
}

func TestAccountCategorizeDeleteInvalidIndex(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
		{"transaction", "category", "create", "1", "20.00", "0.00", "--category", "Groceries"},
	})

	input := strings.Join([]string{
		"d",
		"5",
		"s",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "invalid selection") {
		t.Fatalf("expected invalid selection message, got %q", stdout)
	}
	if !strings.Contains(stdout, "→ skipped") {
		t.Fatalf("expected skip after invalid delete, got %q", stdout)
	}
}

func TestAccountCategorizeDeleteEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
	})

	input := strings.Join([]string{
		"d",
		"s",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "no categorizations to delete") {
		t.Fatalf("expected no-categorizations message, got %q", stdout)
	}
}

func TestAccountCategorizeOkMismatchStaysInLoop(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
		{"transaction", "category", "create", "1", "20.00", "0.00", "--category", "Groceries"},
	})

	input := strings.Join([]string{
		"o",
		"a",
		"@Savings",
		"",
		"",
		"o",
		"n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "totals don't match: remaining 22.00 out / 0.00 in") {
		t.Fatalf("expected totals-don't-match message, got %q", stdout)
	}
	if !strings.Contains(stdout, "Done — 1 categorized, 0 skipped") {
		t.Fatalf("expected Done summary, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "@Savings") {
		t.Fatalf("expected @Savings persisted after fix, got %q", show)
	}
	if !strings.Contains(show, "Groceries") {
		t.Fatalf("expected Groceries persisted, got %q", show)
	}
}

func TestAccountCategorizeOkPersistsOnExistingPartial(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
		{"transaction", "category", "create", "1", "40.00", "0.00", "--category", "Groceries"},
	})

	input := strings.Join([]string{
		"a",
		"Household",
		"",
		"",
		"o",
		"n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "Done — 1 categorized, 0 skipped") {
		t.Fatalf("expected Done summary, got %q", stdout)
	}

	second := runCategorize(t, dbPath, "", "Checking")
	if !strings.Contains(second, "No transactions need categorization") {
		t.Fatalf("expected transaction complete after ok, got %q", second)
	}
}

func TestAccountCategorizeSaveDefaultNo(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
	})

	input := strings.Join([]string{
		"a",
		"Groceries",
		"",
		"",
		"o",
		"n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "→ categorized") {
		t.Fatalf("expected categorized message, got %q", stdout)
	}
	if strings.Contains(stdout, "saved default") {
		t.Fatalf("did not expect default saved, got %q", stdout)
	}

	runCommands(t, dbPath, [][]string{
		{"transaction", "create", "2026-08-29", "Checking", "Market", "42.00", "0.00", ""},
	})

	second := runCategorize(t, dbPath, "q\n", "Checking")
	if strings.Contains(second, "(pre-filled from payee default)") {
		t.Fatalf("did not expect pre-fill since default not saved, got %q", second)
	}
}

func TestAccountCategorizeSaveDefaultPercentRounding(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "A"},
		{"category", "create", "B"},
		{"category", "create", "C"},
		{"transaction", "create", "2026-08-28", "Checking", "Three Way", "10.00", "0.00", ""},
	})

	input := strings.Join([]string{
		"a", "A", "3.33", "",
		"a", "B", "3.33", "",
		"a", "C", "3.34", "",
		"o", "y",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, `saved default for "Three Way"`) {
		t.Fatalf("expected saved default message, got %q", stdout)
	}

	runCommands(t, dbPath, [][]string{
		{"transaction", "create", "2026-08-29", "Checking", "Three Way", "10.00", "0.00", ""},
	})

	second := runCategorize(t, dbPath, "o\n", "Checking")
	if !strings.Contains(second, "(pre-filled from payee default)") {
		t.Fatalf("expected pre-fill from saved default, got %q", second)
	}
	if !strings.Contains(second, "matches existing default") {
		t.Fatalf("expected matches existing default, got %q", second)
	}
}

func TestAccountCategorizeSkipLeavesUnchanged(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
		{"transaction", "category", "create", "1", "20.00", "0.00", "--category", "Groceries"},
	})

	stdout := runCategorize(t, dbPath, "s\n", "Checking")
	if !strings.Contains(stdout, "Done — 0 categorized, 1 skipped") {
		t.Fatalf("expected Done summary, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Groceries") {
		t.Fatalf("expected existing category unchanged after skip, got %q", show)
	}
}

func TestAccountCategorizeQuitMidFlow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-29", "Checking", "Market", "42.00", "0.00", ""},
		{"transaction", "create", "2026-08-28", "Checking", "Cafe", "42.00", "0.00", ""},
	})

	input := strings.Join([]string{
		"a", "Groceries", "", "",
		"o", "n",
		"q",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "→ quit (1 categorized, 0 skipped)") {
		t.Fatalf("expected quit summary, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if !strings.Contains(show, "Groceries") {
		t.Fatalf("expected first transaction categorized, got %q", show)
	}
}

func TestAccountCategorizeEOFQuits(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "42.00", "0.00", ""},
	})

	stdout := runCategorize(t, dbPath, "", "Checking")
	if !strings.Contains(stdout, "→ quit") {
		t.Fatalf("expected quit on EOF, got %q", stdout)
	}
	if strings.Contains(stdout, "Done —") {
		t.Fatalf("did not expect Done summary on EOF, got %q", stdout)
	}

	show := showAccount(t, dbPath, "Checking")
	if strings.Contains(show, "Groceries") {
		t.Fatalf("expected transaction to remain uncategorized after EOF, got %q", show)
	}
}

func TestAccountCategorizeOrderingAndCounts(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	runCommands(t, dbPath, [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"category", "create", "Groceries"},
		{"transaction", "create", "2026-08-30", "Checking", "Newest", "42.00", "0.00", ""},
		{"transaction", "create", "2026-08-29", "Checking", "Middle", "42.00", "0.00", ""},
		{"transaction", "create", "2026-08-28", "Checking", "Oldest", "42.00", "0.00", ""},
	})

	input := strings.Join([]string{
		"a", "Groceries", "", "",
		"o", "n",
		"s",
		"a", "Groceries", "", "",
		"o", "n",
	}, "\n")

	stdout := runCategorize(t, dbPath, input, "Checking")
	if !strings.Contains(stdout, "Done — 2 categorized, 1 skipped") {
		t.Fatalf("expected Done summary, got %q", stdout)
	}
	if !strings.Contains(stdout, "2026-08-30") {
		t.Fatalf("expected newest transaction processed first, got %q", stdout)
	}
}

func invoke(t *testing.T, dbPath string, args ...string) (string, string, int) {
	t.Helper()

	return invokeWithInput(t, dbPath, "", args...)
}

func invokeWithInput(t *testing.T, dbPath, input string, args ...string) (string, string, int) {
	t.Helper()

	restore := chdirToRepoRoot(t)
	defer restore()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := run(append([]string{"--db", dbPath}, args...), strings.NewReader(input), &stdout, &stderr)
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
