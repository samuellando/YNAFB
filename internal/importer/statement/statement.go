package statement

import "time"

// Entry is a single normalized transaction line from a statement.
type Entry struct {
	// TransDate is the transaction date as shown on the statement.
	TransDate time.Time
	// PostedDate is the posting date as shown on the statement.
	PostedDate time.Time
	// Type is the bank-provided entry type (for example "purchase" or "payment").
	Type string
	// Payee is the merchant or counterparty name.
	Payee string
	// Note carries any extra detail such as foreign currency conversion info.
	Note string
	// Outflow and Inflow are amounts in the smallest currency unit (cents).
	// Exactly one of them is non-zero for a normal entry.
	Outflow int64
	Inflow  int64
}

// Statement is a normalized view of a bank statement.
type Statement struct {
	// Institution is the name of the issuing bank.
	Institution string
	// Account is a statement-provided account hint (for example last four digits).
	Account string
	// Start and End are the statement period.
	Start time.Time
	End   time.Time
	// Entries are the parsed transactions in statement order.
	Entries []Entry
}
