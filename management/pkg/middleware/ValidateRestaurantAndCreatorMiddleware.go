package middleware

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
	"strconv"
)

func ValidateRestaurantAndCreator(db database.Database) mux.MiddlewareFunc{
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request) {

			prevContext:= rq.Context().Value("context")
			prevCtx := prevContext.(context.Context)
			span, newCtx := tracing.GetSpanFromContext(prevCtx, "checking_valid_restaurant_and_creator")
			defer span.Finish()
			tags := tracing.TraceTags{FuncName: "ValidateRestaurantAndCreator", ServiceName: tracing.ServiceName, RequestID: span.BaggageItem("requestID")}
			tracing.SetTags(span, tags)

			value:=  rq.Context().Value("userAuth")
			userAuth := value.(*models.UserAuth)

			vars:=mux.Vars(rq)
			res:=vars["resID"]
			resID, _ := strconv.Atoi(res)
			var chkResCtr context.Context
			var chkResOwner context.Context
			var err error
			if userAuth.Role == Admin {
				chkResCtr, err = db.CheckRestaurantCreator(newCtx, userAuth.ID, resID)
				if err != nil {
					if err != database.ErrInternal {
						w.WriteHeader(http.StatusUnauthorized)
						bodyMap:=models.DefaultMap{
							"error": err.Error(),
						}
						w.Write(bodyMap.ConvertToByteArray())
						return
					}
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else if userAuth.Role == Owner {
				chkResOwner, err = db.CheckRestaurantOwner(newCtx, userAuth.ID, resID)
				if err != nil {
					if err != database.ErrInternal {
						w.WriteHeader(http.StatusUnauthorized)
						bodyMap:=models.DefaultMap{
							"error": err.Error(),
						}
						w.Write(bodyMap.ConvertToByteArray())
						return
					}
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			var reqCtx context.Context
			if chkResOwner == nil && chkResCtr == nil {
				reqCtx=context.WithValue(rq.Context(),"context", newCtx)
			} else if chkResOwner == nil {
				reqCtx=context.WithValue(rq.Context(),"context", chkResCtr)
			} else {
				reqCtx=context.WithValue(rq.Context(),"context", chkResOwner)
			}
			reqCtx=context.WithValue(reqCtx,"restaurantID", resID)
			next.ServeHTTP(w,rq.WithContext(reqCtx))
		})
	}
}
