package attrmeta

import "bytes"

// WriteDiff write a "diff" represenation of attr to buffer.
func WriteDiff(buffer *bytes.Buffer, attr Attr) {
	if attr.Value == "" {
		Write(buffer, attr)
		return
	}

	if attr.CreatedAt == attr.UpdatedAt {
		buffer.WriteString("+")
	}

	buffer.WriteString(attr.Key)
	if attr.IsFrozen {
		buffer.WriteString("@")
	} else {
		buffer.WriteString("=")
	}
	buffer.WriteString(attr.Value)
}

// Write writes a representation of attr to the buffer.
// Non-frozen attributes with empty-values are omitted.
func Write(buffer *bytes.Buffer, attr Attr) {
	if attr.Value == "" {
		if attr.IsFrozen {
			buffer.WriteString("!")
		} else {
			buffer.WriteString("-")
		}
		buffer.WriteString(attr.Key)
	} else {
		buffer.WriteString(attr.Key)
		if attr.IsFrozen {
			buffer.WriteString("@")
		} else {
			buffer.WriteString("=")
		}
		buffer.WriteString(attr.Value)
	}
}
