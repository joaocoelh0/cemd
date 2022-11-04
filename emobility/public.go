package emobility

import (
	"github.com/DerAndereAndi/eebus-go-cem/util"
	"github.com/DerAndereAndi/eebus-go/features"
	"github.com/DerAndereAndi/eebus-go/spine/model"
)

var phaseMapping = []string{"a", "b", "c"}

// return the current charge sate of the EV
func (e *EMobilityImpl) EVCurrentChargeState() (EVChargeStateType, error) {
	diagnosisState, err := e.deviceDiagnosis[e.evEntity].GetState()
	if err != nil {
		return EVChargeStateTypeUnkown, err
	}

	switch diagnosisState.OperatingState {
	case model.DeviceDiagnosisOperatingStateTypeNormalOperation:
		return EVChargeStateTypeActive, nil
	case model.DeviceDiagnosisOperatingStateTypeStandby:
		return EVChargeStateTypePaused, nil
	case model.DeviceDiagnosisOperatingStateTypeFailure:
		return EVChargeStateTypeError, nil
	case model.DeviceDiagnosisOperatingStateTypeFinished:
		return EVChargeStateTypeFinished, nil
	}

	return EVChargeStateTypeUnkown, nil
}

// return the number of ac connected phases of the EV or 0 if it is unknown
func (e *EMobilityImpl) EVConnectedPhases() (uint, error) {
	phases, err := e.evElectricalConnection.GetConnectedPhases()
	if err != nil {
		return 0, err
	}

	return phases, nil
}

// return the last current measurement for each phase of the connected EV
//
// possible errors:
//   - ErrDataNotAvailable if no such measurement is (yet) available
//   - and others
func (e *EMobilityImpl) EVCurrents() ([]float64, error) {
	data, err := e.evMeasurement.GetCurrents(e.evElectricalConnection)
	if err != nil {
		return nil, err
	}

	var result []float64

	for _, phase := range phaseMapping {
		value := 0.0
		if theValue, exists := data[phase]; exists {
			value = theValue
		}
		result = append(result, value)
	}

	return result, nil
}

// return the min, max, default limits for each phase of the connected EV
//
// possible errors:
//   - ErrDataNotAvailable if no such measurement is (yet) available
//   - and others
func (e *EMobilityImpl) EVCurrentLimits() ([]float64, []float64, []float64, error) {
	dataMin, dataMax, dataDefault, err := e.evElectricalConnection.GetCurrentsLimits()
	if err != nil {
		return nil, nil, nil, err
	}

	var resultMin, resultMax, resultDefault []float64

	for _, phase := range phaseMapping {
		value := 0.0
		if theValue, exists := dataMin[phase]; exists {
			value = theValue
		}
		resultMin = append(resultMin, value)

		value = 0.0
		if theValue, exists := dataMax[phase]; exists {
			value = theValue
		}
		resultMax = append(resultMax, value)

		value = 0.0
		if theValue, exists := dataDefault[phase]; exists {
			value = theValue
		}
		resultDefault = append(resultDefault, value)
	}

	return resultMin, resultMax, resultDefault, nil
}

