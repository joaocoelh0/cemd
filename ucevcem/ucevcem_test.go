package ucevcem

import (
	eebusutil "github.com/enbility/eebus-go/util"
	"github.com/enbility/spine-go/model"
	"github.com/stretchr/testify/assert"
)

func (s *UCEVCEMSuite) Test_IsUseCaseSupported() {
	data, err := s.sut.IsUseCaseSupported(s.mockRemoteEntity)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), false, data)

	data, err = s.sut.IsUseCaseSupported(s.evEntity)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, data)

	ucData := &model.NodeManagementUseCaseDataType{
		UseCaseInformation: []model.UseCaseInformationDataType{
			{
				Actor: eebusutil.Ptr(model.UseCaseActorTypeEV),
				UseCaseSupport: []model.UseCaseSupportType{
					{
						UseCaseName:      eebusutil.Ptr(model.UseCaseNameTypeMeasurementOfElectricityDuringEVCharging),
						UseCaseAvailable: eebusutil.Ptr(true),
						ScenarioSupport:  []model.UseCaseScenarioSupportType{1},
					},
				},
			},
		},
	}

	nodemgmtEntity := s.remoteDevice.Entity([]model.AddressEntityType{0})
	nodeFeature := s.remoteDevice.FeatureByEntityTypeAndRole(nodemgmtEntity, model.FeatureTypeTypeNodeManagement, model.RoleTypeSpecial)
	fErr := nodeFeature.UpdateData(model.FunctionTypeNodeManagementUseCaseData, ucData, nil, nil)
	assert.Nil(s.T(), fErr)

	data, err = s.sut.IsUseCaseSupported(s.evEntity)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), true, data)
}
