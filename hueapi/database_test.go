package hueapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func setupTestDB(t *testing.T) (*HueApi, string) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "hueapi-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to open test database: %v", err)
	}

	h := &HueApi{boltDb: db}
	err = h.SetupBolt()
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to setup bolt DB: %v", err)
	}

	return h, tmpDir
}

func teardownTestDB(h *HueApi, tmpDir string) {
	if h.boltDb != nil {
		h.boltDb.Close()
	}
	os.RemoveAll(tmpDir)
}

func TestSetupBolt(t *testing.T) {
	h, tmpDir := setupTestDB(t)
	defer teardownTestDB(h, tmpDir)

	// Verify the lights bucket exists
	err := h.boltDb.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("lights"))
		if bucket == nil {
			t.Error("Lights bucket was not created")
		}
		return nil
	})
	assert.NoError(t, err)
}

func TestPutAndGetLight(t *testing.T) {
	h, tmpDir := setupTestDB(t)
	defer teardownTestDB(h, tmpDir)

	// Test putting a new light
	light := &LightInfo{
		Name: "Test Light",
		Type: "Extended color light",
	}

	savedLight, lightId, err := h.PutLight(light)
	assert.NoError(t, err)
	assert.NotEmpty(t, lightId)
	assert.Equal(t, "1", lightId)
	assert.Equal(t, "Test Light", savedLight.Name)
	assert.Equal(t, "Extended color light", savedLight.Type)
	assert.Equal(t, "LCT007", savedLight.ModelID)
	assert.Equal(t, "00:17:88:01:00:bd:c7:b9-01", savedLight.UniqueID)
	assert.Equal(t, "66012040", savedLight.SWVersion)
	assert.Equal(t, "none", savedLight.State.Alert)

	// Test getting the light back
	fetchedLight, err := h.GetLight(lightId)
	assert.NoError(t, err)
	assert.Equal(t, savedLight.Name, fetchedLight.Name)
	assert.Equal(t, savedLight.Type, fetchedLight.Type)
	assert.Equal(t, "66012040", fetchedLight.SWVersion)
	assert.Equal(t, "none", fetchedLight.State.Alert)

}

func TestGetNonExistentLight(t *testing.T) {
	h, tmpDir := setupTestDB(t)
	defer teardownTestDB(h, tmpDir)

	_, err := h.GetLight("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "light nonexistent not found")
}

func TestPutLightWithDuplicateName(t *testing.T) {
	h, tmpDir := setupTestDB(t)
	defer teardownTestDB(h, tmpDir)

	light1 := &LightInfo{
		Name: "Test Light",
		Type: "Extended color light",
	}

	_, _, err := h.PutLight(light1)
	assert.NoError(t, err)

	light2 := &LightInfo{
		Name: "Test Light", // Same name as light1
		Type: "Extended color light",
	}

	_, _, err = h.PutLight(light2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "light with name Test Light already exists")
}

func TestGetLights(t *testing.T) {
	h, tmpDir := setupTestDB(t)
	defer teardownTestDB(h, tmpDir)

	// Add multiple lights
	lights := []*LightInfo{
		{Name: "Light 1", Type: "Extended color light"},
		{Name: "Light 2", Type: "Dimmable light"},
		{Name: "Light 3", Type: "Color light"},
	}

	for _, light := range lights {
		_, _, err := h.PutLight(light)
		assert.NoError(t, err)
	}

	// Test getting all lights
	allLights, err := h.GetLights()
	assert.NoError(t, err)
	assert.Equal(t, len(lights), len(allLights))

	// Verify all lights are present with correct IDs
	lightsMap := make(map[string]*LightInfo)
	for id, light := range allLights {
		lightsMap[light.Name] = light
		expectedId := string(rune('0' + len(lightsMap))) // Convert number to string
		assert.Equal(t, expectedId, id, "Light %s should have ID %s", light.Name, expectedId)
	}

	for _, light := range lights {
		assert.NotNil(t, lightsMap[light.Name], "Light %s should be present in results", light.Name)
	}
}
