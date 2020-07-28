package cassandra

import (
	"strings"
)

// EncodeCassandraID Encode primary and partition keys into a single unit
func EncodeCassandraID(id, external_id string) string {
	return external_id + id
}

// DecodeCassandraID Decode custom ID into partition and primary keys
//
// Encoded ID structure
// First 16 bits -> NanoID = External ID,
// Last 128 bits -> UUIDv1 = Cassandra's timeuuid
func DecodeCassandraID(encodedID string) (id, external_id string) {
	splitStr := strings.Split(encodedID, "")
	external_id = strings.Join(splitStr[:16], "")
	id = strings.Join(splitStr[17:], "")
	return
}
