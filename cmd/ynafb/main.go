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
	"time"

	"samuellando.com/YNAFB/data"
	dbutil "samuellando.com/YNAFB/internal/db"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("ynafb", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dbPath := fs.String("db", "./ynafb.db", "SQLite database path")
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
	if action != "create" {
		return fail(stderr, fmt.Errorf("unsupported action %q", action))
	}

	db, err := dbutil.Open(*dbPath)
	if err != nil {
		return fail(stderr, err)
	}
	defer db.Close()

	queries := data.New(db)
	ctx := context.Background()

	if err := createResource(ctx, queries, resource, remaining[2:], stdout); err != nil {
		return fail(stderr, err)
	}

	return 0
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] budget create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] account create [budget] [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] allocation create [budget] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] category create [budget] [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] goal create [budget] [name] [type] [start] [end|null] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] payee create [budget] [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] payee-default-split create [payee] [to_account] [from_account] [category] [outflow] [inflow]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] transaction create [date] [account] [payee] [reconciled] [note]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] transaction-split create [transaction] [to_account] [from_account] [category] [outflow] [inflow]\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Dates accept RFC3339 or YYYY-MM-DD. Use null for goal end dates.\n")
}

func createResource(ctx context.Context, queries *data.Queries, resource string, args []string, stdout io.Writer) error {
	switch resource {
	case "budget":
		if len(args) != 1 {
			return fmt.Errorf("budget create requires [name]")
		}

		result, err := queries.CreateBudget(ctx, args[0])
		if err != nil {
			return err
		}

		printCreated(stdout, "budget", result.ID)
		return nil

	case "account":
		if len(args) != 2 {
			return fmt.Errorf("account create requires [budget] [name]")
		}

		budget, err := parseInt64("budget", args[0])
		if err != nil {
			return err
		}

		result, err := queries.CreateAccount(ctx, data.CreateAccountParams{
			Budget: budget,
			Name:   args[1],
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "account", result.ID)
		return nil

	case "allocation":
		if len(args) != 3 {
			return fmt.Errorf("allocation create requires [budget] [category] [amount]")
		}

		budget, err := parseInt64("budget", args[0])
		if err != nil {
			return err
		}

		category, err := parseInt64("category", args[1])
		if err != nil {
			return err
		}

		amount, err := parseInt64("amount", args[2])
		if err != nil {
			return err
		}

		result, err := queries.CreateAllocation(ctx, data.CreateAllocationParams{
			Budget:   budget,
			Category: category,
			Amount:   amount,
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "allocation", result.ID)
		return nil

	case "category":
		if len(args) != 2 {
			return fmt.Errorf("category create requires [budget] [name]")
		}

		budget, err := parseInt64("budget", args[0])
		if err != nil {
			return err
		}

		result, err := queries.CreateCategory(ctx, data.CreateCategoryParams{
			Budget: budget,
			Name:   args[1],
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "category", result.ID)
		return nil

	case "goal":
		if len(args) != 7 {
			return fmt.Errorf("goal create requires [budget] [name] [type] [start] [end|null] [category] [amount]")
		}

		budget, err := parseInt64("budget", args[0])
		if err != nil {
			return err
		}

		start, err := parseTime("start", args[3])
		if err != nil {
			return err
		}

		end, err := parseNullableTime("end", args[4])
		if err != nil {
			return err
		}

		category, err := parseInt64("category", args[5])
		if err != nil {
			return err
		}

		amount, err := parseInt64("amount", args[6])
		if err != nil {
			return err
		}

		result, err := queries.CreateGoal(ctx, data.CreateGoalParams{
			Budget:   budget,
			Name:     args[1],
			Type:     args[2],
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
		if len(args) != 2 {
			return fmt.Errorf("payee create requires [budget] [name]")
		}

		budget, err := parseInt64("budget", args[0])
		if err != nil {
			return err
		}

		result, err := queries.CreatePayee(ctx, data.CreatePayeeParams{
			Budget: budget,
			Name:   args[1],
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "payee", result.ID)
		return nil

	case "payee-default-split":
		if len(args) != 6 {
			return fmt.Errorf("payee-default-split create requires [payee] [to_account] [from_account] [category] [outflow] [inflow]")
		}

		payee, err := parseInt64("payee", args[0])
		if err != nil {
			return err
		}

		toAccount, err := parseInt64("to_account", args[1])
		if err != nil {
			return err
		}

		fromAccount, err := parseInt64("from_account", args[2])
		if err != nil {
			return err
		}

		category, err := parseInt64("category", args[3])
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
		if len(args) != 5 {
			return fmt.Errorf("transaction create requires [date] [account] [payee] [reconciled] [note]")
		}

		date, err := parseTime("date", args[0])
		if err != nil {
			return err
		}

		account, err := parseInt64("account", args[1])
		if err != nil {
			return err
		}

		payee, err := parseInt64("payee", args[2])
		if err != nil {
			return err
		}

		reconciled, err := strconv.ParseBool(args[3])
		if err != nil {
			return fmt.Errorf("parse reconciled: %w", err)
		}

		result, err := queries.CreateTransaction(ctx, data.CreateTransactionParams{
			Date:       date,
			Account:    account,
			Payee:      payee,
			Reconciled: reconciled,
			Note:       args[4],
		})
		if err != nil {
			return err
		}

		printCreated(stdout, "transaction", result.ID)
		return nil

	case "transaction-split":
		if len(args) != 6 {
			return fmt.Errorf("transaction-split create requires [transaction] [to_account] [from_account] [category] [outflow] [inflow]")
		}

		transactionID, err := parseInt64("transaction", args[0])
		if err != nil {
			return err
		}

		toAccount, err := parseInt64("to_account", args[1])
		if err != nil {
			return err
		}

		fromAccount, err := parseInt64("from_account", args[2])
		if err != nil {
			return err
		}

		category, err := parseInt64("category", args[3])
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

		result, err := queries.CreateTransactionSplit(ctx, data.CreateTransactionSplitParams{
			Transaction: transactionID,
			ToAccount:   toAccount,
			FromAccount: fromAccount,
			Category:    category,
			Outflow:     outflow,
			Inflow:      inflow,
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

func printCreated(w io.Writer, resource string, id int64) {
	fmt.Fprintf(w, "created %s %d\n", resource, id)
}

func fail(w io.Writer, err error) int {
	fmt.Fprintf(w, "error: %v\n", err)
	return 1
}
