package pocketbase

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// collectionInfo holds the result of resolving a collection from PocketBase.
type collectionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// getCollectionID fetches the internal PocketBase ID for a collection by name.
func (r *Repository) getCollectionID(name string) (string, error) {
	endpoint := fmt.Sprintf("api/collections/%s", name)
	resp, err := r.client.SendRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("collection %s not found (status %d)", name, resp.StatusCode)
	}
	var info collectionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.ID, nil
}

// ensureCollection creates a collection if it doesn't already exist.
// Returns the internal PocketBase ID of the (existing or newly created) collection.
func (r *Repository) ensureCollection(schema map[string]interface{}) (string, error) {
	name := schema["name"].(string)

	endpoint := fmt.Sprintf("api/collections/%s", name)
	resp, err := r.client.SendRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to check collection %s: %w", name, err)
	}

	if resp.StatusCode == http.StatusOK {
		// Collection already exists — read its ID and return.
		var info collectionInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("failed to decode existing collection %s: %w", name, err)
		}
		resp.Body.Close()
		log.Printf("⏭️  [MIGRATE] Collection %q already exists (id=%s). Skipping.", name, info.ID)
		return info.ID, nil
	}
	resp.Body.Close()

	// Create the collection.
	log.Printf("📦 [MIGRATE] Creating collection %q...", name)
	createResp, err := r.client.SendRequest("POST", "api/collections", schema)
	if err != nil {
		return "", fmt.Errorf("failed to POST collection %s: %w", name, err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusOK && createResp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(createResp.Body)
		return "", fmt.Errorf("failed to create collection %s (status %d): %s", name, createResp.StatusCode, string(bodyBytes))
	}

	var created collectionInfo
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		return "", fmt.Errorf("failed to decode created collection %s: %w", name, err)
	}

	log.Printf("✅ [MIGRATE] Collection %q created (id=%s).", name, created.ID)
	return created.ID, nil
}

