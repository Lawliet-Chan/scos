package pkg

import "sync"

// 价格管理
type PriceManager struct {
	prices map[string]float64
	mutex  sync.RWMutex
}

func NewPriceManager() *PriceManager {
	return &PriceManager{
		prices: make(map[string]float64),
	}
}

func (pm *PriceManager) SetPrice(token string, price float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.prices[token] = price
}

func (pm *PriceManager) IncreasePrice(token string, price float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.prices[token] += price
}

func (pm *PriceManager) DecreasePrice(token string, price float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.prices[token] -= price
	if pm.prices[token] < 0 {
		pm.prices[token] = 0
	}
}

func (pm *PriceManager) GetPrice(token string) float64 {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.prices[token]
}
