package emobility

import (
	"sync"

	"github.com/enbility/cemd/scenarios"
	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/util"
	shipapi "github.com/enbility/ship-go/api"
	spineapi "github.com/enbility/spine-go/api"
	"github.com/enbility/spine-go/model"
)

type EmobilityScenarioImpl struct {
	*scenarios.ScenarioImpl

	remoteDevices map[string]*EMobilityImpl

	mux sync.Mutex

	currency      model.CurrencyType
	configuration EmobilityConfiguration
}

var _ scenarios.ScenariosI = (*EmobilityScenarioImpl)(nil)

func NewEMobilityScenario(service api.EEBUSService, currency model.CurrencyType, configuration EmobilityConfiguration) *EmobilityScenarioImpl {
	return &EmobilityScenarioImpl{
		ScenarioImpl:  scenarios.NewScenarioImpl(service),
		remoteDevices: make(map[string]*EMobilityImpl),
		currency:      currency,
		configuration: configuration,
	}
}

// adds all the supported features to the local entity
func (e *EmobilityScenarioImpl) AddFeatures() {
	localEntity := e.Service.LocalDevice().EntityForType(model.EntityTypeTypeCEM)

	// server features
	{
		f := localEntity.GetOrAddFeature(model.FeatureTypeTypeDeviceDiagnosis, model.RoleTypeServer)
		f.AddResultHandler(e)
		f.AddFunctionType(model.FunctionTypeDeviceDiagnosisStateData, true, false)

		// Set the initial state
		deviceDiagnosisStateDate := &model.DeviceDiagnosisStateDataType{
			OperatingState: util.Ptr(model.DeviceDiagnosisOperatingStateTypeNormalOperation),
		}
		f.SetData(model.FunctionTypeDeviceDiagnosisStateData, deviceDiagnosisStateDate)

		f.AddFunctionType(model.FunctionTypeDeviceDiagnosisHeartbeatData, true, false)
	}

	// client features
	var clientFeatures = []model.FeatureTypeType{
		model.FeatureTypeTypeDeviceDiagnosis,
		model.FeatureTypeTypeDeviceClassification,
		model.FeatureTypeTypeDeviceConfiguration,
		model.FeatureTypeTypeElectricalConnection,
		model.FeatureTypeTypeMeasurement,
		model.FeatureTypeTypeLoadControl,
		model.FeatureTypeTypeIdentification,
	}

	if e.configuration.CoordinatedChargingEnabled {
		clientFeatures = append(clientFeatures, model.FeatureTypeTypeTimeSeries)
		clientFeatures = append(clientFeatures, model.FeatureTypeTypeIncentiveTable)
	}
	for _, feature := range clientFeatures {
		f := localEntity.GetOrAddFeature(feature, model.RoleTypeClient)
		f.AddResultHandler(e)
	}
}

// add supported e-mobility usecases
func (e *EmobilityScenarioImpl) AddUseCases() {
	localEntity := e.Service.LocalDevice().EntityForType(model.EntityTypeTypeCEM)

	localEntity.AddUseCaseSupport(
		model.UseCaseActorTypeCEM,
		model.UseCaseNameTypeEVSECommissioningAndConfiguration,
		model.SpecificationVersionType("1.0.1"),
		"",
		true,
		[]model.UseCaseScenarioSupportType{1, 2})

	localEntity.AddUseCaseSupport(
		model.UseCaseActorTypeCEM,
		model.UseCaseNameTypeEVCommissioningAndConfiguration,
		model.SpecificationVersionType("1.0.1"),
		"",
		true,
		[]model.UseCaseScenarioSupportType{1, 2, 3, 4, 5, 6, 7, 8})

	localEntity.AddUseCaseSupport(
		model.UseCaseActorTypeCEM,
		model.UseCaseNameTypeMeasurementOfElectricityDuringEVCharging,
		model.SpecificationVersionType("1.0.1"),
		"",
		true,
		[]model.UseCaseScenarioSupportType{1, 2, 3})

	localEntity.AddUseCaseSupport(
		model.UseCaseActorTypeCEM,
		model.UseCaseNameTypeOverloadProtectionByEVChargingCurrentCurtailment,
		model.SpecificationVersionType("1.0.1b"),
		"",
		true,
		[]model.UseCaseScenarioSupportType{1, 2, 3})

	localEntity.AddUseCaseSupport(
		model.UseCaseActorTypeMonitoringAppliance,
		model.UseCaseNameTypeEVStateOfCharge,
		model.SpecificationVersionType("1.0.0"),
		"",
		true,
		[]model.UseCaseScenarioSupportType{1, 2, 3, 4})

	localEntity.AddUseCaseSupport(
		model.UseCaseActorTypeCEM,
		model.UseCaseNameTypeOptimizationOfSelfConsumptionDuringEVCharging,
		model.SpecificationVersionType("1.0.1b"),
		"",
		true,
		[]model.UseCaseScenarioSupportType{1, 2, 3})

	if e.configuration.CoordinatedChargingEnabled {
		localEntity.AddUseCaseSupport(
			model.UseCaseActorTypeCEM,
			model.UseCaseNameTypeCoordinatedEVCharging,
			model.SpecificationVersionType("1.0.1"),
			"",
			true,
			[]model.UseCaseScenarioSupportType{1, 2, 3, 4, 5, 6, 7, 8})
	}
}

func (e *EmobilityScenarioImpl) RegisterRemoteDevice(details *shipapi.ServiceDetails, dataProvider any) any {
	// TODO: emobility should be stored per remote SKI and
	// only be set for the SKI if the device supports it
	e.mux.Lock()
	defer e.mux.Unlock()

	if em, ok := e.remoteDevices[details.SKI()]; ok {
		return em
	}

	var provider EmobilityDataProvider
	if dataProvider != nil {
		provider = dataProvider.(EmobilityDataProvider)
	}
	emobility := NewEMobility(e.Service, details, e.currency, e.configuration, provider)
	e.remoteDevices[details.SKI()] = emobility
	return emobility
}

func (e *EmobilityScenarioImpl) UnRegisterRemoteDevice(remoteDeviceSki string) {
	e.mux.Lock()
	defer e.mux.Unlock()

	delete(e.remoteDevices, remoteDeviceSki)

	e.Service.RegisterRemoteSKI(remoteDeviceSki, false)
}

func (e *EmobilityScenarioImpl) HandleResult(errorMsg spineapi.ResultMessage) {
	e.mux.Lock()
	defer e.mux.Unlock()

	if errorMsg.DeviceRemote == nil {
		return
	}

	em, ok := e.remoteDevices[errorMsg.DeviceRemote.Ski()]
	if !ok {
		return
	}

	em.HandleResult(errorMsg)
}
