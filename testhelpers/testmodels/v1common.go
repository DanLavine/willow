package testmodels

import (
	. "github.com/onsi/gomega"

	v1common "github.com/DanLavine/willow/pkg/models/api/common/v1"
	"github.com/DanLavine/willow/pkg/models/datatypes"
)

func NoTypeRestrictions(g *GomegaWithT) v1common.TypeRestrictions {
	typeRestrictions := v1common.TypeRestrictions{
		MinDataType: datatypes.MinDataType,
		MaxDataType: datatypes.MaxDataType,
	}

	g.Expect(typeRestrictions.Validate()).ToNot(HaveOccurred())
	return typeRestrictions
}

func TypeRestrictionsGeneral(g *GomegaWithT) v1common.TypeRestrictions {
	typeRestrictions := v1common.TypeRestrictions{
		MinDataType: datatypes.MinDataType,
		MaxDataType: datatypes.MaxWithoutAnyDataType,
	}

	g.Expect(typeRestrictions.Validate()).ToNot(HaveOccurred())
	return typeRestrictions
}

func TypeRestrictions(g *GomegaWithT, minDataType, maxDataType datatypes.DataType) v1common.TypeRestrictions {
	typeRestrictions := v1common.TypeRestrictions{
		MinDataType: minDataType,
		MaxDataType: maxDataType,
	}

	g.Expect(typeRestrictions.Validate()).ToNot(HaveOccurred())
	return typeRestrictions
}
