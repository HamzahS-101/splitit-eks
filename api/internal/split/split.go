package split

import (
	"errors"
	"sort"
)

var (
	ErrNoMembers       = errors.New("at least one member must share the expense")
	ErrDuplicateMember = errors.New("a member appears more than once")
	ErrNegativeShare   = errors.New("shares cannot be negative")
	ErrSharesMismatch  = errors.New("shares must add up to the expense amount")
)

type Share struct {
	MemberID    string `json:"member_id"`
	AmountPence int64  `json:"amount_pence"`
}

type Balance struct {
	MemberID   string `json:"member_id"`
	Name       string `json:"name"`
	PaidPence  int64  `json:"paid_pence"`
	SharePence int64  `json:"share_pence"`
	NetPence   int64  `json:"net_pence"`
}

type Transfer struct {
	From        string `json:"from_member"`
	To          string `json:"to_member"`
	AmountPence int64  `json:"amount_pence"`
}

func Equal(amount int64, memberIDs []string) ([]Share, error) {
	if len(memberIDs) == 0 {
		return nil, ErrNoMembers
	}
	if hasDuplicates(memberIDs) {
		return nil, ErrDuplicateMember
	}

	n := int64(len(memberIDs))
	base, remainder := amount/n, amount%n

	shares := make([]Share, len(memberIDs))
	for i, id := range memberIDs {
		a := base
		if int64(i) < remainder {
			a++
		}
		shares[i] = Share{MemberID: id, AmountPence: a}
	}
	return shares, nil
}

func Custom(amount int64, shares []Share) ([]Share, error) {
	if len(shares) == 0 {
		return nil, ErrNoMembers
	}

	ids := make([]string, len(shares))
	var total int64
	for i, s := range shares {
		if s.AmountPence < 0 {
			return nil, ErrNegativeShare
		}
		ids[i] = s.MemberID
		total += s.AmountPence
	}
	if hasDuplicates(ids) {
		return nil, ErrDuplicateMember
	}
	if total != amount {
		return nil, ErrSharesMismatch
	}
	return shares, nil
}

func SettleUp(balances []Balance) []Transfer {
	type entry struct {
		id     string
		amount int64
	}

	var creditors, debtors []entry
	for _, b := range balances {
		switch {
		case b.NetPence > 0:
			creditors = append(creditors, entry{b.MemberID, b.NetPence})
		case b.NetPence < 0:
			debtors = append(debtors, entry{b.MemberID, -b.NetPence})
		}
	}

	largestFirst := func(e []entry) {
		sort.SliceStable(e, func(i, j int) bool {
			if e[i].amount != e[j].amount {
				return e[i].amount > e[j].amount
			}
			return e[i].id < e[j].id
		})
	}
	largestFirst(creditors)
	largestFirst(debtors)

	transfers := []Transfer{}
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		amount := min(debtors[i].amount, creditors[j].amount)
		transfers = append(transfers, Transfer{From: debtors[i].id, To: creditors[j].id, AmountPence: amount})
		debtors[i].amount -= amount
		creditors[j].amount -= amount
		if debtors[i].amount == 0 {
			i++
		}
		if creditors[j].amount == 0 {
			j++
		}
	}
	return transfers
}

func hasDuplicates(ids []string) bool {
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return true
		}
		seen[id] = true
	}
	return false
}
