package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"samuellando.com/YNAFB/data"
	dbutil "samuellando.com/YNAFB/internal/db"
	"samuellando.com/YNAFB/internal/importer"
)

type budgetContext struct {
	ID   int64
	Name string
}

type splitTargetArgs struct {
	OtherAccount string
	Category     string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("ynafb", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dbPath := fs.String("db", "./ynafb.db", "SQLite database path")
	budgetName := fs.String("budget", "", "Budget name")
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

	if err := executeResourceAction(ctx, db, queries, resource, action, *budgetName, remaining[2:], stdout); err != nil {
		return fail(stderr, err)
	}

	return 0
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] budget create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account list-transactions [account]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account import [account] [pdf]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] allocation create [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] category create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] goal create [name] [type] [start] [end|null] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] payee create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] payee-default-split create [payee] [to_account] [from_account] [category] [outflow] [inflow]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] transaction create [date] [account] [payee] [total_out] [total_in] [note]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] transaction-split create [transaction] [outflow] [inflow] [--category name | --other-account name]\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Dates accept RFC3339 or YYYY-MM-DD. Use null for goal end dates.\n")
	fmt.Fprintf(w, "Omit --budget only when exactly one budget exists.\n")
}

func executeResourceAction(ctx context.Context, db *sql.DB, queries *data.Queries, resource, action, budgetName string, args []string, stdout io.Writer) error {
	if resource == "budget" {
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		if len(args) != 1 {
			return fmt.Errorf("budget create requires [name]")
		}

		result, err := queries.CreateBudget(ctx, args[0])
		if err != nil {
			return err
		}

		printCreated(stdout, "budget", result.ID)
		return nil
	}

	switch resource {
	case "account":
		switch action {
		case "create":
			if len(args) != 1 {
				return fmt.Errorf("account create requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			result, err := queries.CreateAccount(ctx, data.CreateAccountParams{
				Budget: budget.ID,
				Name:   args[0],
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "account", result.ID)
			return nil

		case "list-transactions":
			if len(args) != 1 {
				return fmt.Errorf("account list-transactions requires [account]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			transactions, err := queries.ListAccountTransactions(ctx, data.ListAccountTransactionsParams{
				Account:      accountID,
				OtherAccount: sql.NullInt64{Int64: accountID, Valid: true},
			})
			if err != nil {
				return err
			}

			return printAccountTransactions(stdout, accountID, transactions)

		case "import":
			if len(args) != 2 {
				return fmt.Errorf("account import requires [account] [pdf]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			return importAccountTransactions(ctx, db, queries, budget, accountID, args[1], stdout)

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "allocation":
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		if len(args) != 2 {
			return fmt.Errorf("allocation create requires [category] [amount]")
		}

		budget, err := resolveBudget(ctx, queries, budgetName)
		if err != nil {
			return err
		}

		category, err := resolveCategoryID(ctx, queries, budget, args[0])
		if err != nil {
			return err
		}

		amount, err := parseInt64("amount", args[1])
		if err != nil {
			return err
		}

		result, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
			Budget:   budget.ID,
			Category: category,
			Amount:   amount,
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "allocation", result.ID)
		return nil

	case "category":
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		if len(args) != 1 {
			return fmt.Errorf("category create requires [name]")
		}

		budget, err := resolveBudget(ctx, queries, budgetName)
		if err != nil {
			return err
		}

		result, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
			Budget: budget.ID,
			Name:   args[0],
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "category", result.ID)
		return nil

	case "goal":
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		if len(args) != 6 {
			return fmt.Errorf("goal create requires [name] [type] [start] [end|null] [category] [amount]")
		}

		budget, err := resolveBudget(ctx, queries, budgetName)
		if err != nil {
			return err
		}

		start, err := parseTime("start", args[2])
		if err != nil {
			return err
		}

		end, err := parseNullableTime("end", args[3])
		if err != nil {
			return err
		}

		category, err := resolveCategoryID(ctx, queries, budget, args[4])
		if err != nil {
			return err
		}

		amount, err := parseInt64("amount", args[5])
		if err != nil {
			return err
		}

		result, err := queries.CreateGoal(ctx, data.CreateGoalParams{
			Budget:   budget.ID,
			Name:     args[0],
			Type:     args[1],
			Start:    start,
			End:      end,
			Category: category,
			Amount:   amount,
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "goal", result.ID)
		return nil

	case "payee":
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		if len(args) != 1 {
			return fmt.Errorf("payee create requires [name]")
		}

		budget, err := resolveBudget(ctx, queries, budgetName)
		if err != nil {
			return err
		}

		result, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
			Budget: budget.ID,
			Name:   args[0],
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "payee", result.ID)
		return nil

	case "payee-default-split":
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		if len(args) != 6 {
			return fmt.Errorf("payee-default-split create requires [payee] [to_account] [from_account] [category] [outflow] [inflow]")
		}

		budget, err := resolveBudget(ctx, queries, budgetName)
		if err != nil {
			return err
		}

		payee, err := resolvePayeeID(ctx, queries, budget, args[0])
		if err != nil {
			return err
		}

		toAccount, err := resolveAccountID(ctx, queries, budget, args[1])
		if err != nil {
			return err
		}

		fromAccount, err := resolveAccountID(ctx, queries, budget, args[2])
		if err != nil {
			return err
		}

		category, err := resolveCategoryID(ctx, queries, budget, args[3])
		if err != nil {
			return err
		}

		outflow, err := parseInt64("outflow", args[4])
		if err != nil {
			return err
		}

		inflow, err := parseInt64("inflow", args[5])
		if err != nil {
			return err
		}

		result, err := queries.CreatePayeeDefaultSplit(ctx, data.CreatePayeeDefaultSplitParams{
			Payee:       payee,
			ToAccount:   toAccount,
			FromAccount: fromAccount,
			Category:    category,
			Outflow:     outflow,
			Inflow:      inflow,
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "payee-default-split", result.ID)
		return nil

	case "transaction":
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		if len(args) != 6 {
			return fmt.Errorf("transaction create requires [date] [account] [payee] [total_out] [total_in] [note]")
		}

		budget, err := resolveBudget(ctx, queries, budgetName)
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

		totalOutflow, err := parseInt64("total_out", args[3])
		if err != nil {
			return err
		}

		totalInflow, err := parseInt64("total_in", args[4])
		if err != nil {
			return err
		}

		result, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
			Date:         date,
			Account:      account,
			Payee:        payee,
			TotalOutflow: totalOutflow,
			TotalInflow:  totalInflow,
			Reconciled:   false,
			Note:         args[5],
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "transaction", result.ID)
		return nil

	case "transaction-split":
		if action != "create" {
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

		positionals, targetArgs, err := parseSplitTargetArgs(args)
		if err != nil {
			return err
		}

		if len(positionals) != 3 {
			return fmt.Errorf("transaction-split create requires [transaction] [outflow] [inflow] and exactly one of --category or --other-account")
		}

		budget, err := resolveBudget(ctx, queries, budgetName)
		if err != nil {
			return err
		}

		transactionID, err := parseInt64("transaction", positionals[0])
		if err != nil {
			return err
		}

		outflow, err := parseInt64("outflow", positionals[1])
		if err != nil {
			return err
		}

		inflow, err := parseInt64("inflow", positionals[2])
		if err != nil {
			return err
		}

		otherAccount, category, err := resolveSplitTargets(ctx, queries, budget, outflow, inflow, targetArgs)
		if err != nil {
			return err
		}

		result, err := queries.CreateTransactionSplit(ctx, data.CreateTransactionSplitParams{
			Transaction:  transactionID,
			OtherAccount: sql.NullInt64{Int64: otherAccount, Valid: otherAccount != 0},
			Category:     category,
			Outflow:      outflow,
			Inflow:       inflow,
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "transaction-split", result.ID)
		return nil

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

func parseNullableTime(name, value string) (sql.NullTime, error) {
	if strings.EqualFold(value, "null") {
		return sql.NullTime{}, nil
	}

	parsed, err := parseTime(name, value)
	if err != nil {
		return sql.NullTime{}, err
	}

	return sql.NullTime{Time: parsed, Valid: true}, nil
}

func parseSplitTargetArgs(args []string) ([]string, splitTargetArgs, error) {
	positionals := make([]string, 0, len(args))
	targets := splitTargetArgs{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			positionals = append(positionals, arg)
			continue
		}

		if i+1 >= len(args) {
			return nil, splitTargetArgs{}, fmt.Errorf("missing value for %s", arg)
		}

		value := args[i+1]
		i++

		switch arg {
		case "--other-account":
			if targets.OtherAccount != "" {
				return nil, splitTargetArgs{}, fmt.Errorf("--other-account may only be set once")
			}
			targets.OtherAccount = value
		case "--category":
			if targets.Category != "" {
				return nil, splitTargetArgs{}, fmt.Errorf("--category may only be set once")
			}
			targets.Category = value
		default:
			return nil, splitTargetArgs{}, fmt.Errorf("unsupported flag %q", arg)
		}
	}

	targetCount := 0
	if targets.OtherAccount != "" {
		targetCount++
	}
	if targets.Category != "" {
		targetCount++
	}

	if targetCount != 1 {
		return nil, splitTargetArgs{}, fmt.Errorf("set exactly one of --category or --other-account")
	}

	return positionals, targets, nil
}

func resolveBudget(ctx context.Context, queries *data.Queries, budgetName string) (budgetContext, error) {
	if budgetName != "" {
		budget, err := queries.GetBudgetByName(ctx, budgetName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return budgetContext{}, fmt.Errorf("unknown budget %q", budgetName)
			}

			return budgetContext{}, err
		}

		return budgetContext{ID: budget.ID, Name: stringValue(budget.Name)}, nil
	}

	budgets, err := queries.ListBudgets(ctx)
	if err != nil {
		return budgetContext{}, err
	}

	switch len(budgets) {
	case 0:
		return budgetContext{}, fmt.Errorf("no budgets exist; create one with `ynafb budget create [name]`")
	case 1:
		return budgetContext{ID: budgets[0].ID, Name: stringValue(budgets[0].Name)}, nil
	default:
		return budgetContext{}, fmt.Errorf("multiple budgets exist; pass --budget [name]")
	}
}

func resolveAccountID(ctx context.Context, queries *data.Queries, budget budgetContext, accountName string) (int64, error) {
	account, err := queries.GetAccountByName(ctx, data.GetAccountByNameParams{
		Budget: budget.ID,
		Name:   accountName,
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
		Budget: budget.ID,
		Name:   categoryName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("unknown category %q in budget %q", categoryName, budget.Name)
		}

		return 0, err
	}

	return category.ID, nil
}

func resolvePayeeID(ctx context.Context, queries *data.Queries, budget budgetContext, payeeName string) (int64, error) {
	payee, err := queries.GetPayeeByName(ctx, data.GetPayeeByNameParams{
		Budget: budget.ID,
		Name:   payeeName,
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
		Budget: budget.ID,
		Name:   payeeName,
	})
	if err == nil {
		return payee.ID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	created, createErr := queries.CreatePayee(ctx, data.CreatePayeeParams{
		Budget: budget.ID,
		Name:   payeeName,
	})
	if createErr != nil {
		return 0, createErr
	}

	return created.ID, nil
}

func importAccountTransactions(ctx context.Context, db *sql.DB, queries *data.Queries, budget budgetContext, accountID int64, path string, stdout io.Writer) error {
	stmt, err := importer.ImportFile(path)
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

		_, err = txQueries.CreateTransaction(ctx, data.CreateTransactionParams{
			Date:         entry.TransDate,
			Account:      accountID,
			Payee:        payeeID,
			TotalOutflow: entry.Outflow,
			TotalInflow:  entry.Inflow,
			Reconciled:   false,
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

func resolveSplitTargets(ctx context.Context, queries *data.Queries, budget budgetContext, outflow, inflow int64, args splitTargetArgs) (int64, sql.NullInt64, error) {
	if args.Category != "" {
		categoryID, err := resolveCategoryID(ctx, queries, budget, args.Category)
		if err != nil {
			return 0, sql.NullInt64{}, err
		}

		return 0, sql.NullInt64{Int64: categoryID, Valid: true}, nil
	}

	if (outflow > 0) == (inflow > 0) {
		return 0, sql.NullInt64{}, fmt.Errorf("transfer split with --other-account requires exactly one of outflow or inflow to be non-zero")
	}

	otherAccountID, err := resolveAccountID(ctx, queries, budget, args.OtherAccount)
	if err != nil {
		return 0, sql.NullInt64{}, err
	}

	return otherAccountID, sql.NullInt64{}, nil
}

func printCreated(w io.Writer, resource string, id int64) {
	fmt.Fprintf(w, "created %s %d\n", resource, id)
}

func printAccountTransactions(w io.Writer, accountID int64, transactions []data.ListAccountTransactionsRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tDATE\tPAYEE\tTARGET\tOUTFLOW\tINFLOW\tRECONCILED\tNOTE"); err != nil {
		return err
	}

	for i := 0; i < len(transactions); {
		j := i + 1
		for j < len(transactions) && transactions[j].ID == transactions[i].ID {
			j++
		}

		splitCount := actualSplitCount(transactions[i:j])
		if splitCount > 1 || hasMismatchedSingleSplit(accountID, transactions[i:j]) {
			transaction := transactions[i]
			totalOutflow, totalInflow := displayedTotals(accountID, transactions[i:j])
			if splitCount == 1 && transaction.TransactionAccount == accountID {
				totalOutflow = transaction.TotalOutflow
				totalInflow = transaction.TotalInflow
			}
			if err := writeAccountTransactionRow(
				tw,
				strconv.FormatInt(transaction.ID, 10),
				transaction.Date.Format("2006-01-02"),
				stringValue(transaction.PayeeName),
				"split",
				strconv.FormatInt(totalOutflow, 10),
				strconv.FormatInt(totalInflow, 10),
				strconv.FormatBool(transaction.Reconciled),
				stringValue(transaction.Note),
			); err != nil {
				return err
			}

			for k := i; k < j; k++ {
				transaction := transactions[k]
				outflow, inflow := displayedSplitAmounts(accountID, transaction)
				if err := writeAccountTransactionRow(
					tw,
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

			i = j
			continue
		}

		transaction := transactions[i]
		target := ""
		outflow := strconv.FormatInt(transaction.TotalOutflow, 10)
		inflow := strconv.FormatInt(transaction.TotalInflow, 10)
		if splitCount == 1 {
			target = transactionTarget(accountID, transaction)
			outflow, inflow = displayedSplitAmounts(accountID, transaction)
		}

		if err := writeAccountTransactionRow(
			tw,
			strconv.FormatInt(transaction.ID, 10),
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

		i = j
	}

	return tw.Flush()
}

func displayedTotals(accountID int64, transactions []data.ListAccountTransactionsRow) (int64, int64) {
	var outflow int64
	var inflow int64

	for _, transaction := range transactions {
		if isMirroredTransferRow(accountID, transaction) {
			outflow += nullableInt64Value(transaction.Inflow)
			inflow += nullableInt64Value(transaction.Outflow)
			continue
		}

		outflow += nullableInt64Value(transaction.Outflow)
		inflow += nullableInt64Value(transaction.Inflow)
	}

	return outflow, inflow
}

func actualSplitCount(transactions []data.ListAccountTransactionsRow) int {
	count := 0
	for _, transaction := range transactions {
		if transaction.SplitID.Valid {
			count++
		}
	}

	return count
}

func hasMismatchedSingleSplit(accountID int64, transactions []data.ListAccountTransactionsRow) bool {
	if actualSplitCount(transactions) != 1 {
		return false
	}

	transaction := transactions[0]
	if transaction.TransactionAccount != accountID {
		return false
	}

	return nullableInt64Value(transaction.Outflow) != transaction.TotalOutflow || nullableInt64Value(transaction.Inflow) != transaction.TotalInflow
}

func writeAccountTransactionRow(w io.Writer, id, date, payee, target, outflow, inflow, reconciled, note string) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", id, date, payee, target, outflow, inflow, reconciled, note)
	return err
}

func transactionTarget(accountID int64, transaction data.ListAccountTransactionsRow) string {
	if isMirroredTransferRow(accountID, transaction) {
		return stringValue(transaction.TransactionAccountName)
	}

	if transaction.OtherAccount.Valid {
		return stringValue(transaction.OtherAccountName)
	}

	return stringValue(transaction.CategoryName)
}

func displayedSplitAmounts(accountID int64, transaction data.ListAccountTransactionsRow) (string, string) {
	if isMirroredTransferRow(accountID, transaction) {
		return nullableIntString(transaction.Inflow), nullableIntString(transaction.Outflow)
	}

	return nullableIntString(transaction.Outflow), nullableIntString(transaction.Inflow)
}

func isMirroredTransferRow(accountID int64, transaction data.ListAccountTransactionsRow) bool {
	return transaction.TransactionAccount != accountID && transaction.OtherAccount.Valid && transaction.OtherAccount.Int64 == accountID
}

func nullableIntString(value sql.NullInt64) string {
	if !value.Valid {
		return ""
	}

	return strconv.FormatInt(value.Int64, 10)
}

func nullableInt64Value(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}

	return value.Int64
}

func stringValue(value interface{}) string {
	if value == nil {
		return ""
	}

	return fmt.Sprint(value)
}

func fail(w io.Writer, err error) int {
	fmt.Fprintf(w, "error: %v\n", err)
	return 1
}
