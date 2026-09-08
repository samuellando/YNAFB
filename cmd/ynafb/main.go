package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"samuellando.com/YNAFB/data"
	"samuellando.com/YNAFB/internal/auth"
	dbutil "samuellando.com/YNAFB/internal/db"
	"samuellando.com/YNAFB/internal/db/types"
	"samuellando.com/YNAFB/internal/importer"
)

const defaultLogin = "cli"

type loginContext struct {
	ID       int64
	Username string
}

type budgetContext struct {
	ID      int64
	LoginID int64
	Name    string
}

type categoryTargetArgs struct {
	OtherAccount string
	Category     string
	Income       bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("ynafb", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dbPath := fs.String("db", "./ynafb.db", "SQLite database path")
	budgetName := fs.String("budget", "", "Budget name")
	loginName := fs.String("login", "", "Login username (default: cli)")
	fs.Usage = func() {
		usage(stderr)
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return fail(stderr, err)
	}

	remaining := fs.Args()
	if len(remaining) < 2 {
		usage(stderr)
		return 1
	}

	resource := remaining[0]
	action := remaining[1]

	db, err := dbutil.Open(*dbPath)
	if err != nil {
		return fail(stderr, err)
	}
	defer db.Close()

	queries := data.New(db)
	ctx := context.Background()

	if resource == "login" {
		if err := executeLogin(ctx, queries, action, remaining[2:], stderr, stdout); err != nil {
			return fail(stderr, err)
		}
		return 0
	}

	login, err := resolveLogin(ctx, queries, *loginName)
	if err != nil {
		return fail(stderr, err)
	}

	if err := executeResourceAction(ctx, db, queries, login, resource, action, *budgetName, stdin, remaining[2:], stdout); err != nil {
		return fail(stderr, err)
	}

	return 0
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] login create [username] [password]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] login list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] login delete [username]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] budget create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] budget update [name] [new_name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] budget list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] budget delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] budget show [budget_name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] budget show [budget_name] [month]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account update [name] [new_name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account show [account]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account reconcile [account] [date]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account import [account] [pdf]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] account categorize [account]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] allocation create [month] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] allocation update [month] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] category create [name] [--group name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] category update [name] [new_name] [--group name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] category list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] category delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] group create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] group update [name] [new_name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] group list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] group delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] goal create [type] [start] [end|null] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] goal update [category] [type] [start] [end|null] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] goal list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] goal delete [category]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] payee create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] payee update [name] [new_name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] payee list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] payee delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] payee default-category create [payee] [percent] [--category name | --other-account name | --income]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] payee default-category update [id] [payee] [percent] [--category name | --other-account name | --income]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] payee default-category delete [id]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] transaction create [date] [account] [payee] [total_out] [total_in] [note]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] transaction update [id] [date] [account] [payee] [total_out] [total_in] [note]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] transaction list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] transaction delete [id]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] [--budget name] transaction category create [transaction] [outflow] [inflow] [--category name | --other-account name | --income]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] transaction category update [id] [transaction] [outflow] [inflow] [--category name | --other-account name | --income]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--login username] transaction category delete [id]\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Dates accept RFC3339 or YYYY-MM-DD. Use null for goal end dates (monthly goals).\n")
	fmt.Fprintf(w, "Goal months are YYYY-MM; save goals require an end month.\n")
	fmt.Fprintf(w, "Amounts are entered in dollars, e.g. 25.00 or 12.50.\n")
	fmt.Fprintf(w, "Omit --budget only when exactly one budget exists.\n")
	fmt.Fprintf(w, "Omit --login to use the default %q login.\n", defaultLogin)
}

func executeLogin(ctx context.Context, queries *data.Queries, action string, args []string, stderr, stdout io.Writer) error {
	switch action {
	case "create":
		if len(args) < 1 || len(args) > 2 {
			return fmt.Errorf("login create requires [username] [password]")
		}

		password := ""
		if len(args) == 2 {
			password = args[1]
		}
		if password == "" {
			generated, err := generatePassword()
			if err != nil {
				return err
			}
			password = generated
			fmt.Fprintf(stderr, "generated password for %q: %s\n", args[0], generated)
		}

		if len(password) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}

		result, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: args[0], Password: hash})
		if err != nil {
			return err
		}

		printCreated(stdout, "login", result.ID)
		return nil

	case "list":
		logins, err := queries.ListLogins(ctx)
		if err != nil {
			return err
		}

		return printLogins(stdout, logins)

	case "delete":
		if len(args) != 1 {
			return fmt.Errorf("login delete requires [username]")
		}

		login, err := queries.GetLoginByUsername(ctx, data.GetLoginByUsernameParams{Username: args[0]})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unknown login %q", args[0])
			}
			return err
		}

		if err := queries.DeleteLogin(ctx, data.DeleteLoginParams{ID: login.ID}); err != nil {
			return err
		}

		fmt.Fprintf(stdout, "deleted login %q\n", args[0])
		return nil

	default:
		return fmt.Errorf("unsupported action %q for resource %q", action, "login")
	}
}

func resolveLogin(ctx context.Context, queries *data.Queries, loginName string) (loginContext, error) {
	name := loginName
	if name == "" {
		name = defaultLogin
	}

	login, err := queries.GetLoginByUsername(ctx, data.GetLoginByUsernameParams{Username: name})
	if err == nil {
		return loginContext{ID: login.ID, Username: login.Username}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return loginContext{}, err
	}
	if loginName != "" {
		return loginContext{}, fmt.Errorf("unknown login %q", loginName)
	}

	hash, err := defaultLoginPasswordHash()
	if err != nil {
		return loginContext{}, err
	}

	created, err := queries.CreateLogin(ctx, data.CreateLoginParams{Username: name, Password: hash})
	if err != nil {
		return loginContext{}, err
	}

	return loginContext{ID: created.ID, Username: created.Username}, nil
}

var (
	defaultLoginHashOnce sync.Once
	defaultLoginHash     string
	defaultLoginHashErr  error
)

func defaultLoginPasswordHash() (string, error) {
	defaultLoginHashOnce.Do(func() {
		password, err := generatePassword()
		if err != nil {
			defaultLoginHashErr = err
			return
		}
		defaultLoginHash, defaultLoginHashErr = auth.HashPassword(password)
	})
	return defaultLoginHash, defaultLoginHashErr
}

