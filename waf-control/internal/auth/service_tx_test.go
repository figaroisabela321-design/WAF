package auth

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gov-waf/waf-control/internal/httpx"
)

// fakeUserRepo simulates transactional CreateWithRoles / UpdateWithRoles.
// DOCUMENTATION: postgres integration tests are preferred; this fake proves
// the service+repo contract (COMMIT on success / ROLLBACK semantics) without Postgres.
type fakeUserRepo struct {
	mu                  sync.Mutex
	users               map[uuid.UUID]*User
	byName              map[string]uuid.UUID
	roles               map[uuid.UUID][]uuid.UUID
	failCreateAfterUser bool
	failUpdateRoles     bool
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:  map[uuid.UUID]*User{},
		byName: map[string]uuid.UUID{},
		roles:  map[uuid.UUID][]uuid.UUID{},
	}
}

func (f *fakeUserRepo) Create(ctx context.Context, u *User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *u
	f.users[u.ID] = &cp
	f.byName[u.Username] = u.ID
	return nil
}

func (f *fakeUserRepo) CreateWithRoles(ctx context.Context, u *User, roleIDs []uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failCreateAfterUser {
		// Simulate rollback: do not persist user
		return errors.New("role assign failed — rolled back")
	}
	cp := *u
	f.users[u.ID] = &cp
	f.byName[u.Username] = u.ID
	if roleIDs != nil {
		f.roles[u.ID] = append([]uuid.UUID{}, roleIDs...)
	}
	return nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u := f.users[id]
	if u == nil {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.byName[username]
	if !ok {
		return nil, nil
	}
	cp := *f.users[id]
	return &cp, nil
}

func (f *fakeUserRepo) Update(ctx context.Context, u *User) error {
	return f.UpdateWithRoles(ctx, u, nil)
}

func (f *fakeUserRepo) UpdateWithRoles(ctx context.Context, u *User, roleIDs *[]uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failUpdateRoles {
		return errors.New("update roles failed — rolled back")
	}
	if _, ok := f.users[u.ID]; !ok {
		return errors.New("not found")
	}
	cp := *u
	f.users[u.ID] = &cp
	if roleIDs != nil {
		f.roles[u.ID] = append([]uuid.UUID{}, *roleIDs...)
	}
	return nil
}

func (f *fakeUserRepo) List(ctx context.Context, offset, limit int) ([]User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) SetUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.roles[userID] = append([]uuid.UUID{}, roleIDs...)
	return nil
}
func (f *fakeUserRepo) GetUserRoleCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return []string{}, nil
}
func (f *fakeUserRepo) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return []string{}, nil
}
func (f *fakeUserRepo) CountAdmins(ctx context.Context) (int, error) { return 0, nil }

type fakeRoleRepo struct {
	codes map[string]uuid.UUID
}

func (f *fakeRoleRepo) ListRoles(ctx context.Context) ([]Role, error)             { return nil, nil }
func (f *fakeRoleRepo) ListPermissions(ctx context.Context) ([]Permission, error) { return nil, nil }
func (f *fakeRoleRepo) GetRoleByCode(ctx context.Context, code string) (*Role, error) {
	id, ok := f.codes[code]
	if !ok {
		return nil, nil
	}
	return &Role{ID: id, Code: code, Name: code}, nil
}
func (f *fakeRoleRepo) CreateRole(ctx context.Context, role *Role, permissionCodes []string) error {
	return nil
}
func (f *fakeRoleRepo) GetRoleIDsByCodes(ctx context.Context, codes []string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for _, c := range codes {
		id, ok := f.codes[c]
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

type noopHasher struct{}

func (noopHasher) Hash(p string) (string, error) { return "hash:" + p, nil }
func (noopHasher) Compare(hash, p string) error {
	if hash == "hash:"+p {
		return nil
	}
	return errors.New("mismatch")
}

type noopTokens struct{}

func (noopTokens) Issue(c *Claims, h int) (string, int, error) { return "tok", 3600, nil }
func (noopTokens) Parse(t string) (*Claims, error)             { return nil, errors.New("n/a") }

func TestCreateUserSuccessCommits(t *testing.T) {
	users := newFakeUserRepo()
	rid := uuid.New()
	roles := &fakeRoleRepo{codes: map[string]uuid.UUID{"viewer": rid}}
	svc := NewService(users, roles, noopHasher{}, noopTokens{}, 24)
	resp, err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "u1", Password: "password12345", RoleCodes: []string{"viewer"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, users.users, 1)
	assert.Equal(t, []uuid.UUID{rid}, users.roles[mustParse(t, resp.ID)])
}

func TestCreateUserInvalidRoleNoLeftover(t *testing.T) {
	users := newFakeUserRepo()
	roles := &fakeRoleRepo{codes: map[string]uuid.UUID{}}
	svc := NewService(users, roles, noopHasher{}, noopTokens{}, 24)
	_, err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "u2", Password: "password12345", RoleCodes: []string{"nope"},
	})
	require.Error(t, err)
	ae, ok := httpx.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, httpx.CodeValidation, ae.Code)
	assert.Empty(t, users.users) // validation before tx — no leftover
}

func TestCreateUserTxRollbackNoLeftover(t *testing.T) {
	users := newFakeUserRepo()
	users.failCreateAfterUser = true
	rid := uuid.New()
	roles := &fakeRoleRepo{codes: map[string]uuid.UUID{"viewer": rid}}
	svc := NewService(users, roles, noopHasher{}, noopTokens{}, 24)
	_, err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "u3", Password: "password12345", RoleCodes: []string{"viewer"},
	})
	require.Error(t, err)
	assert.Empty(t, users.users)
}

func TestUpdateUserFailureLeavesOriginal(t *testing.T) {
	users := newFakeUserRepo()
	rid := uuid.New()
	roles := &fakeRoleRepo{codes: map[string]uuid.UUID{"viewer": rid, "admin": uuid.New()}}
	svc := NewService(users, roles, noopHasher{}, noopTokens{}, 24)
	resp, err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "u4", Password: "password12345", Status: "enabled", RoleCodes: []string{"viewer"},
	})
	require.NoError(t, err)
	origHash := users.users[mustParse(t, resp.ID)].PasswordHash
	origRoles := append([]uuid.UUID{}, users.roles[mustParse(t, resp.ID)]...)

	users.failUpdateRoles = true
	pw := "newpassword123"
	_, err = svc.UpdateUser(context.Background(), resp.ID, &UpdateUserRequest{
		Password: &pw, RoleCodes: []string{"admin"},
	})
	require.Error(t, err)
	assert.Equal(t, origHash, users.users[mustParse(t, resp.ID)].PasswordHash)
	assert.Equal(t, origRoles, users.roles[mustParse(t, resp.ID)])
}

func TestCreateUserPasswordMinLength(t *testing.T) {
	users := newFakeUserRepo()
	roles := &fakeRoleRepo{codes: map[string]uuid.UUID{}}
	svc := NewService(users, roles, noopHasher{}, noopTokens{}, 24)
	_, err := svc.CreateUser(context.Background(), &CreateUserRequest{
		Username: "u5", Password: "short",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 12")
}

func mustParse(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	require.NoError(t, err)
	return id
}
