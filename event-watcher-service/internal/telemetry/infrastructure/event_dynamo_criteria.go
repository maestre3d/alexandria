package infrastructure

import "gocloud.dev/docstore"

func QueryCriteriaDynamo(value string, query *docstore.Query) *docstore.Query {
	query.Where("service_name", "=", value).Where("transaction_id", "=", value).Where("event_type", "=", value).
		Where("importance", "=", value).Where("provider", "=", value)
	return query
}
