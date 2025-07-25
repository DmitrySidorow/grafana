package query

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/apis/data/v0alpha1"
	"github.com/grafana/grafana/pkg/registry/apis/query/clientapi"
	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
	errorsK8s "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	query "github.com/grafana/grafana/pkg/apis/query/v0alpha1"
	"github.com/grafana/grafana/pkg/expr/mathexp"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/web"
)

type queryREST struct {
	logger  log.Logger
	builder *QueryAPIBuilder
}

var (
	_ rest.Storage              = (*queryREST)(nil)
	_ rest.SingularNameProvider = (*queryREST)(nil)
	_ rest.Connecter            = (*queryREST)(nil)
	_ rest.Scoper               = (*queryREST)(nil)
	_ rest.StorageMetadata      = (*queryREST)(nil)
)

func newQueryREST(builder *QueryAPIBuilder) *queryREST {
	return &queryREST{
		logger:  log.New("query"),
		builder: builder,
	}
}

func (r *queryREST) New() runtime.Object {
	// This is added as the "ResponseType" regardless what ProducesObject() says :)
	return &query.QueryDataResponse{}
}

func (r *queryREST) Destroy() {}

func (r *queryREST) NamespaceScoped() bool {
	return true
}

func (r *queryREST) GetSingularName() string {
	return "QueryResults" // Used for the
}

func (r *queryREST) ProducesMIMETypes(verb string) []string {
	return []string{"application/json"} // and parquet!
}

func (r *queryREST) ProducesObject(verb string) interface{} {
	return &query.QueryDataResponse{}
}

func (r *queryREST) ConnectMethods() []string {
	return []string{"POST"}
}

func (r *queryREST) NewConnectOptions() (runtime.Object, bool, string) {
	return nil, false, "" // true means you can use the trailing path as a variable
}

func (r *queryREST) Connect(connectCtx context.Context, name string, _ runtime.Object, incomingResponder rest.Responder) (http.Handler, error) {
	// See: /pkg/services/apiserver/builder/helper.go#L34
	// The name is set with a rewriter hack
	if name != "name" {
		r.logger.Debug("Connect name is not name")
		return nil, errorsK8s.NewNotFound(schema.GroupResource{}, name)
	}
	b := r.builder

	return http.HandlerFunc(func(w http.ResponseWriter, httpreq *http.Request) {
		ctx, span := b.tracer.Start(httpreq.Context(), "QueryService.Query")
		defer span.End()
		ctx = request.WithNamespace(ctx, request.NamespaceValue(connectCtx))
		traceId := span.SpanContext().TraceID()
		connectLogger := b.log.New("traceId", traceId.String())
		responder := newResponderWrapper(incomingResponder,
			func(statusCode *int, obj runtime.Object) {
				if *statusCode/100 == 4 {
					span.SetStatus(codes.Error, strconv.Itoa(*statusCode))
				}

				if *statusCode >= 500 {
					o, ok := obj.(*query.QueryDataResponse)
					if ok && o.Responses != nil {
						for refId, response := range o.Responses {
							if response.ErrorSource == backend.ErrorSourceDownstream {
								*statusCode = http.StatusBadRequest //force this to be a 400 since it's downstream
								span.SetStatus(codes.Error, strconv.Itoa(*statusCode))
								span.SetAttributes(attribute.String("error.source", "downstream"))
								break
							} else if response.Error != nil {
								connectLogger.Debug("500 error without downstream error source", "error", response.Error, "errorSource", response.ErrorSource, "refId", refId)
								span.SetStatus(codes.Error, "500 error without downstream error source")
							} else {
								span.SetStatus(codes.Error, "500 error without downstream error source and no Error message")
								span.SetAttributes(attribute.String("error.ref_id", refId))
							}
						}
					}
				}
			},

			func(err error) {
				connectLogger.Error("error caught in handler", "err", err)
				span.SetStatus(codes.Error, "query error")

				if err == nil {
					return
				}

				span.RecordError(err)
			})

		raw := &query.QueryDataRequest{}
		err := web.Bind(httpreq, raw)
		if err != nil {
			b.log.Error("Hit unexpected error when reading query", "err", err)
			err = errorsK8s.NewBadRequest("error reading query")
			// TODO: can we wrap the error so details are not lost?!
			// errutil.BadRequest(
			// 	"query.bind",
			// 	errutil.WithPublicMessage("Error reading query")).
			// 	Errorf("error reading: %w", err)
			responder.Error(err)
			return
		}
		// Parses the request and splits it into multiple sub queries (if necessary)
		req, err := b.parser.parseRequest(ctx, raw)
		if err != nil {
			var refError ErrorWithRefID
			statusCode := http.StatusBadRequest
			message := err
			refID := "A"

			if errors.Is(err, datasources.ErrDataSourceNotFound) {
				statusCode = http.StatusNotFound
				message = errors.New("datasource not found")
			}

			if errors.As(err, &refError) {
				refID = refError.refId
			}

			qdr := &query.QueryDataResponse{
				QueryDataResponse: backend.QueryDataResponse{
					Responses: backend.Responses{
						refID: {
							Error:  message,
							Status: backend.Status(statusCode),
						},
					},
				},
			}

			b.log.Error("Error parsing query", "refId", refID, "message", message)

			responder.Object(statusCode, qdr)
			return
		}

		logEmptyRefids(raw.Queries, b.log)

		for i := range req.Requests {
			req.Requests[i].Headers = ExtractKnownHeaders(httpreq.Header)
		}

		// Fetch information on the grafana instance (e.g. feature toggles)
		instanceConfig, err := b.clientSupplier.GetInstanceConfigurationSettings(ctx)
		if err != nil {
			msg := "failed to get instance configuration settings"
			b.log.Error(msg, "err", err)
			responder.Error(errors.New(msg))
			return
		}

		// Actually run the query (includes expressions)
		rsp, err := b.execute(ctx, req, instanceConfig)
		if err != nil {
			b.log.Error("execute error", "http code", query.GetResponseCode(rsp), "err", err)
			if rsp != nil { // if we have a response, we assume the err is set in the response
				responder.Object(query.GetResponseCode(rsp), &query.QueryDataResponse{
					QueryDataResponse: *rsp,
				})
				return
			} else {
				// return the error to the client, will send all non k8s errors as a k8 unexpected error
				b.log.Error("hit unexpected error while executing query, this will show as an unhandled k8s status error", "err", err)
				responder.Error(err)
				return
			}
		}

		responder.Object(query.GetResponseCode(rsp), &query.QueryDataResponse{
			QueryDataResponse: *rsp, // wrap the backend response as a QueryDataResponse
		})
	}), nil
}

