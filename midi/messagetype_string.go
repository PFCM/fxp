// Code generated by "stringer -type=MessageType,CV1MessageType"; DO NOT EDIT.

package midi

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MTUtility-0]
	_ = x[MTSystem-1]
	_ = x[MTChannelVoice1-2]
	_ = x[MTData-3]
	_ = x[MTChannelVoice2-4]
	_ = x[MTLongData-5]
	_ = x[MTFlexData-13]
	_ = x[MTUMPStream-15]
}

const (
	_MessageType_name_0 = "MTUtilityMTSystemMTChannelVoice1MTDataMTChannelVoice2MTLongData"
	_MessageType_name_1 = "MTFlexData"
	_MessageType_name_2 = "MTUMPStream"
)

var (
	_MessageType_index_0 = [...]uint8{0, 9, 17, 32, 38, 53, 63}
)

func (i MessageType) String() string {
	switch {
	case i <= 5:
		return _MessageType_name_0[_MessageType_index_0[i]:_MessageType_index_0[i+1]]
	case i == 13:
		return _MessageType_name_1
	case i == 15:
		return _MessageType_name_2
	default:
		return "MessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CV1NoteOff-8]
	_ = x[CV1NoteOn-9]
	_ = x[CV1PolyPressure-10]
	_ = x[CV1ControlChange-11]
	_ = x[CV1ProgramChange-12]
	_ = x[CV1ChannelPressure-13]
	_ = x[CV1PitchBend-14]
}

const _CV1MessageType_name = "CV1NoteOffCV1NoteOnCV1PolyPressureCV1ControlChangeCV1ProgramChangeCV1ChannelPressureCV1PitchBend"

var _CV1MessageType_index = [...]uint8{0, 10, 19, 34, 50, 66, 84, 96}

func (i CV1MessageType) String() string {
	i -= 8
	if i >= CV1MessageType(len(_CV1MessageType_index)-1) {
		return "CV1MessageType(" + strconv.FormatInt(int64(i+8), 10) + ")"
	}
	return _CV1MessageType_name[_CV1MessageType_index[i]:_CV1MessageType_index[i+1]]
}
