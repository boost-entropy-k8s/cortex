// Code generated by go-swagger; DO NOT EDIT.

// Copyright Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package general

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetStatusHandlerFunc turns a function with the right signature into a get status handler
type GetStatusHandlerFunc func(GetStatusParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetStatusHandlerFunc) Handle(params GetStatusParams) middleware.Responder {
	return fn(params)
}

// GetStatusHandler interface for that can handle valid get status params
type GetStatusHandler interface {
	Handle(GetStatusParams) middleware.Responder
}

// NewGetStatus creates a new http.Handler for the get status operation
func NewGetStatus(ctx *middleware.Context, handler GetStatusHandler) *GetStatus {
	return &GetStatus{Context: ctx, Handler: handler}
}

/*
	GetStatus swagger:route GET /status general getStatus

Get current status of an Alertmanager instance and its cluster
*/
type GetStatus struct {
	Context *middleware.Context
	Handler GetStatusHandler
}

func (o *GetStatus) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetStatusParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