func logEmptyRefids(queries []v0alpha1.DataQuery, logger log.Logger) {
	emptyCount := 0

	for _, q := range queries {
		if q.RefID == "" {
			emptyCount += 1
		}
	}

	if emptyCount > 0 {
		logger.Info("empty refid found", "empty_count", emptyCount, "query_count", len(queries))
	}
}

func (b *QueryAPIBuilder) execute(ctx context.Context, req parsedRequestInfo, instanceConfig clientapi.InstanceConfigurationSettings) (qdr *backend.QueryDataResponse, err error) {
	switch len(req.Requests) {
	case 0:
		b.log.Debug("executing empty query")
		qdr = &backend.QueryDataResponse{}
	case 1:
		b.log.Debug("executing single query")
		qdr, err = b.handleQuerySingleDatasource(ctx, req.Requests[0], instanceConfig)
		if err != nil {
			b.log.Debug("handleQuerySingleDatasource failed", err)
		}
		if err == nil && isSingleAlertQuery(req) {
			b.log.Debug("handling alert query with single query")
			qdr, err = b.convertQueryFromAlerting(ctx, req.Requests[0], qdr)
			if err != nil {
				b.log.Debug("convertQueryFromAlerting failed", "err", err)
			}
		}
	default:
		b.log.Debug("executing concurrent queries")
		qdr, err = b.executeConcurrentQueries(ctx, req.Requests, instanceConfig)
		if err != nil {
			b.log.Debug("error in executeConcurrentQueries", "err", err)
		}
	}
	if err != nil {
		b.log.Debug("error in query phase, skipping expressions", "error", err)
		return qdr, err //return early here to prevent expressions from being executed if we got an error during the query phase
	}

	if len(req.Expressions) > 0 {
		b.log.Debug("executing expressions")
		qdr, err = b.handleExpressions(ctx, req, qdr)
		if err != nil {
			b.log.Debug("handleExpressions failed", "err", err)
		}
	}

	// Remove hidden results
	for _, refId := range req.HideBeforeReturn {
		r, ok := qdr.Responses[refId]
		if ok && r.Error == nil {
			delete(qdr.Responses, refId)
		}
	}

	return qdr, err
}

