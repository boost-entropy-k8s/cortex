package ruler

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/ruler/rulespb"
	"github.com/cortexproject/cortex/pkg/ruler/rulestore"
)

type mockRuleStore struct {
	rules    map[string]rulespb.RuleGroupList
	errorMap map[string]error
	mtx      sync.Mutex
}

var (
	delim               = "/"
	interval, _         = time.ParseDuration("10s")
	mockRulesNamespaces = map[string]rulespb.RuleGroupList{
		"user1": {
			&rulespb.RuleGroupDesc{
				Name:      "group1",
				Namespace: "namespace1",
				User:      "user1",
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP_RULE",
						Expr:   "up",
					},
					{
						Alert: "UP_ALERT",
						Expr:  "up < 1",
					},
				},
				Interval: interval,
			},
			&rulespb.RuleGroupDesc{
				Name:      "fail",
				Namespace: "namespace2",
				User:      "user1",
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP2_RULE",
						Expr:   "up",
					},
					{
						Alert: "UP2_ALERT",
						Expr:  "up < 1",
					},
				},
				Interval: interval,
			},
		},
	}
	mockRules = map[string]rulespb.RuleGroupList{
		"user1": {
			&rulespb.RuleGroupDesc{
				Name:      "group1",
				Namespace: "namespace1",
				User:      "user1",
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP_RULE",
						Expr:   "up",
					},
					{
						Alert: "UP_ALERT",
						Expr:  "up < 1",
					},
				},
				Interval: interval,
			},
		},
		"user2": {
			&rulespb.RuleGroupDesc{
				Name:      "group1",
				Namespace: "namespace1",
				User:      "user2",
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP_RULE",
						Expr:   "up",
					},
				},
				Interval: interval,
			},
		},
	}

	mockSpecialCharRules = map[string]rulespb.RuleGroupList{
		"user1": {
			&rulespb.RuleGroupDesc{
				Name:      ")(_+?/|group1+/?",
				Namespace: ")(_+?/|namespace1+/?",
				User:      "user1",
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP_RULE",
						Expr:   "up",
					},
					{
						Alert: "UP_ALERT",
						Expr:  "up < 1",
					},
				},
				Interval: interval,
			},
		},
	}
	mockRulesLimit = map[string]rulespb.RuleGroupList{
		"user1": {
			&rulespb.RuleGroupDesc{
				Name:      "group1",
				Namespace: "namespace1",
				User:      "user1",
				Limit:     5,
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP_RULE",
						Expr:   "up",
					},
					{
						Alert: "UP_ALERT",
						Expr:  "up < 1",
					},
				},
				Interval: interval,
			},
		},
	}
	queryOffset          = 2 * time.Minute
	mockRulesQueryOffset = map[string]rulespb.RuleGroupList{
		"user1": {
			&rulespb.RuleGroupDesc{
				Name:      "group1",
				Namespace: "namespace1",
				User:      "user1",
				Limit:     5,
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP_RULE",
						Expr:   "up",
					},
					{
						Alert: "UP_ALERT",
						Expr:  "up < 1",
					},
				},
				Interval: interval,
			},
		},
		"user2": {
			&rulespb.RuleGroupDesc{
				Name:      "group1",
				Namespace: "namespace1",
				User:      "user2",
				Rules: []*rulespb.RuleDesc{
					{
						Record: "UP_RULE",
						Expr:   "up",
					},
				},
				Interval:    interval,
				QueryOffset: &queryOffset,
			},
		},
	}
)

func newMockRuleStore(rules map[string]rulespb.RuleGroupList, errorMap map[string]error) *mockRuleStore {
	return &mockRuleStore{
		rules:    rules,
		errorMap: errorMap,
	}
}

func (m *mockRuleStore) ListAllUsers(_ context.Context) ([]string, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var result []string
	for u := range m.rules {
		result = append(result, u)
	}
	return result, nil
}

func (m *mockRuleStore) ListAllRuleGroups(_ context.Context) (map[string]rulespb.RuleGroupList, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	result := make(map[string]rulespb.RuleGroupList)
	for k, v := range m.rules {
		for _, r := range v {
			result[k] = append(result[k], &rulespb.RuleGroupDesc{
				Namespace: r.Namespace,
				Name:      r.Name,
				User:      k,
				Interval:  r.Interval,
			})
		}
	}

	return result, nil
}

