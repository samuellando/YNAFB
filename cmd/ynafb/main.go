package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
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

type categoryTargetArgs struct {
	OtherAccount string
	Category     string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
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

	if err := executeResourceAction(ctx, db, queries, resource, action, *budgetName, stdin, remaining[2:], stdout); err != nil {
		return fail(stderr, err)
	}

	return 0
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] budget create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] budget list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] budget delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] budget show [budget_name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] budget show [budget_name] [month]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account show [account]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account reconcile [account] [date]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account import [account] [pdf]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] account categorize [account]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] allocation create [month] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] allocation list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] allocation delete [month] [category]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] category create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] category list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] category delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] goal create [type] [start] [end|null] [category] [amount]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] goal list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] goal delete [category]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] payee create [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] payee list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] payee delete [name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] payee default-category create [payee] [percent] [--category name | --other-account name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] payee default-category delete [id]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] transaction create [date] [account] [payee] [total_out] [total_in] [note]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] transaction list\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] transaction delete [id]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] [--budget name] transaction category create [transaction] [outflow] [inflow] [--category name | --other-account name]\n")
	fmt.Fprintf(w, "  ynafb [--db ./ynafb.db] transaction category delete [id]\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Dates accept RFC3339 or YYYY-MM-DD. Use null for goal end dates (monthly goals).\n")
	fmt.Fprintf(w, "Goal months are YYYY-MM; save goals require an end month.\n")
	fmt.Fprintf(w, "Omit --budget only when exactly one budget exists.\n")
}

