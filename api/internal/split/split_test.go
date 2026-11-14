package split

import (
	"errors"
	"testing"
)

func TestEqualGivesRemainderToFirstMembers(t *testing.T) {
	shares, err := Equal(1000, []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{334, 333, 333}
	for i, s := range shares {
		if s.AmountPence != want[i] {
			t.Errorf("share %d = %d, want %d", i, s.AmountPence, want[i])
		}
	}
}

func TestEqualRejectsBadInput(t *testing.T) {
	if _, err := Equal(1000, nil); !errors.Is(err, ErrNoMembers) {
		t.Errorf("no members: got %v", err)
	}
	if _, err := Equal(1000, []string{"a", "a"}); !errors.Is(err, ErrDuplicateMember) {
		t.Errorf("duplicate: got %v", err)
	}
}

func TestCustom(t *testing.T) {
	tests := []struct {
		name   string
		shares []Share
		want   error
	}{
		{"exact", []Share{{"a", 3000}, {"b", 1000}, {"c", 2000}}, nil},
		{"zero share allowed", []Share{{"a", 6000}, {"b", 0}}, nil},
		{"short", []Share{{"a", 3000}, {"b", 1000}}, ErrSharesMismatch},
		{"over", []Share{{"a", 5000}, {"b", 2000}}, ErrSharesMismatch},
		{"negative", []Share{{"a", 7000}, {"b", -1000}}, ErrNegativeShare},
		{"duplicate", []Share{{"a", 3000}, {"a", 3000}}, ErrDuplicateMember},
		{"empty", nil, ErrNoMembers},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Custom(6000, tt.shares)
			if !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestSettleUpClearsAllBalances(t *testing.T) {
	tests := []struct {
		name     string
		nets     map[string]int64
		maxTrans int
	}{
		{"one payer", map[string]int64{"alice": 4000, "bob": -2000, "cara": -2000}, 2},
		{"two creditors", map[string]int64{"a": 5000, "b": 1000, "c": -3000, "d": -3000}, 3},
		{"already settled", map[string]int64{"a": 0, "b": 0}, 0},
		{"odd pennies", map[string]int64{"a": 667, "b": -334, "c": -333}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var balances []Balance
			for id, net := range tt.nets {
				balances = append(balances, Balance{MemberID: id, NetPence: net})
			}

			transfers := SettleUp(balances)
			if len(transfers) > tt.maxTrans {
				t.Errorf("got %d transfers, want at most %d", len(transfers), tt.maxTrans)
			}

			remaining := map[string]int64{}
			for id, net := range tt.nets {
				remaining[id] = net
			}
			for _, tr := range transfers {
				if tr.AmountPence <= 0 {
					t.Errorf("non-positive transfer %+v", tr)
				}
				remaining[tr.From] += tr.AmountPence
				remaining[tr.To] -= tr.AmountPence
			}
			for id, v := range remaining {
				if v != 0 {
					t.Errorf("%s left with %d after settling", id, v)
				}
			}
		})
	}
}