func (m *mockRuleStore) ListRuleGroupsForUserAndNamespace(_ context.Context, userID, namespace string) (rulespb.RuleGroupList, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var result rulespb.RuleGroupList
	for _, r := range m.rules[userID] {
		if namespace != "" && namespace != r.Namespace {
			continue
		}

		result = append(result, &rulespb.RuleGroupDesc{
			Namespace: r.Namespace,
			Name:      r.Name,
			User:      userID,
			Interval:  r.Interval,
		})
	}
	return result, nil
}

func (m *mockRuleStore) LoadRuleGroups(ctx context.Context, groupsToLoad map[string]rulespb.RuleGroupList) (map[string]rulespb.RuleGroupList, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	result := make(map[string]rulespb.RuleGroupList, len(groupsToLoad))
	var err error

	gm := make(map[string]*rulespb.RuleGroupDesc)
	for _, gs := range m.rules {
		for _, gr := range gs {
			user, namespace, name := gr.GetUser(), gr.GetNamespace(), gr.GetName()
			key := user + delim + base64.URLEncoding.EncodeToString([]byte(namespace)) + delim + base64.URLEncoding.EncodeToString([]byte(name))
			gm[key] = gr
		}
	}

	for _, gs := range groupsToLoad {
		for _, gr := range gs {
			user, namespace, name := gr.GetUser(), gr.GetNamespace(), gr.GetName()
			if e, ok := m.errorMap[user]; ok {
				err = e
				continue
			}
			key := user + delim + base64.URLEncoding.EncodeToString([]byte(namespace)) + delim + base64.URLEncoding.EncodeToString([]byte(name))
			mgr, ok := gm[key]
			if !ok {
				return nil, fmt.Errorf("failed to get rule group user %s", gr.GetUser())
			}
			*gr = *mgr
			result[user] = append(result[user], gr)
		}
	}
	return result, err
}

func (m *mockRuleStore) GetRuleGroup(_ context.Context, userID string, namespace string, group string) (*rulespb.RuleGroupDesc, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	userRules, exists := m.rules[userID]
	if !exists {
		return nil, rulestore.ErrUserNotFound
	}

	if namespace == "" {
		return nil, rulestore.ErrGroupNamespaceNotFound
	}

	for _, rg := range userRules {
		if rg.Namespace == namespace && rg.Name == group {
			return rg, nil
		}
	}

	return nil, rulestore.ErrGroupNotFound
}

func (m *mockRuleStore) SetRuleGroup(ctx context.Context, userID string, namespace string, group *rulespb.RuleGroupDesc) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	userRules, exists := m.rules[userID]
	if !exists {
		userRules = rulespb.RuleGroupList{}
		m.rules[userID] = userRules
	}

	if namespace == "" {
		return rulestore.ErrGroupNamespaceNotFound
	}

	for i, rg := range userRules {
		if rg.Namespace == namespace && rg.Name == group.Name {
			userRules[i] = group
			return nil
		}
	}

	m.rules[userID] = append(userRules, group)
	return nil
}

func (m *mockRuleStore) DeleteRuleGroup(ctx context.Context, userID string, namespace string, group string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	userRules, exists := m.rules[userID]
	if !exists {
		userRules = rulespb.RuleGroupList{}
		m.rules[userID] = userRules
	}

	if namespace == "" {
		return rulestore.ErrGroupNamespaceNotFound
	}

	for i, rg := range userRules {
		if rg.Namespace == namespace && rg.Name == group {
			m.rules[userID] = append(userRules[:i], userRules[:i+1]...)
			return nil
		}
	}

	return nil
}

func (m *mockRuleStore) DeleteNamespace(ctx context.Context, userID, namespace string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	userRules, exists := m.rules[userID]
	if !exists {
		userRules = rulespb.RuleGroupList{}
		m.rules[userID] = userRules
	}

	if namespace == "" {
		return rulestore.ErrGroupNamespaceNotFound
	}

	for i, rg := range userRules {
		if rg.Namespace == namespace {

			// Only here to assert on partial failures.
			if rg.Name == "fail" {
				return fmt.Errorf("unable to delete rg")
			}

			m.rules[userID] = append(userRules[:i], userRules[i+1:]...)
		}
	}

	return nil
}