func executeResourceAction(ctx context.Context, db *sql.DB, queries *data.Queries, resource, action, budgetName string, stdin io.Reader, args []string, stdout io.Writer) error {
	if resource == "budget" {
		switch action {
		case "create":
			if len(args) != 1 {
				return fmt.Errorf("budget create requires [name]")
			}

			result, err := queries.CreateBudget(ctx, args[0])
			if err != nil {
				return err
			}

			printCreated(stdout, "budget", result.ID)
			return nil

		case "list":
			budgets, err := queries.ListBudgetBalances(ctx)
			if err != nil {
				return err
			}

			return printBudgets(stdout, budgets)

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("budget delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeleteBudget(ctx, budget.ID); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted budget %q\n", budget.Name)
			return nil

		case "show":
			if len(args) != 1 && len(args) != 2 {
				return fmt.Errorf("budget show requires [budget_name] or [budget_name] [month]")
			}

			budget, err := resolveBudget(ctx, queries, args[0])
			if err != nil {
				return err
			}

			if len(args) == 1 {
				allocMonths, err := queries.ListAllocationMonthsByBudget(ctx, budget.ID)
				if err != nil {
					return err
				}

				txDates, err := queries.ListTransactionDatesByBudget(ctx, budget.ID)
				if err != nil {
					return err
				}

				return printBudgetMonths(stdout, distinctBudgetMonths(allocMonths, txDates))
			}

			month, err := parseMonth("month", args[1])
			if err != nil {
				return err
			}

			rows, err := queries.ListBudgetMonthCategories(ctx, data.ListBudgetMonthCategoriesParams{
				Budget: budget.ID,
				Start:  month,
				End:    month.AddDate(0, 1, 0),
			})
			if err != nil {
				return err
			}

			goals, err := queries.ListGoalsByBudget(ctx, budget.ID)
			if err != nil {
				return err
			}

			allocations, err := queries.ListAllocationsByBudget(ctx, budget.ID)
			if err != nil {
				return err
			}

			return printBudgetMonthCategories(stdout, rows, goals, buildCategoryAllocations(allocations), month)

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

		case "list":
			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			accounts, err := queries.ListAccountBalances(ctx, budget.ID)
			if err != nil {
				return err
			}

			return printAccountBalances(stdout, accounts)

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("account delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeleteAccount(ctx, accountID); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted account %q\n", args[0])
			return nil

		case "show":
			if len(args) != 1 {
				return fmt.Errorf("account show requires [account]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			accountID, err := resolveAccountID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			balances, err := queries.GetAccountBalances(ctx, accountID)
			if err != nil {
				return err
			}

			if err := printAccountSummary(stdout, args[0], balances.Balance, balances.ReconciledBalance); err != nil {
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

		case "reconcile":
			if len(args) != 2 {
				return fmt.Errorf("account reconcile requires [account] [date]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
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

			budget, err := resolveBudget(ctx, queries, budgetName)
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

			budget, err := resolveBudget(ctx, queries, budgetName)
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

			budget, err := resolveBudget(ctx, queries, budgetName)
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
				Budget:   budget.ID,
				Category: category,
				Month:    month,
				Amount:   amount,
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "allocation", result.ID)
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			allocations, err := queries.ListAllocationsByBudget(ctx, budget.ID)
			if err != nil {
				return err
			}

			return printAllocations(stdout, allocations)

		case "delete":
			if len(args) != 2 {
				return fmt.Errorf("allocation delete requires [month] [category]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
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

			if err := queries.DeleteAllocationByCategoryAndMonth(ctx, data.DeleteAllocationByCategoryAndMonthParams{
				Budget:   budget.ID,
				Category: category,
				Month:    month,
			}); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted allocation %s %q\n", month.Format("2006-01"), args[1])
			return nil

		default:
			return fmt.Errorf("unsupported action %q for resource %q", action, resource)
		}

	case "category":
		switch action {
		case "create":
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

		case "list":
			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			categories, err := queries.ListCategoriesByBudget(ctx, budget.ID)
			if err != nil {
				return err
			}

			return printNames(stdout, categories, func(c data.Category) string { return stringValue(c.Name) })

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("category delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			categoryID, err := resolveCategoryID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeleteCategory(ctx, categoryID); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted category %q\n", args[0])
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

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			goalType := strings.ToLower(args[0])
			if goalType != "monthly" && goalType != "save" {
				return fmt.Errorf("goal create: unknown goal type %q (expected monthly or save)", args[0])
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
				Budget:   budget.ID,
				Type:     goalType,
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

		case "list":
			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			goals, err := queries.ListGoalsByBudget(ctx, budget.ID)
			if err != nil {
				return err
			}

			allocations, err := queries.ListAllocationsByBudget(ctx, budget.ID)
			if err != nil {
				return err
			}

			return printGoals(stdout, goals, buildCategoryAllocations(allocations), monthStart(time.Now()))

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("goal delete requires [category]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			categoryID, err := resolveCategoryID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			goal, err := queries.GetGoalByCategory(ctx, data.GetGoalByCategoryParams{
				Budget:   budget.ID,
				Category: categoryID,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("no goal for category %q in budget %q", args[0], budget.Name)
				}
				return err
			}

			if err := queries.DeleteGoal(ctx, goal.ID); err != nil {
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

		case "list":
			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			payees, err := queries.ListPayeesByBudget(ctx, budget.ID)
			if err != nil {
				return err
			}

			return printNames(stdout, payees, func(p data.Payee) string { return stringValue(p.Name) })

		case "delete":
			if len(args) != 1 {
				return fmt.Errorf("payee delete requires [name]")
			}

			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			payeeID, err := resolvePayeeID(ctx, queries, budget, args[0])
			if err != nil {
				return err
			}

			if err := queries.DeletePayee(ctx, payeeID); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted payee %q\n", args[0])
			return nil

		case "default-category":
			if len(args) < 1 {
				return fmt.Errorf("payee default-category requires a subcommand (create|delete)")
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

				budget, err := resolveBudget(ctx, queries, budgetName)
				if err != nil {
					return err
				}

				payee, err := resolvePayeeID(ctx, queries, budget, subArgs[0])
				if err != nil {
					return err
				}

				percent, err := parseInt64("percent", positionals[1])
				if err != nil {
					return err
				}

				otherAccount, category, err := resolveCategoryTargets(ctx, queries, budget, targetArgs)
				if err != nil {
					return err
				}

				result, err := queries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
					Payee:        payee,
					OtherAccount: sql.NullInt64{Int64: otherAccount, Valid: otherAccount != 0},
					Category:     category,
					Percent:      percent,
				})
				if err != nil {
					return err
				}

				printCreated(stdout, "payee default-category", result.ID)
				return nil

			case "delete":
				if len(subArgs) != 1 {
					return fmt.Errorf("payee default-category delete requires [id]")
				}

				id, err := parseInt64("id", subArgs[0])
				if err != nil {
					return err
				}

				if err := queries.DeletePayeeDefaultCategory(ctx, id); err != nil {
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
				Note:         args[5],
			})
			if err != nil {
				return err
			}

			printCreated(stdout, "transaction", result.ID)
			return nil

		case "list":
			budget, err := resolveBudget(ctx, queries, budgetName)
			if err != nil {
				return err
			}

			transactions, err := queries.ListTransactionsByBudget(ctx, budget.ID)
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

			if err := queries.DeleteTransaction(ctx, id); err != nil {
				return err
			}

			fmt.Fprintf(stdout, "deleted transaction %d\n", id)
			return nil

		case "category":
			if len(args) < 1 {
				return fmt.Errorf("transaction category requires a subcommand (create|delete)")
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

				otherAccount, category, err := resolveCategoryTargets(ctx, queries, budget, targetArgs)
				if err != nil {
					return err
				}

				result, err := queries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
					Transaction:  transactionID,
					OtherAccount: sql.NullInt64{Int64: otherAccount, Valid: otherAccount != 0},
					Category:     category,
					Outflow:      outflow,
					Inflow:       inflow,
				})
				if err != nil {
					return err
				}

				printCreated(stdout, "transaction category", result.ID)
				return nil

			case "delete":
				if len(subArgs) != 1 {
					return fmt.Errorf("transaction category delete requires [id]")
				}

				id, err := parseInt64("id", subArgs[0])
				if err != nil {
					return err
				}

				if err := queries.DeleteTransactionCategory(ctx, id); err != nil {
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

		if i+1 >= len(args) {
			return nil, categoryTargetArgs{}, fmt.Errorf("missing value for %s", arg)
		}

		value := args[i+1]
		i++

		switch arg {
		case "--other-account":
			if targets.OtherAccount != "" {
				return nil, categoryTargetArgs{}, fmt.Errorf("--other-account may only be set once")
			}
			targets.OtherAccount = value
		case "--category":
			if targets.Category != "" {
				return nil, categoryTargetArgs{}, fmt.Errorf("--category may only be set once")
			}
			targets.Category = value
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

	if targetCount != 1 {
		return nil, categoryTargetArgs{}, fmt.Errorf("set exactly one of --category or --other-account")
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

		_, err = txQueries.CreateTransaction(ctx, data.CreateTransactionParams{
			Date:         entry.TransDate,
			Account:      accountID,
			Payee:        payeeID,
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
		AccountID: accountID,
		Date:      date,
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
		AccountID: accountID,
		Date:      date,
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
	targetID   int64
	targetName string
	outflow    int64
	inflow     int64
}

func categorizeAccount(ctx context.Context, db *sql.DB, queries *data.Queries, budget budgetContext, accountID int64, accountName string, stdin io.Reader, stdout io.Writer) error {
	transactions, err := queries.ListAccountTransactions(ctx, data.ListAccountTransactionsParams{
		Account:      accountID,
		OtherAccount: sql.NullInt64{Int64: accountID, Valid: true},
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
				if err := replaceTransactionCategories(ctx, db, queries, tx.ID, working); err != nil {
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
		for j < len(transactions) && transactions[j].ID == transactions[i].ID {
			j++
		}

		rows := transactions[i:j]
		if categorizeNeedsAttention(accountID, rows) {
			t := rows[0]
			queue = append(queue, categorizeTransaction{
				ID:                     t.ID,
				Date:                   t.Date,
				TransactionAccount:     t.TransactionAccount,
				TransactionAccountName: stringValue(t.TransactionAccountName),
				PayeeName:              stringValue(t.PayeeName),
				TotalOutflow:           t.TotalOutflow,
				TotalInflow:            t.TotalInflow,
				Reconciled:             t.Reconciled,
				Note:                   stringValue(t.Note),
				rows:                   rows,
			})
		}

		i = j
	}
	return queue
}

func categorizeNeedsAttention(accountID int64, rows []data.ListAccountTransactionsRow) bool {
	if rows[0].TransactionAccount != accountID {
		return false
	}

	var out, in int64
	for _, r := range rows {
		if r.CategoryID.Valid {
			out += nullableInt64Value(r.Outflow)
			in += nullableInt64Value(r.Inflow)
		}
	}

	return out != rows[0].TotalOutflow || in != rows[0].TotalInflow
}

func loadCategorizations(rows []data.ListAccountTransactionsRow) []categorizeRow {
	var working []categorizeRow
	for _, r := range rows {
		if !r.CategoryID.Valid {
			continue
		}

		row := categorizeRow{
			id:      r.CategoryID.Int64,
			outflow: nullableInt64Value(r.Outflow),
			inflow:  nullableInt64Value(r.Inflow),
		}
		if r.OtherAccount.Valid {
			row.transfer = true
			row.targetID = r.OtherAccount.Int64
			row.targetName = stringValue(r.OtherAccountName)
		} else {
			row.transfer = false
			row.targetID = r.Category.Int64
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

	defaults, err := queries.ListPayeeDefaultCategoriesByPayee(ctx, payeeID)
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
		if d.OtherAccount.Valid {
			row.transfer = true
			row.targetID = d.OtherAccount.Int64
			row.targetName = stringValue(d.OtherAccountName)
		} else {
			row.transfer = false
			row.targetID = d.Category.Int64
			row.targetName = stringValue(d.CategoryName)
		}
		working = append(working, row)
	}
	return working, true
}

func defaultAmounts(defaults []data.ListPayeeDefaultCategoriesByPayeeRow, total int64) []int64 {
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
		ID:                     tx.ID,
		Date:                   tx.Date,
		TransactionAccount:     tx.TransactionAccount,
		TransactionAccountName: tx.TransactionAccountName,
		PayeeName:              tx.PayeeName,
		TotalOutflow:           tx.TotalOutflow,
		TotalInflow:            tx.TotalInflow,
		Reconciled:             tx.Reconciled,
		Note:                   tx.Note,
	}

	if len(working) == 0 {
		return []data.ListAccountTransactionsRow{base}
	}

	rows := make([]data.ListAccountTransactionsRow, 0, len(working))
	for i, row := range working {
		r := base
		r.CategoryID = sql.NullInt64{Int64: int64(i + 1), Valid: true}
		r.Outflow = sql.NullInt64{Int64: row.outflow, Valid: true}
		r.Inflow = sql.NullInt64{Int64: row.inflow, Valid: true}
		if row.transfer {
			r.OtherAccount = sql.NullInt64{Int64: row.targetID, Valid: true}
			r.OtherAccountName = row.targetName
		} else {
			r.Category = sql.NullInt64{Int64: row.targetID, Valid: true}
			r.CategoryName = row.targetName
		}
		rows = append(rows, r)
	}
	return rows
}

func categorizeAdd(ctx context.Context, queries *data.Queries, budget budgetContext, reader *bufio.Reader, stdout io.Writer, tx categorizeTransaction, working *[]categorizeRow) error {
	target, err := prompt(reader, stdout, "  Target (category, or @account for a transfer): ")
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
	category, err := queries.GetCategoryByName(ctx, data.GetCategoryByNameParams{Budget: budget.ID, Name: name})
	if err == nil {
		return category.ID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	ans, err := prompt(reader, stdout, fmt.Sprintf("  Category %q doesn't exist. Create it? [y/n]: ", name))
	if err != nil {
		return 0, err
	}
	if !strings.EqualFold(ans, "y") {
		return 0, fmt.Errorf("category %q not created", name)
	}

	created, err := queries.CreateCategory(ctx, data.CreateCategoryParams{Budget: budget.ID, Name: name})
	if err != nil {
		return 0, err
	}

	fmt.Fprintf(stdout, "  → created category %q\n", name)
	return created.ID, nil
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

	defaults, err := queries.ListPayeeDefaultCategoriesByPayee(ctx, payeeID)
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

	if err := replacePayeeDefaults(ctx, db, queries, payeeID, working, percents); err != nil {
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
	return row.targetName
}

func categorizeKey(row categorizeRow) string {
	if row.transfer {
		return fmt.Sprintf("account:%d", row.targetID)
	}
	return fmt.Sprintf("category:%d", row.targetID)
}

func defaultsMatch(tx categorizeTransaction, defaults []data.ListPayeeDefaultCategoriesByPayeeRow, working []categorizeRow) bool {
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
		if d.OtherAccount.Valid {
			key = fmt.Sprintf("account:%d", d.OtherAccount.Int64)
		} else {
			key = fmt.Sprintf("category:%d", d.Category.Int64)
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

func replaceTransactionCategories(ctx context.Context, db *sql.DB, queries *data.Queries, transactionID int64, working []categorizeRow) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQueries := queries.WithTx(tx)
	if err := txQueries.DeleteTransactionCategoriesByTransaction(ctx, transactionID); err != nil {
		return err
	}
	for _, row := range working {
		if _, err := txQueries.CreateTransactionCategory(ctx, data.CreateTransactionCategoryParams{
			Transaction:  transactionID,
			OtherAccount: categorizeOtherAccount(row),
			Category:     categorizeCategory(row),
			Outflow:      row.outflow,
			Inflow:       row.inflow,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func replacePayeeDefaults(ctx context.Context, db *sql.DB, queries *data.Queries, payeeID int64, working []categorizeRow, percents map[string]int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txQueries := queries.WithTx(tx)
	if err := txQueries.DeletePayeeDefaultCategoriesByPayee(ctx, payeeID); err != nil {
		return err
	}
	for _, row := range working {
		if _, err := txQueries.CreatePayeeDefaultCategory(ctx, data.CreatePayeeDefaultCategoryParams{
			Payee:        payeeID,
			OtherAccount: categorizeOtherAccount(row),
			Category:     categorizeCategory(row),
			Percent:      percents[categorizeKey(row)],
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
	if row.transfer {
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

func resolveCategoryTargets(ctx context.Context, queries *data.Queries, budget budgetContext, args categoryTargetArgs) (int64, sql.NullInt64, error) {
	if args.Category != "" {
		categoryID, err := resolveCategoryID(ctx, queries, budget, args.Category)
		if err != nil {
			return 0, sql.NullInt64{}, err
		}

		return 0, sql.NullInt64{Int64: categoryID, Valid: true}, nil
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

func distinctBudgetMonths(allocMonths []time.Time, txDates []time.Time) []string {
	seen := make(map[string]struct{})
	for _, m := range allocMonths {
		seen[m.Format("2006-01")] = struct{}{}
	}
	for _, d := range txDates {
		seen[d.Format("2006-01")] = struct{}{}
	}

	months := make([]string, 0, len(seen))
	for m := range seen {
		months = append(months, m)
	}
	sort.Strings(months)
	return months
}

func printBudgetMonths(w io.Writer, months []string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "MONTH"); err != nil {
		return err
	}
	for _, m := range months {
		if _, err := fmt.Fprintf(tw, "%s\n", m); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printBudgetMonthCategories(w io.Writer, rows []data.ListBudgetMonthCategoriesRow, goals []data.ListGoalsByBudgetRow, allocations map[int64]map[time.Time]int64, month time.Time) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "CATEGORY\tGOAL\tALLOCATED\tSPENT\tREMAINING"); err != nil {
		return err
	}

	goalsByCategory := make(map[int64]data.ListGoalsByBudgetRow)
	for _, g := range goals {
		goalsByCategory[g.CategoryID] = g
	}

	for _, r := range rows {
		goal := ""
		if g, ok := goalsByCategory[r.ID]; ok && goalActiveInMonth(g, month) {
			goal = formatCents(goalMonthlyValue(g, allocations, month))
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", stringValue(r.Name), goal, formatCents(r.Allocated), formatCents(r.Spent), formatCents(r.Allocated-r.Spent)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printBudgets(w io.Writer, budgets []data.ListBudgetBalancesRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tBALANCE\tRECONCILED BALANCE"); err != nil {
		return err
	}
	for _, b := range budgets {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", stringValue(b.Name), formatCents(b.Balance), formatCents(b.ReconciledBalance)); err != nil {
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

func printAccountBalances(w io.Writer, accounts []data.ListAccountBalancesRow) error {
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

func printAllocations(w io.Writer, allocations []data.ListAllocationsByBudgetRow) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "MONTH\tCATEGORY\tAMOUNT"); err != nil {
		return err
	}
	for _, a := range allocations {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", a.Month.Format("2006-01"), stringValue(a.CategoryName), formatCents(a.Amount)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printGoals(w io.Writer, goals []data.ListGoalsByBudgetRow, allocations map[int64]map[time.Time]int64, now time.Time) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "TYPE\tCATEGORY\tSTART\tEND\tAMOUNT\tMONTHLY"); err != nil {
		return err
	}
	for _, g := range goals {
		end := ""
		if g.End.Valid {
			end = g.End.Time.Format("2006-01")
		}
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			stringValue(g.Type),
			stringValue(g.CategoryName),
			g.Start.Format("2006-01"),
			end,
			formatCents(g.Amount),
			formatCents(goalMonthlyValue(g, allocations, now)),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func buildCategoryAllocations(rows []data.ListAllocationsByBudgetRow) map[int64]map[time.Time]int64 {
	byCategory := make(map[int64]map[time.Time]int64)
	for _, r := range rows {
		if byCategory[r.CategoryID] == nil {
			byCategory[r.CategoryID] = make(map[time.Time]int64)
		}
		byCategory[r.CategoryID][r.Month] = r.Amount
	}
	return byCategory
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func monthsRemaining(from, end time.Time) int {
	if from.After(end) {
		return 0
	}
	return (end.Year()-from.Year())*12 + int(end.Month()-from.Month()) + 1
}

func goalActiveInMonth(g data.ListGoalsByBudgetRow, month time.Time) bool {
	if month.Before(g.Start) {
		return false
	}
	if g.End.Valid && g.End.Time.Before(month) {
		return false
	}
	return true
}

func goalMonthlyValue(g data.ListGoalsByBudgetRow, allocations map[int64]map[time.Time]int64, month time.Time) int64 {
	if !goalActiveInMonth(g, month) {
		return 0
	}

	if stringValue(g.Type) == "save" {
		if !g.End.Valid {
			return 0
		}

		var allocated int64
		for allocMonth, amount := range allocations[g.CategoryID] {
			if !allocMonth.Before(g.Start) && allocMonth.Before(month) {
				allocated += amount
			}
		}

		months := monthsRemaining(month, g.End.Time)
		if months <= 0 {
			return 0
		}

		value := (g.Amount - allocated) / int64(months)
		if value < 0 {
			value = 0
		}
		return value
	}

	return g.Amount
}

func printTransactionsByBudget(w io.Writer, transactions []data.ListTransactionsByBudgetRow) error {
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
			stringValue(t.AccountName),
			stringValue(t.PayeeName),
			formatCents(t.TotalOutflow),
			formatCents(t.TotalInflow),
			stringValue(t.Note),
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
		for j < len(transactions) && transactions[j].ID == transactions[i].ID {
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
		if categoryCount == 1 && transaction.TransactionAccount == accountID {
			totalOutflow = transaction.TotalOutflow
			totalInflow = transaction.TotalInflow
		}
		if err := writeAccountTransactionRow(
			w,
			strconv.FormatInt(transaction.ID, 10),
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
	outflow := formatCents(transaction.TotalOutflow)
	inflow := formatCents(transaction.TotalInflow)
	if categoryCount == 1 {
		target = transactionTarget(accountID, transaction)
		outflow, inflow = displayedCategoryAmounts(accountID, transaction)
	}

	if err := writeAccountTransactionRow(
		w,
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

	return nil
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
		return "@" + stringValue(transaction.TransactionAccountName)
	}

	if transaction.OtherAccount.Valid {
		return "@" + stringValue(transaction.OtherAccountName)
	}

	return stringValue(transaction.CategoryName)
}

func displayedCategoryAmounts(accountID int64, transaction data.ListAccountTransactionsRow) (string, string) {
	if isMirroredTransferRow(accountID, transaction) {
		return nullableCentsString(transaction.Inflow), nullableCentsString(transaction.Outflow)
	}

	return nullableCentsString(transaction.Outflow), nullableCentsString(transaction.Inflow)
}

func isMirroredTransferRow(accountID int64, transaction data.ListAccountTransactionsRow) bool {
	return transaction.TransactionAccount != accountID && transaction.OtherAccount.Valid && transaction.OtherAccount.Int64 == accountID
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
	if value == nil {
		return ""
	}

	return fmt.Sprint(value)
}

func fail(w io.Writer, err error) int {
	fmt.Fprintf(w, "error: %v\n", err)
	return 1
}
