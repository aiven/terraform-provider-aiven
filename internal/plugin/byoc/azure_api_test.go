package byoc_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
)

const environmentsPath = "/v1/organization/org1234567890/custom-cloud-environments"

// Terraform and the generated HTTP client share this API across CLI invocations.
// Each test chooses when provisioning and deletion reach a terminal state.
type cloudAPI struct {
	mu             sync.Mutex
	environments   map[string]map[string]any
	nextID         int
	mutations      []string
	provisionState string // defaults to active
	deleteState    string // defaults to a GET 404 after DELETE
	beforeDelete   func()
}

func (a *cloudAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	if path != environmentsPath && !strings.HasPrefix(path, environmentsPath+"/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		a.mutations = append(a.mutations, r.Method+" "+path)
	}
	var body map[string]any
	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if path == environmentsPath && r.Method == http.MethodPost {
		id := a.add(body)
		writeEnvironment(w, a.environments[id])
		return
	}

	id, action, _ := strings.Cut(strings.TrimPrefix(path, environmentsPath+"/"), "/")
	env := a.environments[id]
	if env == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"Unknown environment","status":404}]}`))
		return
	}
	switch {
	case r.Method == http.MethodPost && action == "provision":
		if env["state"] != "draft" {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"errors":[{"message":"Already provisioned","status":409}]}`))
			return
		}
		maps.Copy(env, body)
		env["state"] = "active"
		if a.provisionState != "" {
			env["state"] = a.provisionState
		}
	case r.Method == http.MethodPut && action == "":
		// An omitted field is unchanged. An explicit empty tags object clears it.
		maps.Copy(env, body)
	case r.Method == http.MethodDelete && action == "":
		if a.beforeDelete != nil {
			a.beforeDelete()
		}
		if a.deleteState == "" {
			delete(a.environments, id)
		} else {
			env["state"] = a.deleteState
		}
		_, _ = w.Write([]byte("{}"))
		return
	case r.Method == http.MethodGet && action == "":
	default:
		http.NotFound(w, r)
		return
	}
	writeEnvironment(w, env)
}

func writeEnvironment(w http.ResponseWriter, env map[string]any) {
	_ = json.NewEncoder(w).Encode(map[string]any{"custom_cloud_environment": env})
}

// add is called during fixture setup or under mu by the create endpoint.
func (a *cloudAPI) add(env map[string]any) string {
	if a.environments == nil {
		a.environments = make(map[string]map[string]any)
	}
	a.nextID++
	id := fmt.Sprintf("00000000-0000-4000-8000-%012d", a.nextID)
	env["custom_cloud_environment_id"] = id
	if _, ok := env["state"]; !ok {
		env["state"] = "draft"
	}
	if _, ok := env["contact_emails"]; !ok {
		env["contact_emails"] = []map[string]string{{"email": "admin@example.com", "role": "Admin"}}
	}
	if _, ok := env["tags"]; !ok {
		env["tags"] = map[string]any{}
	}
	a.environments[id] = env
	return id
}

func (a *cloudAPI) set(id, key string, value any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.environments[id][key] = value
}

func (a *cloudAPI) get(id string) map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()
	return maps.Clone(a.environments[id])
}

func (a *cloudAPI) allowDelete() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.deleteState = ""
}

func (a *cloudAPI) writes() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return slices.Clone(a.mutations)
}
