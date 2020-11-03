package types

import "sync"

// OrderMap is used for storing orders by their order id
type OrderMap map[uint64]Order

func (m OrderMap) Add(o Order) {
	m[o.OrderID] = o
}

func (m OrderMap) Delete(orderID uint64) {
	delete(m, orderID)
}

func (m OrderMap) IDs() (ids []uint64) {
	for id := range m {
		ids = append(ids, id)
	}

	return ids
}

func (m OrderMap) Exists(orderID uint64) bool {
	_, ok := m[orderID]
	return ok
}

func (m OrderMap) FindByStatus(status OrderStatus) (orders OrderSlice) {
	for _, o := range m {
		if o.Status == status {
			orders = append(orders, o)
		}
	}

	return orders
}

func (m OrderMap) Filled() OrderSlice {
	return m.FindByStatus(OrderStatusFilled)
}

func (m OrderMap) Canceled() OrderSlice {
	return m.FindByStatus(OrderStatusCanceled)
}

func (m OrderMap) Orders() (orders OrderSlice) {
	for _, o := range m {
		orders = append(orders, o)
	}
	return orders
}

type SyncOrderMap struct {
	orders OrderMap

	sync.RWMutex
}

func NewSyncOrderMap() *SyncOrderMap {
	return &SyncOrderMap{
		orders: make(OrderMap),
	}
}

func (m *SyncOrderMap) Delete(orderID uint64) {
	m.Lock()
	defer m.Unlock()

	m.orders.Delete(orderID)
}

func (m *SyncOrderMap) Add(o Order) {
	m.Lock()
	defer m.Unlock()

	m.orders.Add(o)
}

func (m *SyncOrderMap) Iterate(it func(id uint64, order Order) bool) {
	m.Lock()
	defer m.Unlock()

	for id := range m.orders {
		if it(id, m.orders[id]) {
			break
		}
	}
}

func (m *SyncOrderMap) Exists(orderID uint64) bool {
	m.RLock()
	defer m.RUnlock()

	return m.orders.Exists(orderID)
}

func (m *SyncOrderMap) Len() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.orders)
}

func (m *SyncOrderMap) IDs() []uint64 {
	m.RLock()
	defer m.RUnlock()

	return m.orders.IDs()
}

func (m *SyncOrderMap) FindByStatus(status OrderStatus) OrderSlice {
	m.RLock()
	defer m.RUnlock()

	return m.orders.FindByStatus(status)
}

func (m *SyncOrderMap) Filled() OrderSlice {
	return m.FindByStatus(OrderStatusFilled)
}

// AnyFilled find any order is filled and stop iterating the order map
func (m *SyncOrderMap) AnyFilled() (order Order, ok bool) {
	m.RLock()
	defer m.RUnlock()

	for _, o := range m.orders {
		if o.Status == OrderStatusFilled {
			ok = true
			order = o
			return order, ok
		}
	}

	return
}

func (m *SyncOrderMap) Canceled() OrderSlice {
	return m.FindByStatus(OrderStatusCanceled)
}

func (m *SyncOrderMap) Orders() OrderSlice {
	m.RLock()
	defer m.RUnlock()
	return m.orders.Orders()
}

type OrderSlice []Order

func (s OrderSlice) IDs() (ids []uint64) {
	for _, o := range s {
		ids = append(ids, o.OrderID)
	}
	return ids
}