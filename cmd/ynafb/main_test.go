package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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
			args:  []string{"allocation", "create", "Groceries", "2500"},
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
			args:  []string{"goal", "create", "Emergency Fund", "target", "2026-08-28", "null", "Savings", "5000"},
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
			args:  []string{"transaction", "create", "2026-08-28", "Checking", "Market", "2500", "0", "weekly shop"},
			want:  "created transaction 1\n",
		},
		{
			name:  "transaction category",
			setup: [][]string{{"budget", "create", "Home Budget"}, {"account", "create", "Checking"}, {"category", "create", "Groceries"}, {"payee", "create", "Market"}, {"transaction", "create", "2026-08-28", "Checking", "Market", "2500", "0", "weekly shop"}},
			args:  []string{"transaction", "category", "create", "1", "2500", "0", "--category", "Groceries"},
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

	stdout, stderr, exitCode := invoke(t, dbPath, "transaction", "create", "2026-08-28", "Checking", "Corner Store", "1250", "0", "snacks")
	if exitCode != 0 {
		t.Fatalf("transaction create failed with exit code %d, stdout=%q stderr=%q", exitCode, stdout, stderr)
	}

	if stdout != "created transaction 1\n" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	stdout, stderr, exitCode = invoke(t, dbPath, "transaction", "category", "create", "1", "1250", "0", "--category", "Groceries")
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
		{"transaction", "create", "2026-08-28", "ws-visa", "Online Shop", "4200", "0", "pending import"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "4200", "0", "import mismatch"},
		{"transaction", "category", "create", "1", "4000", "0", "--category", "Groceries"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Cafe", "1000", "0", "coffee"},
		{"transaction", "category", "create", "1", "1000", "0", "--category", "Groceries"},
		{"transaction", "create", "2026-08-29", "Checking", "Bank", "1500", "0", "move money"},
		{"transaction", "category", "create", "2", "1500", "0", "--other-account", "Savings"},
		{"transaction", "create", "2026-08-30", "Checking", "Market", "2500", "0", "weekly shop"},
		{"transaction", "category", "create", "3", "2000", "0", "--category", "Groceries"},
		{"transaction", "category", "create", "3", "500", "0", "--category", "Household"},
		{"transaction", "create", "2026-08-31", "Savings", "Market", "9999", "0", "should not appear"},
		{"transaction", "category", "create", "4", "9999", "0", "--category", "Groceries"},
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
		{"transaction", "create", "2026-08-09", "ws-visa", "superC", "100", "0", ""},
		{"transaction", "category", "create", "1", "25", "0", "--category", "groceries"},
		{"transaction", "category", "create", "1", "75", "0", "--other-account", "merry"},
		{"transaction", "create", "2026-08-09", "ws-visa", "superC", "200", "0", ""},
		{"transaction", "category", "create", "2", "100", "0", "--category", "groceries"},
		{"transaction", "category", "create", "2", "100", "0", "--other-account", "merry"},
		{"transaction", "create", "2026-08-09", "ws-visa", "superC", "250", "0", ""},
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

func TestTransactionCategoryCreateRejectsMultipleTargets(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"payee", "create", "Market"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "2500", "0", "weekly shop"},
	}

	for _, args := range setup {
		stdout, stderr, exitCode := invoke(t, dbPath, args...)
		if exitCode != 0 {
			t.Fatalf("setup command %q failed with exit code %d, stdout=%q stderr=%q", strings.Join(args, " "), exitCode, stdout, stderr)
		}
	}

	stdout, stderr, exitCode := invoke(t, dbPath, "transaction", "category", "create", "1", "2500", "0", "--category", "Groceries", "--other-account", "Savings")
	if exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}

	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}

	want := "error: set exactly one of --category or --other-account\n"
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
		{"allocation", "create", "Groceries", "5000"},
		{"allocation", "create", "Household", "2500"},
		{"goal", "create", "Vacation", "target", "2026-09-01", "null", "Groceries", "100000"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "2500", "0", "shop"},
		{"transaction", "create", "2026-08-29", "Savings", "Cafe", "1000", "0", "coffee"},
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
		{[]string{"allocation", "list"}, "CATEGORY   AMOUNT\nGroceries  50.00\nHousehold  25.00\n"},
		{[]string{"goal", "list"}, "NAME      TYPE    START       END  CATEGORY   AMOUNT\nVacation  target  2026-09-01       Groceries  1000.00\n"},
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
		{"allocation", "create", "Groceries", "5000"},
		{"goal", "create", "Vacation", "target", "2026-09-01", "null", "Groceries", "100000"},
		{"transaction", "create", "2026-08-28", "Checking", "Market", "2500", "0", "shop"},
		{"transaction", "category", "create", "1", "2000", "0", "--category", "Groceries"},
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
		{[]string{"allocation", "delete", "Groceries"}, "deleted allocation \"Groceries\"\n"},
		{[]string{"goal", "delete", "Vacation"}, "deleted goal \"Vacation\"\n"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "2500", "0", "shop"},
		{"transaction", "category", "create", "1", "2500", "0", "--category", "Groceries"},
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

func TestAccountCategorize(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ynafb.db")

	setup := [][]string{
		{"budget", "create", "Home Budget"},
		{"account", "create", "Checking"},
		{"account", "create", "Savings"},
		{"category", "create", "Groceries"},
		{"category", "create", "Household"},
		{"transaction", "create", "2026-08-29", "Checking", "Grocery Mart", "4200", "0", ""},
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "4200", "0", "import mismatch"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "4200", "0", ""},
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
		{"transaction", "create", "2026-07-26", "Checking", "IGA #8644", "2957", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Online Shop", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
		{"transaction", "category", "create", "1", "4200", "0", "--category", "Groceries"},
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
		{"category", "create", "Income"},
		{"transaction", "create", "2026-08-28", "Checking", "Employer", "0", "4200", "paycheck"},
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
		{"transaction", "create", "2026-08-28", "Savings", "Me", "0", "500", ""},
		{"transaction", "category", "create", "1", "0", "500", "--other-account", "Checking"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Bank", "1500", "0", "move"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
		{"transaction", "category", "create", "1", "2000", "0", "--category", "Groceries"},
		{"transaction", "category", "create", "1", "1000", "0", "--category", "Household"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
		{"transaction", "category", "create", "1", "2000", "0", "--category", "Groceries"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
		{"transaction", "category", "create", "1", "2000", "0", "--category", "Groceries"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
		{"transaction", "category", "create", "1", "4000", "0", "--category", "Groceries"},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-29", "Checking", "Market", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Three Way", "1000", "0", ""},
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
		{"transaction", "create", "2026-08-29", "Checking", "Three Way", "1000", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
		{"transaction", "category", "create", "1", "2000", "0", "--category", "Groceries"},
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
		{"transaction", "create", "2026-08-29", "Checking", "Market", "4200", "0", ""},
		{"transaction", "create", "2026-08-28", "Checking", "Cafe", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-28", "Checking", "Market", "4200", "0", ""},
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
		{"transaction", "create", "2026-08-30", "Checking", "Newest", "4200", "0", ""},
		{"transaction", "create", "2026-08-29", "Checking", "Middle", "4200", "0", ""},
		{"transaction", "create", "2026-08-28", "Checking", "Oldest", "4200", "0", ""},
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
