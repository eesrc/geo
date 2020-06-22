package validation

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxLimit     = 2500
	defaultLimit = 255
)

// FilterParams contains parameters regarding listing normal list parameters
type FilterParams struct {
	Since  int64
	Until  int64
	Limit  int64
	Offset int64
}

// NewFilterParamsFromQueryParams returns FilterParams from given Query values.
// It returns a validation error if invalid filter paramters is provided
func NewFilterParamsFromQueryParams(values url.Values) (FilterParams, error) {
	parameterMap := make(HandlerParameterMap)

	// Values allows for multiple values, consolidate into one value joining ","
	for key, val := range values {
		parameterMap[key] = strings.Join(val, ",")
	}

	filterParams := FilterParams{
		Since:  0,
		Until:  time.Now().UnixNano() / int64(time.Millisecond),
		Limit:  defaultLimit,
		Offset: 0,
	}

	if limit, err := parameterMap.AsInt64("limit"); err == nil {
		if limit > maxLimit {
			return filterParams, newError(
				NewErrorResponse(
					http.StatusBadRequest,
					NewParameterErrorDetail("limit", fmt.Sprintf("The provided limit is bigger than allowed max of %d items", maxLimit)),
				),
			)
		}

		if limit < 1 {
			return filterParams, getTooLowValidationError("limit")
		}

		filterParams.Limit = limit
	} else {
		// Limit is optional, however if it's an invalid number we return 400
		if _, ok := err.(*KeyNotFoundError); !ok {
			return filterParams, getNonNumberValidationError("limit")
		}
	}

	if offset, err := parameterMap.AsInt64("offset"); err == nil {
		if offset < 0 {
			return filterParams, getTooLowValidationError("offset")
		}

		filterParams.Offset = offset
	} else {
		// Offset is optional, however if it's an invalid number we return 400
		if _, ok := err.(*KeyNotFoundError); !ok {
			return filterParams, getNonNumberValidationError("offset")
		}
	}

	if since, err := parameterMap.AsInt64("since"); err == nil {
		filterParams.Since = since
	} else {
		// since is optional, however if it's an invalid number we return 400
		if _, ok := err.(*KeyNotFoundError); !ok {
			return filterParams, getNonTimestampValidationError("since")
		}
	}

	if until, err := parameterMap.AsInt64("until"); err == nil {
		filterParams.Until = until
	} else {
		// until is optional, however if it's an invalid number we return 400
		if _, ok := err.(*KeyNotFoundError); !ok {
			return filterParams, getNonTimestampValidationError("until")
		}
	}

	// Validate dates
	if filterParams.Since > filterParams.Until {
		return filterParams, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail(
					"since",
					"Provided since is after until",
				),
				NewParameterErrorDetail(
					"until",
					"Provided until is before since",
				),
			),
		)
	}

	return filterParams, nil
}

func getTooLowValidationError(key string) error {
	return newError(
		NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail(
				key,
				fmt.Sprintf("The provided %s is below or equal to 0", key),
			),
		),
	)
}

func getNonNumberValidationError(key string) error {
	return newError(
		NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail(
				key,
				fmt.Sprintf("The provided %s is not a number", key),
			),
		),
	)
}

func getNonTimestampValidationError(field string) error {
	return newError(
		NewErrorResponse(
			http.StatusBadRequest,
			NewParameterErrorDetail(
				field,
				fmt.Sprintf("The provided %s is not a unix timestamp", field),
			),
		),
	)
}