// Process a single request
// See: https://github.com/grafana/grafana/blob/v10.2.3/pkg/services/query/query.go#L242
func (b *QueryAPIBuilder) handleQuerySingleDatasource(ctx context.Context, req datasourceRequest, instanceConfig clientapi.InstanceConfigurationSettings) (*backend.QueryDataResponse, error) {
	ctx, span := b.tracer.Start(ctx, "Query.handleQuerySingleDatasource")
	defer span.End()
	span.SetAttributes(
		attribute.String("datasource.type", req.PluginId),
		attribute.String("datasource.uid", req.UID),
	)

	allHidden := true
	for idx := range req.Request.Queries {
		if !req.Request.Queries[idx].Hide {
			allHidden = false
			break
		}
	}
	if allHidden {
		return &backend.QueryDataResponse{}, nil
	}

	client, err := b.clientSupplier.GetDataSourceClient(
		ctx,
		v0alpha1.DataSourceRef{
			Type: req.PluginId,
			UID:  req.UID,
		},
		req.Headers,
		instanceConfig,
	)
	if err != nil {
		b.log.Debug("error getting single datasource client", "error", err, "reqUid", req.UID)
		qdr := buildErrorResponse(err, req)
		return qdr, err
	}

	rsp, err := client.QueryData(ctx, *req.Request)

	if err == nil && rsp != nil {
		for _, q := range req.Request.Queries {
			if q.ResultAssertions != nil {
				result, ok := rsp.Responses[q.RefID]
				if ok && result.Error == nil {
					err = q.ResultAssertions.Validate(result.Frames)
					if err != nil {
						b.log.Error("Validate failed", "err", err)
						result.Error = err
						result.ErrorSource = backend.ErrorSourceDownstream
						rsp.Responses[q.RefID] = result
					}
				}
			}
		}
	}

	if err != nil {
		b.log.Debug("error in single datasource query", "error", err)
		rsp = buildErrorResponse(err, req)
	}

	return rsp, err
}

// buildErrorResponses applies the provided error to each query response in the list. These queries should all belong to the same datasource.
func buildErrorResponse(err error, req datasourceRequest) *backend.QueryDataResponse {
	rsp := backend.NewQueryDataResponse()
	for _, query := range req.Request.Queries {
		rsp.Responses[query.RefID] = backend.DataResponse{
			Error: err,
		}
	}
	return rsp
}