// send new LoadControlLimits to the remote EV
//
// parameters:
//   - obligations: Overload Protection Limits per phase in A
//   - recommendations: Self Consumption recommendations per phase in A
//
// obligations:
// Sets a maximum A limit for each phase that the EV may not exceed.
// Mainly used for implementing overload protection of the site or limiting the
// maximum charge power of EVs when the EV and EVSE communicate via IEC61851
// and with ISO15118 if the EV does not support the Optimization of Self Consumption
// usecase.
//
// recommendations:
// Sets a recommended charge power in A for each phase. This is mainly
// used if the EV and EVSE communicate via ISO15118 to support charging excess solar power.
// The EV either needs to support the Optimization of Self Consumption usecase or
// the EVSE needs to be able map the recommendations into oligation limits which then
// works for all EVs communication either via IEC61851 or ISO15118.
//
// note:
// For obligations to work for optimizing solar excess power, the EV needs to
// have an energy demand. Recommendations work even if the EV does not have an active
// energy demand, given it communicated with the EVSE via ISO15118 and supports the usecase.
// In ISO15118-2 the usecase is only supported via VAS extensions which are vendor specific
// and needs to have specific EVSE support for the specific EV brand.
// In ISO15118-20 this is a standard feature which does not need special support on the EVSE.
func (e *EMobilityImpl) EVWriteLoadControlLimits(obligations, recommendations []float64) error {
	electricalDesc, _, err := e.evElectricalConnection.GetParamDescriptionListData()
	if err != nil {
		return features.ErrMetadataNotAvailable
	}

	elLimits, err := e.evElectricalConnection.GetEVLimitValues()
	if err != nil {
		return features.ErrMetadataNotAvailable
	}

	limitDesc, err := e.evLoadControl.GetLimitDescription()
	if err != nil {
		return err
	}

	currentLimits, err := e.evLoadControl.GetLimitValues()
	if err != nil {
		return err
	}

	var limitData []model.LoadControlLimitDataType

	for scopeTypes := 0; scopeTypes < 2; scopeTypes++ {
		currentsPerPhase := obligations
		if scopeTypes == 1 {
			currentsPerPhase = recommendations
		}

		for index, limit := range currentsPerPhase {
			phase := phaseMapping[index]

			var limitId *model.LoadControlLimitIdType
			var elConnectionid *model.ElectricalConnectionIdType

			for _, lDesc := range limitDesc {
				if lDesc.LimitCategory == nil || lDesc.MeasurementId == nil {
					continue
				}

				if *lDesc.LimitCategory != model.LoadControlCategoryTypeObligation {
					continue
				}

				elDesc, exists := electricalDesc[*lDesc.MeasurementId]
				if !exists {
					continue
				}
				if elDesc.ElectricalConnectionId == nil || elDesc.AcMeasuredPhases == nil || string(*elDesc.AcMeasuredPhases) != phase {
					continue
				}

				limitId = lDesc.LimitId
				elConnectionid = elDesc.ElectricalConnectionId
				break
			}

			if limitId == nil || elConnectionid == nil {
				continue
			}

			var currentLimitsForID features.LoadControlLimitType
			var found bool
			for _, item := range currentLimits {
				if uint(*limitId) != item.LimitId {
					continue
				}
				currentLimitsForID = item
				found = true
				break
			}
			if !found || !currentLimitsForID.IsChangeable {
				continue
			}

			limitValue := model.NewScaledNumberType(limit)
			for _, elLimit := range elLimits {
				if elLimit.ConnectionID != uint(*elConnectionid) {
					continue
				}
				if elLimit.Scope != model.ScopeTypeTypeACCurrent {
					continue
				}
				if limit < elLimit.Min {
					limitValue = model.NewScaledNumberType(elLimit.Min)
				}
				if limit > elLimit.Max {
					limitValue = model.NewScaledNumberType(elLimit.Max)
				}
			}

			active := true
			newLimit := model.LoadControlLimitDataType{
				LimitId:       limitId,
				IsLimitActive: &active,
				Value:         limitValue,
			}
			limitData = append(limitData, newLimit)
		}
	}

	_, err = e.evLoadControl.WriteLimitValues(limitData)

	return err
}

// return the current communication standard type used to communicate between EVSE and EV
//
// if an EV is connected via IEC61851, no ISO15118 specific data can be provided!
// sometimes the connection starts with IEC61851 before it switches
// to ISO15118, and sometimes it falls back again. so the error return is
// never absolut for the whole connection time, except if the use case
// is not supported
//
// the values are not constant and can change due to communication problems, bugs, and
// sometimes communication starts with IEC61851 before it switches to ISO
//
// possible errors:
//   - ErrDataNotAvailable if that information is not (yet) available
//   - ErrNotSupported if getting the communication standard is not supported
//   - and others
func (e *EMobilityImpl) EVCommunicationStandard() (EVCommunicationStandardType, error) {
	// check if device configuration descriptions has an communication standard key name
	support, err := e.evDeviceConfiguration.GetDescriptionKeyNameSupport(model.DeviceConfigurationKeyNameTypeCommunicationsStandard)
	if err != nil {
		return EVCommunicationStandardTypeUnknown, err
	}
	if !support {
		return EVCommunicationStandardTypeUnknown, features.ErrNotSupported
	}

	data, err := e.evDeviceConfiguration.GetEVCommunicationStandard()
	if err != nil {
		return EVCommunicationStandardTypeUnknown, err
	}

	return EVCommunicationStandardType(*data), err
}

