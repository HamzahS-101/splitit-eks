package store

import (
	"context"
	"errors"
	"time"

	"github.com/HamzahS-101/splitit-eks/api/internal/split"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("already exists")
	ErrInvalidReference = errors.New("member does not belong to this group")
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

type Member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Group struct {
	ID        string    `json:"-"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	Members   []Member  `json:"members"`
}

type Expense struct {
	ID          string        `json:"id"`
	PaidBy      string        `json:"paid_by"`
	Description string        `json:"description"`
	AmountPence int64         `json:"amount_pence"`
	SplitType   string        `json:"split_type"`
	SpentOn     string        `json:"spent_on"`
	CreatedAt   time.Time     `json:"created_at"`
	Shares      []split.Share `json:"shares"`
}

type Settlement struct {
	ID          string    `json:"id"`
	From        string    `json:"from_member"`
	To          string    `json:"to_member"`
	AmountPence int64     `json:"amount_pence"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Store) CreateGroup(ctx context.Context, slug, name string, members []string) (Group, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Group{}, err
	}
	defer tx.Rollback(ctx)

	g := Group{Slug: slug, Name: name, Members: []Member{}}
	err = tx.QueryRow(ctx,
		`INSERT INTO groups (slug, name) VALUES ($1, $2) RETURNING id::text, currency, created_at`,
		slug, name,
	).Scan(&g.ID, &g.Currency, &g.CreatedAt)
	if err != nil {
		return Group{}, mapErr(err)
	}

	for _, m := range members {
		var mem Member
		err := tx.QueryRow(ctx,
			`INSERT INTO members (group_id, name) VALUES ($1, $2) RETURNING id::text, name`,
			g.ID, m,
		).Scan(&mem.ID, &mem.Name)
		if err != nil {
			return Group{}, mapErr(err)
		}
		g.Members = append(g.Members, mem)
	}

	return g, tx.Commit(ctx)
}

func (s *Store) GroupID(ctx context.Context, slug string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id::text FROM groups WHERE slug = $1`, slug).Scan(&id)
	return id, mapErr(err)
}

func (s *Store) GetGroup(ctx context.Context, slug string) (Group, error) {
	g := Group{Slug: slug, Members: []Member{}}
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, name, currency, created_at FROM groups WHERE slug = $1`, slug,
	).Scan(&g.ID, &g.Name, &g.Currency, &g.CreatedAt)
	if err != nil {
		return Group{}, mapErr(err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id::text, name FROM members WHERE group_id = $1 ORDER BY name`, g.ID)
	if err != nil {
		return Group{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return Group{}, err
		}
		g.Members = append(g.Members, m)
	}
	return g, rows.Err()
}

func (s *Store) AddMember(ctx context.Context, groupID, name string) (Member, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Member{}, err
	}
	defer tx.Rollback(ctx)

	var m Member
	err = tx.QueryRow(ctx,
		`INSERT INTO members (group_id, name) VALUES ($1, $2) RETURNING id::text, name`,
		groupID, name,
	).Scan(&m.ID, &m.Name)
	if err != nil {
		return Member{}, mapErr(err)
	}
	if err := touch(ctx, tx, groupID); err != nil {
		return Member{}, err
	}
	return m, tx.Commit(ctx)
}

func (s *Store) ListExpenses(ctx context.Context, groupID string) ([]Expense, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, paid_by::text, description, amount_pence, split_type, spent_on::text, created_at
		FROM expenses
		WHERE group_id = $1
		ORDER BY spent_on DESC, created_at DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := []Expense{}
	index := map[string]int{}
	for rows.Next() {
		e := Expense{Shares: []split.Share{}}
		if err := rows.Scan(&e.ID, &e.PaidBy, &e.Description, &e.AmountPence, &e.SplitType, &e.SpentOn, &e.CreatedAt); err != nil {
			return nil, err
		}
		index[e.ID] = len(expenses)
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	shareRows, err := s.pool.Query(ctx,
		`SELECT expense_id::text, member_id::text, amount_pence FROM expense_shares WHERE group_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer shareRows.Close()
	for shareRows.Next() {
		var expenseID string
		var sh split.Share
		if err := shareRows.Scan(&expenseID, &sh.MemberID, &sh.AmountPence); err != nil {
			return nil, err
		}
		if i, ok := index[expenseID]; ok {
			expenses[i].Shares = append(expenses[i].Shares, sh)
		}
	}
	return expenses, shareRows.Err()
}

func (s *Store) CreateExpense(ctx context.Context, groupID string, e Expense) (Expense, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Expense{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO expenses (group_id, paid_by, description, amount_pence, split_type, spent_on)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6::date, current_date))
		RETURNING id::text, spent_on::text, created_at`,
		groupID, e.PaidBy, e.Description, e.AmountPence, e.SplitType, nullIfEmpty(e.SpentOn),
	).Scan(&e.ID, &e.SpentOn, &e.CreatedAt)
	if err != nil {
		return Expense{}, mapErr(err)
	}

	if err := insertShares(ctx, tx, groupID, e.ID, e.Shares); err != nil {
		return Expense{}, err
	}
	if err := touch(ctx, tx, groupID); err != nil {
		return Expense{}, err
	}
	return e, tx.Commit(ctx)
}

func (s *Store) UpdateExpense(ctx context.Context, groupID string, e Expense) (Expense, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Expense{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE expenses
		SET paid_by = $3, description = $4, amount_pence = $5, split_type = $6,
		    spent_on = COALESCE($7::date, spent_on)
		WHERE id = $1 AND group_id = $2
		RETURNING spent_on::text, created_at`,
		e.ID, groupID, e.PaidBy, e.Description, e.AmountPence, e.SplitType, nullIfEmpty(e.SpentOn),
	).Scan(&e.SpentOn, &e.CreatedAt)
	if err != nil {
		return Expense{}, mapErr(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM expense_shares WHERE expense_id = $1`, e.ID); err != nil {
		return Expense{}, err
	}
	if err := insertShares(ctx, tx, groupID, e.ID, e.Shares); err != nil {
		return Expense{}, err
	}
	if err := touch(ctx, tx, groupID); err != nil {
		return Expense{}, err
	}
	return e, tx.Commit(ctx)
}

func (s *Store) DeleteExpense(ctx context.Context, groupID, id string) error {
	return s.deleteInGroup(ctx, `DELETE FROM expenses WHERE id = $1 AND group_id = $2`, groupID, id)
}

func (s *Store) ListSettlements(ctx context.Context, groupID string) ([]Settlement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, from_member::text, to_member::text, amount_pence, created_at
		FROM settlements
		WHERE group_id = $1
		ORDER BY created_at DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settlements := []Settlement{}
	for rows.Next() {
		var st Settlement
		if err := rows.Scan(&st.ID, &st.From, &st.To, &st.AmountPence, &st.CreatedAt); err != nil {
			return nil, err
		}
		settlements = append(settlements, st)
	}
	return settlements, rows.Err()
}

func (s *Store) CreateSettlement(ctx context.Context, groupID string, st Settlement) (Settlement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Settlement{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO settlements (group_id, from_member, to_member, amount_pence)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, created_at`,
		groupID, st.From, st.To, st.AmountPence,
	).Scan(&st.ID, &st.CreatedAt)
	if err != nil {
		return Settlement{}, mapErr(err)
	}
	if err := touch(ctx, tx, groupID); err != nil {
		return Settlement{}, err
	}
	return st, tx.Commit(ctx)
}

