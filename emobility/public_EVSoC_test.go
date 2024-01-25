package emobility

import (
	"testing"

	"github.com/enbility/eebus-go/util"
	"github.com/enbility/spine-go/mocks"
	"github.com/enbility/spine-go/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_EVSoC(t *testing.T) {
	emobilty, eebusService := setupEmobility(t)

	mockRemoteDevice := mocks.NewDeviceRemoteInterface(t)
	mockRemoteEntity := mocks.NewEntityRemoteInterface(t)
	mockRemoteFeature := mocks.NewFeatureRemoteInterface(t)
	mockRemoteDevice.EXPECT().FeatureByEntityTypeAndRole(mock.Anything, mock.Anything, mock.Anything).Return(mockRemoteFeature)
	mockRemoteEntity.EXPECT().Device().Return(mockRemoteDevice)
	data, err := emobilty.EVSoC(mockRemoteEntity)
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, data)

	localDevice, localEntity, remoteDevice, entites, _ := setupDevices(eebusService)
	emobilty.evseEntity = entites[0]
	emobilty.evEntity = entites[1]

	data, err = emobilty.EVSoC(emobilty.evEntity)
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, data)

	data, err = emobilty.EVSoC(emobilty.evEntity)
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, data)

	datagram := datagramForEntityAndFeatures(true, localDevice, localEntity, nil, model.FeatureTypeTypeNodeManagement, model.RoleTypeSpecial, model.RoleTypeSpecial)

	cmd := []model.CmdType{{
		NodeManagementUseCaseData: &model.NodeManagementUseCaseDataType{
			UseCaseInformation: []model.UseCaseInformationDataType{
				{
					Actor: util.Ptr(model.UseCaseActorTypeEV),
					UseCaseSupport: []model.UseCaseSupportType{
						{
							UseCaseName:      util.Ptr(model.UseCaseNameTypeEVStateOfCharge),
							UseCaseAvailable: util.Ptr(true),
							ScenarioSupport:  []model.UseCaseScenarioSupportType{1},
						},
					},
				},
			},
		}}}
	datagram.Payload.Cmd = cmd

	err = localDevice.ProcessCmd(datagram, remoteDevice)
	assert.Nil(t, err)

	data, err = emobilty.EVSoC(emobilty.evEntity)
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, data)

	datagram = datagramForEntityAndFeatures(false, localDevice, localEntity, emobilty.evEntity, model.FeatureTypeTypeMeasurement, model.RoleTypeServer, model.RoleTypeClient)

	cmd = []model.CmdType{{
		MeasurementDescriptionListData: &model.MeasurementDescriptionListDataType{
			MeasurementDescriptionData: []model.MeasurementDescriptionDataType{
				{
					MeasurementId:   util.Ptr(model.MeasurementIdType(0)),
					MeasurementType: util.Ptr(model.MeasurementTypeTypePercentage),
					CommodityType:   util.Ptr(model.CommodityTypeTypeElectricity),
					ScopeType:       util.Ptr(model.ScopeTypeTypeStateOfCharge),
				},
			},
		}}}
	datagram.Payload.Cmd = cmd

	err = localDevice.ProcessCmd(datagram, remoteDevice)
	assert.Nil(t, err)

	data, err = emobilty.EVSoC(emobilty.evEntity)
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, data)

	datagram = datagramForEntityAndFeatures(false, localDevice, localEntity, emobilty.evEntity, model.FeatureTypeTypeMeasurement, model.RoleTypeServer, model.RoleTypeClient)

	cmd = []model.CmdType{{
		MeasurementListData: &model.MeasurementListDataType{
			MeasurementData: []model.MeasurementDataType{
				{
					MeasurementId: util.Ptr(model.MeasurementIdType(0)),
				},
			},
		}}}
	datagram.Payload.Cmd = cmd

	err = localDevice.ProcessCmd(datagram, remoteDevice)
	assert.Nil(t, err)

	data, err = emobilty.EVSoC(emobilty.evEntity)
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, data)

	datagram = datagramForEntityAndFeatures(false, localDevice, localEntity, emobilty.evEntity, model.FeatureTypeTypeMeasurement, model.RoleTypeServer, model.RoleTypeClient)

	cmd = []model.CmdType{{
		MeasurementListData: &model.MeasurementListDataType{
			MeasurementData: []model.MeasurementDataType{
				{
					MeasurementId: util.Ptr(model.MeasurementIdType(0)),
					Value:         model.NewScaledNumberType(80),
				},
			},
		}}}
	datagram.Payload.Cmd = cmd

	err = localDevice.ProcessCmd(datagram, remoteDevice)
	assert.Nil(t, err)

	data, err = emobilty.EVSoC(emobilty.evEntity)
	assert.Nil(t, err)
	assert.Equal(t, 80.0, data)
}
