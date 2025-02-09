package cem

import (
	"github.com/enbility/cemd/api"
	eebusapi "github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/service"
	"github.com/enbility/ship-go/logging"
	"github.com/enbility/spine-go/model"
	"github.com/enbility/spine-go/spine"
)

// Generic CEM implementation
type Cem struct {
	Service eebusapi.ServiceInterface

	Currency model.CurrencyType

	eventCB api.EventHandlerCB

	usecases []api.UseCaseInterface
}

func NewCEM(
	serviceDescription *eebusapi.Configuration,
	serviceHandler eebusapi.ServiceReaderInterface,
	eventCB api.EventHandlerCB,
	log logging.LoggingInterface) *Cem {
	cem := &Cem{
		Service:  service.NewService(serviceDescription, serviceHandler),
		Currency: model.CurrencyTypeEur,
		eventCB:  eventCB,
	}

	cem.Service.SetLogging(log)

	_ = spine.Events.Subscribe(cem)

	return cem
}

var _ api.CemInterface = (*Cem)(nil)

// Set up the eebus service
func (h *Cem) Setup() error {
	return h.Service.Setup()
}

// Start the EEBUS service
func (h *Cem) Start() {
	h.Service.Start()
}

// Shutdown the EEBUS servic
func (h *Cem) Shutdown() {
	h.Service.Shutdown()
}

// Add a use case implementation
func (h *Cem) AddUseCase(usecase api.UseCaseInterface) {
	h.usecases = append(h.usecases, usecase)

	usecase.AddFeatures()
	usecase.AddUseCase()
}
