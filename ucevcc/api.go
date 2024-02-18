package ucevcc

import (
	"github.com/enbility/cemd/api"
	spineapi "github.com/enbility/spine-go/api"
	"github.com/enbility/spine-go/model"
)

//go:generate mockery

// interface for the EV Commissioning and Configuration UseCase
type UCEvCCInterface interface {
	api.UseCaseInterface

	// Scenario 1 & 8

	// return if the EV is connected
	EVConnected(entity spineapi.EntityRemoteInterface) bool

	// Scenario 2

	// return the current communication standard type used to communicate between EVSE and EV
	EVCommunicationStandard(entity spineapi.EntityRemoteInterface) (string, error)

	// Scenario 3

	// return if the EV supports asymmetric charging
	EVAsymmetricChargingSupported(entity spineapi.EntityRemoteInterface) (bool, error)

	// Scenario 4

	// return the identifications of the currently connected EV or nil if not available
	// these can be multiple, e.g. PCID, Mac Address, RFID
	EVIdentifications(entity spineapi.EntityRemoteInterface) ([]IdentificationItem, error)

	// Scenario 5

	// the manufacturer data of an EVSE
	// returns deviceName, serialNumber, error
	EVManufacturerData(ski string, entity spineapi.EntityRemoteInterface) (string, string, error)

	// Scenario 6

	// return the number of ac connected phases of the EV or 0 if it is unknown
	EVConnectedPhases(entity spineapi.EntityRemoteInterface) (uint, error)

	// return the min, max, default limits for each phase of the connected EV
	EVCurrentLimits(entity spineapi.EntityRemoteInterface) ([]float64, []float64, []float64, error)

	// Scenario 7

	// is the EV in sleep mode
	EVInSleepMode(ski string, entity spineapi.EntityRemoteInterface) (bool, error)
}

// EV identification
type IdentificationItem struct {
	// the identification value
	Value string

	// the type of the identification value, e.g.
	ValueType model.IdentificationTypeType
}

const (
	// An EV was connected
	UCEvCCEventConnected api.UseCaseEventType = "ucEvConnected"

	// An EV was disconnected
	UCEvCCEventDisconnected api.UseCaseEventType = "ucEvDisonnected"

	// EV device configuration data was updated (CommunicationStandard, Asymmetric charging)
	UCEvCCEventConfigurationUdpate api.UseCaseEventType = "ucEvConfigurationUpdate"

	// EV manufacturer data was updated
	UCEvCCEventManufacturerUpdate api.UseCaseEventType = "ucEvManufacturerUpdate"

	// EV charging power limits updated
	UCEvCCEventChargingPowerLimitsUpdate api.UseCaseEventType = "ucEvPowerLimitsUpdate"
)
