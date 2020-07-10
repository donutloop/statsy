package handler

import (
	"encoding/json"
	"fmt"
	"github.com/donutloop/statsy/internal/dao"
	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

// HandlerFunc is an http.HandlerFunc with an Context.
type HandlerDomainFunc func(Context, interface{}) (interface{}, error)

type Context struct {
	LogError    func(v ...interface{})
	LogInfo     func(v ...interface{})
	DAO         *dao.DAO
	RouteParams chi.RouteParams
}

func (x *Context) URLParam(key string) string {
	for k := len(x.RouteParams.Keys) - 1; k >= 0; k-- {
		if x.RouteParams.Keys[k] == key {
			return x.RouteParams.Values[k]
		}
	}
	return ""
}

func StatsHandler(HandlerDomainFunc HandlerDomainFunc, ctx Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		domainReq := &StatsRequest{}
		if err := UnmarshalRequest(r.Body, domainReq); err != nil {
			ctx.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if domainReq.CustomerID <= 0 {
			ctx.LogError("customer id is bad")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := domainReq.Validate(); err != nil {
			body, err := json.Marshal(err)
			if err != nil {
				ctx.LogError(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusBadRequest, "missing required fields", domainReq.Timestamp)
			w.Write(body)
			return
		}

		active, err := ctx.DAO.IsCustomerActive(domainReq.CustomerID)
		if err != nil {
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusInternalServerError, err.Error(), domainReq.Timestamp)
			return
		}

		if !active {
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusBadRequest, fmt.Sprintf("customer is not active %v", domainReq.CustomerID), domainReq.Timestamp)
			return
		}

		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusBadRequest, "User-Agent header is missing", domainReq.Timestamp)
			return
		}

		found, err := ctx.DAO.FindBlackedlistedUserAgent(userAgent)
		if err != nil {
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusInternalServerError, err.Error(), domainReq.Timestamp)
			return
		}

		if found {
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusBadRequest, fmt.Sprintf("found blacked listed user agent %v", userAgent), domainReq.Timestamp)
			return
		}

		found, err = ctx.DAO.FindBlacklistedIP(domainReq.RemoteIP)
		if err != nil {
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusInternalServerError, err.Error(), domainReq.Timestamp)
			return
		}

		if found {
			handleInvalidRequest(w, ctx, domainReq.CustomerID, http.StatusBadRequest, fmt.Sprintf("found blacked listed ip %v", domainReq.RemoteIP), domainReq.Timestamp)
			return
		}

		if err := ctx.DAO.InsertOrUpdateCustomerStats(domainReq.CustomerID, true, domainReq.Timestamp); err != nil {
			ctx.LogError(err)
		}

		_, err = HandlerDomainFunc(ctx, domainReq)
		if err != nil {
			ctx.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func handleInvalidRequest(resp http.ResponseWriter, ctx Context, customerID int64, statusCode int, errorMsg string, time int64) {
	ctx.LogInfo(errorMsg)
	resp.WriteHeader(statusCode)
	if err := ctx.DAO.InsertOrUpdateCustomerStats(customerID, false, time); err != nil {
		ctx.LogError(err)
	}
	return
}

func WrapHandler(HandlerDomainFunc HandlerDomainFunc, ctx Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routeContext := chi.RouteContext(r.Context())
		routeParams := routeContext.URLParams
		ctx.RouteParams = routeParams
		domainResp, err := HandlerDomainFunc(ctx, nil)
		if err != nil {
			ctx.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respBody, err := json.Marshal(domainResp)
		if err != nil {
			ctx.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(respBody)
		if err != nil {
			ctx.LogError(err)
		}
	})
}

func UnmarshalRequest(requestBody io.ReadCloser, v interface{}) error {
	b, err := ioutil.ReadAll(requestBody)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}
	return nil
}

type StatsRequest struct {
	CustomerID int64  `json:"customerID"`
	RemoteIP   string `json:"remoteIP"`
	TagID      int64  `json:"tagID"`
	Timestamp  int64  `json:"timestamp"`
	UserID     string `json:"userID"`
}

func (c *StatsRequest) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.CustomerID, validation.Required),
		validation.Field(&c.RemoteIP, validation.Required),
		validation.Field(&c.TagID, validation.Required),
		validation.Field(&c.Timestamp, validation.Required),
		validation.Field(&c.UserID, validation.Required),
	)
}

// todo(marcel) Right now! this is just a stub handler
func HandleCustomer(ctx Context, domainRequestRaw interface{}) (interface{}, error) {
	return nil, nil
}

func GetStatsByCustomerID(ctx Context, domainRequestRaw interface{}) (interface{}, error) {

	customerID, err := strconv.Atoi(ctx.URLParam("customerID"))
	if err != nil {
		return nil, err
	}

	dayUnix, err := strconv.Atoi(ctx.URLParam("day"))
	if err != nil {
		return nil, err
	}

	cs, err := ctx.DAO.GetCutsomerRequestCountByDayAndID(int64(customerID), int64(dayUnix))
	if err != nil {
		return nil, err
	}

	return cs, nil
}