func (s *Store) DeleteSettlement(ctx context.Context, groupID, id string) error {
	return s.deleteInGroup(ctx, `DELETE FROM settlements WHERE id = $1 AND group_id = $2`, groupID, id)
}

func (s *Store) Balances(ctx context.Context, groupID string) ([]split.Balance, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT m.id::text, m.name,
		       COALESCE((SELECT sum(e.amount_pence) FROM expenses e WHERE e.paid_by = m.id), 0)::bigint,
		       COALESCE((SELECT sum(sh.amount_pence) FROM expense_shares sh WHERE sh.member_id = m.id), 0)::bigint,
		       COALESCE((SELECT sum(x.amount_pence) FROM settlements x WHERE x.from_member = m.id), 0)::bigint,
		       COALESCE((SELECT sum(x.amount_pence) FROM settlements x WHERE x.to_member = m.id), 0)::bigint
		FROM members m
		WHERE m.group_id = $1
		ORDER BY m.name`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := []split.Balance{}
	for rows.Next() {
		var b split.Balance
		var sent, received int64
		if err := rows.Scan(&b.MemberID, &b.Name, &b.PaidPence, &b.SharePence, &sent, &received); err != nil {
			return nil, err
		}
		b.NetPence = b.PaidPence - b.SharePence + sent - received
		balances = append(balances, b)
	}
	return balances, rows.Err()
}

func (s *Store) DeleteInactiveGroups(ctx context.Context, days int) (int64, error) {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM groups WHERE last_activity_at < now() - make_interval(days => $1)`, days)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (s *Store) deleteInGroup(ctx context.Context, query, groupID, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, query, id, groupID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if err := touch(ctx, tx, groupID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertShares(ctx context.Context, tx pgx.Tx, groupID, expenseID string, shares []split.Share) error {
	for _, sh := range shares {
		_, err := tx.Exec(ctx,
			`INSERT INTO expense_shares (expense_id, group_id, member_id, amount_pence) VALUES ($1, $2, $3, $4)`,
			expenseID, groupID, sh.MemberID, sh.AmountPence)
		if err != nil {
			return mapErr(err)
		}
	}
	return nil
}

func touch(ctx context.Context, tx pgx.Tx, groupID string) error {
	_, err := tx.Exec(ctx, `UPDATE groups SET last_activity_at = now() WHERE id = $1`, groupID)
	return err
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrConflict
		case "23503":
			return ErrInvalidReference
		}
	}
	return err
}
