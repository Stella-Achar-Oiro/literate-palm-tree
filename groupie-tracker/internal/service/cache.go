// internal/service/cache.go
package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"groupie-tracker/internal/models"
)

type CacheService struct {
	data      models.Datas
	expiresAt time.Time
	duration  time.Duration
	mutex     sync.RWMutex
}

func NewCacheService(duration time.Duration) *CacheService {
	return &CacheService{
		duration: duration,
	}
}

func (c *CacheService) RefreshCache() error {
	var newData models.Datas
	if err := c.fetchAllData(&newData); err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = newData
	c.expiresAt = time.Now().Add(c.duration)

	return nil
}

func (c *CacheService) GetCachedData() (models.Datas, error) {
	c.mutex.RLock()
	if time.Now().Before(c.expiresAt) {
		defer c.mutex.RUnlock()
		return c.data, nil
	}
	c.mutex.RUnlock()

	if err := c.RefreshCache(); err != nil {
		return models.Datas{}, fmt.Errorf("failed to refresh cache: %w", err)
	}

	return c.data, nil
}

func (c *CacheService) fetchAllData(data *models.Datas) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	wg.Add(4)
	go c.fetchData(models.ArtistsAPI, &data.ArtistsData, &wg, errChan)
	go c.fetchData(models.LocationsAPI, &data.LocationsData, &wg, errChan)
	go c.fetchData(models.DatesAPI, &data.DatesData, &wg, errChan)
	go c.fetchData(models.RelationsAPI, &data.RelationsData, &wg, errChan)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CacheService) fetchData(url string, target interface{}, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		errChan <- fmt.Errorf("failed to fetch data from %s: %w", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("received non-200 status code from %s: %d", url, resp.StatusCode)
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		errChan <- fmt.Errorf("failed to decode data from %s: %w", url, err)
	}
}