// returns if the EVSE and EV combination support optimzation of self consumption
//
// possible errors:
//   - ErrDataNotAvailable if that information is not (yet) available
//   - and others
func (e *EMobilityImpl) EVOptimizationOfSelfConsumptionSupported() (bool, error) {
	evEntity, err := util.EntityOfTypeForSki(e.service, model.EntityTypeTypeEV, e.ski)
	if err != nil {
		return false, err
	}

	// check if the Optimization of self consumption usecase is supported
	if !util.IsUsecaseSupported(model.UseCaseNameTypeOptimizationOfSelfConsumptionDuringEVCharging, model.UseCaseActorTypeEV, evEntity.Device()) {
		return false, nil
	}

	// check if loadcontrol limit descriptions contains a recommendation category
	support, err := e.evLoadControl.GetLimitDescriptionCategorySupport(model.LoadControlCategoryTypeRecommendation)
	if err != nil {
		return false, err
	}
	return support, nil
}

// return if the EVSE and EV combination support providing an SoC
//
// requires EVSoCSupported to return true
// only works with a current ISO15118-2 with VAS or ISO15118-20
// communication between EVSE and EV
//
// possible errors:
//   - ErrDataNotAvailable if no such measurement is (yet) available
//   - and others
func (e *EMobilityImpl) EVSoCSupported() (bool, error) {
	evEntity, err := util.EntityOfTypeForSki(e.service, model.EntityTypeTypeEV, e.ski)
	if err != nil {
		return false, err
	}

	// check if the SoC usecase is supported
	if !util.IsUsecaseSupported(model.UseCaseNameTypeEVStateOfCharge, model.UseCaseActorTypeEV, evEntity.Device()) {
		return false, nil
	}

	// check if measurement descriptions has an SoC scope type
	desc, err := e.evMeasurement.GetDescriptionForScope(model.ScopeTypeTypeStateOfCharge)
	if err != nil {
		return false, err
	}
	if len(desc) == 0 {
		return false, features.ErrDataNotAvailable
	}

	return true, nil
}

// return the last known SoC of the connected EV
//
// requires EVSoCSupported to return true
// only works with a current ISO15118-2 with VAS or ISO15118-20
// communication between EVSE and EV
//
// possible errors:
//   - ErrNotSupported if support for SoC is not possible
//   - ErrDataNotAvailable if no such measurement is (yet) available
//   - and others
func (e *EMobilityImpl) EVSoC() (float64, error) {
	// check if the SoC is supported
	support, err := e.EVSoCSupported()
	if err != nil {
		return 0.0, err
	}
	if !support {
		return 0.0, features.ErrNotSupported
	}

	return e.evMeasurement.GetSoC()
}

// returns if the EVSE and EV combination support coordinated charging
//
// possible errors:
//   - ErrDataNotAvailable if that information is not (yet) available
//   - and others
func (e *EMobilityImpl) EVCoordinatedChargingSupported() (bool, error) {
	evEntity, err := util.EntityOfTypeForSki(e.service, model.EntityTypeTypeEV, e.ski)
	if err != nil {
		return false, err
	}

	// check if the Coordinated charging usecase is supported
	if !util.IsUsecaseSupported(model.UseCaseNameTypeCoordinatedEVCharging, model.UseCaseActorTypeEV, evEntity.Device()) {
		return false, nil
	}

	return true, nil
}
