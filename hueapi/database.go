package hueapi

import (
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
)

func (h *HueApi) SetupBolt() error {
	return h.boltDb.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("lights"))
		return err
	})
}

func (h *HueApi) GetLights() (map[string]*LightInfo, error) {
	result := make(map[string]*LightInfo)
	err := h.boltDb.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("lights"))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}
		return bucket.ForEach(func(k, v []byte) error {
			light := &LightInfo{}
			if err := json.Unmarshal(v, light); err != nil {
				return err
			}
			result[string(k)] = light
			return nil
		})
	})
	return result, err
}

func (h *HueApi) GetLight(id string) (*LightInfo, error) {
	result := new(LightInfo)
	err := h.boltDb.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("lights"))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}
		v := bucket.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("light %s not found", id)
		}
		return json.Unmarshal(v, result)
	})
	return result, err
}

func (h *HueApi) PutLight(light *LightInfo) (*LightInfo, string, error) {
	var lightId string
	err := h.boltDb.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("lights"))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}

		// Check for duplicate names
		nameExists := false
		err := bucket.ForEach(func(k, v []byte) error {
			existing := &LightInfo{}
			if err := json.Unmarshal(v, existing); err != nil {
				return err
			}
			if existing.Name == light.Name {
				nameExists = true
			}
			return nil
		})
		if err != nil {
			return err
		}
		if nameExists {
			return fmt.Errorf("light with name %s already exists", light.Name)
		}

		// Generate unique ID starting from 1
		for i := 1; ; i++ {
			id := fmt.Sprintf("%d", i)
			if bucket.Get([]byte(id)) == nil {
				lightId = id
				break
			}
		}

		light.Defaults(lightId)

		data, err := json.Marshal(light)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(lightId), data)
	})
	return light, lightId, err
}

func (h *HueApi) DeleteLight(lightId string) error {
	return h.boltDb.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("lights"))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}
		if bucket.Get([]byte(lightId)) == nil {
			return fmt.Errorf("light %s not found", lightId)
		}
		return bucket.Delete([]byte(lightId))
	})
}

func (h *HueApi) UpdateLight(lightId string, light *LightInfo) error {
	return h.boltDb.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("lights"))
		if bucket == nil {
			return fmt.Errorf("bucket does not exist")
		}
		if bucket.Get([]byte(lightId)) == nil {
			return fmt.Errorf("light %s not found", lightId)
		}
		jsonData, err := json.Marshal(light)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(lightId), jsonData)
	})
}
