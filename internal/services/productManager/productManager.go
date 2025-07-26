package productManager

import (
	"log/slog"
	"prodLoaderREST/internal/broker"
	"prodLoaderREST/internal/services/consumer/vk"
	"prodLoaderREST/internal/storage"
)

type Manager struct {
	VK *vk.Consumer
	//Ucoz ucoz.Consumer
	// Avito avito.Consumer
	log     *slog.Logger
	storage storage.Storage
	broker  *broker.Exchanger
}

func New(log *slog.Logger, VK *vk.Consumer, broker *broker.Exchanger, storage storage.Storage) *Manager {
	return &Manager{
		log:     log,
		VK:      VK,
		storage: storage,
		broker:  broker,
	}
}

func (m *Manager) Listen() {
	go m.VK.ListenLoad(broker.VKaddProductChannel)
	go m.VK.ListenDelete(broker.VKdeleteProductChannel)
}
