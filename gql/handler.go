package gql

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/graph-gophers/graphql-go"
)

type persistedQuery struct {
	Sha256Hash string `json:"sha256Hash"`
}

type queryExtensions struct {
	PersistedQuery persistedQuery `json:"persistedQuery"`
}

type queryParams struct {
	Extensions    queryExtensions        `json:"extensions"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

type Handler struct {
	Schema *graphql.Schema
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		//TODO
		log.Fatal(err)
	}

	if len(body) == 0 {
		//TODO
		return
	}

	batchedQuery := false
	queries := []*queryParams{}
	switch string(body[0]) {
	case "{":
		var single queryParams
		err = json.Unmarshal(body, &single)
		if err != nil {
			log.Fatal(err)
		}

		queries = append(queries, &single)
	case "[":
		batchedQuery = true
		err = json.Unmarshal(body, &queries)
		if err != nil {
			log.Fatal(err)
		}
	}

	responses := make([]*graphql.Response, len(queries))
	wg := &sync.WaitGroup{}
	wg.Add(len(responses))

	for i, query := range queries {
		go func(i int, query *queryParams) {
			defer wg.Done()
			response := h.Schema.Exec(ctx, query.Query, query.OperationName, query.Variables)

			if isCtxDone(ctx) || ctx.Err() == context.Canceled {
				return
			}

			if response.Extensions == nil {
				response.Extensions = make(map[string]interface{})
			}

			if response == nil {
				return
			}

			responses[i] = response
			if response.Errors != nil {
				for _, e := range response.Errors {
					if e.Extensions == nil {
						e.Extensions = make(map[string]interface{})
					}

					if e.ResolverError != nil {
						msg := sanitizeResolverError(e.ResolverError)
						e.Message = msg
					}
				}
			}
		}(i, query)
	}

	wg.Wait()

	if batchedQuery {
		WriteJSON(w, responses)
	} else {
		WriteJSON(w, responses[0])
	}

}

func WriteJSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	w.Header().Add("Content-Length", strconv.Itoa(len(b)))
	_, err = w.Write(b)
	return err
}

func isCtxDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// type ResolverError struct {
// 	Code    int    `json:"code"`
// 	Message string `json:"message"`
// }

// // NewResolverError creates a new resolver error
// func NewResolverError(ctx context.Context, err error, code int) *ResolverError {
// 	stacktrace := ""
// 	err = errors.WithStack(err)
// 	if e, ok := err.(stackTracer); ok {
// 		for _, f := range e.StackTrace() {
// 			stacktrace += fmt.Sprintf("%+s:%d\n", f, f)
// 		}
// 	}

// 	return &ResolverError{
// 		Code:    code,
// 		Message: err.Error(),
// 	}
// }

// // Error resolves the formatted error message
// func (r *ResolverError) Error() string {
// 	return fmt.Sprintf("[%d] %s", r.Code, r.Message)
// }

// // Extensions resolves the additional extension info for the error
// func (r *ResolverError) Extensions() map[string]interface{} {
// 	return map[string]interface{}{
// 		"code": r.Code,
// 	}
// }

// type stackTracer interface {
// 	StackTrace() errors.StackTrace
// }

// func resultsWithKeys(keys dataloader.Keys, m map[string]*dataloader.Result) []*dataloader.Result {
// 	results := make([]*dataloader.Result, 0, len(keys))

// 	for _, key := range keys {
// 		result, found := m[key.String()]
// 		if !found {
// 			result = &dataloader.Result{}
// 		}
// 		results = append(results, result)
// 	}

// 	return results
// }
