package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type CacheException struct {
	RealtimeGatewayException
}

func NewCacheException(domain string) CacheException {
	return CacheException{
		RealtimeGatewayException: RealtimeGatewayException{
			Domain: domain,
		},
	}
}

func (e CacheException) Unavailable(cause error) *cexceptions.Exception {
	exception := cexceptions.New("CacheClientUnavailable", e.Domain, "AccessCache", "The realtime cache is unavailable", http.StatusServiceUnavailable, true)
	exception.Retryable = true
	return exception.WithOrigin(cause)
}

func (e CacheException) NotFound(cause error) *cexceptions.Exception {
	return cexceptions.New("NotFound", e.Domain, "GetCache", "The realtime cache record was not found", http.StatusNotFound).WithOrigin(cause)
}

func (e CacheException) DeserializationFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("DeserializationFailed", e.Domain, "ReadCache", "Failed to deserialize the realtime cache record", http.StatusInternalServerError, true).WithOrigin(cause)
}

func (e CacheException) SerializationFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("SerializationFailed", e.Domain, "WriteCache", "Failed to serialize the realtime cache record", http.StatusInternalServerError, true).WithOrigin(cause)
}

func (e CacheException) CreateFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("FailedToCreate", e.Domain, "CreateCache", "Failed to create the realtime cache record", http.StatusInternalServerError, true).WithOrigin(cause)
}

func (e CacheException) InvalidRateLimitTokenCount() *cexceptions.Exception {
	return cexceptions.New("InvalidRateLimitTokenCount", e.Domain, "SynchronizeCache", "The rate-limit token count is invalid", http.StatusBadRequest)
}

func (e CacheException) UpdateFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("FailedToUpdate", e.Domain, "UpdateCache", "Failed to update the realtime cache record", http.StatusInternalServerError, true).WithOrigin(cause)
}

func (e CacheException) DeleteFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("FailedToDelete", e.Domain, "DeleteCache", "Failed to delete the realtime cache record", http.StatusInternalServerError, true).WithOrigin(cause)
}