// executeConcurrentQueries executes queries to multiple datasources concurrently and returns the aggregate result.
func (b *QueryAPIBuilder) executeConcurrentQueries(ctx context.Context, requests []datasourceRequest, instanceConfig clientapi.InstanceConfigurationSettings) (*backend.QueryDataResponse, error) {
	ctx, span := b.tracer.Start(ctx, "Query.executeConcurrentQueries")
	defer span.End()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(b.concurrentQueryLimit) // prevent too many concurrent requests
	rchan := make(chan *backend.QueryDataResponse, len(requests))

	// Create panic recovery function for loop below
	recoveryFn := func(req datasourceRequest) {
		if r := recover(); r != nil {
			var err error
			b.log.Error("query datasource panic", "error", r, "stack", log.Stack(1))
			if theErr, ok := r.(error); ok {
				err = theErr
			} else if theErrString, ok := r.(string); ok {
				err = errors.New(theErrString)
			} else {
				err = fmt.Errorf("unexpected error - %s", b.userFacingDefaultError)
			}
			// Due to the panic, there is no valid response for any query for this datasource. Append an error for each one.
			rchan <- buildErrorResponse(err, req)
		}
	}

	// Query each datasource concurrently
	for idx := range requests {
		req := requests[idx]
		g.Go(func() error {
			defer recoveryFn(req)

			dqr, err := b.handleQuerySingleDatasource(ctx, req, instanceConfig)
			if err == nil {
				rchan <- dqr
			} else {
				rchan <- buildErrorResponse(err, req)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(rchan)

	// Merge the results from each response
	resp := backend.NewQueryDataResponse()
	for result := range rchan {
		for refId, dataResponse := range result.Responses {
			resp.Responses[refId] = dataResponse
		}
	}

	return resp, nil
}

// Unlike the implementation in expr/node.go, all datasource queries have been processed first
func (b *QueryAPIBuilder) handleExpressions(ctx context.Context, req parsedRequestInfo, data *backend.QueryDataResponse) (qdr *backend.QueryDataResponse, err error) {
	start := time.Now()
	ctx, span := b.tracer.Start(ctx, "Query.handleExpressions")
	traceId := span.SpanContext().TraceID()
	expressionsLogger := b.log.New("traceId", traceId.String())
	expressionsLogger.Debug("handling expressions")
	defer func() {
		var respStatus string
		switch err {
		case nil:
			respStatus = "success"
		default:
			respStatus = "failure"
		}
		duration := float64(time.Since(start).Nanoseconds()) / float64(time.Millisecond)
		b.metrics.ExpressionsQuerySummary.WithLabelValues(respStatus).Observe(duration)

		span.End()
	}()

	qdr = data
	if qdr == nil {
		qdr = &backend.QueryDataResponse{}
	}
	if qdr.Responses == nil {
		qdr.Responses = make(backend.Responses) // avoid NPE for lookup
	}
	now := start // <<< this should come from the original query parser
	vars := make(mathexp.Vars)
	for _, expression := range req.Expressions {
		// Setup the variables
		for _, refId := range expression.Command.NeedsVars() {
			_, ok := vars[refId]
			if !ok {
				dr, ok := qdr.Responses[refId]
				if ok {
					_, isSqlInput := req.SqlInputs[refId]

					_, res, err := b.converter.Convert(ctx, req.RefIDTypes[refId], dr.Frames, isSqlInput)
					if err != nil {
						expressionsLogger.Error("error converting frames for expressions", "error", err)
						res.Error = err
					}

					vars[refId] = res
				} else {
					expressionsLogger.Error("missing variable in handle expressions", "refId", refId, "expressionRefId", expression.RefID)
					// This should error in the parsing phase
					err := fmt.Errorf("missing variable %s for %s", refId, expression.RefID)
					qdr.Responses[refId] = backend.DataResponse{
						Error: err,
					}
					return qdr, err
				}
			}
		}

		refId := expression.RefID
		results, err := expression.Command.Execute(ctx, now, vars, b.tracer, b.metrics)
		if err != nil {
			expressionsLogger.Error("error executing expression", "error", err)
			results.Error = err
		}
		qdr.Responses[refId] = backend.DataResponse{
			Error:  results.Error,
			Frames: results.Values.AsDataFrames(refId),
		}
	}
	return qdr, nil
}

func (b *QueryAPIBuilder) convertQueryFromAlerting(ctx context.Context, req datasourceRequest,
	qdr *backend.QueryDataResponse) (*backend.QueryDataResponse, error) {
	if len(req.Request.Queries) == 0 {
		return nil, errors.New("no queries to convert")
	}
	if qdr == nil {
		b.log.Debug("unexpected response of nil from datasource", "datasource.type", req.PluginId, "datasource.uid", req.UID)
		return nil, errors.New("unexpected response of nil from datasource")
	}
	refID := req.Request.Queries[0].RefID
	if _, exist := qdr.Responses[refID]; !exist {
		return nil, fmt.Errorf("refID '%s' does not exist", refID)
	}
	frames := qdr.Responses[refID].Frames
	_, results, err := b.converter.Convert(ctx, req.PluginId, frames, false)
	if err != nil {
		b.log.Error("issue converting query from alerting", "err", err)
		results.Error = err
	}
	qdr = &backend.QueryDataResponse{
		Responses: map[string]backend.DataResponse{
			refID: {
				Frames: results.Values.AsDataFrames(refID),
				Error:  results.Error,
			},
		},
	}
	return qdr, err
}

type responderWrapper struct {
	wrapped    rest.Responder
	onObjectFn func(statusCode *int, obj runtime.Object)
	onErrorFn  func(err error)
}

func newResponderWrapper(responder rest.Responder, onObjectFn func(statusCode *int, obj runtime.Object), onErrorFn func(err error)) *responderWrapper {
	return &responderWrapper{
		wrapped:    responder,
		onObjectFn: onObjectFn,
		onErrorFn:  onErrorFn,
	}
}

func (r responderWrapper) Object(statusCode int, obj runtime.Object) {
	if r.onObjectFn != nil {
		r.onObjectFn(&statusCode, obj)
	}

	r.wrapped.Object(statusCode, obj)
}

func (r responderWrapper) Error(err error) {
	if r.onErrorFn != nil {
		r.onErrorFn(err)
	}

	r.wrapped.Error(err)
}

// Checks if the request only contains a single query and is from Alerting
func isSingleAlertQuery(req parsedRequestInfo) bool {
	if len(req.Requests) != 1 {
		return false
	}
	headers := req.Requests[0].Headers
	_, exist := headers[models.FromAlertHeaderName]
	if exist && len(req.Requests[0].Request.Queries) == 1 {
		return true
	}
	return false
}
