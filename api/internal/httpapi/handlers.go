package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HamzahS-101/splitit-eks/api/internal/split"
	"github.com/HamzahS-101/splitit-eks/api/internal/store"
)

const (
	maxMembers     = 20
	maxAmountPence = 100_000_000
)

type createGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type nameRequest struct {
	Name string `json:"name"`
}

type expenseRequest struct {
	Description string        `json:"description"`
	AmountPence int64         `json:"amount_pence"`
	PaidBy      string        `json:"paid_by"`
	SpentOn     string        `json:"spent_on"`
	SplitType   string        `json:"split_type"`
	MemberIDs   []string      `json:"member_ids"`
	Shares      []split.Share `json:"shares"`
}

type settlementRequest struct {
	From        string `json:"from_member"`
	To          string `json:"to_member"`
	AmountPence int64  `json:"amount_pence"`
}

type balancesResponse struct {
	Balances  []split.Balance  `json:"balances"`
	Transfers []split.Transfer `json:"transfers"`
}

func (s *Server) createGroup(w http.ResponseWriter, r *http.Request) {
	var req createGroupRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 80 {
		writeError(w, http.StatusBadRequest, "group name must be 1 to 80 characters")
		return
	}
	members, msg := cleanMemberNames(req.Members)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	for range 3 {
		g, err := s.store.CreateGroup(r.Context(), newSlug(), name, members)
		if errors.Is(err, store.ErrConflict) {
			continue
		}
		if err != nil {
			s.storeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, g)
		return
	}
	writeError(w, http.StatusInternalServerError, "could not create group")
}

func (s *Server) getGroup(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !validSlug(slug) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	g, err := s.store.GetGroup(r.Context(), slug)
	if err != nil {
		s.storeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (s *Server) addMember(w http.ResponseWriter, r *http.Request) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}
	var req nameRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 40 {
		writeError(w, http.StatusBadRequest, "member name must be 1 to 40 characters")
		return
	}
	m, err := s.store.AddMember(r.Context(), groupID, name)
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "a member with that name already exists")
		return
	}
	if err != nil {
		s.storeError(w, err)
		return
	}
	s.cache.Invalidate(r.Context(), groupID)
	writeJSON(w, http.StatusCreated, m)
}

func (s *Server) listExpenses(w http.ResponseWriter, r *http.Request) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}
	expenses, err := s.store.ListExpenses(r.Context(), groupID)
	if err != nil {
		s.storeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"expenses": expenses})
}

func (s *Server) createExpense(w http.ResponseWriter, r *http.Request) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}
	e, ok := s.parseExpense(w, r)
	if !ok {
		return
	}
	created, err := s.store.CreateExpense(r.Context(), groupID, e)
	if err != nil {
		s.storeError(w, err)
		return
	}
	s.cache.Invalidate(r.Context(), groupID)
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateExpense(w http.ResponseWriter, r *http.Request) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !validUUID(id) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	e, ok := s.parseExpense(w, r)
	if !ok {
		return
	}
	e.ID = id
	updated, err := s.store.UpdateExpense(r.Context(), groupID, e)
	if err != nil {
		s.storeError(w, err)
		return
	}
	s.cache.Invalidate(r.Context(), groupID)
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteExpense(w http.ResponseWriter, r *http.Request) {
	s.deleteByID(w, r, s.store.DeleteExpense)
}

func (s *Server) getBalances(w http.ResponseWriter, r *http.Request) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}

	if b, hit := s.cache.GetBalances(r.Context(), groupID); hit {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(b)
		return
	}

	balances, err := s.store.Balances(r.Context(), groupID)
	if err != nil {
		s.storeError(w, err)
		return
	}
	b, err := json.Marshal(balancesResponse{Balances: balances, Transfers: split.SettleUp(balances)})
	if err != nil {
		s.storeError(w, err)
		return
	}
	s.cache.SetBalances(r.Context(), groupID, b)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(b)
}

func (s *Server) listSettlements(w http.ResponseWriter, r *http.Request) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}
	settlements, err := s.store.ListSettlements(r.Context(), groupID)
	if err != nil {
		s.storeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"settlements": settlements})
}

