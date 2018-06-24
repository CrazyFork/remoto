// Code generated by Remoto; DO NOT EDIT.

// Package suggestionbox contains the HTTP server for suggestionbox services.
package suggestionbox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/matryer/remoto/go/remotohttp"
	"github.com/matryer/remoto/remototypes"
	"github.com/pkg/errors"
)

type Suggestionbox interface {
	CreateModel(context.Context, *CreateModelRequest) (*CreateModelResponse, error)

	DeleteModel(context.Context, *DeleteModelRequest) (*DeleteModelResponse, error)

	GetState(context.Context, *GetStateRequest) (*remototypes.FileResponse, error)

	ListModels(context.Context, *ListModelsRequest) (*ListModelsResponse, error)

	Predict(context.Context, *PredictRequest) (*PredictResponse, error)

	PutState(context.Context, *PutStateRequest) (*PutStateResponse, error)

	Reward(context.Context, *RewardRequest) (*RewardResponse, error)
}

// Run is the simplest way to run the services.
func Run(addr string,
	suggestionbox Suggestionbox,
) error {
	server := New(
		suggestionbox,
	)
	if err := server.Describe(os.Stdout); err != nil {
		return errors.Wrap(err, "describe service")
	}
	if err := http.ListenAndServe(addr, server); err != nil {
		return err
	}
	return nil
}

// New makes a new remotohttp.Server with the specified services
// registered.
func New(
	suggestionbox Suggestionbox,
) *remotohttp.Server {
	server := &remotohttp.Server{
		OnErr: func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", r.Method, r.URL.Path, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
		NotFound: http.NotFoundHandler(),
	}

	RegisterSuggestionboxServer(server, suggestionbox)
	return server
}

// RegisterSuggestionboxServer registers a Suggestionbox with a remotohttp.Server.
func RegisterSuggestionboxServer(server *remotohttp.Server, service Suggestionbox) {
	srv := &httpSuggestionboxServer{
		service: service,
		server:  server,
	}
	server.Register("/remoto/Suggestionbox.CreateModel", http.HandlerFunc(srv.handleCreateModel))
	server.Register("/remoto/Suggestionbox.DeleteModel", http.HandlerFunc(srv.handleDeleteModel))
	server.Register("/remoto/Suggestionbox.GetState", http.HandlerFunc(srv.handleGetState))
	server.Register("/remoto/Suggestionbox.ListModels", http.HandlerFunc(srv.handleListModels))
	server.Register("/remoto/Suggestionbox.Predict", http.HandlerFunc(srv.handlePredict))
	server.Register("/remoto/Suggestionbox.PutState", http.HandlerFunc(srv.handlePutState))
	server.Register("/remoto/Suggestionbox.Reward", http.HandlerFunc(srv.handleReward))

}

type Choice struct {
	ID string `json:"id"`

	Features []Feature `json:"features"`
}

type CreateModelRequest struct {
	Model Model `json:"model"`
}

type CreateModelResponse struct {

	// Error is an error message if one occurred.
	Error string `json:"error"`
}

type DeleteModelRequest struct {
	ModelID string `json:"model_id"`
}

type DeleteModelResponse struct {

	// Error is an error message if one occurred.
	Error string `json:"error"`
}

type Feature struct {
	Key string `json:"key"`

	Type string `json:"type"`

	Value string `json:"value"`

	File remototypes.File `json:"file"`
}

type GetStateRequest struct {
}

type ListModelsRequest struct {
}

type ListModelsResponse struct {
	Models []Model `json:"models"`

	// Error is an error message if one occurred.
	Error string `json:"error"`
}

type Model struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Options ModelOptions `json:"options"`

	Choices []Choice `json:"choices"`
}

type ModelOptions struct {
	RewardExpirationSeconds int `json:"reward_expiration_seconds"`

	Ngrams int `json:"ngrams"`

	Skipgrams int `json:"skipgrams"`

	Mode string `json:"mode"`

	Epsilon float64 `json:"epsilon"`

	Cover float64 `json:"cover"`
}

type PredictRequest struct {
	ModelID string `json:"model_id"`

	Limit int `json:"limit"`

	Inputs []Feature `json:"inputs"`
}

type PredictResponse struct {
	Choices []PredictedChoice `json:"choices"`

	// Error is an error message if one occurred.
	Error string `json:"error"`
}

type PredictedChoice struct {
	ID string `json:"id"`

	Features []Feature `json:"features"`

	RewardID string `json:"reward_id"`
}

type PutStateRequest struct {
	StateFile remototypes.File `json:"state_file"`
}

type PutStateResponse struct {

	// Error is an error message if one occurred.
	Error string `json:"error"`
}

type RewardRequest struct {
	ModelID string `json:"model_id"`

	RewardID string `json:"reward_id"`

	Value int `json:"value"`
}

type RewardResponse struct {

	// Error is an error message if one occurred.
	Error string `json:"error"`
}

// httpSuggestionboxServer is an internal type that provides an
// HTTP wrapper around Suggestionbox.
type httpSuggestionboxServer struct {
	// service is the Suggestionbox being exposed by this
	// server.
	service Suggestionbox
	// server is the remotohttp.Server that this server is
	// registered with.
	server *remotohttp.Server
}

