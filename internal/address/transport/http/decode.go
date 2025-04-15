package http

import (
	"context"
	"encoding/json"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	httpError "github.com/nurdsoft/nurd-commerce-core/shared/errors/http"
	"github.com/pkg/errors"
)

type RequestBodyType interface {
	entities.AddressRequestBody
}

func decodeBodyFromRequest[T RequestBodyType](req *T, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return errors.WithMessage(httpError.ErrBadRequestBody, err.Error())
	}

	defer r.Body.Close()

	return nil
}

func decodeAddAddressRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqBody := &entities.AddressRequestBody{}
	err := decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.AddAddressRequest{
		Address: reqBody,
	}, nil
}

func decodeGetAddressesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeGetAddressRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	addressID, err := uuid.Parse(params["address_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "address_id is not valid")
	}

	return &entities.GetAddressRequest{
		AddressID: addressID,
	}, nil
}

func decodeUpdateAddressRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	addressID, err := uuid.Parse(params["address_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "address_id is not valid")
	}

	reqBody := &entities.AddressRequestBody{}

	err = decodeBodyFromRequest(reqBody, r)
	if err != nil {
		return nil, err
	}

	return &entities.UpdateAddressRequest{
		AddressID: addressID,
		Address:   reqBody,
	}, nil
}

func decodeDeleteAddressRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	addressID, err := uuid.Parse(params["address_id"])
	if err != nil {
		return nil, moduleErrors.NewAPIError("VALIDATION_ERROR", "address_id is not valid")
	}

	return &entities.DeleteAddressRequest{
		AddressID: addressID,
	}, nil
}