// patchRelationFields issues a PATCH to add/update relation fields to an existing collection.
func (r *Repository) patchRelationFields(collectionID string, fields []map[string]interface{}) error {
	patch := map[string]interface{}{
		"schema": fields,
	}
	endpoint := fmt.Sprintf("api/collections/%s", collectionID)
	resp, err := r.client.SendRequest("PATCH", endpoint, patch)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("patch failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// MigrateCollections ensures that all required PocketBase collections exist.
// Uses a two-pass strategy:
//  1. Create all base collections without relation fields.
//  2. Resolve real PocketBase IDs and patch in relation fields.
func (r *Repository) MigrateCollections() error {
	log.Println("🛠️  [MIGRATE] Starting PocketBase Schema Auto-Migration...")

	// ── PASS 1: Create base collections (no relations) ──────────────────────

	baseSchemas := []map[string]interface{}{
		{
			"name": "guilds",
			"type": "base",
			"schema": []map[string]interface{}{
				{"name": "discord_id", "type": "text", "required": true},
				{"name": "name", "type": "text"},
				{"name": "status", "type": "select", "options": map[string]interface{}{"values": []string{"active", "archived"}}},
				{"name": "announcement_channel_id", "type": "text"},
				{"name": "squad_roles", "type": "json"},
				{"name": "mentor_roles", "type": "json"},
				{"name": "skill_roles", "type": "json"},
			},
		},
		{
			"name": "roles",
			"type": "base",
			"schema": []map[string]interface{}{
				{"name": "discord_id", "type": "text", "required": true},
				{"name": "name", "type": "text"},
				{"name": "shift", "type": "text"},
				{"name": "check_in_time", "type": "text"},
				{"name": "checkout_cooldown", "type": "number"},
				{"name": "is_monitored", "type": "bool"},
				{"name": "is_active", "type": "bool"},
				{"name": "is_staff", "type": "bool"},
				{"name": "squad_channel_id", "type": "text"},
			},
		},
		{
			"name": "managers",
			"type": "base",
			"schema": []map[string]interface{}{
				{"name": "discord_id", "type": "text", "required": true},
				{"name": "name", "type": "text"},
				{"name": "role", "type": "select", "options": map[string]interface{}{"values": []string{"admin", "staff", "mentor"}}},
			},
		},
		{
			"name": "students",
			"type": "base",
			"schema": []map[string]interface{}{
				{"name": "discord_id", "type": "text", "required": true},
				{"name": "username", "type": "text"},
				{"name": "nickname", "type": "text"},
				{"name": "status", "type": "select", "options": map[string]interface{}{"values": []string{"active", "inactive"}}},
			},
		},
		{
			"name": "meetings",
			"type": "base",
			"schema": []map[string]interface{}{
				{"name": "started_at", "type": "date"},
				{"name": "ended_at", "type": "date"},
				{"name": "status", "type": "select", "options": map[string]interface{}{"values": []string{"ongoing", "finished", "cancelled"}}},
			},
		},
	}

	// Track IDs for pass 2
	ids := make(map[string]string)

	for _, schema := range baseSchemas {
		id, err := r.ensureCollection(schema)
		if err != nil {
			return err
		}
		ids[schema["name"].(string)] = id
	}

	// ── PASS 2: Patch in relation fields using resolved IDs ──────────────────

	log.Println("🔗 [MIGRATE] Patching relation fields...")

	// roles → guilds
	rolesID := ids["roles"]
	guildsID := ids["guilds"]
	managersID := ids["managers"]
	studentsID := ids["students"]
	meetingsID := ids["meetings"]

	// roles.guild_id → guilds
	if err := r.addRelationFieldIfMissing("roles", rolesID, "guild_id", guildsID, 1); err != nil {
		log.Printf("⚠️  [MIGRATE] roles.guild_id patch: %v", err)
	}

	// managers.guilds → guilds (multi)
	if err := r.addRelationFieldIfMissing("managers", managersID, "guilds", guildsID, 0); err != nil {
		log.Printf("⚠️  [MIGRATE] managers.guilds patch: %v", err)
	}

	// students.role_id → roles
	if err := r.addRelationFieldIfMissing("students", studentsID, "role_id", rolesID, 1); err != nil {
		log.Printf("⚠️  [MIGRATE] students.role_id patch: %v", err)
	}
	// students.secondary_roles → roles (multi)
	if err := r.addRelationFieldIfMissing("students", studentsID, "secondary_roles", rolesID, 0); err != nil {
		log.Printf("⚠️  [MIGRATE] students.secondary_roles patch: %v", err)
	}
	// students.guild_id → guilds
	if err := r.addRelationFieldIfMissing("students", studentsID, "guild_id", guildsID, 1); err != nil {
		log.Printf("⚠️  [MIGRATE] students.guild_id patch: %v", err)
	}

	// meetings relations
	if err := r.addRelationFieldIfMissing("meetings", meetingsID, "student_id", studentsID, 1); err != nil {
		log.Printf("⚠️  [MIGRATE] meetings.student_id patch: %v", err)
	}
	if err := r.addRelationFieldIfMissing("meetings", meetingsID, "manager_id", managersID, 1); err != nil {
		log.Printf("⚠️  [MIGRATE] meetings.manager_id patch: %v", err)
	}
	if err := r.addRelationFieldIfMissing("meetings", meetingsID, "guild_id", guildsID, 1); err != nil {
		log.Printf("⚠️  [MIGRATE] meetings.guild_id patch: %v", err)
	}

	log.Println("✅ [MIGRATE] PocketBase Schema Auto-Migration complete.")
	return nil
}

// addRelationFieldIfMissing checks whether a relation field already exists on the collection,
// and if not, patches it in. maxSelect=0 means unlimited (multi-relation).
func (r *Repository) addRelationFieldIfMissing(collName, collID, fieldName, targetCollID string, maxSelect int) error {
	// Fetch current schema
	endpoint := fmt.Sprintf("api/collections/%s", collID)
	resp, err := r.client.SendRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data struct {
		Schema []map[string]interface{} `json:"schema"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	// Check if field already exists
	for _, f := range data.Schema {
		if f["name"] == fieldName {
			log.Printf("⏭️  [MIGRATE] %s.%s already exists. Skipping.", collName, fieldName)
			return nil
		}
	}

	// Build new field definition
	opts := map[string]interface{}{
		"collectionId": targetCollID,
	}
	if maxSelect > 0 {
		opts["maxSelect"] = maxSelect
	}

	newField := map[string]interface{}{
		"name":    fieldName,
		"type":    "relation",
		"options": opts,
	}

	allFields := append(data.Schema, newField)
	patch := map[string]interface{}{"schema": allFields}

	patchResp, err := r.client.SendRequest("PATCH", fmt.Sprintf("api/collections/%s", collID), patch)
	if err != nil {
		return err
	}
	defer patchResp.Body.Close()

	if patchResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(patchResp.Body)
		return fmt.Errorf("PATCH %s.%s failed (status %d): %s", collName, fieldName, patchResp.StatusCode, string(bodyBytes))
	}

	log.Printf("✅ [MIGRATE] %s.%s relation field added.", collName, fieldName)
	return nil
}