// handleCreateModel is an http.Handler wrapper for Suggestionbox.CreateModel.
func (srv *httpSuggestionboxServer) handleCreateModel(w http.ResponseWriter, r *http.Request) {
	var reqs []*CreateModelRequest
	if err := remotohttp.Decode(r, &reqs); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

	resps := make([]CreateModelResponse, len(reqs))
	for i := range reqs {
		resp, err := srv.service.CreateModel(r.Context(), reqs[i])
		if err != nil {
			resps[i].Error = err.Error()
			continue
		}
		resps[i] = *resp
	}
	if err := remotohttp.Encode(w, r, http.StatusOK, resps); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

}

// handleDeleteModel is an http.Handler wrapper for Suggestionbox.DeleteModel.
func (srv *httpSuggestionboxServer) handleDeleteModel(w http.ResponseWriter, r *http.Request) {
	var reqs []*DeleteModelRequest
	if err := remotohttp.Decode(r, &reqs); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

	resps := make([]DeleteModelResponse, len(reqs))
	for i := range reqs {
		resp, err := srv.service.DeleteModel(r.Context(), reqs[i])
		if err != nil {
			resps[i].Error = err.Error()
			continue
		}
		resps[i] = *resp
	}
	if err := remotohttp.Encode(w, r, http.StatusOK, resps); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

}

// handleGetState is an http.Handler wrapper for Suggestionbox.GetState.
func (srv *httpSuggestionboxServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	var reqs []*GetStateRequest
	if err := remotohttp.Decode(r, &reqs); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

	// single file response

	if len(reqs) != 1 {
		if err := remotohttp.EncodeErr(w, r, errors.New("only single requests supported for file response endpoints")); err != nil {
			srv.server.OnErr(w, r, err)
			return
		}
		return
	}

	resp, err := srv.service.GetState(r.Context(), reqs[0])
	if err != nil {
		resp.Error = err.Error()
		if err := remotohttp.Encode(w, r, http.StatusOK, []interface{}{resp}); err != nil {
			srv.server.OnErr(w, r, err)
			return
		}
	}
	if resp.ContentType == "" {
		resp.ContentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.QuoteToASCII(resp.Filename))
	if resp.ContentLength > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(resp.ContentLength))
	}
	if _, err := io.Copy(w, resp.Data); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

}

// handleListModels is an http.Handler wrapper for Suggestionbox.ListModels.
func (srv *httpSuggestionboxServer) handleListModels(w http.ResponseWriter, r *http.Request) {
	var reqs []*ListModelsRequest
	if err := remotohttp.Decode(r, &reqs); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

	resps := make([]ListModelsResponse, len(reqs))
	for i := range reqs {
		resp, err := srv.service.ListModels(r.Context(), reqs[i])
		if err != nil {
			resps[i].Error = err.Error()
			continue
		}
		resps[i] = *resp
	}
	if err := remotohttp.Encode(w, r, http.StatusOK, resps); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

}

// handlePredict is an http.Handler wrapper for Suggestionbox.Predict.
func (srv *httpSuggestionboxServer) handlePredict(w http.ResponseWriter, r *http.Request) {
	var reqs []*PredictRequest
	if err := remotohttp.Decode(r, &reqs); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

	resps := make([]PredictResponse, len(reqs))
	for i := range reqs {
		resp, err := srv.service.Predict(r.Context(), reqs[i])
		if err != nil {
			resps[i].Error = err.Error()
			continue
		}
		resps[i] = *resp
	}
	if err := remotohttp.Encode(w, r, http.StatusOK, resps); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

}

// handlePutState is an http.Handler wrapper for Suggestionbox.PutState.
func (srv *httpSuggestionboxServer) handlePutState(w http.ResponseWriter, r *http.Request) {
	var reqs []*PutStateRequest
	if err := remotohttp.Decode(r, &reqs); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

	resps := make([]PutStateResponse, len(reqs))
	for i := range reqs {
		resp, err := srv.service.PutState(r.Context(), reqs[i])
		if err != nil {
			resps[i].Error = err.Error()
			continue
		}
		resps[i] = *resp
	}
	if err := remotohttp.Encode(w, r, http.StatusOK, resps); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

}

// handleReward is an http.Handler wrapper for Suggestionbox.Reward.
func (srv *httpSuggestionboxServer) handleReward(w http.ResponseWriter, r *http.Request) {
	var reqs []*RewardRequest
	if err := remotohttp.Decode(r, &reqs); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

	resps := make([]RewardResponse, len(reqs))
	for i := range reqs {
		resp, err := srv.service.Reward(r.Context(), reqs[i])
		if err != nil {
			resps[i].Error = err.Error()
			continue
		}
		resps[i] = *resp
	}
	if err := remotohttp.Encode(w, r, http.StatusOK, resps); err != nil {
		srv.server.OnErr(w, r, err)
		return
	}

}

// this is here so we don't get a compiler complaints.
func init() {
	var _ = remototypes.File{}
	var _ = strconv.Itoa(0)
	var _ = io.EOF
}
