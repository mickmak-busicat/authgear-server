// Code generated by "stringer -type=DataType"; DO NOT EDIT.

package skydb

import "strconv"

const _DataType_name = "TypeStringTypeNumberTypeBooleanTypeJSONTypeReferenceTypeLocationTypeDateTimeTypeAssetTypeACLTypeIntegerTypeSequenceTypeGeometryTypeUnknown"

var _DataType_index = [...]uint8{0, 10, 20, 31, 39, 52, 64, 76, 85, 92, 103, 115, 127, 138}

func (i DataType) String() string {
	i -= 1
	if i >= DataType(len(_DataType_index)-1) {
		return "DataType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _DataType_name[_DataType_index[i]:_DataType_index[i+1]]
}