func generatePassword() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func executeResourceAction(ctx context.Context, db *sql.DB, queries *data.Queries, login loginContext, resource, action, budgetName string, stdin io.Reader, args []string, stdout io.Writer) error {
	if resource == "budget" {
		switch action {
		case "create":
			if len(args) != 1 {
				return fmt.Errorf("budget create requires [name]")
			}

			result, err := queries.CreateBudget(ctx, data.CreateBudgetParams{
				LoginID: login.ID,
				Name:    args[0],
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "budget", result.ID)
			return nil

		case "update":
			if len(args) != 2 {
				return fmt.Errorf("budget update requires [name] [new_name]")
			}

			budget, err := resolveBudget(ctx, queries, login, args[0])
			if err != nil {
				return err
			}

			_, err = queries.UpdateBudget(ctx, data.UpdateBudgetParams{
				Name:    args[1],
				ID:      budget.ID,
				LoginID: login.ID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unknown budget %q", args[0])
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated budget %q\n", args[1])
			return nil

		case "list":
			budgets, err := queries.ListBudgets(ctx, data.ListBudgetsParams{LoginID: login.ID})
			if err != nil {
				return err
			}

			return printBudgets(stdout, budgets)

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("budget delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeleteBudget(ctx, data.DeleteBudgetParams{ID: budget.ID, LoginID: login.ID}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted budget %q\n", budget.Name)
			return nil

		case "show":
			positional := make([]string, 0, 2)
			for _, a := range args {
				if strings.HasPrefix(a, "--") {
					return fmt.Errorf("unsupported flag %q", a)
				}
				positional = append(positional, a)
			}
			if len(positional) != 1 && len(positional) != 2 {
				return fmt.Errorf("budget show requires [budget_name] or [budget_name] [month]")
			}

		budget, err := resolveBudget(ctx, queries, login, positional[0])
		if err != nil {
			return err
		}

		month := time.Now()
		if len(positional) == 2 {
			month, err = parseMonth("month", positional[1])
			if err != nil {
				return err
			}
		} else {
			month, err = parseMonth("month", time.Now().Format("2006-01"))
			if err != nil {
				return err
			}
		}

		rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
			ID:      budget.ID,
			LoginID: budget.LoginID,
			Month:   types.UnixTime{Time: month},
		})
		if err != nil {
			return err
		}

		goals, err := queries.ListGoalsValues(ctx, data.ListGoalsValuesParams{
			ID:      budget.ID,
			LoginID: budget.LoginID,
			Month:   types.UnixTime{Time: month},
		})
		if err != nil {
			return err
		}

		var goalsTotal int64 = 0
		for _, goal := range goals {
			goalsTotal += goal.AmountForMonth
		}

		summary, err := queries.GetBudgetMonthSummary(ctx, data.GetBudgetMonthSummaryParams{
			ID:      budget.ID,
			LoginID: budget.LoginID,
			Month:   types.UnixTime{Time: month},
		})
			if err != nil {
				return err
			}

			if err := printBudgetMonthSummary(stdout, summary, goalsTotal); err != nil {
				return err
			}

			return printBudgetMonthCategories(stdout, rows, goals)

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}
	}

	switch resource {
	case "account":
		switch action {
		case "create":
			if len(args) != 1 {
				return fmt.Errorf("account create requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			result, err := queries.CreateAccount(ctx, data.CreateAccountParams{
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
				Name:     args[0],
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "account", result.ID)
			return nil

		case "update":
			if len(args) != 2 {
				return fmt.Errorf("account update requires [name] [new_name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			_, err = queries.UpdateAccount(ctx, data.UpdateAccountParams{
				Name:     args[1],
				ID:       accountID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unknown account %q", args[0])
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated account %q\n", args[1])
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			accounts, err := queries.ListAccountsBalances(ctx, data.ListAccountsBalancesParams{BudgetID: budget.ID, LoginID: budget.LoginID})
			if err != nil {
				return err
			}

			return printAccountBalances(stdout, accounts)

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("account delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeleteAccount(ctx, data.DeleteAccountParams{
				ID:       accountID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted account %q\n", args[0])
			return nil

		case "show":
			if len(args) != 1 {
				return fmt.Errorf("account show requires [account]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			balances, err := queries.GetAccountBalances(ctx, data.GetAccountBalancesParams{
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
				ID:       accountID,
			})
			if err != nil {
				return err
			}

			if err := printAccountSummary(stdout, args[0], balances.Balance, balances.ReconciledBalance); err != nil {
				return err
			}

			transactions, err := queries.ListAccountTransactions(ctx, data.ListAccountTransactionsParams{
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
				ID:       accountID,
			})
			if err != nil {
				return err
			}

			return printAccountTransactions(stdout, accountID, transactions)

		case "reconcile":
			if len(args) != 2 {
				return fmt.Errorf("account reconcile requires [account] [date]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			date, err := parseTime("date", args[1])
			if err != nil {
				return err
			}

			return reconcileAccount(ctx, queries, budget, accountID, args[0], date, stdin, stdout)

		case "import":
			if len(args) != 2 {
				return fmt.Errorf("account import requires [account] [pdf]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			return importAccountTransactions(ctx, db, queries, budget, accountID, args[1], stdout)

		case "categorize":
			if len(args) != 1 {
				return fmt.Errorf("account categorize requires [account]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			return categorizeAccount(ctx, db, queries, budget, accountID, args[0], stdin, stdout)

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "allocation":
		switch action {
		case "create":
			if len(args) != 3 {
				return fmt.Errorf("allocation create requires [month] [category] [amount]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			month, err := parseMonth("month", args[0])
			if err != nil {
				return err
			}

			category, err := resolveCategoryID(ctx, queries, budget, args[1])
			if err != nil {
				return err
			}

			amount, err := parseAmount(args[2])
			if err != nil {
				return err
			}

			result, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
				BudgetID:   budget.ID,
				LoginID:    budget.LoginID,
				CategoryID: category,
				Month:      types.UnixTime{Time: month},
				Amount:     amount,
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "allocation", result.ID)
			return nil

		case "update":
			if len(args) != 3 {
				return fmt.Errorf("allocation update requires [month] [category] [amount]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			month, err := parseMonth("month", args[0])
			if err != nil {
				return err
			}

			category, err := resolveCategoryID(ctx, queries, budget, args[1])
			if err != nil {
				return err
			}

			amount, err := parseAmount(args[2])
			if err != nil {
				return err
			}

			_, err = queries.UpdateAllocation(ctx, data.UpdateAllocationParams{
				Amount:     amount,
				BudgetID:   budget.ID,
				LoginID:    budget.LoginID,
				CategoryID: category,
				Month:      types.UnixTime{Time: month},
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("no allocation for %s %q in budget %q", month.Format("2006-01"), args[1], budget.Name)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated allocation %s %q\n", month.Format("2006-01"), args[1])
			return nil

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "category":
		switch action {
		case "create":
			var name, group string
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--group":
					if i+1 >= len(args) {
						return fmt.Errorf("missing value for --group")
					}
					group = args[i+1]
					i++
				default:
					if strings.HasPrefix(args[i], "--") {
						return fmt.Errorf("unsupported flag %q", args[i])
					}
					if name != "" {
						return fmt.Errorf("category create requires [name]")
					}
					name = args[i]
				}
			}

			if name == "" {
				return fmt.Errorf("category create requires [name]")
			}

			if strings.EqualFold(name, "income") {
				return fmt.Errorf("category name %q is reserved", name)
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			var categoryGroup sql.NullInt64
			if group != "" {
				groupID, err := resolveOrCreateCategoryGroup(ctx, queries, budget, group)
				if err != nil {
					return err
				}
				categoryGroup = sql.NullInt64{Int64: groupID, Valid: true}
			}

			result, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
				BudgetID:        budget.ID,
				LoginID:         budget.LoginID,
				Name:            name,
				CategoryGroupID: categoryGroup,
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "category", result.ID)
			return nil

		case "update":
			var name, newName, group string
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--group":
					if i+1 >= len(args) {
						return fmt.Errorf("missing value for --group")
					}
					group = args[i+1]
					i++
				default:
					if strings.HasPrefix(args[i], "--") {
						return fmt.Errorf("unsupported flag %q", args[i])
					}
					if name == "" {
						name = args[i]
					} else if newName == "" {
						newName = args[i]
					} else {
						return fmt.Errorf("category update requires [name] [new_name]")
					}
				}
			}

			if name == "" || newName == "" {
				return fmt.Errorf("category update requires [name] [new_name]")
			}

			if strings.EqualFold(newName, "income") {
				return fmt.Errorf("category name %q is reserved", newName)
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			categoryID, err := resolveCategoryID(ctx, queries, budget, name)
			if err != nil {
				return err
			}

			var categoryGroup sql.NullInt64
			if group != "" {
				groupID, err := resolveOrCreateCategoryGroup(ctx, queries, budget, group)
				if err != nil {
					return err
				}
				categoryGroup = sql.NullInt64{Int64: groupID, Valid: true}
			}

			_, err = queries.UpdateCategory(ctx, data.UpdateCategoryParams{
				Name:            newName,
				CategoryGroupID: categoryGroup,
				ID:              categoryID,
				BudgetID:        budget.ID,
				LoginID:         budget.LoginID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unknown category %q", name)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated category %q\n", newName)
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			categories, err := queries.ListCategories(ctx, data.ListCategoriesParams{BudgetID: budget.ID, LoginID: budget.LoginID})
			if err != nil {
				return err
			}

			return printNames(stdout, categories, func(c data.ListCategoriesRow) string { return stringValue(c.Name) })

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("category delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			categoryID, err := resolveCategoryID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeleteCategory(ctx, data.DeleteCategoryParams{
				ID:       categoryID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted category %q\n", args[0])
			return nil

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "group":
		switch action {
		case "create":
			if len(args) != 1 {
				return fmt.Errorf("group create requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			result, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
				Name:     args[0],
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "group", result)
			return nil

		case "update":
			if len(args) != 2 {
				return fmt.Errorf("group update requires [name] [new_name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			groupID, err := resolveCategoryGroupID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			_, err = queries.UpdateCategoryGroup(ctx, data.UpdateCategoryGroupParams{
				Name:     args[1],
				ID:       groupID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unknown group %q", args[0])
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated group %q\n", args[1])
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			groups, err := queries.ListCategoryGroups(ctx, data.ListCategoryGroupsParams{BudgetID: budget.ID, LoginID: budget.LoginID})
			if err != nil {
				return err
			}

			return printNames(stdout, groups, func(g data.CategoryGroup) string { return stringValue(g.Name) })

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("group delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			groupID, err := resolveCategoryGroupID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeleteCategoryGroup(ctx, data.DeleteCategoryGroupParams{
				ID:       groupID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted group %q\n", args[0])
			return nil

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "goal":
		switch action {
		case "create":
			if len(args) != 5 {
				return fmt.Errorf("goal create requires [type] [start] [end|null] [category] [amount]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			goalType := strings.ToLower(args[0])
			if !validGoalType(goalType) {
				return fmt.Errorf("goal create: unknown goal type %q (expected monthly, save, or refill)", args[0])
			}

			start, err := parseMonth("start", args[1])
			if err != nil {
				return err
			}

			end, err := parseNullableMonth("end", args[2])
			if err != nil {
				return err
			}

			if goalType == "save" && !end.Valid {
				return fmt.Errorf("goal create: save goals require an end month")
			}

			if end.Valid && end.Time.Before(start) {
				return fmt.Errorf("goal create: end month must not be before start month")
			}

			category, err := resolveCategoryID(ctx, queries, budget, args[3])
			if err != nil {
				return err
			}

			amount, err := parseAmount(args[4])
			if err != nil {
				return err
			}

			result, err := queries.CreateGoal(ctx, data.CreateGoalParams{
				BudgetID:   budget.ID,
				LoginID:    budget.LoginID,
				Type:       goalType,
				StartDate:  types.UnixTime{Time: start},
				EndDate:    types.NullUnixTime{Time: end.Time, Valid: end.Valid},
				CategoryID: category,
				Amount:     amount,
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "goal", result.ID)
			return nil

		case "update":
			if len(args) != 5 {
				return fmt.Errorf("goal update requires [category] [type] [start] [end|null] [amount]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			category, err := resolveCategoryID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			goalType := strings.ToLower(args[1])
			if !validGoalType(goalType) {
				return fmt.Errorf("goal update: unknown goal type %q (expected monthly, save, or refill)", args[1])
			}

			start, err := parseMonth("start", args[2])
			if err != nil {
				return err
			}

			end, err := parseNullableMonth("end", args[3])
			if err != nil {
				return err
			}

			if goalType == "save" && !end.Valid {
				return fmt.Errorf("goal update: save goals require an end month")
			}

			if end.Valid && end.Time.Before(start) {
				return fmt.Errorf("goal update: end month must not be before start month")
			}

			amount, err := parseAmount(args[4])
			if err != nil {
				return err
			}

			_, err = queries.UpdateGoal(ctx, data.UpdateGoalParams{
				Type:       goalType,
				StartDate:  types.UnixTime{Time: start},
				EndDate:    types.NullUnixTime{Time: end.Time, Valid: end.Valid},
				Amount:     amount,
				BudgetID:   budget.ID,
				LoginID:    budget.LoginID,
				CategoryID: category,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("no goal for category %q in budget %q", args[0], budget.Name)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated goal for category %q\n", args[0])
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			goals, err := queries.ListGoals(ctx, data.ListGoalsParams{BudgetID: budget.ID, LoginID: budget.LoginID})
			if err != nil {
				return err
			}

			return printGoals(stdout, goals)

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("goal delete requires [category]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			categoryID, err := resolveCategoryID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			goal, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
				BudgetID:   budget.ID,
				LoginID:    budget.LoginID,
				CategoryID: categoryID,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("no goal for category %q in budget %q", args[0], budget.Name)
				}
				return err
			}

			if err := queries.DeleteGoal(ctx, data.DeleteGoalParams{
				ID:       goal.ID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted goal for category %q\n", args[0])
			return nil

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "payee":
		switch action {
		case "create":
			if len(args) != 1 {
				return fmt.Errorf("payee create requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			result, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
				Name:     args[0],
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "payee", result.ID)
			return nil

		case "update":
			if len(args) != 2 {
				return fmt.Errorf("payee update requires [name] [new_name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			payeeID, err := resolvePayeeID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			_, err = queries.UpdatePayee(ctx, data.UpdatePayeeParams{
				Name:     args[1],
				ID:       payeeID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unknown payee %q", args[0])
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated payee %q\n", args[1])
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			payees, err := queries.ListPayees(ctx, data.ListPayeesParams{BudgetID: budget.ID, LoginID: budget.LoginID})
			if err != nil {
				return err
			}

			return printNames(stdout, payees, func(p data.Payee) string { return stringValue(p.Name) })

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("payee delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			payeeID, err := resolvePayeeID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeletePayee(ctx, data.DeletePayeeParams{
				ID:       payeeID,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted payee %q\n", args[0])
			return nil

		case "default-category":
			if len(args) < 1 {
				return fmt.Errorf("payee default-category requires a subcommand (create|update|delete)")
			}

			subAction := args[0]
			subArgs := args[1:]
			switch subAction {
			case "create":
				positionals, targetArgs, err := parseCategoryTargetArgs(subArgs)
				if err != nil {
					return err
				}

				if len(positionals) != 2 {
					return fmt.Errorf("payee default-category create requires [payee] [percent] and exactly one of --category or --other-account")
				}

				budget, err := resolveBudget(ctx, queries, login, budgetName)
				if err != nil {
					return err
				}

				payee, err := resolvePayeeID(ctx, queries, budget, positionals[0])
				if err != nil {
					return err
				}

				percent, err := parseInt64("percent", positionals[1])
				if err != nil {
					return err
				}

				target, err := resolveCategoryTargets(ctx, queries, budget, targetArgs)
				if err != nil {
					return err
				}

				result, err := queries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
					BudgetID:      budget.ID,
					LoginID:       budget.LoginID,
					PayeeID:       payee,
					DestAccountID: sql.NullInt64{Int64: target.otherAccount, Valid: target.otherAccount != 0},
					CategoryID:    target.category,
					Income:        target.income,
					Percent:       percent,
				})
				if err != nil {
					return err
				}

				printCreated(stdout, "payee default-category", result.ID)
				return nil

			case "update":
				positionals, targetArgs, err := parseCategoryTargetArgs(subArgs)
				if err != nil {
					return err
				}

				if len(positionals) != 3 {
					return fmt.Errorf("payee default-category update requires [id] [payee] [percent] and exactly one of --category, --other-account, or --income")
				}

				id, err := parseInt64("id", positionals[0])
				if err != nil {
					return err
				}

				budget, err := resolveBudget(ctx, queries, login, budgetName)
				if err != nil {
					return err
				}

				payee, err := resolvePayeeID(ctx, queries, budget, positionals[1])
				if err != nil {
					return err
				}

				percent, err := parseInt64("percent", positionals[2])
				if err != nil {
					return err
				}

				target, err := resolveCategoryTargets(ctx, queries, budget, targetArgs)
				if err != nil {
					return err
				}

				_, err = queries.UpdatePayeeDefaultLine(ctx, data.UpdatePayeeDefaultLineParams{
					PayeeID:       payee,
					DestAccountID: sql.NullInt64{Int64: target.otherAccount, Valid: target.otherAccount != 0},
					CategoryID:    target.category,
					Income:        target.income,
					Percent:       percent,
					ID:            id,
					BudgetID:      budget.ID,
					LoginID:       budget.LoginID,
				})
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("unknown payee default-category %d", id)
				}
				if err != nil {
					return err
				}

				fmt.Fprintf(stdout, "updated payee default-category %d\n", id)
				return nil

			case "delete":
				if len(subArgs) != 1 {
					return fmt.Errorf("payee default-category delete requires [id]")
				}

				id, err := parseInt64("id", subArgs[0])
				if err != nil {
					return err
				}

				budget, err := resolveBudget(ctx, queries, login, budgetName)
				if err != nil {
					return err
				}

				if err := queries.DeletePayeeDefaultLine(ctx, data.DeletePayeeDefaultLineParams{
					ID:       id,
					BudgetID: budget.ID,
					LoginID:  budget.LoginID,
				}); err != nil {
					return err
				}

				fmt.Fprintf(stdout, "deleted payee default-category %d\n", id)
				return nil

			default:
				return fmt.Errorf("unsupported action %q for resource %q", subAction, "payee default-category")
			}

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "transaction":
		switch action {
		case "create":
			if len(args) != 6 {
				return fmt.Errorf("transaction create requires [date] [account] [payee] [total_out] [total_in] [note]")
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			date, err := parseTime("date", args[0])
			if err != nil {
				return err
			}

			account, err := resolveAccountID(ctx, queries, budget, args[1])
			if err != nil {
				return err
			}

			payee, err := resolveOrCreatePayeeID(ctx, queries, budget, args[2])
			if err != nil {
				return err
			}

			totalOutflow, err := parseAmount(args[3])
			if err != nil {
				return fmt.Errorf("parse total_out: %w", err)
			}

			totalInflow, err := parseAmount(args[4])
			if err != nil {
				return fmt.Errorf("parse total_in: %w", err)
			}

			result, err := queries.CreateTrx(ctx, data.CreateTrxParams{
				BudgetID:     budget.ID,
				LoginID:      budget.LoginID,
				Date:         types.UnixTime{Time: date},
				AccountID:    account,
				PayeeID:      payee,
				TotalOutflow: totalOutflow,
				TotalInflow:  totalInflow,
				Note:         args[5],
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "transaction", result.ID)
			return nil

		case "update":
			if len(args) != 7 {
				return fmt.Errorf("transaction update requires [id] [date] [account] [payee] [total_out] [total_in] [note]")
			}

			id, err := parseInt64("id", args[0])
			if err != nil {
				return err
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			date, err := parseTime("date", args[1])
			if err != nil {
				return err
			}

			account, err := resolveAccountID(ctx, queries, budget, args[2])
			if err != nil {
				return err
			}

			payee, err := resolveOrCreatePayeeID(ctx, queries, budget, args[3])
			if err != nil {
				return err
			}

			totalOutflow, err := parseAmount(args[4])
			if err != nil {
				return fmt.Errorf("parse total_out: %w", err)
			}

			totalInflow, err := parseAmount(args[5])
			if err != nil {
				return fmt.Errorf("parse total_in: %w", err)
			}

			_, err = queries.UpdateTrx(ctx, data.UpdateTrxParams{
				Date:         types.UnixTime{Time: date},
				AccountID:    account,
				PayeeID:      payee,
				TotalOutflow: totalOutflow,
				TotalInflow:  totalInflow,
				Note:         args[6],
				ID:           id,
				BudgetID:     budget.ID,
				LoginID:      budget.LoginID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("unknown transaction %d", id)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(stdout, "updated transaction %d\n", id)
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			transactions, err := queries.ListTrxs(ctx, data.ListTrxsParams{BudgetID: budget.ID, LoginID: budget.LoginID})
			if err != nil {
				return err
			}

			return printTransactionsByBudget(stdout, transactions)

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("transaction delete requires [id]")
			}

			id, err := parseInt64("id", args[0])
			if err != nil {
				return err
			}

			budget, err := resolveBudget(ctx, queries, login, budgetName)
			if err != nil {
				return err
			}

			if err := queries.DeleteTrx(ctx, data.DeleteTrxParams{
				ID:       id,
				BudgetID: budget.ID,
				LoginID:  budget.LoginID,
			}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted transaction %d\n", id)
			return nil

		case "category":
			if len(args) < 1 {
				return fmt.Errorf("transaction category requires a subcommand (create|update|delete)")
			}

			subAction := args[0]
			subArgs := args[1:]
			switch subAction {
			case "create":
				positionals, targetArgs, err := parseCategoryTargetArgs(subArgs)
				if err != nil {
					return err
				}

				if len(positionals) != 3 {
					return fmt.Errorf("transaction category create requires [transaction] [outflow] [inflow] and exactly one of --category or --other-account")
				}

				budget, err := resolveBudget(ctx, queries, login, budgetName)
				if err != nil {
					return err
				}

				transactionID, err := parseInt64("transaction", positionals[0])
				if err != nil {
					return err
				}

				outflow, err := parseAmount(positionals[1])
				if err != nil {
					return fmt.Errorf("parse outflow: %w", err)
				}

				inflow, err := parseAmount(positionals[2])
				if err != nil {
					return fmt.Errorf("parse inflow: %w", err)
				}

				target, err := resolveCategoryTargets(ctx, queries, budget, targetArgs)
				if err != nil {
					return err
				}

				result, err := queries.CreateTrxLine(ctx, data.CreateTrxLineParams{
					BudgetID:      budget.ID,
					LoginID:       budget.LoginID,
					TrxID:         transactionID,
					DestAccountID: sql.NullInt64{Int64: target.otherAccount, Valid: target.otherAccount != 0},
					CategoryID:    target.category,
					Income:        target.income,
					Outflow:       outflow,
					Inflow:        inflow,
				})
				if err != nil {
					return err
				}

				printCreated(stdout, "transaction category", result.ID)
				return nil

			case "update":
				positionals, targetArgs, err := parseCategoryTargetArgs(subArgs)
				if err != nil {
					return err
				}

				if len(positionals) != 4 {
					return fmt.Errorf("transaction category update requires [id] [transaction] [outflow] [inflow] and exactly one of --category, --other-account, or --income")
				}

				id, err := parseInt64("id", positionals[0])
				if err != nil {
					return err
				}

				budget, err := resolveBudget(ctx, queries, login, budgetName)
				if err != nil {
					return err
				}

				transactionID, err := parseInt64("transaction", positionals[1])
				if err != nil {
					return err
				}

				outflow, err := parseAmount(positionals[2])
				if err != nil {
					return fmt.Errorf("parse outflow: %w", err)
				}

				inflow, err := parseAmount(positionals[3])
				if err != nil {
					return fmt.Errorf("parse inflow: %w", err)
				}

				target, err := resolveCategoryTargets(ctx, queries, budget, targetArgs)
				if err != nil {
					return err
				}

				_, err = queries.UpdateTrxLine(ctx, data.UpdateTrxLineParams{
					TrxID:         transactionID,
					DestAccountID: sql.NullInt64{Int64: target.otherAccount, Valid: target.otherAccount != 0},
					CategoryID:    target.category,
					Income:        target.income,
					Outflow:       outflow,
					Inflow:        inflow,
					ID:            id,
					BudgetID:      budget.ID,
					LoginID:       budget.LoginID,
				})
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("unknown transaction category %d", id)
				}
				if err != nil {
					return err
				}

				fmt.Fprintf(stdout, "updated transaction category %d\n", id)
				return nil

			case "delete":
				if len(subArgs) != 1 {
					return fmt.Errorf("transaction category delete requires [id]")
				}

				id, err := parseInt64("id", subArgs[0])
				if err != nil {
					return err
				}

				budget, err := resolveBudget(ctx, queries, login, budgetName)
				if err != nil {
					return err
				}

				if err := queries.DeleteTrxLine(ctx, data.DeleteTrxLineParams{
					ID:       id,
					BudgetID: budget.ID,
					LoginID:  budget.LoginID,
				}); err != nil {
					return err
				}

				fmt.Fprintf(stdout, "deleted transaction category %d\n", id)
				return nil

			default:
				return fmt.Errorf("unsupported action %q for resource %q", subAction, "transaction category")
			}

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	default:
		return fmt.Errorf("unsupported resource %q", resource)
	}
}

func parseInt64(name, value string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}

	return parsed, nil
}

func parseTime(name, value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("parse %s: expected RFC3339 or YYYY-MM-DD", name)
}

func parseMonth(name, value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s: expected YYYY-MM", name)
	}

	return parsed, nil
}

func parseNullableMonth(name, value string) (sql.NullTime, error) {
	if strings.EqualFold(value, "null") {
		return sql.NullTime{}, nil
	}

	parsed, err := parseMonth(name, value)
	if err != nil {
		return sql.NullTime{}, err
	}

	return sql.NullTime{Time: parsed, Valid: true}, nil
}

func parseCategoryTargetArgs(args []string) ([]string, categoryTargetArgs, error) {
	positionals := make([]string, 0, len(args))
	targets := categoryTargetArgs{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			positionals = append(positionals, arg)
			continue
		}

		switch arg {
		case "--other-account":
			if i+1 >= len(args) {
				return nil, categoryTargetArgs{}, fmt.Errorf("missing value for %s", arg)
			}
			if targets.OtherAccount != "" {
				return nil, categoryTargetArgs{}, fmt.Errorf("--other-account may only be set once")
			}
			i++
			targets.OtherAccount = args[i]
		case "--category":
			if i+1 >= len(args) {
				return nil, categoryTargetArgs{}, fmt.Errorf("missing value for %s", arg)
			}
			if targets.Category != "" {
				return nil, categoryTargetArgs{}, fmt.Errorf("--category may only be set once")
			}
			i++
			targets.Category = args[i]
		case "--income":
			if targets.Income {
				return nil, categoryTargetArgs{}, fmt.Errorf("--income may only be set once")
			}
			targets.Income = true
		default:
			return nil, categoryTargetArgs{}, fmt.Errorf("unsupported flag %q", arg)
		}
	}

	targetCount := 0
	if targets.OtherAccount != "" {
		targetCount++
	}
	if targets.Category != "" {
		targetCount++
	}
	if targets.Income {
		targetCount++
	}

	if targetCount != 1 {
		return nil, categoryTargetArgs{}, fmt.Errorf("set exactly one of --category, --other-account, or --income")
	}

	return positionals, targets, nil
}

func resolveBudget(ctx context.Context, queries *data.Queries, login loginContext, budgetName string) (budgetContext, error) {
	if budgetName != "" {
		budget, err := queries.GetBudgetByName(ctx, data.GetBudgetByNameParams{
			LoginID: login.ID,
			Name:    budgetName,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return budgetContext{}, fmt.Errorf("unknown budget %q", budgetName)
			}

			return budgetContext{}, err
		}

		return budgetContext{ID: budget.ID, LoginID: login.ID, Name: stringValue(budget.Name)}, nil
	}

	budgets, err := queries.ListBudgets(ctx, data.ListBudgetsParams{LoginID: login.ID})
	if err != nil {
		return budgetContext{}, err
	}

	switch len(budgets) {
	case 0:
		return budgetContext{}, fmt.Errorf("no budgets exist; create one with `ynafb budget create [name]`")
	case 1:
		return budgetContext{ID: budgets[0].ID, LoginID: login.ID, Name: stringValue(budgets[0].Name)}, nil
	default:
		return budgetContext{}, fmt.Errorf("multiple budgets exist; pass --budget [name]")
	}
}

func resolveAccountID(ctx context.Context, queries *data.Queries, budget budgetContext, accountName string) (int64, error) {
	account, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     accountName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("unknown account %q in budget %q", accountName, budget.Name)
		}

		return 0, err
	}

	return account.ID, nil
}

func resolveCategoryID(ctx context.Context, queries *data.Queries, budget budgetContext, categoryName string) (int64, error) {
	category, err := queries.GetCategoryByName(ctx, data.GetCategoryByNameParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     categoryName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("unknown category %q in budget %q", categoryName, budget.Name)
		}

		return 0, err
	}

	return category.ID, nil
}

func resolveCategoryGroupID(ctx context.Context, queries *data.Queries, budget budgetContext, groupName string) (int64, error) {
	group, err := queries.GetCategoryGroupByName(ctx, data.GetCategoryGroupByNameParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     groupName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("unknown group %q in budget %q", groupName, budget.Name)
		}

		return 0, err
	}

	return group.ID, nil
}

func resolvePayeeID(ctx context.Context, queries *data.Queries, budget budgetContext, payeeName string) (int64, error) {
	payee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     payeeName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("unknown payee %q in budget %q", payeeName, budget.Name)
		}

		return 0, err
	}

	return payee.ID, nil
}

func resolveOrCreatePayeeID(ctx context.Context, queries *data.Queries, budget budgetContext, payeeName string) (int64, error) {
	payee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     payeeName,
	})
	if err == nil {
		return payee.ID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	created, createErr := queries.CreatePayee(ctx, data.CreatePayeeParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		Name:     payeeName,
	})
	if createErr != nil {
		return 0, createErr
	}

	return created.ID, nil
}

func importAccountTransactions(ctx context.Context, db *sql.DB, queries *data.Queries, budget budgetContext, accountID int64, path string, stdout io.Writer) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	stmt, err := importer.Parse(contents)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQueries := queries.WithTx(tx)
	for _, entry := range stmt.Entries {
		payeeName := strings.TrimSpace(entry.Payee)
		if payeeName == "" {
			payeeName = "unknown"
		}

		payeeID, err := resolveOrCreatePayeeID(ctx, txQueries, budget, payeeName)
		if err != nil {
			return err
		}

		_, err = txQueries.CreateTrx(ctx, data.CreateTrxParams{
			BudgetID:     budget.ID,
			LoginID:      budget.LoginID,
			Date:         types.UnixTime{Time: entry.TransDate},
			AccountID:    accountID,
			PayeeID:      payeeID,
			TotalOutflow: entry.Outflow,
			TotalInflow:  entry.Inflow,
			Note:         entry.Note,
		})
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "imported %d transactions\n", len(stmt.Entries))
	return nil
}

func reconcileAccount(ctx context.Context, queries *data.Queries, budget budgetContext, accountID int64, accountName string, date time.Time, stdin io.Reader, stdout io.Writer) error {
	balance, err := queries.GetAccountBalanceAsOf(ctx, data.GetAccountBalanceAsOfParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		ID:       accountID,
		Date:     types.UnixTime{Time: date},
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(stdout, "Balance as of %s: %s\n", date.Format("2006-01-02"), formatCents(balance)); err != nil {
		return err
	}

	reader := bufio.NewReader(stdin)
	ans, err := prompt(reader, stdout, "Statement balance: ")
	if err != nil {
		return err
	}
	if ans == "" {
		if _, err := fmt.Fprintln(stdout, "cancelled"); err != nil {
			return err
		}
		return nil
	}

	statement, err := parseAmount(ans)
	if err != nil {
		return err
	}

	if statement != balance {
		return fmt.Errorf("balance mismatch: statement %s != calculated %s", formatCents(statement), formatCents(balance))
	}

	n, err := queries.ReconcileAccountTransactions(ctx, data.ReconcileAccountTransactionsParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		ID:       accountID,
		Date:     types.UnixTime{Time: date},
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "reconciled %d transactions through %s in %q\n", n, date.Format("2006-01-02"), accountName)
	return nil
}

type categorizeTransaction struct {
	ID                     int64
	Date                   time.Time
	TransactionAccount     int64
	TransactionAccountName string
	PayeeName              string
	TotalOutflow           int64
	TotalInflow            int64
	Reconciled             bool
	Note                   string
	rows                   []data.ListAccountTransactionsRow
}

type categorizeRow struct {
	id         int64
	transfer   bool
	income     bool
	targetID   int64
	targetName string
	outflow    int64
	inflow     int64
}

func categorizeAccount(ctx context.Context, db *sql.DB, queries *data.Queries, budget budgetContext, accountID int64, accountName string, stdin io.Reader, stdout io.Writer) error {
	transactions, err := queries.ListAccountTransactions(ctx, data.ListAccountTransactionsParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		ID:       accountID,
	})
	if err != nil {
		return err
	}

	queue := categorizeQueue(accountID, transactions)
	if len(queue) == 0 {
		fmt.Fprintln(stdout, "No transactions need categorization")
		return nil
	}

	reader := bufio.NewReader(stdin)
	fmt.Fprintf(stdout, "Categorizing %d transactions in %q\n", len(queue), accountName)

	var categorized, skipped int
	for i, tx := range queue {
		if _, err := fmt.Fprintln(stdout, "────────────────────────────────────────"); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "[%d/%d]\n", i+1, len(queue))

		working := loadCategorizations(tx.rows)
		prefilled := false
		if len(working) == 0 {
			working, prefilled = prefillFromDefaults(ctx, queries, budget, tx)
		}

	transactionLoop:
		for {
			if err := printCategorizeTransaction(stdout, accountID, tx, working); err != nil {
				return err
			}
			remainingOut, remainingIn := categorizeRemaining(tx, working)
			if _, err := fmt.Fprintf(stdout, "Remaining: %s out / %s in\n", formatCents(remainingOut), formatCents(remainingIn)); err != nil {
				return err
			}
			if prefilled {
				if _, err := fmt.Fprintln(stdout, "(pre-filled from payee default)"); err != nil {
					return err
				}
			}

			cmd, err := prompt(reader, stdout, "(a)dd  (d)elete  (s)kip  (o)k  (q)uit > ")
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Fprintln(stdout, "→ quit")
					return nil
				}
				return err
			}

			switch cmd {
			case "a", "add":
				if err := categorizeAdd(ctx, queries, budget, reader, stdout, tx, &working); err != nil {
					return err
				}
			case "d", "delete":
				if err := categorizeDelete(reader, stdout, &working); err != nil {
					return err
				}
			case "s", "skip":
				if _, err := fmt.Fprintln(stdout, "→ skipped"); err != nil {
					return err
				}
				skipped++
				break transactionLoop
			case "o", "ok":
				remainingOut, remainingIn := categorizeRemaining(tx, working)
				if remainingOut != 0 || remainingIn != 0 {
					if _, err := fmt.Fprintf(stdout, "totals don't match: remaining %s out / %s in\n", formatCents(remainingOut), formatCents(remainingIn)); err != nil {
						return err
					}
					continue
				}
				if err := replaceTransactionCategories(ctx, db, queries, budget, tx.ID, working); err != nil {
					return err
				}
				categorized++
				if err := categorizeSaveDefault(ctx, db, queries, budget, reader, stdout, tx, working); err != nil {
					return err
				}
				break transactionLoop
			case "q", "quit":
				if _, err := fmt.Fprintf(stdout, "→ quit (%d categorized, %d skipped)\n", categorized, skipped); err != nil {
					return err
				}
				return nil
			default:
				if _, err := fmt.Fprintln(stdout, "invalid option"); err != nil {
					return err
				}
			}
		}
	}

	fmt.Fprintf(stdout, "\nDone — %d categorized, %d skipped\n", categorized, skipped)
	return nil
}

func categorizeQueue(accountID int64, transactions []data.ListAccountTransactionsRow) []categorizeTransaction {
	var queue []categorizeTransaction
	for i := 0; i < len(transactions); {
		j := i + 1
		for j < len(transactions) && transactions[j].TrxID == transactions[i].TrxID {
			j++
		}

		rows := transactions[i:j]
		if categorizeNeedsAttention(accountID, rows) {
			t := rows[0]
			queue = append(queue, categorizeTransaction{
				ID:                     t.TrxID,
				Date:                   t.Date.Time,
				TransactionAccount:     t.AccountID,
				TransactionAccountName: stringValue(t.AccountName),
				PayeeName:              stringValue(t.PayeeName),
				TotalOutflow:           t.Outflow,
				TotalInflow:            t.Inflow,
				Reconciled:             t.Reconciled,
				Note:                   t.Note,
				rows:                   rows,
			})
		}

		i = j
	}
	return queue
}

func categorizeNeedsAttention(accountID int64, rows []data.ListAccountTransactionsRow) bool {
	if rows[0].AccountID != accountID {
		return false
	}

	// Incoming mirrored transfers are already categorized on the source side.
	if rows[0].SourceAccountID.Valid {
		return false
	}

	var out, in int64
	for _, r := range rows {
		if r.CategoryID.Valid || r.DestAccountID.Valid || r.Income {
			out += nullableInt64Value(r.LineOutflow)
			in += nullableInt64Value(r.LineInflow)
		}
	}

	return out != rows[0].Outflow || in != rows[0].Inflow
}

func loadCategorizations(rows []data.ListAccountTransactionsRow) []categorizeRow {
	var working []categorizeRow
	for _, r := range rows {
		if !r.CategoryID.Valid && !r.DestAccountID.Valid && !r.Income {
			continue
		}

		row := categorizeRow{
			id:      r.CategoryID.Int64,
			outflow: nullableInt64Value(r.LineOutflow),
			inflow:  nullableInt64Value(r.LineInflow),
		}
		if r.SourceAccountID.Valid {
			row.transfer = true
			row.targetID = r.SourceAccountID.Int64
			row.targetName = stringValue(r.SourceAccountName)
		} else if r.DestAccountID.Valid {
			row.transfer = true
			row.targetID = r.DestAccountID.Int64
			row.targetName = stringValue(r.DestAccountName)
		} else if r.Income {
			row.income = true
			row.targetName = "Income"
		} else {
			row.transfer = false
			row.targetID = r.CategoryID.Int64
			row.targetName = stringValue(r.CategoryName)
		}
		working = append(working, row)
	}
	return working
}

func prefillFromDefaults(ctx context.Context, queries *data.Queries, budget budgetContext, tx categorizeTransaction) ([]categorizeRow, bool) {
	payeeID, err := resolvePayeeID(ctx, queries, budget, tx.PayeeName)
	if err != nil {
		return nil, false
	}

	defaults, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{
		PayeeID:  payeeID,
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
	})
	if err != nil || len(defaults) == 0 {
		return nil, false
	}

	total := tx.TotalOutflow
	useOutflow := true
	if total == 0 {
		total = tx.TotalInflow
		useOutflow = false
	}
	if total <= 0 {
		return nil, false
	}

	amounts := defaultAmounts(defaults, total)
	working := make([]categorizeRow, 0, len(defaults))
	for i, d := range defaults {
		row := categorizeRow{}
		if useOutflow {
			row.outflow = amounts[i]
		} else {
			row.inflow = amounts[i]
		}
		if d.DestAccountID.Valid {
			row.transfer = true
			row.targetID = d.DestAccountID.Int64
			row.targetName = stringValue(d.DestAccountName)
		} else if d.Income {
			row.income = true
			row.targetName = "Income"
		} else {
			row.transfer = false
			row.targetID = d.CategoryID.Int64
			row.targetName = stringValue(d.CategoryName)
		}
		working = append(working, row)
	}
	return working, true
}

func defaultAmounts(defaults []data.ListPayeeDefaultLinesByPayeeRow, total int64) []int64 {
	amounts := make([]int64, len(defaults))
	var allocated int64
	for i, d := range defaults {
		if i == len(defaults)-1 {
			amounts[i] = total - allocated
		} else {
			amounts[i] = total * d.Percent / 100
			allocated += amounts[i]
		}
	}
	return amounts
}

func categorizeRemaining(tx categorizeTransaction, working []categorizeRow) (int64, int64) {
	var out, in int64
	for _, row := range working {
		out += row.outflow
		in += row.inflow
	}
	return tx.TotalOutflow - out, tx.TotalInflow - in
}

func printCategorizeTransaction(stdout io.Writer, accountID int64, tx categorizeTransaction, working []categorizeRow) error {
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tDATE\tPAYEE\tTARGET\tOUTFLOW\tINFLOW\tRECONCILED\tNOTE"); err != nil {
		return err
	}
	if err := printTransaction(tw, accountID, synthesizeRows(tx, working)); err != nil {
		return err
	}
	return tw.Flush()
}

func synthesizeRows(tx categorizeTransaction, working []categorizeRow) []data.ListAccountTransactionsRow {
	base := data.ListAccountTransactionsRow{
		TrxID:       tx.ID,
		Date:        types.UnixTime{Time: tx.Date},
		AccountID:   tx.TransactionAccount,
		AccountName: sql.NullString{String: tx.TransactionAccountName, Valid: true},
		PayeeName:   sql.NullString{String: tx.PayeeName, Valid: true},
		Outflow:     tx.TotalOutflow,
		Inflow:      tx.TotalInflow,
		Reconciled:  tx.Reconciled,
		Note:        tx.Note,
	}

	if len(working) == 0 {
		return []data.ListAccountTransactionsRow{base}
	}

	rows := make([]data.ListAccountTransactionsRow, 0, len(working))
	for i, row := range working {
		r := base
		r.CategoryID = sql.NullInt64{Int64: int64(i + 1), Valid: true}
		r.LineOutflow = sql.NullInt64{Int64: row.outflow, Valid: true}
		r.LineInflow = sql.NullInt64{Int64: row.inflow, Valid: true}
		if row.transfer {
			r.DestAccountID = sql.NullInt64{Int64: row.targetID, Valid: true}
			r.DestAccountName = sql.NullString{String: row.targetName, Valid: true}
		} else if row.income {
			r.Income = true
		} else {
			r.CategoryID = sql.NullInt64{Int64: row.targetID, Valid: true}
			r.CategoryName = sql.NullString{String: row.targetName, Valid: true}
		}
		rows = append(rows, r)
	}
	return rows
}

func categorizeAdd(ctx context.Context, queries *data.Queries, budget budgetContext, reader *bufio.Reader, stdout io.Writer, tx categorizeTransaction, working *[]categorizeRow) error {
	target, err := prompt(reader, stdout, "  Target (category, @account for a transfer, or income): ")
	if err != nil {
		return err
	}

	row := categorizeRow{}
	if strings.HasPrefix(target, "@") {
		name := strings.TrimPrefix(target, "@")
		id, err := resolveAccountID(ctx, queries, budget, name)
		if err != nil {
			fmt.Fprintf(stdout, "  %v\n", err)
			return nil
		}
		row.transfer = true
		row.targetID = id
		row.targetName = name
	} else if strings.EqualFold(target, "income") {
		row.income = true
		row.targetName = "Income"
	} else {
		if target == "" {
			fmt.Fprintln(stdout, "  target required")
			return nil
		}
		id, err := resolveOrCreateCategoryID(ctx, queries, budget, target, reader, stdout)
		if err != nil {
			fmt.Fprintf(stdout, "  %v\n", err)
			return nil
		}
		row.transfer = false
		row.targetID = id
		row.targetName = target
	}

	remainingOut, remainingIn := categorizeRemaining(tx, *working)
	outflow, err := promptAmount(reader, stdout, fmt.Sprintf("  Outflow [%s]: ", formatCents(remainingOut)), remainingOut)
	if err != nil {
		return err
	}
	inflow, err := promptAmount(reader, stdout, fmt.Sprintf("  Inflow [%s]: ", formatCents(remainingIn)), remainingIn)
	if err != nil {
		return err
	}
	row.outflow = outflow
	row.inflow = inflow
	*working = append(*working, row)

	fmt.Fprintf(stdout, "  → added: %s  %s out / %s in\n", categorizeTargetLabel(row), formatCents(row.outflow), formatCents(row.inflow))
	return nil
}

func resolveOrCreateCategoryID(ctx context.Context, queries *data.Queries, budget budgetContext, name string, reader *bufio.Reader, stdout io.Writer) (int64, error) {
	category, err := queries.GetCategoryByName(ctx, data.GetCategoryByNameParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     name,
	})
	if err == nil {
		return category.ID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if strings.EqualFold(name, "income") {
		return 0, fmt.Errorf("category name %q is reserved", name)
	}

	ans, err := prompt(reader, stdout, fmt.Sprintf("  Category %q doesn't exist. Create it? [y/n]: ", name))
	if err != nil {
		return 0, err
	}
	if !strings.EqualFold(ans, "y") {
		return 0, fmt.Errorf("category %q not created", name)
	}

	created, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
		BudgetID:        budget.ID,
		LoginID:         budget.LoginID,
		Name:            name,
		CategoryGroupID: sql.NullInt64{},
	})
	if err != nil {
		return 0, err
	}

	fmt.Fprintf(stdout, "  → created category %q\n", name)
	return created.ID, nil
}

func resolveOrCreateCategoryGroup(ctx context.Context, queries *data.Queries, budget budgetContext, name string) (int64, error) {
	group, err := queries.GetCategoryGroupByName(ctx, data.GetCategoryGroupByNameParams{
		LoginID:  budget.LoginID,
		BudgetID: budget.ID,
		Name:     name,
	})
	if err == nil {
		return group.ID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	created, err := queries.CreateCategoryGroup(ctx, data.CreateCategoryGroupParams{
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
		Name:     name,
	})
	if err != nil {
		return 0, err
	}
	return created, nil
}

func categorizeDelete(reader *bufio.Reader, stdout io.Writer, working *[]categorizeRow) error {
	if len(*working) == 0 {
		fmt.Fprintln(stdout, "  no categorizations to delete")
		return nil
	}

	for i, row := range *working {
		fmt.Fprintf(stdout, "  %d. %s  %s out / %s in\n", i+1, categorizeTargetLabel(row), formatCents(row.outflow), formatCents(row.inflow))
	}

	ans, err := prompt(reader, stdout, fmt.Sprintf("  Delete which? [1-%d]: ", len(*working)))
	if err != nil {
		return err
	}
	n, err := strconv.Atoi(ans)
	if err != nil || n < 1 || n > len(*working) {
		fmt.Fprintln(stdout, "  invalid selection")
		return nil
	}

	removed := (*working)[n-1]
	*working = append((*working)[:n-1], (*working)[n:]...)
	fmt.Fprintf(stdout, "  → removed: %s\n", categorizeTargetLabel(removed))
	return nil
}

func categorizeSaveDefault(ctx context.Context, db *sql.DB, queries *data.Queries, budget budgetContext, reader *bufio.Reader, stdout io.Writer, tx categorizeTransaction, working []categorizeRow) error {
	if len(working) == 0 {
		fmt.Fprintln(stdout, "→ categorized")
		return nil
	}

	payeeID, err := resolvePayeeID(ctx, queries, budget, tx.PayeeName)
	if err != nil {
		fmt.Fprintln(stdout, "→ categorized")
		return nil
	}

	total := tx.TotalOutflow
	if total == 0 {
		total = tx.TotalInflow
	}
	if total <= 0 {
		fmt.Fprintln(stdout, "→ categorized")
		return nil
	}

	percents := categorizePercents(tx, working)

	defaults, err := queries.ListPayeeDefaultLinesByPayee(ctx, data.ListPayeeDefaultLinesByPayeeParams{
		PayeeID:  payeeID,
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
	})
	if err != nil {
		return err
	}

	if defaultsMatch(tx, defaults, working) {
		fmt.Fprintln(stdout, "→ categorized (matches existing default)")
		return nil
	}

	ans, err := prompt(reader, stdout, fmt.Sprintf("  Save as default categorization for %q? [y/n]: ", tx.PayeeName))
	if err != nil {
		return err
	}
	if !strings.EqualFold(ans, "y") {
		fmt.Fprintln(stdout, "→ categorized")
		return nil
	}

	if err := replacePayeeDefaults(ctx, db, queries, budget, payeeID, working, percents); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "→ saved default for %q\n", tx.PayeeName)
	return nil
}

func categorizePercents(tx categorizeTransaction, working []categorizeRow) map[string]int64 {
	total := tx.TotalOutflow
	useOutflow := true
	if total == 0 {
		total = tx.TotalInflow
		useOutflow = false
	}

	percents := make(map[string]int64, len(working))
	var allocated int64
	for i, row := range working {
		var amount int64
		if useOutflow {
			amount = row.outflow
		} else {
			amount = row.inflow
		}

		if i == len(working)-1 {
			percents[categorizeKey(row)] = 100 - allocated
		} else {
			p := amount * 100 / total
			allocated += p
			percents[categorizeKey(row)] = p
		}
	}
	return percents
}

func categorizeTargetLabel(row categorizeRow) string {
	if row.transfer {
		return "@" + row.targetName
	}
	if row.income {
		return "Income"
	}
	return row.targetName
}

func categorizeKey(row categorizeRow) string {
	if row.transfer {
		return fmt.Sprintf("account:%d", row.targetID)
	}
	if row.income {
		return "income"
	}
	return fmt.Sprintf("category:%d", row.targetID)
}

func defaultsMatch(tx categorizeTransaction, defaults []data.ListPayeeDefaultLinesByPayeeRow, working []categorizeRow) bool {
	if len(defaults) != len(working) {
		return false
	}

	total := tx.TotalOutflow
	useOutflow := true
	if total == 0 {
		total = tx.TotalInflow
		useOutflow = false
	}
	if total <= 0 {
		return false
	}

	amounts := defaultAmounts(defaults, total)

	expected := make(map[string]int64, len(defaults))
	for i, d := range defaults {
		var key string
		if d.DestAccountID.Valid {
			key = fmt.Sprintf("account:%d", d.DestAccountID.Int64)
		} else if d.Income {
			key = "income"
		} else {
			key = fmt.Sprintf("category:%d", d.CategoryID.Int64)
		}
		expected[key] = amounts[i]
	}

	actual := make(map[string]int64, len(working))
	for _, row := range working {
		amount := row.outflow
		if !useOutflow {
			amount = row.inflow
		}
		actual[categorizeKey(row)] = amount
	}

	for key, amount := range expected {
		if actual[key] != amount {
			return false
		}
	}
	return true
}

func replaceTransactionCategories(ctx context.Context, db *sql.DB, queries *data.Queries, budget budgetContext, transactionID int64, working []categorizeRow) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQueries := queries.WithTx(tx)
	if err := txQueries.DeleteTrxLinesByTrx(ctx, data.DeleteTrxLinesByTrxParams{
		TrxID:    transactionID,
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
	}); err != nil {
		return err
	}
	for _, row := range working {
		if _, err := txQueries.CreateTrxLine(ctx, data.CreateTrxLineParams{
			BudgetID:      budget.ID,
			LoginID:       budget.LoginID,
			TrxID:         transactionID,
			DestAccountID: categorizeOtherAccount(row),
			CategoryID:    categorizeCategory(row),
			Income:        row.income,
			Outflow:       row.outflow,
			Inflow:        row.inflow,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func replacePayeeDefaults(ctx context.Context, db *sql.DB, queries *data.Queries, budget budgetContext, payeeID int64, working []categorizeRow, percents map[string]int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQueries := queries.WithTx(tx)
	if err := txQueries.DeletePayeeDefaultLinesByPayee(ctx, data.DeletePayeeDefaultLinesByPayeeParams{
		PayeeID:  payeeID,
		BudgetID: budget.ID,
		LoginID:  budget.LoginID,
	}); err != nil {
		return err
	}
	for _, row := range working {
		if _, err := txQueries.CreatePayeeDefaultLine(ctx, data.CreatePayeeDefaultLineParams{
			BudgetID:      budget.ID,
			LoginID:       budget.LoginID,
			PayeeID:       payeeID,
			DestAccountID: categorizeOtherAccount(row),
			CategoryID:    categorizeCategory(row),
			Income:        row.income,
			Percent:       percents[categorizeKey(row)],
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func categorizeOtherAccount(row categorizeRow) sql.NullInt64 {
	if row.transfer {
		return sql.NullInt64{Int64: row.targetID, Valid: true}
	}
	return sql.NullInt64{}
}

func categorizeCategory(row categorizeRow) sql.NullInt64 {
	if row.transfer || row.income {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: row.targetID, Valid: true}
}

func prompt(reader *bufio.Reader, stdout io.Writer, msg string) (string, error) {
	if _, err := fmt.Fprint(stdout, msg); err != nil {
		return "", err
	}
	line, err := reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && line != "") {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptAmount(reader *bufio.Reader, stdout io.Writer, msg string, defaultVal int64) (int64, error) {
	s, err := prompt(reader, stdout, msg)
	if err != nil {
		return 0, err
	}
	if s == "" {
		return defaultVal, nil
	}
	return parseAmount(s)
}

func parseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}

	parts := strings.SplitN(s, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q", s)
	}

	fraction := int64(0)
	if len(parts) == 2 {
		frac := parts[1]
		if len(frac) > 2 {
			frac = frac[:2]
		}
		for len(frac) < 2 {
			frac += "0"
		}
		f, err := strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid amount %q", s)
		}
		fraction = f
	}

	cents := whole*100 + fraction
	if negative {
		cents = -cents
	}
	return cents, nil
}

type categoryTargetResolved struct {
	otherAccount int64
	category     sql.NullInt64
	income       bool
}

func resolveCategoryTargets(ctx context.Context, queries *data.Queries, budget budgetContext, args categoryTargetArgs) (categoryTargetResolved, error) {
	if args.Income {
		return categoryTargetResolved{income: true}, nil
	}

	if args.Category != "" {
		categoryID, err := resolveCategoryID(ctx, queries, budget, args.Category)
		if err != nil {
			return categoryTargetResolved{}, err
		}

		return categoryTargetResolved{category: sql.NullInt64{Int64: categoryID, Valid: true}}, nil
	}

	otherAccountID, err := resolveAccountID(ctx, queries, budget, args.OtherAccount)
	if err != nil {
		return categoryTargetResolved{}, err
	}

	return categoryTargetResolved{otherAccount: otherAccountID}, nil
}

func printCreated(w io.Writer, resource string, id int64) {
	fmt.Fprintf(w, "created %s %d\n", resource, id)
}

func printLogins(w io.Writer, logins []data.Login) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "USERNAME"); err != nil {
		return err
	}
	for _, l := range logins {
		if _, err := fmt.Fprintf(tw, "%s\n", l.Username); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printBudgetMonthSummary(w io.Writer, summary data.GetBudgetMonthSummaryRow, goalsTotal int64) error {
	if _, err := fmt.Fprintf(w, "Ready to assign: %s\n", formatCents(summary.ReadyToAssign)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Income: %s\n", formatCents(summary.Income)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Goals: %s\n", formatCents(goalsTotal)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Allocated: %s\n", formatCents(summary.Allocated)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Spent: %s\n", formatCents(summary.Spent)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Available: %s\n", formatCents(summary.Available)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Uncategorized: %s\n", formatCents(summary.Uncategorized)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

const ungroupedLabel = "No group"

func printBudgetMonthCategories(w io.Writer, rows []data.ListBudgetMonthCategoriesRow, goals []data.ListGoalsValuesRow) error {
	goalsByCategory := make(map[int64]data.ListGoalsValuesRow)
	for _, g := range goals {
		goalsByCategory[g.CategoryID] = g
	}

	group := "No group"
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "CATEGORY\tGOAL\tALLOCATED\tSPENT\tAVAILABLE\tWARNING"); err != nil {
		return err
	}
	count := 0
	for _, row := range rows {
		if row.CategoryGroupName.Valid && row.CategoryGroupName.String != group {
			if count > 0 {
				if _, err := fmt.Fprintf(w, "%s:\n", group); err != nil {
					return err
				}
				if err := tw.Flush(); err != nil {
					return err
				}
				var buf bytes.Buffer
				tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintln(tw, "CATEGORY\tGOAL\tALLOCATED\tSPENT\tAVAILABLE\tWARNING"); err != nil {
					return err
				}
				for _, line := range strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n") {
					if _, err := fmt.Fprintln(w, strings.TrimRight(line, " ")); err != nil {
						return err
					}
				}
			}
			group = row.CategoryGroupName.String
		}
		goal, ok := goalsByCategory[row.CategoryID]
		amountForMonth := ""
		warning := ""
		if ok {
			amountForMonth = formatCents(goal.AmountForMonth)
			warning = goalWarning(goal)
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			stringValue(row.CategoryName),
			amountForMonth,
			formatCents(row.Allocated),
			formatCents(row.Spent),
			formatCents(row.Available),
			warning); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "%s:\n", group); err != nil {
		return err
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	for _, line := range strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n") {
		if _, err := fmt.Fprintln(w, strings.TrimRight(line, " ")); err != nil {
			return err
		}
	}
	return nil
}

func printBudgets(w io.Writer, budgets []data.Budget) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME"); err != nil {
		return err
	}
	for _, b := range budgets {
		if _, err := fmt.Fprintf(tw, "%s\n", stringValue(b.Name)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printNames[T any](w io.Writer, rows []T, name func(T) string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\n", name(row)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printAccountBalances(w io.Writer, accounts []data.ListAccountsBalancesRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tBALANCE\tRECONCILED BALANCE"); err != nil {
		return err
	}
	for _, a := range accounts {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", stringValue(a.Name), formatCents(a.Balance), formatCents(a.ReconciledBalance)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printGoals(w io.Writer, goals []data.ListGoalsRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "TYPE\tCATEGORY\tSTART\tEND\tAMOUNT"); err != nil {
		return err
	}
	for _, g := range goals {
		end := ""
		if g.EndDate.Valid {
			end = g.EndDate.Time.Format("2006-01")
		}
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\n",
			stringValue(g.Type),
			stringValue(g.CategoryName),
			g.StartDate.Format("2006-01"),
			end,
			formatCents(g.Amount),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func validGoalType(t string) bool {
	return t == "monthly" || t == "save" || t == "refill"
}

func goalWarning(goal data.ListGoalsValuesRow) string {
	s := ""
	if goal.Gap == 0 {
		return ""
	}
	if goal.Gap >= 0 {
		s += "Deallocate "
		s += formatCents(goal.Gap)
		s += " to match goal amount"
	} else {
		s += "Allocate "
		s += formatCents(goal.Gap * -1)
		s += " more to "
		switch goal.Type {
		case "refill":
			s += "refill category"
		case "save":
			s += "stay on track for saving goal"
		case "monthly":
			s += "meet monthly allocation"
		}
	}
	return s
}

func printTransactionsByBudget(w io.Writer, transactions []data.ListTrxsRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tDATE\tACCOUNT\tPAYEE\tOUTFLOW\tINFLOW\tNOTE"); err != nil {
		return err
	}
	for _, t := range transactions {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			t.Date.Format("2006-01-02"),
			t.AccountName,
			t.PayeeName,
			formatCents(t.TotalOutflow),
			formatCents(t.TotalInflow),
			t.Note,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printAccountSummary(w io.Writer, name string, balance, reconciledBalance int64) error {
	if _, err := fmt.Fprintf(w, "Name: %s\n", name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Balance: %s\n", formatCents(balance)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Reconciled balance: %s\n", formatCents(reconciledBalance)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

func printAccountTransactions(w io.Writer, accountID int64, transactions []data.ListAccountTransactionsRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tDATE\tPAYEE\tTARGET\tOUTFLOW\tINFLOW\tRECONCILED\tNOTE"); err != nil {
		return err
	}

	for i := 0; i < len(transactions); {
		j := i + 1
		for j < len(transactions) && transactions[j].TrxID == transactions[i].TrxID {
			j++
		}

		if err := printTransaction(tw, accountID, transactions[i:j]); err != nil {
			return err
		}

		i = j
	}

	return tw.Flush()
}

func printTransaction(w io.Writer, accountID int64, transactions []data.ListAccountTransactionsRow) error {
	categoryCount := actualCategoryCount(transactions)
	if categoryCount > 1 || hasMismatchedSingleCategory(accountID, transactions) {
		transaction := transactions[0]
		totalOutflow, totalInflow := displayedTotals(accountID, transactions)
		if categoryCount == 1 && transaction.AccountID == accountID {
			totalOutflow = transaction.Outflow
			totalInflow = transaction.Inflow
		}
		if err := writeAccountTransactionRow(
			w,
			strconv.FormatInt(transaction.TrxID, 10),
			transaction.Date.Format("2006-01-02"),
			stringValue(transaction.PayeeName),
			"category",
			formatCents(totalOutflow),
			formatCents(totalInflow),
			strconv.FormatBool(transaction.Reconciled),
			stringValue(transaction.Note),
		); err != nil {
			return err
		}

		for _, transaction := range transactions {
			outflow, inflow := displayedCategoryAmounts(accountID, transaction)
			if err := writeAccountTransactionRow(
				w,
				"",
				"",
				"",
				transactionTarget(accountID, transaction),
				outflow,
				inflow,
				"",
				"",
			); err != nil {
				return err
			}
		}

		return nil
	}

	transaction := transactions[0]
	target := ""
	outflow := formatCents(transaction.Outflow)
	inflow := formatCents(transaction.Inflow)
	if len(transactions) == 1 && transactionHasTarget(transaction) {
		target = transactionTarget(accountID, transaction)
		outflow, inflow = displayedCategoryAmounts(accountID, transaction)
	}

	if err := writeAccountTransactionRow(
		w,
		strconv.FormatInt(transaction.TrxID, 10),
		transaction.Date.Format("2006-01-02"),
		stringValue(transaction.PayeeName),
		target,
		outflow,
		inflow,
		strconv.FormatBool(transaction.Reconciled),
		stringValue(transaction.Note),
	); err != nil {
		return err
	}

	return nil
}

func displayedTotals(accountID int64, transactions []data.ListAccountTransactionsRow) (int64, int64) {
	var outflow int64
	var inflow int64

	for _, transaction := range transactions {
		if isMirroredTransferRow(accountID, transaction) {
			outflow += nullableInt64Value(transaction.LineInflow)
			inflow += nullableInt64Value(transaction.LineOutflow)
			continue
		}

		outflow += nullableInt64Value(transaction.LineOutflow)
		inflow += nullableInt64Value(transaction.LineInflow)
	}

	return outflow, inflow
}

func transactionHasTarget(transaction data.ListAccountTransactionsRow) bool {
	return transaction.CategoryID.Valid || transaction.DestAccountID.Valid || transaction.Income || transaction.SourceAccountID.Valid
}

func actualCategoryCount(transactions []data.ListAccountTransactionsRow) int {
	count := 0
	for _, transaction := range transactions {
		if transaction.CategoryID.Valid {
			count++
		}
	}

	return count
}

func hasMismatchedSingleCategory(accountID int64, transactions []data.ListAccountTransactionsRow) bool {
	if actualCategoryCount(transactions) != 1 {
		return false
	}

	transaction := transactions[0]
	if transaction.AccountID != accountID {
		return false
	}

	return nullableInt64Value(transaction.LineOutflow) != transaction.Outflow || nullableInt64Value(transaction.LineInflow) != transaction.Inflow
}

func writeAccountTransactionRow(w io.Writer, id, date, payee, target, outflow, inflow, reconciled, note string) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", id, date, payee, target, outflow, inflow, reconciled, note)
	return err
}

func transactionTarget(accountID int64, transaction data.ListAccountTransactionsRow) string {
	if isMirroredTransferRow(accountID, transaction) {
		return "@" + stringValue(transaction.SourceAccountName)
	}

	if transaction.DestAccountID.Valid {
		return "@" + stringValue(transaction.DestAccountName)
	}

	if transaction.Income {
		return "Income"
	}

	return stringValue(transaction.CategoryName)
}

func displayedCategoryAmounts(accountID int64, transaction data.ListAccountTransactionsRow) (string, string) {
	if isMirroredTransferRow(accountID, transaction) {
		return formatCents(transaction.Outflow), formatCents(transaction.Inflow)
	}

	return nullableCentsString(transaction.LineOutflow), nullableCentsString(transaction.LineInflow)
}

func isMirroredTransferRow(accountID int64, transaction data.ListAccountTransactionsRow) bool {
	return transaction.SourceAccountID.Valid
}

func nullableCentsString(value sql.NullInt64) string {
	if !value.Valid {
		return ""
	}

	return formatCents(value.Int64)
}

func formatCents(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	s := fmt.Sprintf("%d.%02d", cents/100, cents%100)
	if negative {
		s = "-" + s
	}
	return s
}

func nullableInt64Value(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}

	return value.Int64
}

func stringValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case sql.NullString:
		if !v.Valid {
			return ""
		}
		return v.String
	case sql.NullInt64:
		if !v.Valid {
			return ""
		}
		return strconv.FormatInt(v.Int64, 10)
	case sql.NullBool:
		if !v.Valid {
			return ""
		}
		return strconv.FormatBool(v.Bool)
	}

	return fmt.Sprint(value)
}

func fail(w io.Writer, err error) int {
	fmt.Fprintf(w, "error: %v\n", err)
	return 1
}