func (s *Server) createSettlement(w http.ResponseWriter, r *http.Request) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}
	var req settlementRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if !validUUID(req.From) || !validUUID(req.To) {
		writeError(w, http.StatusBadRequest, "from_member and to_member must be member IDs")
		return
	}
	if req.From == req.To {
		writeError(w, http.StatusBadRequest, "a member cannot pay themselves")
		return
	}
	if req.AmountPence <= 0 || req.AmountPence > maxAmountPence {
		writeError(w, http.StatusBadRequest, "amount must be between 1p and £1,000,000")
		return
	}
	st, err := s.store.CreateSettlement(r.Context(), groupID, store.Settlement{
		From: req.From, To: req.To, AmountPence: req.AmountPence,
	})
	if err != nil {
		s.storeError(w, err)
		return
	}
	s.cache.Invalidate(r.Context(), groupID)
	writeJSON(w, http.StatusCreated, st)
}

func (s *Server) deleteSettlement(w http.ResponseWriter, r *http.Request) {
	s.deleteByID(w, r, s.store.DeleteSettlement)
}

func (s *Server) groupID(w http.ResponseWriter, r *http.Request) (string, bool) {
	slug := r.PathValue("slug")
	if !validSlug(slug) {
		writeError(w, http.StatusNotFound, "not found")
		return "", false
	}
	id, err := s.store.GroupID(r.Context(), slug)
	if err != nil {
		s.storeError(w, err)
		return "", false
	}
	return id, true
}

func (s *Server) deleteByID(w http.ResponseWriter, r *http.Request, del func(ctx context.Context, groupID, id string) error) {
	groupID, ok := s.groupID(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !validUUID(id) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err := del(r.Context(), groupID, id); err != nil {
		s.storeError(w, err)
		return
	}
	s.cache.Invalidate(r.Context(), groupID)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) parseExpense(w http.ResponseWriter, r *http.Request) (store.Expense, bool) {
	var req expenseRequest
	if !decodeJSON(w, r, &req) {
		return store.Expense{}, false
	}

	description := strings.TrimSpace(req.Description)
	if description == "" || len(description) > 120 {
		writeError(w, http.StatusBadRequest, "description must be 1 to 120 characters")
		return store.Expense{}, false
	}
	if req.AmountPence <= 0 || req.AmountPence > maxAmountPence {
		writeError(w, http.StatusBadRequest, "amount must be between 1p and £1,000,000")
		return store.Expense{}, false
	}
	if !validUUID(req.PaidBy) {
		writeError(w, http.StatusBadRequest, "paid_by must be a member ID")
		return store.Expense{}, false
	}
	if req.SpentOn != "" {
		if _, err := time.Parse(time.DateOnly, req.SpentOn); err != nil {
			writeError(w, http.StatusBadRequest, "spent_on must be YYYY-MM-DD")
			return store.Expense{}, false
		}
	}

	var shares []split.Share
	var err error
	switch req.SplitType {
	case "equal":
		for _, id := range req.MemberIDs {
			if !validUUID(id) {
				writeError(w, http.StatusBadRequest, "member_ids must be member IDs")
				return store.Expense{}, false
			}
		}
		shares, err = split.Equal(req.AmountPence, req.MemberIDs)
	case "custom":
		for _, sh := range req.Shares {
			if !validUUID(sh.MemberID) {
				writeError(w, http.StatusBadRequest, "shares must use member IDs")
				return store.Expense{}, false
			}
		}
		shares, err = split.Custom(req.AmountPence, req.Shares)
	default:
		writeError(w, http.StatusBadRequest, `split_type must be "equal" or "custom"`)
		return store.Expense{}, false
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return store.Expense{}, false
	}

	return store.Expense{
		PaidBy:      req.PaidBy,
		Description: description,
		AmountPence: req.AmountPence,
		SplitType:   req.SplitType,
		SpentOn:     req.SpentOn,
		Shares:      shares,
	}, true
}

func cleanMemberNames(names []string) ([]string, string) {
	if len(names) < 2 || len(names) > maxMembers {
		return nil, "a group needs 2 to 20 members"
	}
	seen := map[string]bool{}
	cleaned := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || len(n) > 40 {
			return nil, "member names must be 1 to 40 characters"
		}
		key := strings.ToLower(n)
		if seen[key] {
			return nil, "member names must be unique"
		}
		seen[key] = true
		cleaned = append(cleaned, n)
	}
	return cleaned, ""
}
